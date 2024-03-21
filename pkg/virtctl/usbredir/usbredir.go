/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2017, 2021 Red Hat, Inc.
 *
 */

package usbredir

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	"kubevirt.io/client-go/kubecli"

	"kubevirt.io/kubevirt/pkg/virtctl/templates"
)

const usbredirClient = "usbredirect"

func NewCommand(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "usbredir (vendor:product)|(bus-device) (VMI)",
		Short:   "Redirect an USB device to a virtual machine instance.",
		Example: usage(),
		Args:    templates.ExactArgs("usb", 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := usbredirCommand{clientConfig: clientConfig}
			return c.Run(cmd, args)
		},
	}
	cmd.SetUsageTemplate(templates.UsageTemplate())
	return cmd
}

type usbredirCommand struct {
	clientConfig clientcmd.ClientConfig
}

func (usbredirCmd *usbredirCommand) Run(command *cobra.Command, args []string) error {
	if _, err := exec.LookPath(usbredirClient); err != nil {
		return fmt.Errorf("Error on finding %s in $PATH: %s", usbredirClient, err.Error())
	}

	namespace, _, err := usbredirCmd.clientConfig.Namespace()
	if err != nil {
		return err
	}

	virtCli, err := kubecli.GetKubevirtClientFromClientConfig(usbredirCmd.clientConfig)
	if err != nil {
		return err
	}

	vmiArg := args[1]
	usbdeviceArg := args[0]

	// Get connection to the websocket for usbredir subresource
	usbredirVMI, err := virtCli.VirtualMachineInstance(namespace).USBRedir(vmiArg)
	if err != nil {
		return fmt.Errorf("Can't access VMI %s: %s", vmiArg, err.Error())
	}

	usbredirClient := NewUSBRedirClient().
		WithRemoteVMIStream(usbredirVMI).
		WithLocalTCPClient("localhost:0")

	usbredirClient.ConnectRemote()
	usbredirClient.ConnectLocal(usbdeviceArg)
	return usbredirClient.Run()
}

func usage() string {
	return `# Find the device you want to redirect (linux):
	$ lsusb | grep Kingston
	Bus 002 Device 003: ID 0951:1666 Kingston Technology DataTraveler 100 G3/G4/SE9 G2/50

	# Redirect it with vendor:product to testvmi:
    {{ProgramName}} usbredir 0951:1666 testvmi

	# Redirect it with bus-device:
    {{ProgramName}} usbredir 02-03 testvmi
	`
}
