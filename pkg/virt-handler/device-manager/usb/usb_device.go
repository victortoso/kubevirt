package usb

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"google.golang.org/grpc"
	"kubevirt.io/client-go/log"

	"kubevirt.io/kubevirt/pkg/safepath"
	"kubevirt.io/kubevirt/pkg/util"
	device_manager "kubevirt.io/kubevirt/pkg/virt-handler/device-manager"
	devicepluginapi "kubevirt.io/kubevirt/pkg/virt-handler/device-manager/deviceplugin/v1beta1"
)

var _ Plugin = &usbDevicePlugin{}

type Plugin interface {
	Start(stop <-chan struct{}) (err error)
	Name() string
}

type factoryFuncType func(resourceName string, usbdevs []*usbDevice) Plugin

// The sysfs metadata wrapper for the USB devices
type usbDevice struct {
	Name         string
	Manufacturer string
	Vendor       int
	Product      int
	BCD          int
	Bus          int
	DeviceNumber int
	Serial       string
	DevicePath   string
}

// The uniqueness in the system comes from bus and device number but having the vendor:product
// information can help a lot. Not all usb devices provide or export a serial number.
func (dev *usbDevice) GetID() string {
	return fmt.Sprintf("%04x:%04x-%02d:%02d", dev.Vendor, dev.Product, dev.Bus, dev.DeviceNumber)
}

func (dev *usbDevice) toKubeVirtDevicePlugin() *devicepluginapi.Device {
	return &devicepluginapi.Device{
		ID:       dev.GetID(),
		Health:   devicepluginapi.Healthy,
		Topology: nil,
	}
}

// The actual plugin
type usbDevicePlugin struct {
	socketPath   string
	stop         <-chan struct{}
	done         chan struct{}
	deregistered chan struct{}
	server       *grpc.Server
	resourceName string
	devices      []*usbDevice
	logger       *log.FilteredLogger
}

func (plugin *usbDevicePlugin) devicesToKubeVirtDevicePlugin() []*devicepluginapi.Device {
	devices := make([]*devicepluginapi.Device, 0, len(plugin.devices))
	for _, usb := range plugin.devices {
		devices = append(devices, usb.toKubeVirtDevicePlugin())
	}
	return devices
}

var _ devicepluginapi.DevicePluginServer = &usbDevicePlugin{}

func (plugin *usbDevicePlugin) Name() string {
	return plugin.resourceName
}

func (plugin *usbDevicePlugin) stopDevicePlugin() error {
	defer func() {
		select {
		case <-plugin.done:
			return
		default:
			close(plugin.done)
		}
	}()

	// Give the device plugin one second to properly deregister
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	select {
	case <-plugin.deregistered:
	case <-ticker.C:
	}

	plugin.server.Stop()
	return plugin.cleanup()
}

func (plugin *usbDevicePlugin) Start(stop <-chan struct{}) error {
	plugin.stop = stop
	plugin.done = make(chan struct{})
	plugin.deregistered = make(chan struct{})

	err := plugin.cleanup()
	if err != nil {
		return fmt.Errorf("error on cleanup: %v", err)
	}

	sock, err := net.Listen("unix", plugin.socketPath)
	if err != nil {
		return fmt.Errorf("error creating GRPC server socket: %v", err)
	}

	plugin.server = grpc.NewServer([]grpc.ServerOption{}...)
	defer plugin.stopDevicePlugin()

	devicepluginapi.RegisterDevicePluginServer(plugin.server, plugin)

	errChan := make(chan error, 2)

	go func() {
		errChan <- plugin.server.Serve(sock)
	}()

	err = device_manager.WaitForGRPCServer(plugin.socketPath, 5*time.Second)
	if err != nil {
		return fmt.Errorf("error starting the GRPC server: %v", err)
	}

	err = plugin.register()
	if err != nil {
		return fmt.Errorf("error registering with device plugin manager: %v", err)
	}

	plugin.logger.Infof("%s device plugin started", plugin.resourceName)
	err = <-errChan
	return err
}

func (plugin *usbDevicePlugin) cleanup() error {
	err := os.Remove(plugin.socketPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (plugin *usbDevicePlugin) register() error {
	conn, err := grpc.Dial(devicepluginapi.KubeletSocket,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := devicepluginapi.NewRegistrationClient(conn)
	reqt := &devicepluginapi.RegisterRequest{
		Version:      devicepluginapi.Version,
		Endpoint:     path.Base(plugin.socketPath),
		ResourceName: plugin.Name(),
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

func (plugin *usbDevicePlugin) GetDevicePluginOptions(ctx context.Context, _ *devicepluginapi.Empty) (*devicepluginapi.DevicePluginOptions, error) {
	return &devicepluginapi.DevicePluginOptions{
		PreStartRequired: false,
	}, nil
}

// Interface to expose Devices: IDs, health and Topology
func (plugin *usbDevicePlugin) ListAndWatch(_ *devicepluginapi.Empty, lws devicepluginapi.DevicePlugin_ListAndWatchServer) error {
	response := devicepluginapi.ListAndWatchResponse{
		Devices: plugin.devicesToKubeVirtDevicePlugin(),
	}
	if err := lws.Send(&response); err != nil {
		plugin.logger.Reason(err).Warningf("Failed to send device plugin %s",
			plugin.resourceName)
		return err
	}

	// At the moment, we only send the list of devices upon creation. Unplugged devices should be
	// marked as unhealthy and USBDevicesConfig changes should delete/create new Plugin
	done := false
	for !done {
		select {
		// TODO add a health check, e.g usb was unplugged
		case <-plugin.stop:
			done = true
		}
	}

	response = devicepluginapi.ListAndWatchResponse{
		Devices: []*devicepluginapi.Device{},
	}
	if err := lws.Send(&response); err != nil {
		plugin.logger.Reason(err).Warningf("Failed to deregister device plugin %s",
			plugin.resourceName)
	}
	close(plugin.deregistered)
	return nil
}

// Interface to allocate requested Device, exported by ListAndWatch
func (plugin *usbDevicePlugin) Allocate(_ context.Context, allocRequest *devicepluginapi.AllocateRequest) (*devicepluginapi.AllocateResponse, error) {
	allocResponse := new(devicepluginapi.AllocateResponse)
	env := make(map[string]string)
	for _, request := range allocRequest.ContainerRequests {
		containerResponse := &devicepluginapi.ContainerAllocateResponse{}
		for _, id := range request.DevicesIDs {
			plugin.logger.V(5).Infof("usb device id: %s", id)

			var dev *usbDevice
			for _, usb := range plugin.devices {
				if usb.GetID() == id {
					dev = usb
					break
				}
			}

			if dev == nil {
				plugin.logger.V(5).Infof("usb disappeared: %s", id)
				continue
			}

			spath, err := safepath.JoinAndResolveWithRelativeRoot(dev.DevicePath)
			if err != nil {
				return nil, fmt.Errorf("error opening the socket %s: %v", dev.DevicePath, err)
			}

			err = safepath.ChownAtNoFollow(spath, util.NonRootUID, util.NonRootUID)
			if err != nil {
				return nil, fmt.Errorf("error setting the permission the socket %s: %v", dev.DevicePath, err)
			}

			// We might have more than one USB device per resource name
			key := util.ResourceNameToEnvVar("USB", plugin.resourceName)
			value := fmt.Sprintf("%d:%d", dev.Bus, dev.DeviceNumber)
			if previous, exist := env[key]; exist {
				env[key] = fmt.Sprintf("%s,%s", previous, value)
			} else {
				env[key] = value
			}
			containerResponse.Envs = env
			containerResponse.Devices = append(containerResponse.Devices,
				&devicepluginapi.DeviceSpec{
					ContainerPath: dev.DevicePath,
					HostPath:      dev.DevicePath,
					Permissions:   "mrw",
				})
		}
		allocResponse.ContainerResponses = append(allocResponse.ContainerResponses, containerResponse)
	}

	return allocResponse, nil
}

func (plugin *usbDevicePlugin) PreStartContainer(context.Context, *devicepluginapi.PreStartContainerRequest) (*devicepluginapi.PreStartContainerResponse, error) {
	return &devicepluginapi.PreStartContainerResponse{}, nil
}

func NewUSBDevicePlugin(resourceName string, usbdevs []*usbDevice) Plugin {
	s := strings.Split(resourceName, "/")
	resourceID := s[0]
	if len(s) > 1 {
		resourceID = s[1]
	}
	loggerID := fmt.Sprintf("usb-device-%s", resourceID)
	plugin := &usbDevicePlugin{
		socketPath:   device_manager.SocketPath(resourceID),
		resourceName: resourceName,
		devices:      usbdevs,
		logger:       log.Log.With("subcomponent", loggerID),
	}
	return plugin
}
