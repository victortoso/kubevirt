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
	lastDeviceIndex := make(map[string]int)
	for _, device := range vmiHostDevices {
		env := util.ResourceNameToEnvVar("USB", device.DeviceName)
		addressString, ok := os.LookupEnv(env)
		if !ok {
			// USB is only part of HostDevices. We can skip if we don't find it.
			log.Log.V(5).Infof("USB environment variable for %s not found", device.DeviceName)
			continue
		}

		index := 0
		if count, exists := lastDeviceIndex[device.DeviceName]; exists {
			index = count + 1
		}
		lastDeviceIndex[device.DeviceName] = index

		values := strings.Split(addressString, ",")
		strs := strings.Split(values[index], ":")
		if len(strs) != 2 {
			log.Log.Warningf("Bad value index=%d in env=%s: %s",
				index, env, addressString)
			continue
		}
		bus, deviceNumber := strs[0], strs[1]
		hostdevices = append(hostdevices,
			api.HostDevice{
				Type:  "usb",
				Mode:  "subsystem",
				Alias: api.NewUserDefinedAlias("usb-host-" + device.Name),
				Source: api.HostDeviceSource{
					Address: &api.Address{
						Bus:    bus,
						Device: deviceNumber,
					},
				},
			})
	}
	return hostdevices, nil
}
