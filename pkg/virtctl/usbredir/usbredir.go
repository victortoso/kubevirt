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
 * Copyright 2017, 2020 Red Hat, Inc.
 *
 */

package usbredir

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"

	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/kubevirt/pkg/virtctl/templates"
)

const (
	USBREDIRSERVER = "usbredirect"
)

func NewCommand(clientConfig clientcmd.ClientConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "usbredir (vendor:product) (resource name) (VMI)",
		Short:   "Redirect a usb device to a virtual machine instance.",
		Example: usage(),
		Args:    templates.ExactArgs("usb", 3),
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
	if _, err := exec.LookPath(USBREDIRSERVER); err != nil {
		return fmt.Errorf("Error on finding %s in $PATH: %s", USBREDIRSERVER, err.Error())
	}

	namespace, _, err := usbredirCmd.clientConfig.Namespace()
	if err != nil {
		return err
	}

	virtCli, err := kubecli.GetKubevirtClientFromClientConfig(usbredirCmd.clientConfig)
	if err != nil {
		return err
	}

	vmiArg := args[2]
	usbNameArg := args[1]
	usbdeviceArg := args[0]

	glog.V(2).Infof("FIXME: Handle USB label name properly: %v", usbNameArg)

	// Get connection to the websocket for usbredir subresource
	usbredirVMI, err := virtCli.VirtualMachineInstance(namespace).USBRedir(vmiArg)
	if err != nil {
		return fmt.Errorf("Can't access VMI %s: %s", vmiArg, err.Error())
	}

	// We will connect the local USB device using a usbredir TCP client to the
	// remote VM using the websocket.
	pipeInReader, pipeInWriter := io.Pipe()
	pipeOutReader, pipeOutWriter := io.Pipe()

	// Configure in/out and start stream with websocket
	k8ResChan := make(chan error)
	go func() {
		defer pipeOutWriter.Close()
		k8ResChan <- usbredirVMI.Stream(kubecli.StreamOptions{
			In:  pipeInReader,
			Out: pipeOutWriter,
		})
	}()

	lnAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:0"))
	if err != nil {
		return fmt.Errorf("Can't resolve the address: %s", err.Error())
	}

	// The local tcp server is used to proxy between remote websocket and local USB
	ln, err := net.ListenTCP("tcp", lnAddr)
	if err != nil {
		return fmt.Errorf("Can't listen on unix socket: %s", err.Error())
	}

	// forward data to/from websocket after usbredir client connects.
	usbredirDoneChan := make(chan struct{}, 1)
	streamResChan := make(chan error)
	streamStop := make(chan error)
	go func() {
		defer pipeInWriter.Close()
		start := time.Now()

		usbredirConn, err := ln.Accept()
		if err != nil {
			glog.V(2).Infof("Failed to accept connection: %s", err.Error())
			streamResChan <- err
			return
		}
		defer usbredirConn.Close()

		glog.V(2).Infof("Connected to usbredirserver %v", time.Now().Sub(start))

		// write to usbredirserver from pipeOutReader
		go func() {
			_, err := io.Copy(usbredirConn, pipeOutReader)
			streamStop <- err
		}()

		// read from usbredirserver towards pipeInWriter
		go func() {
			_, err := io.Copy(pipeInWriter, usbredirConn)
			streamStop <- err
		}()

		// Wait for usbredirserver to complete
		<-usbredirDoneChan
		streamResChan <- err
	}()

	port := ln.Addr().(*net.TCPAddr).Port

	// execute usbredirserver
	usbredirResultChan := make(chan error)
	go func() {
		defer close(usbredirDoneChan)

		bin := USBREDIRSERVER
		args := []string{}
		port_arg := fmt.Sprintf("localhost:%v", port)
		args = append(args, "--device", usbdeviceArg, "--to", port_arg)

		glog.Infof("port_arg: '%s'", port_arg)
		glog.Infof("args: '%v'", args)
		glog.Infof("Executing commandline: '%s %v'", bin, args)

		command := exec.Command(bin, args...)
		output, err := command.CombinedOutput()
		if err != nil {
			glog.Errorf("Failed to execut %v due %v, output: %v", bin, err, string(output))
		} else {
			glog.V(2).Infof("%v output: %v", bin, string(output))
		}
		usbredirResultChan <- err
	}()

	sigStopChan := make(chan struct{}, 1)
	go func() {
		defer close(sigStopChan)
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
	}()

	select {
	case <-sigStopChan:
	case err = <-streamStop:
	case err = <-k8ResChan:
	case err = <-usbredirResultChan:
	case err = <-streamResChan:
	}

	if err != nil {
		return fmt.Errorf("Error encountered: %s", err.Error())
	}
	return nil
}

func usage() string {
	return `  # Redirect a local USB device to the remote VMI:\n"
  {{ProgramName}} usbredir vendor:product testvmi`
}
