// Automatically generated by MockGen. DO NOT EDIT!
// Source: kubevirt.go

package kubecli

import (
	time "time"

	gomock "github.com/golang/mock/gomock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	discovery "k8s.io/client-go/discovery"
	v1alpha1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1alpha1"
	v1beta1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1beta1"
	v10 "k8s.io/client-go/kubernetes/typed/apps/v1"
	v1beta10 "k8s.io/client-go/kubernetes/typed/apps/v1beta1"
	v1beta2 "k8s.io/client-go/kubernetes/typed/apps/v1beta2"
	v11 "k8s.io/client-go/kubernetes/typed/authentication/v1"
	v1beta11 "k8s.io/client-go/kubernetes/typed/authentication/v1beta1"
	v12 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	v1beta12 "k8s.io/client-go/kubernetes/typed/authorization/v1beta1"
	v13 "k8s.io/client-go/kubernetes/typed/autoscaling/v1"
	v2beta1 "k8s.io/client-go/kubernetes/typed/autoscaling/v2beta1"
	v14 "k8s.io/client-go/kubernetes/typed/batch/v1"
	v1beta13 "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	v2alpha1 "k8s.io/client-go/kubernetes/typed/batch/v2alpha1"
	v1beta14 "k8s.io/client-go/kubernetes/typed/certificates/v1beta1"
	v15 "k8s.io/client-go/kubernetes/typed/core/v1"
	v1beta15 "k8s.io/client-go/kubernetes/typed/events/v1beta1"
	v1beta16 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	v16 "k8s.io/client-go/kubernetes/typed/networking/v1"
	v1beta17 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
	v17 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	v1alpha10 "k8s.io/client-go/kubernetes/typed/rbac/v1alpha1"
	v1beta18 "k8s.io/client-go/kubernetes/typed/rbac/v1beta1"
	v1alpha11 "k8s.io/client-go/kubernetes/typed/scheduling/v1alpha1"
	v1alpha12 "k8s.io/client-go/kubernetes/typed/settings/v1alpha1"
	v18 "k8s.io/client-go/kubernetes/typed/storage/v1"
	v1alpha13 "k8s.io/client-go/kubernetes/typed/storage/v1alpha1"
	v1beta19 "k8s.io/client-go/kubernetes/typed/storage/v1beta1"
	rest "k8s.io/client-go/rest"

	versioned "kubevirt.io/containerized-data-importer/pkg/client/clientset/versioned"
	v19 "kubevirt.io/kubevirt/pkg/api/v1"
)

// Mock of KubevirtClient interface
type MockKubevirtClient struct {
	ctrl     *gomock.Controller
	recorder *_MockKubevirtClientRecorder
}

// Recorder for MockKubevirtClient (not exported)
type _MockKubevirtClientRecorder struct {
	mock *MockKubevirtClient
}

func NewMockKubevirtClient(ctrl *gomock.Controller) *MockKubevirtClient {
	mock := &MockKubevirtClient{ctrl: ctrl}
	mock.recorder = &_MockKubevirtClientRecorder{mock}
	return mock
}

func (_m *MockKubevirtClient) EXPECT() *_MockKubevirtClientRecorder {
	return _m.recorder
}

func (_m *MockKubevirtClient) VirtualMachineInstance(namespace string) VirtualMachineInstanceInterface {
	ret := _m.ctrl.Call(_m, "VirtualMachineInstance", namespace)
	ret0, _ := ret[0].(VirtualMachineInstanceInterface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) VirtualMachineInstance(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "VirtualMachineInstance", arg0)
}

func (_m *MockKubevirtClient) ReplicaSet(namespace string) ReplicaSetInterface {
	ret := _m.ctrl.Call(_m, "ReplicaSet", namespace)
	ret0, _ := ret[0].(ReplicaSetInterface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) ReplicaSet(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ReplicaSet", arg0)
}

func (_m *MockKubevirtClient) VirtualMachine(namespace string) VirtualMachineInterface {
	ret := _m.ctrl.Call(_m, "VirtualMachine", namespace)
	ret0, _ := ret[0].(VirtualMachineInterface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) VirtualMachine(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "VirtualMachine", arg0)
}

func (_m *MockKubevirtClient) ServerVersion() *ServerVersion {
	ret := _m.ctrl.Call(_m, "ServerVersion")
	ret0, _ := ret[0].(*ServerVersion)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) ServerVersion() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ServerVersion")
}

func (_m *MockKubevirtClient) RestClient() *rest.RESTClient {
	ret := _m.ctrl.Call(_m, "RestClient")
	ret0, _ := ret[0].(*rest.RESTClient)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) RestClient() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RestClient")
}

func (_m *MockKubevirtClient) CdiClient() versioned.Interface {
	ret := _m.ctrl.Call(_m, "CdiClient")
	ret0, _ := ret[0].(versioned.Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) CdiClient() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CdiClient")
}

func (_m *MockKubevirtClient) Discovery() discovery.DiscoveryInterface {
	ret := _m.ctrl.Call(_m, "Discovery")
	ret0, _ := ret[0].(discovery.DiscoveryInterface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Discovery() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Discovery")
}

func (_m *MockKubevirtClient) AdmissionregistrationV1alpha1() v1alpha1.AdmissionregistrationV1alpha1Interface {
	ret := _m.ctrl.Call(_m, "AdmissionregistrationV1alpha1")
	ret0, _ := ret[0].(v1alpha1.AdmissionregistrationV1alpha1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AdmissionregistrationV1alpha1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AdmissionregistrationV1alpha1")
}

func (_m *MockKubevirtClient) AdmissionregistrationV1beta1() v1beta1.AdmissionregistrationV1beta1Interface {
	ret := _m.ctrl.Call(_m, "AdmissionregistrationV1beta1")
	ret0, _ := ret[0].(v1beta1.AdmissionregistrationV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AdmissionregistrationV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AdmissionregistrationV1beta1")
}

func (_m *MockKubevirtClient) Admissionregistration() v1beta1.AdmissionregistrationV1beta1Interface {
	ret := _m.ctrl.Call(_m, "Admissionregistration")
	ret0, _ := ret[0].(v1beta1.AdmissionregistrationV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Admissionregistration() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Admissionregistration")
}

func (_m *MockKubevirtClient) AppsV1beta1() v1beta10.AppsV1beta1Interface {
	ret := _m.ctrl.Call(_m, "AppsV1beta1")
	ret0, _ := ret[0].(v1beta10.AppsV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AppsV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AppsV1beta1")
}

func (_m *MockKubevirtClient) AppsV1beta2() v1beta2.AppsV1beta2Interface {
	ret := _m.ctrl.Call(_m, "AppsV1beta2")
	ret0, _ := ret[0].(v1beta2.AppsV1beta2Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AppsV1beta2() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AppsV1beta2")
}

func (_m *MockKubevirtClient) AppsV1() v10.AppsV1Interface {
	ret := _m.ctrl.Call(_m, "AppsV1")
	ret0, _ := ret[0].(v10.AppsV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AppsV1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AppsV1")
}

func (_m *MockKubevirtClient) Apps() v10.AppsV1Interface {
	ret := _m.ctrl.Call(_m, "Apps")
	ret0, _ := ret[0].(v10.AppsV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Apps() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Apps")
}

func (_m *MockKubevirtClient) AuthenticationV1() v11.AuthenticationV1Interface {
	ret := _m.ctrl.Call(_m, "AuthenticationV1")
	ret0, _ := ret[0].(v11.AuthenticationV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AuthenticationV1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AuthenticationV1")
}

func (_m *MockKubevirtClient) Authentication() v11.AuthenticationV1Interface {
	ret := _m.ctrl.Call(_m, "Authentication")
	ret0, _ := ret[0].(v11.AuthenticationV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Authentication() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Authentication")
}

func (_m *MockKubevirtClient) AuthenticationV1beta1() v1beta11.AuthenticationV1beta1Interface {
	ret := _m.ctrl.Call(_m, "AuthenticationV1beta1")
	ret0, _ := ret[0].(v1beta11.AuthenticationV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AuthenticationV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AuthenticationV1beta1")
}

func (_m *MockKubevirtClient) AuthorizationV1() v12.AuthorizationV1Interface {
	ret := _m.ctrl.Call(_m, "AuthorizationV1")
	ret0, _ := ret[0].(v12.AuthorizationV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AuthorizationV1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AuthorizationV1")
}

func (_m *MockKubevirtClient) Authorization() v12.AuthorizationV1Interface {
	ret := _m.ctrl.Call(_m, "Authorization")
	ret0, _ := ret[0].(v12.AuthorizationV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Authorization() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Authorization")
}

func (_m *MockKubevirtClient) AuthorizationV1beta1() v1beta12.AuthorizationV1beta1Interface {
	ret := _m.ctrl.Call(_m, "AuthorizationV1beta1")
	ret0, _ := ret[0].(v1beta12.AuthorizationV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AuthorizationV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AuthorizationV1beta1")
}

func (_m *MockKubevirtClient) AutoscalingV1() v13.AutoscalingV1Interface {
	ret := _m.ctrl.Call(_m, "AutoscalingV1")
	ret0, _ := ret[0].(v13.AutoscalingV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AutoscalingV1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AutoscalingV1")
}

func (_m *MockKubevirtClient) Autoscaling() v13.AutoscalingV1Interface {
	ret := _m.ctrl.Call(_m, "Autoscaling")
	ret0, _ := ret[0].(v13.AutoscalingV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Autoscaling() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Autoscaling")
}

func (_m *MockKubevirtClient) AutoscalingV2beta1() v2beta1.AutoscalingV2beta1Interface {
	ret := _m.ctrl.Call(_m, "AutoscalingV2beta1")
	ret0, _ := ret[0].(v2beta1.AutoscalingV2beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) AutoscalingV2beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "AutoscalingV2beta1")
}

func (_m *MockKubevirtClient) BatchV1() v14.BatchV1Interface {
	ret := _m.ctrl.Call(_m, "BatchV1")
	ret0, _ := ret[0].(v14.BatchV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) BatchV1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "BatchV1")
}

func (_m *MockKubevirtClient) Batch() v14.BatchV1Interface {
	ret := _m.ctrl.Call(_m, "Batch")
	ret0, _ := ret[0].(v14.BatchV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Batch() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Batch")
}

func (_m *MockKubevirtClient) BatchV1beta1() v1beta13.BatchV1beta1Interface {
	ret := _m.ctrl.Call(_m, "BatchV1beta1")
	ret0, _ := ret[0].(v1beta13.BatchV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) BatchV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "BatchV1beta1")
}

func (_m *MockKubevirtClient) BatchV2alpha1() v2alpha1.BatchV2alpha1Interface {
	ret := _m.ctrl.Call(_m, "BatchV2alpha1")
	ret0, _ := ret[0].(v2alpha1.BatchV2alpha1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) BatchV2alpha1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "BatchV2alpha1")
}

func (_m *MockKubevirtClient) CertificatesV1beta1() v1beta14.CertificatesV1beta1Interface {
	ret := _m.ctrl.Call(_m, "CertificatesV1beta1")
	ret0, _ := ret[0].(v1beta14.CertificatesV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) CertificatesV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CertificatesV1beta1")
}

func (_m *MockKubevirtClient) Certificates() v1beta14.CertificatesV1beta1Interface {
	ret := _m.ctrl.Call(_m, "Certificates")
	ret0, _ := ret[0].(v1beta14.CertificatesV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Certificates() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Certificates")
}

func (_m *MockKubevirtClient) CoreV1() v15.CoreV1Interface {
	ret := _m.ctrl.Call(_m, "CoreV1")
	ret0, _ := ret[0].(v15.CoreV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) CoreV1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CoreV1")
}

func (_m *MockKubevirtClient) Core() v15.CoreV1Interface {
	ret := _m.ctrl.Call(_m, "Core")
	ret0, _ := ret[0].(v15.CoreV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Core() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Core")
}

func (_m *MockKubevirtClient) EventsV1beta1() v1beta15.EventsV1beta1Interface {
	ret := _m.ctrl.Call(_m, "EventsV1beta1")
	ret0, _ := ret[0].(v1beta15.EventsV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) EventsV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "EventsV1beta1")
}

func (_m *MockKubevirtClient) Events() v1beta15.EventsV1beta1Interface {
	ret := _m.ctrl.Call(_m, "Events")
	ret0, _ := ret[0].(v1beta15.EventsV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Events() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Events")
}

func (_m *MockKubevirtClient) ExtensionsV1beta1() v1beta16.ExtensionsV1beta1Interface {
	ret := _m.ctrl.Call(_m, "ExtensionsV1beta1")
	ret0, _ := ret[0].(v1beta16.ExtensionsV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) ExtensionsV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ExtensionsV1beta1")
}

func (_m *MockKubevirtClient) Extensions() v1beta16.ExtensionsV1beta1Interface {
	ret := _m.ctrl.Call(_m, "Extensions")
	ret0, _ := ret[0].(v1beta16.ExtensionsV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Extensions() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Extensions")
}

func (_m *MockKubevirtClient) NetworkingV1() v16.NetworkingV1Interface {
	ret := _m.ctrl.Call(_m, "NetworkingV1")
	ret0, _ := ret[0].(v16.NetworkingV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) NetworkingV1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "NetworkingV1")
}

func (_m *MockKubevirtClient) Networking() v16.NetworkingV1Interface {
	ret := _m.ctrl.Call(_m, "Networking")
	ret0, _ := ret[0].(v16.NetworkingV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Networking() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Networking")
}

func (_m *MockKubevirtClient) PolicyV1beta1() v1beta17.PolicyV1beta1Interface {
	ret := _m.ctrl.Call(_m, "PolicyV1beta1")
	ret0, _ := ret[0].(v1beta17.PolicyV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) PolicyV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PolicyV1beta1")
}

func (_m *MockKubevirtClient) Policy() v1beta17.PolicyV1beta1Interface {
	ret := _m.ctrl.Call(_m, "Policy")
	ret0, _ := ret[0].(v1beta17.PolicyV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Policy() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Policy")
}

func (_m *MockKubevirtClient) RbacV1() v17.RbacV1Interface {
	ret := _m.ctrl.Call(_m, "RbacV1")
	ret0, _ := ret[0].(v17.RbacV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) RbacV1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RbacV1")
}

func (_m *MockKubevirtClient) Rbac() v17.RbacV1Interface {
	ret := _m.ctrl.Call(_m, "Rbac")
	ret0, _ := ret[0].(v17.RbacV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Rbac() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Rbac")
}

func (_m *MockKubevirtClient) RbacV1beta1() v1beta18.RbacV1beta1Interface {
	ret := _m.ctrl.Call(_m, "RbacV1beta1")
	ret0, _ := ret[0].(v1beta18.RbacV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) RbacV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RbacV1beta1")
}

func (_m *MockKubevirtClient) RbacV1alpha1() v1alpha10.RbacV1alpha1Interface {
	ret := _m.ctrl.Call(_m, "RbacV1alpha1")
	ret0, _ := ret[0].(v1alpha10.RbacV1alpha1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) RbacV1alpha1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "RbacV1alpha1")
}

func (_m *MockKubevirtClient) SchedulingV1alpha1() v1alpha11.SchedulingV1alpha1Interface {
	ret := _m.ctrl.Call(_m, "SchedulingV1alpha1")
	ret0, _ := ret[0].(v1alpha11.SchedulingV1alpha1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) SchedulingV1alpha1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SchedulingV1alpha1")
}

func (_m *MockKubevirtClient) Scheduling() v1alpha11.SchedulingV1alpha1Interface {
	ret := _m.ctrl.Call(_m, "Scheduling")
	ret0, _ := ret[0].(v1alpha11.SchedulingV1alpha1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Scheduling() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Scheduling")
}

func (_m *MockKubevirtClient) SettingsV1alpha1() v1alpha12.SettingsV1alpha1Interface {
	ret := _m.ctrl.Call(_m, "SettingsV1alpha1")
	ret0, _ := ret[0].(v1alpha12.SettingsV1alpha1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) SettingsV1alpha1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SettingsV1alpha1")
}

func (_m *MockKubevirtClient) Settings() v1alpha12.SettingsV1alpha1Interface {
	ret := _m.ctrl.Call(_m, "Settings")
	ret0, _ := ret[0].(v1alpha12.SettingsV1alpha1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Settings() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Settings")
}

func (_m *MockKubevirtClient) StorageV1beta1() v1beta19.StorageV1beta1Interface {
	ret := _m.ctrl.Call(_m, "StorageV1beta1")
	ret0, _ := ret[0].(v1beta19.StorageV1beta1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) StorageV1beta1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "StorageV1beta1")
}

func (_m *MockKubevirtClient) StorageV1() v18.StorageV1Interface {
	ret := _m.ctrl.Call(_m, "StorageV1")
	ret0, _ := ret[0].(v18.StorageV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) StorageV1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "StorageV1")
}

func (_m *MockKubevirtClient) Storage() v18.StorageV1Interface {
	ret := _m.ctrl.Call(_m, "Storage")
	ret0, _ := ret[0].(v18.StorageV1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) Storage() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Storage")
}

func (_m *MockKubevirtClient) StorageV1alpha1() v1alpha13.StorageV1alpha1Interface {
	ret := _m.ctrl.Call(_m, "StorageV1alpha1")
	ret0, _ := ret[0].(v1alpha13.StorageV1alpha1Interface)
	return ret0
}

func (_mr *_MockKubevirtClientRecorder) StorageV1alpha1() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "StorageV1alpha1")
}

// Mock of StreamInterface interface
type MockStreamInterface struct {
	ctrl     *gomock.Controller
	recorder *_MockStreamInterfaceRecorder
}

// Recorder for MockStreamInterface (not exported)
type _MockStreamInterfaceRecorder struct {
	mock *MockStreamInterface
}

func NewMockStreamInterface(ctrl *gomock.Controller) *MockStreamInterface {
	mock := &MockStreamInterface{ctrl: ctrl}
	mock.recorder = &_MockStreamInterfaceRecorder{mock}
	return mock
}

func (_m *MockStreamInterface) EXPECT() *_MockStreamInterfaceRecorder {
	return _m.recorder
}

func (_m *MockStreamInterface) Stream(options StreamOptions) error {
	ret := _m.ctrl.Call(_m, "Stream", options)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockStreamInterfaceRecorder) Stream(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Stream", arg0)
}

// Mock of VirtualMachineInstanceInterface interface
type MockVirtualMachineInstanceInterface struct {
	ctrl     *gomock.Controller
	recorder *_MockVirtualMachineInstanceInterfaceRecorder
}

// Recorder for MockVirtualMachineInstanceInterface (not exported)
type _MockVirtualMachineInstanceInterfaceRecorder struct {
	mock *MockVirtualMachineInstanceInterface
}

func NewMockVirtualMachineInstanceInterface(ctrl *gomock.Controller) *MockVirtualMachineInstanceInterface {
	mock := &MockVirtualMachineInstanceInterface{ctrl: ctrl}
	mock.recorder = &_MockVirtualMachineInstanceInterfaceRecorder{mock}
	return mock
}

func (_m *MockVirtualMachineInstanceInterface) EXPECT() *_MockVirtualMachineInstanceInterfaceRecorder {
	return _m.recorder
}

func (_m *MockVirtualMachineInstanceInterface) Get(name string, options *v1.GetOptions) (*v19.VirtualMachineInstance, error) {
	ret := _m.ctrl.Call(_m, "Get", name, options)
	ret0, _ := ret[0].(*v19.VirtualMachineInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInstanceInterfaceRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get", arg0, arg1)
}

func (_m *MockVirtualMachineInstanceInterface) List(opts *v1.ListOptions) (*v19.VirtualMachineInstanceList, error) {
	ret := _m.ctrl.Call(_m, "List", opts)
	ret0, _ := ret[0].(*v19.VirtualMachineInstanceList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInstanceInterfaceRecorder) List(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "List", arg0)
}

func (_m *MockVirtualMachineInstanceInterface) Create(instance *v19.VirtualMachineInstance) (*v19.VirtualMachineInstance, error) {
	ret := _m.ctrl.Call(_m, "Create", instance)
	ret0, _ := ret[0].(*v19.VirtualMachineInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInstanceInterfaceRecorder) Create(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Create", arg0)
}

func (_m *MockVirtualMachineInstanceInterface) Update(_param0 *v19.VirtualMachineInstance) (*v19.VirtualMachineInstance, error) {
	ret := _m.ctrl.Call(_m, "Update", _param0)
	ret0, _ := ret[0].(*v19.VirtualMachineInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInstanceInterfaceRecorder) Update(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Update", arg0)
}

func (_m *MockVirtualMachineInstanceInterface) Delete(name string, options *v1.DeleteOptions) error {
	ret := _m.ctrl.Call(_m, "Delete", name, options)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockVirtualMachineInstanceInterfaceRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Delete", arg0, arg1)
}

func (_m *MockVirtualMachineInstanceInterface) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v19.VirtualMachineInstance, error) {
	_s := []interface{}{name, pt, data}
	for _, _x := range subresources {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "Patch", _s...)
	ret0, _ := ret[0].(*v19.VirtualMachineInstance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInstanceInterfaceRecorder) Patch(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Patch", _s...)
}

func (_m *MockVirtualMachineInstanceInterface) SerialConsole(name string, timeout time.Duration) (StreamInterface, error) {
	ret := _m.ctrl.Call(_m, "SerialConsole", name, timeout)
	ret0, _ := ret[0].(StreamInterface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInstanceInterfaceRecorder) SerialConsole(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "SerialConsole", arg0, arg1)
}

func (_m *MockVirtualMachineInstanceInterface) VNC(name string) (StreamInterface, error) {
	ret := _m.ctrl.Call(_m, "VNC", name)
	ret0, _ := ret[0].(StreamInterface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInstanceInterfaceRecorder) VNC(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "VNC", arg0)
}

// Mock of ReplicaSetInterface interface
type MockReplicaSetInterface struct {
	ctrl     *gomock.Controller
	recorder *_MockReplicaSetInterfaceRecorder
}

// Recorder for MockReplicaSetInterface (not exported)
type _MockReplicaSetInterfaceRecorder struct {
	mock *MockReplicaSetInterface
}

func NewMockReplicaSetInterface(ctrl *gomock.Controller) *MockReplicaSetInterface {
	mock := &MockReplicaSetInterface{ctrl: ctrl}
	mock.recorder = &_MockReplicaSetInterfaceRecorder{mock}
	return mock
}

func (_m *MockReplicaSetInterface) EXPECT() *_MockReplicaSetInterfaceRecorder {
	return _m.recorder
}

func (_m *MockReplicaSetInterface) Get(name string, options v1.GetOptions) (*v19.VirtualMachineInstanceReplicaSet, error) {
	ret := _m.ctrl.Call(_m, "Get", name, options)
	ret0, _ := ret[0].(*v19.VirtualMachineInstanceReplicaSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockReplicaSetInterfaceRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get", arg0, arg1)
}

func (_m *MockReplicaSetInterface) List(opts v1.ListOptions) (*v19.VirtualMachineInstanceReplicaSetList, error) {
	ret := _m.ctrl.Call(_m, "List", opts)
	ret0, _ := ret[0].(*v19.VirtualMachineInstanceReplicaSetList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockReplicaSetInterfaceRecorder) List(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "List", arg0)
}

func (_m *MockReplicaSetInterface) Create(_param0 *v19.VirtualMachineInstanceReplicaSet) (*v19.VirtualMachineInstanceReplicaSet, error) {
	ret := _m.ctrl.Call(_m, "Create", _param0)
	ret0, _ := ret[0].(*v19.VirtualMachineInstanceReplicaSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockReplicaSetInterfaceRecorder) Create(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Create", arg0)
}

func (_m *MockReplicaSetInterface) Update(_param0 *v19.VirtualMachineInstanceReplicaSet) (*v19.VirtualMachineInstanceReplicaSet, error) {
	ret := _m.ctrl.Call(_m, "Update", _param0)
	ret0, _ := ret[0].(*v19.VirtualMachineInstanceReplicaSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockReplicaSetInterfaceRecorder) Update(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Update", arg0)
}

func (_m *MockReplicaSetInterface) Delete(name string, options *v1.DeleteOptions) error {
	ret := _m.ctrl.Call(_m, "Delete", name, options)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockReplicaSetInterfaceRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Delete", arg0, arg1)
}

// Mock of VMIPresetInterface interface
type MockVMIPresetInterface struct {
	ctrl     *gomock.Controller
	recorder *_MockVMIPresetInterfaceRecorder
}

// Recorder for MockVMIPresetInterface (not exported)
type _MockVMIPresetInterfaceRecorder struct {
	mock *MockVMIPresetInterface
}

func NewMockVMIPresetInterface(ctrl *gomock.Controller) *MockVMIPresetInterface {
	mock := &MockVMIPresetInterface{ctrl: ctrl}
	mock.recorder = &_MockVMIPresetInterfaceRecorder{mock}
	return mock
}

func (_m *MockVMIPresetInterface) EXPECT() *_MockVMIPresetInterfaceRecorder {
	return _m.recorder
}

func (_m *MockVMIPresetInterface) Get(name string, options v1.GetOptions) (*v19.VirtualMachineInstancePreset, error) {
	ret := _m.ctrl.Call(_m, "Get", name, options)
	ret0, _ := ret[0].(*v19.VirtualMachineInstancePreset)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVMIPresetInterfaceRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get", arg0, arg1)
}

func (_m *MockVMIPresetInterface) List(opts v1.ListOptions) (*v19.VirtualMachineInstancePresetList, error) {
	ret := _m.ctrl.Call(_m, "List", opts)
	ret0, _ := ret[0].(*v19.VirtualMachineInstancePresetList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVMIPresetInterfaceRecorder) List(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "List", arg0)
}

func (_m *MockVMIPresetInterface) Create(_param0 *v19.VirtualMachineInstancePreset) (*v19.VirtualMachineInstancePreset, error) {
	ret := _m.ctrl.Call(_m, "Create", _param0)
	ret0, _ := ret[0].(*v19.VirtualMachineInstancePreset)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVMIPresetInterfaceRecorder) Create(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Create", arg0)
}

func (_m *MockVMIPresetInterface) Update(_param0 *v19.VirtualMachineInstancePreset) (*v19.VirtualMachineInstancePreset, error) {
	ret := _m.ctrl.Call(_m, "Update", _param0)
	ret0, _ := ret[0].(*v19.VirtualMachineInstancePreset)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVMIPresetInterfaceRecorder) Update(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Update", arg0)
}

func (_m *MockVMIPresetInterface) Delete(name string, options *v1.DeleteOptions) error {
	ret := _m.ctrl.Call(_m, "Delete", name, options)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockVMIPresetInterfaceRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Delete", arg0, arg1)
}

func (_m *MockVMIPresetInterface) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v19.VirtualMachineInstancePreset, error) {
	_s := []interface{}{name, pt, data}
	for _, _x := range subresources {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "Patch", _s...)
	ret0, _ := ret[0].(*v19.VirtualMachineInstancePreset)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVMIPresetInterfaceRecorder) Patch(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Patch", _s...)
}

// Mock of VirtualMachineInterface interface
type MockVirtualMachineInterface struct {
	ctrl     *gomock.Controller
	recorder *_MockVirtualMachineInterfaceRecorder
}

// Recorder for MockVirtualMachineInterface (not exported)
type _MockVirtualMachineInterfaceRecorder struct {
	mock *MockVirtualMachineInterface
}

func NewMockVirtualMachineInterface(ctrl *gomock.Controller) *MockVirtualMachineInterface {
	mock := &MockVirtualMachineInterface{ctrl: ctrl}
	mock.recorder = &_MockVirtualMachineInterfaceRecorder{mock}
	return mock
}

func (_m *MockVirtualMachineInterface) EXPECT() *_MockVirtualMachineInterfaceRecorder {
	return _m.recorder
}

func (_m *MockVirtualMachineInterface) Get(name string, options *v1.GetOptions) (*v19.VirtualMachine, error) {
	ret := _m.ctrl.Call(_m, "Get", name, options)
	ret0, _ := ret[0].(*v19.VirtualMachine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInterfaceRecorder) Get(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Get", arg0, arg1)
}

func (_m *MockVirtualMachineInterface) List(opts *v1.ListOptions) (*v19.VirtualMachineList, error) {
	ret := _m.ctrl.Call(_m, "List", opts)
	ret0, _ := ret[0].(*v19.VirtualMachineList)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInterfaceRecorder) List(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "List", arg0)
}

func (_m *MockVirtualMachineInterface) Create(_param0 *v19.VirtualMachine) (*v19.VirtualMachine, error) {
	ret := _m.ctrl.Call(_m, "Create", _param0)
	ret0, _ := ret[0].(*v19.VirtualMachine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInterfaceRecorder) Create(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Create", arg0)
}

func (_m *MockVirtualMachineInterface) Update(_param0 *v19.VirtualMachine) (*v19.VirtualMachine, error) {
	ret := _m.ctrl.Call(_m, "Update", _param0)
	ret0, _ := ret[0].(*v19.VirtualMachine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInterfaceRecorder) Update(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Update", arg0)
}

func (_m *MockVirtualMachineInterface) Delete(name string, options *v1.DeleteOptions) error {
	ret := _m.ctrl.Call(_m, "Delete", name, options)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockVirtualMachineInterfaceRecorder) Delete(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Delete", arg0, arg1)
}

func (_m *MockVirtualMachineInterface) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v19.VirtualMachine, error) {
	_s := []interface{}{name, pt, data}
	for _, _x := range subresources {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "Patch", _s...)
	ret0, _ := ret[0].(*v19.VirtualMachine)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVirtualMachineInterfaceRecorder) Patch(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Patch", _s...)
}
