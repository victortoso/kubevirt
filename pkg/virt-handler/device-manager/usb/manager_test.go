package usb

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	framework "k8s.io/client-go/tools/cache/testing"
	"kubevirt.io/api/usb/v1alpha1"
	"kubevirt.io/kubevirt/pkg/testutils"
)

/* Mock feeder */
type usbDevicesConfigFeeder struct {
	MockQueue *testutils.MockWorkQueue
	Source    *framework.FakeControllerSource
}

func (v *usbDevicesConfigFeeder) Add(vmi *v1alpha1.USBDevicesConfig) {
	v.MockQueue.ExpectAdds(1)
	v.Source.Add(vmi)
	v.MockQueue.Wait()
}

func (v *usbDevicesConfigFeeder) Modify(vmi *v1alpha1.USBDevicesConfig) {
	v.MockQueue.ExpectAdds(1)
	v.Source.Modify(vmi)
	v.MockQueue.Wait()
}

func (v *usbDevicesConfigFeeder) Delete(vmi *v1alpha1.USBDevicesConfig) {
	v.MockQueue.ExpectAdds(1)
	v.Source.Delete(vmi)
	v.MockQueue.Wait()
}

func newUSBDevicesConfigFeeder(queue *testutils.MockWorkQueue, source *framework.FakeControllerSource) *usbDevicesConfigFeeder {
	return &usbDevicesConfigFeeder{
		MockQueue: queue,
		Source:    source,
	}
}

/* Mock usbDevicePlugin */
type stub struct {
	resourceName string
	devices      []*usbDevice
}

func (s stub) Name() string                        { return s.resourceName }
func (s stub) Start(_ <-chan struct{}) (err error) { return nil }

var _ = Describe("USB Manager", func() {
	var (
		manager *USBManager
		feeder  *usbDevicesConfigFeeder
	)

	BeforeEach(func() {
		informer, source := testutils.NewFakeInformerFor(&v1alpha1.USBDevicesConfig{})
		manager = NewUSBManager(informer)
		queue := testutils.NewMockWorkQueue(manager.queue)
		manager.queue = queue
		feeder = newUSBDevicesConfigFeeder(queue, source)
		stop := make(chan struct{})
		DeferCleanup(func() { close(stop) })
		go informer.Run(stop)
		Expect(cache.WaitForCacheSync(stop, informer.HasSynced)).To(BeTrue())
	})

	Context("sanity test", func() {
		var (
			resourceName     string
			nodeDevices      []string
			localDevices     []string
			usbDevicesConfig *v1alpha1.USBDevicesConfig
		)

		BeforeEach(func() {
			resourceName = "kubevirt.io/usb-storage"
			nodeDevices = []string{"dead:beef"}
			localDevices = append(nodeDevices, "dead:cafe")
			usbDevicesConfig = USBDevicesConfigHelper(resourceName, nodeDevices)

			manager.factoryFunc = func(resourceName string, devices []*usbDevice) Plugin {
				return &stub{
					resourceName: resourceName,
					devices:      devices,
				}
			}
			manager.discoveryFunc = func() []*usbDevice { return discoveryHelper(localDevices) }

			feeder.Add(usbDevicesConfig)
		})

		It("should start plugin", func() {
			manager.Execute()

			// Check that plugin properly started
			Eventually(func() bool {
				manager.state.lock.Lock()
				defer manager.state.lock.Unlock()
				return manager.state.resourceNameToPluginHandler[resourceName].started
			}, 2*time.Second).Should(BeTrue(), "Plugin should be started")
		})
	})
})

func discoveryHelper(usbs []string) []*usbDevice {
	var ret []*usbDevice
	for _, usb := range usbs {
		values := strings.Split(usb, ":")
		vendor, _ := strconv.ParseInt(values[0], 16, 32)
		product, _ := strconv.ParseInt(values[1], 16, 32)
		ret = append(ret, &usbDevice{
			Vendor:  int(vendor),
			Product: int(product),
		})
	}
	return ret
}

func usbDevicesConfigHelper(name string, devices []string) *v1alpha1.USBDevicesConfig {
	config := &v1alpha1.USBDevicesConfig{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "test",
			Name:      fmt.Sprintf("test-%s", name),
		},
		Spec: v1alpha1.USBDevicesConfigSpec{
			USB: v1alpha1.USB{
				ResourceName: name,
			},
		},
	}
	for _, str := range devices {
		config.Spec.USB.USBHostDevices = append(config.Spec.USB.USBHostDevices,
			v1alpha1.USBHostDevices{
				SelectByVendorProduct: str,
			})
	}
	return config
}
