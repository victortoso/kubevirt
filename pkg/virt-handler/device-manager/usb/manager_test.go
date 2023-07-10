package usb

import (
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
			usbDevicesConfig = usbDevicesConfigHelper(nil, resourceName, nodeDevices)

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
				if plugin, exist := manager.state.plugins[resourceName]; exist {
					return plugin.started
				}
				return false
			}, 2*time.Second).Should(BeTrue(), "Plugin should be started")
		})

		It("should stop plugin", func() {
			manager.Execute()

			// Wait start
			Eventually(func() bool {
				if plugin, exist := manager.state.plugins[resourceName]; exist {
					return plugin.started
				}
				return false
			}, 2*time.Second).Should(BeTrue(), "Plugin should be started")

			feeder.Delete(usbDevicesConfig)
			manager.Execute()

			// Wait stop
			Eventually(func() map[string]*pluginHandler {
				return manager.state.plugins
			}, 2*time.Second).Should(BeEmpty(), "No plugin is running")
		})

		It("should fail due usb devices found in the node", func() {
			manager.discoveryFunc = func() []*usbDevice { return discoveryHelper([]string{}) }
			manager.Execute()
			Expect(manager.state.plugins).To(BeEmpty(), "Should be empty as no usb  was found")
		})

		It("should find the usb device and handle it", func() {
			manager.Execute()

			// Check if resource was added
			Expect(manager.state.plugins).To(HaveKey(resourceName))
			usbPlugin := manager.state.plugins[resourceName].plugin.(*stub)
			Expect(usbPlugin.devices).To(HaveLen(1))
			Expect(usbPlugin.devices).Should(ContainElement(HaveField("Product", toInt("beef"))))

		})

		It("should change the usb device over same config name", func() {
			manager.Execute()

			usbDevicesConfig.Spec.USB[0].USBHostDevices = []v1alpha1.USBHostDevices{
				{
					SelectByVendorProduct: "dead:cafe",
				},
			}
			feeder.Add(usbDevicesConfig)
			manager.Execute()

			Expect(manager.state.plugins).To(HaveKey(resourceName))
			usbPlugin := manager.state.plugins[resourceName].plugin.(*stub)
			Expect(usbPlugin.devices).To(HaveLen(1))
			Expect(usbPlugin.devices).Should(ContainElement(HaveField("Product", toInt("cafe"))))
		})
	})

	Context("multiple configs", func() {
		var (
			usbDevicesConfig *v1alpha1.USBDevicesConfig

			storageName    string
			storageDevices []string

			miscName    string
			miscDevices []string
		)

		BeforeEach(func() {
			// Using stub is not a must here as we don´t wait till the plugin start, in which case
			// it would fail due lack of real sysfs. Still, better be safe instead of introducing
			// possible flaky tests.
			manager.factoryFunc = func(resourceName string, devices []*usbDevice) Plugin {
				return &stub{
					resourceName: resourceName,
					devices:      devices,
				}
			}

			storageName = "kubevirt.io/usb-storage"
			storageDevices = []string{"dead:beef", "dead:cafe", "dead:face"}
			usbDevicesConfig = usbDevicesConfigHelper(nil, storageName, storageDevices)
			feeder.Add(usbDevicesConfig)

			miscName = "kubevirt.io/usb-misc"
			miscDevices = []string{"dead:dead", "babe:cafe", "babe:face"}
		})

		It("should start plugin", func() {
			devices := []string{}
			devices = append(devices, storageDevices...)
			devices = append(devices, miscDevices...)
			manager.discoveryFunc = func() []*usbDevice { return discoveryHelper(devices) }

			// Add the first resource and check
			manager.Execute()
			Expect(manager.state.plugins).To(HaveLen(1))
			Expect(manager.state.plugins).To(HaveKey(storageName))

			// Check usb devices from first resource
			usbPlugin := manager.state.plugins[storageName].plugin.(*stub)
			Expect(usbPlugin.devices).To(HaveLen(3))
			Expect(usbPlugin.devices).Should(ContainElement(HaveField("Product", toInt("beef"))))

			// Add second resource and check
			usbDevicesConfig = usbDevicesConfigHelper(usbDevicesConfig, miscName, miscDevices)
			feeder.Add(usbDevicesConfig)
			manager.Execute()

			Expect(manager.state.plugins).To(HaveLen(2))
			Expect(manager.state.plugins).To(HaveKey(miscName))

			// Check usb devices from second resource
			usbPlugin = manager.state.plugins[miscName].plugin.(*stub)
			Expect(usbPlugin.devices).To(HaveLen(3))
			Expect(usbPlugin.devices).Should(ContainElement(HaveField("Product", toInt("dead"))))
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

func usbDevicesConfigHelper(config *v1alpha1.USBDevicesConfig, resourceName string, devices []string) *v1alpha1.USBDevicesConfig {
	// For this resourceName
	usbs := []v1alpha1.USBHostDevices{}
	for _, str := range devices {
		usbs = append(usbs, v1alpha1.USBHostDevices{
			SelectByVendorProduct: str,
		})
	}

	if config == nil {
		config = &v1alpha1.USBDevicesConfig{
			ObjectMeta: v1.ObjectMeta{
				Namespace: "test",
				Name:      "test-namespaced",
			},
		}
	}
	config.Spec.USB = append(config.Spec.USB, v1alpha1.USB{
		ResourceName:   resourceName,
		USBHostDevices: usbs,
	})
	return config
}

func toInt(str string) int {
	val, _ := strconv.ParseInt(str, 16, 32)
	return int(val)
}
