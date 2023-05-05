package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeConfig represents a subset of Node resources that we want to expose
// to virtual machines.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
// +genclient:nonNamespaced
type NodeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NodeConfigSpec   `json:"spec" valid:"required"`
	Status            NodeConfigStatus `json:"status,omitempty"`
}

type NodeConfigSpec struct {
	USB USB `json:"usb,omitempty"`
}

type USB struct {
	// The resource name which identifies the list of USB host devices.
	// e.g: kubevirt.io/storage for generic storages or kubevirt.io/bootable-usb
	ResourceName string `json:"resourceName"`
	// +listType=atomic
	USBHostDevices []USBHostDevices `json:"usbHostDevices,omitempty"`
}

type USBHostDevices struct {
	// The vendor:product of the devices we want to select.
	// e.g: "0951:1666"
	SelectByVendorProduct string `json:"selectByVendorProduct"`
	// TODO:
	// Optional: If Serial Number is set, use that as a selector.
	// e.g: "E0D55E6CBD23E691583A004D"
	// SelectBySerialNumber:
	// Optional: Use Bus and Device number as selector
	// e.g: "02-03"
	// SelectByBusDeviceNumber
}

// TODO
type NodeConfigStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NodeConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeConfig `json:"items"`
}
