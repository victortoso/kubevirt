package usb

import "kubevirt.io/kubevirt/pkg/controller"

func (manager *USBManager) addFunc(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err == nil {
		manager.queue.Add(key)
	}
}

func (manager *USBManager) updateFunc(_, updatedObj interface{}) {
	key, err := controller.KeyFunc(updatedObj)
	if err == nil {
		manager.queue.Add(key)
	}
}

func (manager *USBManager) deleteFunc(obj interface{}) {
	key, err := controller.KeyFunc(obj)
	if err == nil {
		manager.queue.Add(key)
	}
}
