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

		It("should stop plugin", func() {
			manager.Execute()

			// Wait start
			Eventually(func() bool {
				return manager.state.resourceNameToPluginHandler[resourceName].started
			}, 2*time.Second).Should(BeTrue(), "Plugin should be started")

			feeder.Delete(usbDevicesConfig)
			manager.Execute()

			// Wait stop
			Eventually(func() map[string]*pluginHandler {
				return manager.state.resourceNameToPluginHandler
			}, 2*time.Second).Should(BeEmpty(), "No plugin is running")
		})

		It("should fail due usb devices found in the node", func() {
			manager.discoveryFunc = func() []*usbDevice { return discoveryHelper([]string{}) }
			manager.Execute()
			Expect(manager.state.resourceNameToPluginHandler).To(BeEmpty(), "Should be empty as no usb  was found")
		})

		It("should find the usb device and handle it", func() {
			manager.Execute()

			// Check if resource was added
			Expect(manager.state.resourceNameToPluginHandler).To(HaveKey(resourceName))
			usbPlugin := manager.state.resourceNameToPluginHandler[resourceName].plugin.(*stub)
			Expect(usbPlugin.devices).To(HaveLen(1))
			Expect(usbPlugin.devices).Should(ContainElement(HaveField("Product", toInt("beef"))))

		})

		It("should change the usb device over same config name", func() {
			manager.Execute()

			usbDevicesConfig.Spec.USB.USBHostDevices = []v1alpha1.USBHostDevices{
				{
					SelectByVendorProduct: "dead:cafe",
				},
			}
			feeder.Add(usbDevicesConfig)
			manager.Execute()

			Expect(manager.state.resourceNameToPluginHandler).To(HaveKey(resourceName))
			usbPlugin := manager.state.resourceNameToPluginHandler[resourceName].plugin.(*stub)
			Expect(usbPlugin.devices).To(HaveLen(1))
			Expect(usbPlugin.devices).Should(ContainElement(HaveField("Product", toInt("cafe"))))
		})
	})

	Context("multiple configs", func() {
		var (
			storageName    string
			storageDevices []string
			storageConfig  *v1alpha1.USBDevicesConfig

			miscName    string
			miscDevices []string
			miscConfig  *v1alpha1.USBDevicesConfig
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
			storageConfig = usbDevicesConfigHelper(storageName, storageDevices)
			feeder.Add(storageConfig)

			miscName = "kubevirt.io/usb-misc"
			miscDevices = []string{"dead:dead", "babe:cafe", "babe:face"}
			miscConfig = usbDevicesConfigHelper(miscName, miscDevices)
			feeder.Add(miscConfig)
		})

		It("should start plugin", func() {
			devices := []string{}
			devices = append(devices, storageDevices...)
			devices = append(devices, miscDevices...)
			manager.discoveryFunc = func() []*usbDevice { return discoveryHelper(devices) }

			// Add the first resource and check
			manager.Execute()
			Expect(manager.state.resourceNameToPluginHandler).To(HaveLen(1))
			Expect(manager.state.resourceNameToPluginHandler).To(HaveKey(storageName))

			// Check usb devices from first resource
			usbPlugin := manager.state.resourceNameToPluginHandler[storageName].plugin.(*stub)
			Expect(usbPlugin.devices).To(HaveLen(3))
			Expect(usbPlugin.devices).Should(ContainElement(HaveField("Product", toInt("beef"))))

			// Add second resource and check
			manager.Execute()
			Expect(manager.state.resourceNameToPluginHandler).To(HaveLen(2))
			Expect(manager.state.resourceNameToPluginHandler).To(HaveKey(miscName))

			// Check usb devices from second resource
			usbPlugin = manager.state.resourceNameToPluginHandler[miscName].plugin.(*stub)
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

func toInt(str string) int {
	val, _ := strconv.ParseInt(str, 16, 32)
	return int(val)
}
