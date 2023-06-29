package usb

import (
	"context"
	"fmt"
	"strings"
	"time"

	expect "github.com/google/goexpect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/api/usb/v1alpha1"
	usbv1alpha1 "kubevirt.io/client-go/generated/kubevirt/clientset/versioned/typed/usb/v1alpha1"
	"kubevirt.io/client-go/kubecli"

	"kubevirt.io/kubevirt/tests"
	"kubevirt.io/kubevirt/tests/console"
	"kubevirt.io/kubevirt/tests/decorators"
	"kubevirt.io/kubevirt/tests/framework/cleanup"
	"kubevirt.io/kubevirt/tests/framework/kubevirt"
	"kubevirt.io/kubevirt/tests/libvmi"
	"kubevirt.io/kubevirt/tests/util"
)

var _ = Describe("[sig-compute-realtime][Serial]USB", Serial, decorators.SigCompute, func() {
	var (
		virtClient kubecli.KubevirtClient
		usbClient  usbv1alpha1.UsbV1alpha1Interface
	)

	BeforeEach(func() {
		virtClient = kubevirt.Client()
		usbClient = virtClient.GeneratedKubeVirtClient().UsbV1alpha1()
	})

	Context("valid USBDevicesconfig", func() {

		const resourceName = "kubevirt.io/usb-storage"

		createTestUSBDevicesConfig := func() *v1alpha1.USBDevicesConfig {
			usbDevicesConfig := &v1alpha1.USBDevicesConfig{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test",
					Labels: map[string]string{
						cleanup.TestLabelForNamespace(util.NamespaceTestDefault): "true",
					},
				},
				Spec: v1alpha1.USBDevicesConfigSpec{
					// Validate there is at least one usb
					USB: []v1alpha1.USB{
						{
							ResourceName: resourceName,
							USBHostDevices: []v1alpha1.USBHostDevices{
								{
									// Todo we could do VendorProductSelector and validate it
									SelectByVendorProduct: "46f4:0001",
								},
							},
						},
					},
				},
			}
			usbDevicesConfig, err := usbClient.USBDevicesConfigs(util.NamespaceTestDefault).Create(context.TODO(), usbDevicesConfig, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred())
			return usbDevicesConfig
		}

		It("sanity check", func() {
			usbDevicesConfig := createTestUSBDevicesConfig()

			expectResourceToAppear(virtClient, resourceName)

			By("Deleting node config")
			err := usbClient.USBDevicesConfigs(util.NamespaceTestDefault).Delete(context.TODO(), usbDevicesConfig.Name, metav1.DeleteOptions{})
			Expect(err).ToNot(HaveOccurred())

			expectResourceToDisappear(virtClient, resourceName)
		})

		It("able to consume with VM", func() {
			_ = tests.EnableFeatureGate("HostDevices")

			_ = createTestUSBDevicesConfig()

			expectResourceToAppear(virtClient, resourceName)

			vmi := libvmi.NewCirros(withUsb(resourceName))
			vmi = tests.RunVMIAndExpectLaunch(vmi, 90)

			By("Check there is usb device available inside guest")
			Expect(console.LoginToCirros(vmi)).To(Succeed())
			Expect(console.SafeExpectBatch(vmi, []expect.Batcher{
				&expect.BSnd{S: "dmesg | grep -i vendor=46f4 | wc -l\n"},
				&expect.BExp{R: "1"},
			}, 200)).To(Succeed())
		})

	})
})

func expectResourceToAppear(virtClient kubecli.KubevirtClient, resourceName string) {
	By("Expecting to find usb devices")
	retriveResourceAssertion(virtClient, resourceName).Should(BeNumerically(">", 0), fmt.Sprintf("Failed to find %s on any node", resourceName))
}

func expectResourceToDisappear(virtClient kubecli.KubevirtClient, resourceName string) {
	By("Expecting to not find usb devices")
	retriveResourceAssertion(virtClient, resourceName).Should(BeNumerically("==", 0), fmt.Sprintf("Failed to remove %s on any node", resourceName))
}

func retriveResourceAssertion(virtClient kubecli.KubevirtClient, resourceName string) types.AsyncAssertion {
	return Eventually(func() int64 {
		nodes, err := virtClient.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		Expect(err).ToNot(HaveOccurred())

		for _, node := range nodes.Items {
			for key, amount := range node.Status.Capacity {
				if strings.HasPrefix(string(key), resourceName) {
					value, ok := amount.AsInt64()
					Expect(ok).To(BeTrue())
					return value
				}
			}
		}
		return 0
	}, 90*time.Second, time.Second)
}

func withUsb(resourceName string) libvmi.Option {
	return func(vmi *v1.VirtualMachineInstance) {
		vmi.Spec.Domain.Devices.HostDevices = append(vmi.Spec.Domain.Devices.HostDevices,
			v1.HostDevice{
				Name:       "usb-storage",
				DeviceName: resourceName,
			},
		)
	}
}
