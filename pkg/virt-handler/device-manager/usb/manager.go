package usb

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"kubevirt.io/api/usb/v1alpha1"
	"kubevirt.io/client-go/log"
)

type USBManagerInterface interface {
	Run(stopCh chan struct{})
}

func newState() state {
	return state{
		resourceNameToPluginHandler: map[string]*pluginHandler{},
		nodeconfigToResource:        map[string]string{},
		lock:                        sync.Mutex{},
	}
}

type state struct {
	resourceNameToPluginHandler map[string]*pluginHandler
	nodeconfigToResource        map[string]string
	lock                        sync.Mutex
}

func (s *state) insert(nodeConfig string, plugin Plugin) chan struct{} {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, alreadyExists := s.nodeconfigToResource[nodeConfig]; alreadyExists {
		log.Log.Infof("Could not insert %s: Already exists", nodeConfig)
		return nil
	}
	close := make(chan struct{})
	resourceName := plugin.Name()

	s.nodeconfigToResource[nodeConfig] = resourceName
	s.resourceNameToPluginHandler[resourceName] = &pluginHandler{
		started:  false,
		failed:   false,
		stopChan: close,
		plugin:   plugin,
	}
	log.Log.Infof("Insert %s into manager's state", nodeConfig)
	return close
}

func (s *state) clean(nodeConfig string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	resourceName, configInState := s.nodeconfigToResource[nodeConfig]
	if !configInState {
		return
	}

	handler := s.resourceNameToPluginHandler[resourceName]
	close(handler.stopChan)

	delete(s.nodeconfigToResource, nodeConfig)
	delete(s.resourceNameToPluginHandler, resourceName)
	log.Log.Infof("Removed %s", nodeConfig)
}

func (s *state) list() []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	ret := make([]string, len(s.nodeconfigToResource))
	for nodeConfig := range s.nodeconfigToResource {
		ret = append(ret, nodeConfig)
	}
	return ret
}

func (s *state) updateHandler(resourceName string, started bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	handler, exist := s.resourceNameToPluginHandler[resourceName]
	if !exist {
		log.Log.Warningf("Failed to update %s: resource no longer exists", resourceName)
		return
	}

	if started {
		handler.started = true
	} else {
		handler.failed = true
	}
}

type USBManager struct {
	nodeConfigInformer cache.SharedIndexInformer
	queue              workqueue.RateLimitingInterface
	discoveryFunc      func() []*usbDevice
	factoryFunc        factory
	state              state
	logger             *log.FilteredLogger
}

type pluginHandler struct {
	started  bool
	failed   bool
	stopChan chan struct{}
	plugin   Plugin
}

type usbDeviceSelector struct {
	resourceName string
	vendor       int
	product      int
}

func NewUSBManager(nodeConfigInformer cache.SharedIndexInformer) *USBManager {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "virt-handler-nodeconfig")
	manager := &USBManager{
		nodeConfigInformer: nodeConfigInformer,
		queue:              queue,
		discoveryFunc:      discoverUSBDevices,
		factoryFunc:        NewUSBDevicePlugin,
		state:              newState(),
		logger:             log.Log.With("subcomponent", "usb-manager"),
	}
	nodeConfigInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    manager.addFunc,
		UpdateFunc: manager.updateFunc,
		DeleteFunc: manager.deleteFunc,
	})

	return manager
}

func (manager *USBManager) cleanUpWorker(stop chan struct{}) func() {
	return func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				for _, nodeConfig := range manager.state.list() {
					if _, exists, _ := manager.nodeConfigInformer.GetIndexer().GetByKey(nodeConfig); !exists {
						manager.logger.Infof("Cleaning up plugin for %s node config", nodeConfig)
						manager.state.clean(nodeConfig)
					}
				}
			case <-stop:
				return
			}
		}
	}
}

func (manager *USBManager) Run(stopCh chan struct{}) {
	defer manager.queue.ShutDown()

	log.Log.Info("Starting USB manager")

	cache.WaitForCacheSync(stopCh, manager.nodeConfigInformer.HasSynced)

	// Start the actual work
	go wait.Until(manager.runWorker, time.Second, stopCh)
	// TODO this might be to much
	go wait.Until(manager.cleanUpWorker(stopCh), time.Second, stopCh)
	log.Log.Info("Started USB manager")

	<-stopCh
	log.Log.Info("Stoping USB manager")
}

func (manager *USBManager) Execute() bool {
	key, quit := manager.queue.Get()
	if quit {
		return false
	}
	defer manager.queue.Done(key)

	if err := manager.execute(key.(string)); err != nil {
		log.Log.Reason(err).Infof("re-enqueuing NodeConfig %v", key)
		manager.queue.AddRateLimited(key)
	} else {
		log.Log.V(4).Infof("processed NodeConfig %v", key)
		manager.queue.Forget(key)
	}
	return true
}

func (manager *USBManager) runWorker() {
	for manager.Execute() {
	}
}

func (manager *USBManager) execute(key string) error {
	obj, exists, err := manager.nodeConfigInformer.GetStore().GetByKey(key)
	if err != nil {
		return fmt.Errorf("failed to get object for key %s, %v", key, err)
	}

	if !exists || obj == nil {
		log.Log.Infof("Removing %s (exists: %t) (is nil: %t)", key, exists, obj == nil)
		manager.state.clean(key)
		return nil
	}

	// If key already exists, cleanup before proceeding
	manager.state.clean(key)

	log.Log.Infof("Iterating over %s", key)
	nodeConfig := obj.(*v1alpha1.NodeConfig)
	return manager.syncDevicePlugin(nodeConfig, key)
}

func constructPermittedUSBDevicesMap(nodeConfig *v1alpha1.NodeConfig) map[int][]usbDeviceSelector {
	// Iterate over requested USB Devices and map it vendor:product
	permittedUSBDevices := make(map[int][]usbDeviceSelector)
	resourceName := nodeConfig.Spec.USB.ResourceName
	for index, dev := range nodeConfig.Spec.USB.USBHostDevices {
		sep := strings.Index(dev.SelectByVendorProduct, ":")
		if sep == -1 {
			log.Log.Warningf("Failed to parse USBHostDevices[%d] = %s",
				index, dev.SelectByVendorProduct)
			continue
		}
		val, err := strconv.ParseInt(dev.SelectByVendorProduct[:sep], 16, 32)
		if err != nil {
			log.Log.Warningf("Failed to convert vendor from base16 string to int: %s",
				dev.SelectByVendorProduct[:sep])
			continue
		}
		vendor := int(val)

		val, err = strconv.ParseInt(dev.SelectByVendorProduct[sep+1:], 16, 32)
		if err != nil {
			log.Log.Warningf("Failed to convert product from base16 string to int: %s",
				dev.SelectByVendorProduct[:sep])
			continue
		}
		product := int(val)

		permittedUSBDevices[vendor] = append(permittedUSBDevices[vendor],
			usbDeviceSelector{
				resourceName: resourceName,
				vendor:       vendor,
				product:      product,
			})
	}
	return permittedUSBDevices
}

func (manager *USBManager) syncDevicePlugin(nodeConfig *v1alpha1.NodeConfig, key string) error {
	log.Log.Infof("%s sync", nodeConfig.Name)

	// Sanity check
	if nodeConfig == nil || len(nodeConfig.Spec.USB.USBHostDevices) == 0 {
		log.Log.V(5).Infof("No USB devices")
		return nil
	}

	localDevicesFound := manager.discoveryFunc()
	if len(localDevicesFound) == 0 {
		log.Log.V(5).Info("No USB devices found in this node")
		return nil
	}

	permittedDevicesPerVendor := constructPermittedUSBDevicesMap(nodeConfig)

	// For each device found in this node, compare with those requested in node config
	// to see if we have any matches as we only start the Plugin with those that matched.
	devicesToExport := map[string][]*usbDevice{}
	for _, device := range localDevicesFound {
		permittedDevices, vendorMatched := permittedDevicesPerVendor[device.Vendor]
		if !vendorMatched {
			continue
		}
		for _, permpermittedDevice := range permittedDevices {
			if permpermittedDevice.product != device.Product {
				continue
			}

			resourceName := permpermittedDevice.resourceName

			_, ok := devicesToExport[resourceName]
			if !ok {
				devicesToExport[resourceName] = []*usbDevice{}
			}
			devicesToExport[resourceName] = append(devicesToExport[resourceName], device)
		}
	}

	for resourceName, devices := range devicesToExport {
		log.Log.Infof("%s has %d devices", resourceName, len(devices))
		plugin := manager.factoryFunc(resourceName, devices)
		manager.startPlugin(plugin, key)
	}
	return nil
}

func (manager *USBManager) startPlugin(plugin Plugin, key string) {
	var stop chan struct{}
	if stop = manager.state.insert(key, plugin); stop == nil {
		// No changes in NodeConfig
		log.Log.V(9).Infof("USB plugin %s is already started", plugin.Name())
		return
	}

	log.Log.Infof("USB pluggin %s starting", plugin.Name())
	go manager.startUpPlugin(plugin, stop)
}

func (manager *USBManager) startUpPlugin(plugin Plugin, stop chan struct{}) {
	retries := 0

	tryStartPlugin := func() bool {
		err := plugin.Start(stop)
		if err == nil {
			log.DefaultLogger().Infof("Started %s USB pluggin.", plugin.Name())
			return true
		}
		log.DefaultLogger().Reason(err).Errorf("Error starting %s USB pluggin. Retry #%d",
			plugin.Name(), retries)

		return false
	}

	for {
		if tryStartPlugin() {
			manager.state.updateHandler(plugin.Name(), true)
			return
		}

		retries++
		if retries > 10 {
			manager.state.updateHandler(plugin.Name(), false)
			log.DefaultLogger().Errorf("Unable to start %s USB pluggin", plugin.Name())
			return
		}
		select {
		case <-stop:
			// Start has been cancelled
			return
		case <-time.After(10 * time.Second):
			// Try again
			continue
		}
	}
}

func parseSysUeventFile(path string) *usbDevice {
	// Grab all details we are interested from uevent
	file, err := os.Open(filepath.Join(path, "uevent"))
	if err != nil {
		log.Log.Reason(err).Infof("Unable to access %s/%s", path, "uevent")
		return nil
	}
	defer file.Close()

	u := usbDevice{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		equal := strings.Index(line, "=")
		if strings.HasPrefix(line, "BUSNUM") {
			val, err := strconv.ParseInt(line[equal+1:], 10, 32)
			if err != nil {
				return nil
			}
			u.Bus = int(val)
		} else if strings.HasPrefix(line, "DEVNUM") {
			val, err := strconv.ParseInt(line[equal+1:], 10, 32)
			if err != nil {
				return nil
			}
			u.DeviceNumber = int(val)
		} else if strings.HasPrefix(line, "PRODUCT") {
			values := strings.Split(line[equal+1:], "/")
			if len(values) != 3 {
				return nil
			}

			val, err := strconv.ParseInt(values[0], 16, 32)
			if err != nil {
				return nil
			}
			u.Vendor = int(val)

			val, err = strconv.ParseInt(values[1], 16, 32)
			if err != nil {
				return nil
			}
			u.Product = int(val)

			val, err = strconv.ParseInt(values[2], 16, 32)
			if err != nil {
				return nil
			}
			u.BCD = int(val)
		} else if strings.HasPrefix(line, "DEVNAME") {
			u.DevicePath = "/dev/" + line[equal+1:]
		}
	}
	return &u
}

func discoverUSBDevices() []*usbDevice {
	usbDevices := make([]*usbDevice, 0)
	err := filepath.Walk("/sys/bus/usb/devices", func(path string, info os.FileInfo, err error) error {
		// Ignore named usb controllers
		if strings.HasPrefix(info.Name(), "usb") {
			return nil
		}
		// We are interested in actual USB devices information that
		// contains idVendor and idProduct. We can skip all others.
		if _, err := os.Stat(filepath.Join(path, "idVendor")); err != nil {
			return nil
		}

		device := parseSysUeventFile(path)
		if device == nil {
			return nil
		}
		usbDevices = append(usbDevices, device)

		// FIXME: Check if device is available ?
		return nil
	})

	if err != nil {
		log.Log.Reason(err).Error("Failed when walking usb devices tree")
	}
	return usbDevices
}
