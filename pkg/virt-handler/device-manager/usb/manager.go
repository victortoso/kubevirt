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

// The public facing API for virt-handler
type USBManagerInterface interface {
	Run(stopCh chan struct{})
}

// The handler to store and access Plugin's states
type state struct {
	// FIXME: Should the resourceName be unique across USBDevicesConfigs?
	resourceNameToPluginHandler map[string]*pluginHandler
	usbDevicesConfigToResource  map[string]string
	lock                        sync.Mutex
	logger                      *log.FilteredLogger
}

func newState() state {
	return state{
		resourceNameToPluginHandler: map[string]*pluginHandler{},
		usbDevicesConfigToResource:  map[string]string{},
		lock:                        sync.Mutex{},
		logger:                      log.Log.With("subcomponent", "usb-manager-state"),
	}
}

func (s *state) insert(key string, plugin Plugin) chan struct{} {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, alreadyExists := s.usbDevicesConfigToResource[key]; alreadyExists {
		s.logger.Warningf("Could not insert %s: Already exists", key)
		return nil
	}
	close := make(chan struct{})
	resourceName := plugin.Name()

	s.usbDevicesConfigToResource[key] = resourceName
	s.resourceNameToPluginHandler[resourceName] = &pluginHandler{
		started:  false,
		failed:   false,
		stopChan: close,
		plugin:   plugin,
	}
	s.logger.V(5).Infof("Insert %s", key)
	return close
}

func (s *state) clean(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	resourceName, configInState := s.usbDevicesConfigToResource[key]
	if !configInState {
		return
	}

	handler := s.resourceNameToPluginHandler[resourceName]
	close(handler.stopChan)

	delete(s.usbDevicesConfigToResource, key)
	delete(s.resourceNameToPluginHandler, resourceName)
	s.logger.V(5).Infof("Removed %s", key)
}

func (s *state) list() []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	ret := make([]string, len(s.usbDevicesConfigToResource))
	for key := range s.usbDevicesConfigToResource {
		ret = append(ret, key)
	}
	return ret
}

func (s *state) updateHandler(resourceName string, started bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	handler, exist := s.resourceNameToPluginHandler[resourceName]
	if !exist {
		s.logger.Warningf("Failed to update %s: resource no longer exists", resourceName)
		return
	}

	if started {
		handler.started = true
	} else {
		handler.failed = true
	}
	s.logger.V(5).Infof("%s update: started=%t failed=%t", resourceName, handler.started, handler.failed)
}

type discoveryFuncType func() []*usbDevice

type USBManager struct {
	usbDevicesConfigInformer cache.SharedIndexInformer
	queue                    workqueue.RateLimitingInterface
	discoveryFunc            discoveryFuncType
	factoryFunc              factoryFuncType
	state                    state
	logger                   *log.FilteredLogger
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

func NewUSBManager(usbDevicesConfigInformer cache.SharedIndexInformer) *USBManager {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "virt-handler-usbdevicesconfig")
	manager := &USBManager{
		usbDevicesConfigInformer: usbDevicesConfigInformer,
		queue:                    queue,
		discoveryFunc:            discoverUSBDevices,
		factoryFunc:              NewUSBDevicePlugin,
		state:                    newState(),
		logger:                   log.Log.With("subcomponent", "usb-manager"),
	}
	usbDevicesConfigInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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
				for _, key := range manager.state.list() {
					if _, exists, _ := manager.usbDevicesConfigInformer.GetIndexer().GetByKey(key); !exists {
						manager.logger.Infof("Cleaning up plugin for %s node config", key)
						manager.state.clean(key)
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

	manager.logger.Info("Starting USB manager")

	cache.WaitForCacheSync(stopCh, manager.usbDevicesConfigInformer.HasSynced)

	// Start the actual work
	go wait.Until(manager.runWorker, time.Second, stopCh)
	// TODO: This might not be necessary as execute() should know when a USBDevicesConfig changed or was
	// deleted.
	go wait.Until(manager.cleanUpWorker(stopCh), time.Second, stopCh)
	manager.logger.Info("Started USB manager")

	<-stopCh
	manager.logger.Info("Stoping USB manager")
}

func (manager *USBManager) Execute() bool {
	key, quit := manager.queue.Get()
	if quit {
		return false
	}
	defer manager.queue.Done(key)

	if err := manager.execute(key.(string)); err != nil {
		manager.logger.Reason(err).Infof("re-enqueuing USBDevicesConfig %v", key)
		manager.queue.AddRateLimited(key)
	} else {
		manager.logger.V(5).Infof("processed USBDevicesConfig %v", key)
		manager.queue.Forget(key)
	}
	return true
}

func (manager *USBManager) runWorker() {
	for manager.Execute() {
	}
}

func (manager *USBManager) execute(key string) error {
	obj, exists, err := manager.usbDevicesConfigInformer.GetStore().GetByKey(key)
	if err != nil {
		return fmt.Errorf("failed to get object for key %s, %v", key, err)
	}

	if !exists || obj == nil {
		manager.state.clean(key)
		return nil
	}

	// If key already exists, cleanup before proceeding
	manager.state.clean(key)

	manager.logger.V(5).Infof("Iterating over %s", key)
	usbDevicesConfig := obj.(*v1alpha1.USBDevicesConfig)
	return manager.syncDevicePlugin(usbDevicesConfig, key)
}

func constructPermittedUSBDevicesMap(usbDevicesConfig *v1alpha1.USBDevicesConfig) map[int][]usbDeviceSelector {
	// Iterate over requested USB Devices and map it vendor:product
	permittedUSBDevices := make(map[int][]usbDeviceSelector)
	for _, usb := range usbDevicesConfig.Spec.USB {
		resourceName := usb.ResourceName
		for index, dev := range usb.USBHostDevices {
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

			// TODO: For the moment, we consider only a single resource can hold a product:vendor.
			// we will need more selectors to allow multiple resources to hold same product:vendor
			if selectors, exists := permittedUSBDevices[vendor]; exists {
				otherProductResourceName := ""
				for _, selector := range selectors {
					if selector.product == product {
						otherProductResourceName = selector.resourceName
						break
					}
				}

				if otherProductResourceName != "" {
					log.Log.Warningf("Duplicated USB %s, %s will not receive what is currently attached to %s",
						dev.SelectByVendorProduct, resourceName, otherProductResourceName)
					continue
				}
			}

			permittedUSBDevices[vendor] = append(permittedUSBDevices[vendor],
				usbDeviceSelector{
					resourceName: resourceName,
					vendor:       vendor,
					product:      product,
				})
		}
	}
	return permittedUSBDevices
}

func (manager *USBManager) syncDevicePlugin(usbDevicesConfig *v1alpha1.USBDevicesConfig, key string) error {
	manager.logger.V(5).Infof("%s sync", usbDevicesConfig.Name)

	// Sanity check
	if usbDevicesConfig == nil || len(usbDevicesConfig.Spec.USB) == 0 {
		manager.logger.V(5).Infof("No USB devices")
		return nil
	}

	localDevicesFound := manager.discoveryFunc()
	if len(localDevicesFound) == 0 {
		manager.logger.V(5).Info("No USB devices found in this node")
		return nil
	}

	permittedDevicesPerVendor := constructPermittedUSBDevicesMap(usbDevicesConfig)

	// For each device found in this node, compare with those requested in USBDevicesConfig
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
		manager.logger.V(5).Infof("%s has %d devices", resourceName, len(devices))
		plugin := manager.factoryFunc(resourceName, devices)
		manager.startPlugin(plugin, key)
	}
	return nil
}

func (manager *USBManager) startPlugin(plugin Plugin, key string) {
	var stop chan struct{}
	if stop = manager.state.insert(key, plugin); stop == nil {
		// No changes in USBDevicesConfig
		manager.logger.V(5).Infof("USB plugin %s is already started", plugin.Name())
		return
	}

	manager.logger.Infof("USB plugin %s starting", plugin.Name())
	go manager.startUpPlugin(plugin, stop)
}

func (manager *USBManager) startUpPlugin(plugin Plugin, stop chan struct{}) {
	retries := 0

	tryStartPlugin := func() bool {
		err := plugin.Start(stop)
		if err == nil {
			manager.logger.Infof("Started %s USB plugin.", plugin.Name())
			return true
		}
		manager.logger.Reason(err).Errorf("Error starting %s USB plugin. Retry #%d",
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
			manager.logger.Errorf("Unable to start %s USB plugin", plugin.Name())
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
		return nil
	})

	if err != nil {
		log.Log.Reason(err).Error("Failed when walking usb devices tree")
	}
	return usbDevices
}
