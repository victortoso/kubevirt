package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// USBDevicesConfig represents a subset of USB devices that we want to expose
// to virtual machines.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
type USBDevicesConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              USBDevicesConfigSpec   `json:"spec" valid:"required"`
	Status            USBDevicesConfigStatus `json:"status,omitempty"`
}

type USBDevicesConfigSpec struct {
	// +listType=atomic
	USB []USB `json:"usb,omitempty"`
}

// USB defines a group of USB devices based on its selectors and combine them under a single
// resource name, to be allocated to the VM
type USB struct {
	// Identifies the list of USB host devices.
	// e.g: kubevirt.io/storage, kubevirt.io/bootable-usb, etc
	ResourceName string `json:"resourceName"`
	// +listType=atomic
	USBHostDevices []USBHostDevices `json:"usbHostDevices,omitempty"`
}

type USBHostDevices struct {
	// The vendor:product of the devices we want to select.
	// e.g: "0951:1666"
	SelectByVendorProduct string `json:"selectByVendorProduct"`
}

// TODO
type USBDevicesConfigStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type USBDevicesConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []USBDevicesConfig `json:"items"`
}
