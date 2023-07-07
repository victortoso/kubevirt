package usb

import "kubevirt.io/kubevirt/pkg/controller"

func (manager *USBManager) addFunc(obj interface{}) {
	if key, err := controller.KeyFunc(obj); err == nil {
		manager.queue.Add(key)
	} else {
		manager.logger.Reason(err).Warningf("Failed to add %s", key)
	}
}

func (manager *USBManager) updateFunc(_, updatedObj interface{}) {
	if key, err := controller.KeyFunc(updatedObj); err == nil {
		manager.queue.Add(key)
	} else {
		manager.logger.Reason(err).Warningf("Failed to update %s", key)
	}
}

func (manager *USBManager) deleteFunc(obj interface{}) {
	if key, err := controller.KeyFunc(obj); err == nil {
		manager.queue.Add(key)
	} else {
		manager.logger.Reason(err).Warningf("Failed to delete %s", key)
	}
}
