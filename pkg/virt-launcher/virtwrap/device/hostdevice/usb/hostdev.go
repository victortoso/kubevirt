package usb

import (
	"os"
	"strings"

	v1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/log"

	"kubevirt.io/kubevirt/pkg/util"
	"kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

func CreateHostDevices(vmiHostDevices []v1.HostDevice) ([]api.HostDevice, error) {
	hostdevices := []api.HostDevice{}
	for _, device := range vmiHostDevices {
		env := util.ResourceNameToEnvVar("USB", device.DeviceName)
		addressString, ok := os.LookupEnv(env)
		if !ok {
			// USB is only part of HostDevices. We can skip if we don't find it.
			log.Log.V(5).Infof("USB environment variable for %s not found", device.DeviceName)
			continue
		}
		evnS := strings.Split(addressString, ":")
		bus, device := evnS[0], evnS[1]
		hostdevices = append(hostdevices,
			api.HostDevice{
				Type:  "usb",
				Mode:  "subsystem",
				Alias: api.NewUserDefinedAlias("usb-host"),
				Source: api.HostDeviceSource{
					Address: &api.Address{
						Bus:    bus,
						Device: device,
					},
				},
			})
	}
	return hostdevices, nil
}
