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
	// A handler per resource name
	plugins map[string]*pluginHandler
	lock    sync.Mutex
	logger  *log.FilteredLogger
}

func newState() state {
	return state{
		plugins: map[string]*pluginHandler{},
		lock:    sync.Mutex{},
		logger:  log.Log.With("subcomponent", "usb-manager-state"),
	}
}

func (s *state) insert(plugin Plugin) chan struct{} {
	s.lock.Lock()
	defer s.lock.Unlock()

	close := make(chan struct{})
	resourceName := plugin.Name()
	s.plugins[resourceName] = &pluginHandler{
		started:  false,
		failed:   false,
		stopChan: close,
		plugin:   plugin,
	}
	return close
}

func (s *state) clean() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for resourceName, handler := range s.plugins {
		close(handler.stopChan)
		delete(s.plugins, resourceName)
		s.logger.V(5).Infof("Removed %s", resourceName)
	}
}

func (s *state) updateHandler(resourceName string, started bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	handler, exist := s.plugins[resourceName]
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
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "virt-handler-usb-devices-config")
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

func (manager *USBManager) Run(stopCh chan struct{}) {
	defer manager.queue.ShutDown()

	manager.logger.Info("Starting USB manager")

	cache.WaitForCacheSync(stopCh, manager.usbDevicesConfigInformer.HasSynced)

	//
	go wait.Until(manager.runWorker, time.Second, stopCh)
	manager.logger.Info("Started USB manager")

	<-stopCh
	manager.logger.Info("Stoping USB manager")
}

func (manager *USBManager) runWorker() {
	key, quit := manager.queue.Get()
	if quit {
		manager.logger.V(5).Info("Queue signals to exit")
		return
	}
	defer manager.queue.Done(key)

	err := manager.execute(key.(string))
	if err != nil {
		manager.logger.Reason(err).Infof("re-enqueuing USBDevicesConfig %v", key)
		manager.queue.AddRateLimited(key)
	}

	manager.logger.V(5).Infof("processed USBDevicesConfig %v", key)
	manager.queue.Forget(key)
}

func (manager *USBManager) execute(key string) error {
	obj, exists, err := manager.usbDevicesConfigInformer.GetStore().GetByKey(key)

	if err != nil {
		return fmt.Errorf("failed to get object for key %s, %v", key, err)
	} else if !exists || obj == nil {
		manager.logger.V(5).Infof("processed USBDevicesConfig %v", key)
		manager.state.clean()
		return nil
	}

	// If key already exists, cleanup before proceeding
	manager.state.clean()

	usbDevicesConfig := obj.(*v1alpha1.USBDevicesConfig)
	return manager.syncDevicePlugin(usbDevicesConfig)
}

func constructPermittedUSBDevicesMap(usbDevicesConfig *v1alpha1.USBDevicesConfig) map[int][]usbDeviceSelector {
	// Iterate over requested USB Devices and map it vendor:product
	permittedUSBDevices := make(map[int][]usbDeviceSelector)
	for _, usb := range usbDevicesConfig.Spec.USB {
		resourceName := usb.ResourceName
		for index, dev := range usb.USBHostDevices {
			values := strings.Split(dev.SelectByVendorProduct, ":")
			if len(values) != 2 {
				log.Log.Warningf("Failed to parse USBHostDevices[%d] = %s",
					index, dev.SelectByVendorProduct)
				continue
			}
			val, err := strconv.ParseInt(values[0], 16, 32)
			if err != nil {
				log.Log.Warningf("Failed to convert vendor from base16 string to int: %s",
					dev.SelectByVendorProduct[:sep])
				continue
			}
			vendor := int(val)

			val, err = strconv.ParseInt(values[1], 16, 32)
			if err != nil {
				log.Log.Warningf("Failed to convert product from base16 string to int: %s",
					dev.SelectByVendorProduct[:sep])
				continue
			}
			product := int(val)

			// TODO: For the moment, we consider that only a single resource can hold a product:vendor.
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

func (manager *USBManager) syncDevicePlugin(usbDevicesConfig *v1alpha1.USBDevicesConfig) error {
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
		manager.startPlugin(plugin)
	}
	return nil
}

func (manager *USBManager) startPlugin(plugin Plugin) {
	var stop chan struct{}
	if stop = manager.state.insert(plugin); stop == nil {
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
