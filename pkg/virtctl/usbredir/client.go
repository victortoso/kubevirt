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
 * Copyright the KubeVirt Authors.
 *
 */

package usbredir

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/client-go/log"
)

type ClientConnectFn func(device, address string, stop chan struct{}) error

type Client struct {
	// To connect local USB device buffer to the remote VM using the websocket.
	inputReader  *io.PipeReader
	inputWriter  *io.PipeWriter
	outputReader *io.PipeReader
	outputWriter *io.PipeWriter

	listener *net.TCPListener

	stopped bool

	// channels
	done   chan struct{}
	stop   chan struct{}
	stream chan error
	local  chan error
	remote chan error

	ClientConnect ClientConnectFn
}

func NewUSBRedirClient() *Client {
	inReader, inWriter := io.Pipe()
	outReader, outWriter := io.Pipe()
	return &Client{
		stop:          make(chan struct{}),
		inputReader:   inReader,
		inputWriter:   inWriter,
		outputReader:  outReader,
		outputWriter:  outWriter,
		ClientConnect: clientConnect,
	}
}

func (k *Client) Stop() {
	k.stopped = true
	k.stop <- struct{}{}
}

func (k *Client) WithRemoteVMIStream(usbredirStream kubecli.StreamInterface) *Client {
	k.stream = make(chan error)

	go func() {
		defer k.outputWriter.Close()
		errch := make(chan error)
		go func() {
			errch <- usbredirStream.Stream(
				kubecli.StreamOptions{
					In:  k.inputReader,
					Out: k.outputWriter,
				},
			)
		}()
		select {
		case <-k.stop:
		case err := <-errch:
			k.stream <- err
		}
	}()

	return k
}

func (k *Client) WithLocalTCPClient(address string) *Client {
	lnAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Log.Errorf("Can't resolve the address: %s", err.Error())
		return nil
	}

	// The local tcp server is used to proxy between remote websocket and local USB
	k.listener, err = net.ListenTCP("tcp", lnAddr)
	if err != nil {
		log.Log.Errorf("Can't listen on unix socket: %s", err.Error())
		return nil
	}

	return k
}

func (k *Client) ConnectRemote() {
	// forward data to/from websocket after usbredir client connects.
	k.done = make(chan struct{}, 1)
	k.remote = make(chan error)
	go func() {
		defer k.inputWriter.Close()
		start := time.Now()

		usbredirConn, err := k.listener.Accept()
		if err != nil {
			log.Log.V(2).Infof("Failed to accept connection: %s", err.Error())
			k.remote <- err
			return
		}
		defer usbredirConn.Close()

		log.Log.V(2).Infof("Connected to %s at %v", usbredirClient, time.Now().Sub(start))

		stream := make(chan error)
		// write to local usbredir from pipeOutReader
		go func() {
			_, err := io.Copy(usbredirConn, k.outputReader)
			stream <- err
		}()

		// read from local usbredir towards pipeInWriter
		go func() {
			_, err := io.Copy(k.inputWriter, usbredirConn)
			stream <- err
		}()

		select {
		case <-k.stop: // outside request to stop
		case <-k.done: // Wait for local usbredir to complete
		case err = <-stream: // Wait for remote connection to close
			if err == nil && !k.stopped {
				// Remote connection closed, report this as error
				err = fmt.Errorf("Remote connection has closed.")
			}
		}

		// Wait for local usbredir to complete
		k.remote <- err
	}()
}

func clientConnect(device, address string, stop chan struct{}) error {
	bin := usbredirClient
	args := []string{}
	args = append(args, "--device", device, "--to", address)

	log.Log.Infof("port_arg: '%s'", address)
	log.Log.Infof("args: '%v'", args)
	log.Log.Infof("Executing commandline: '%s %v'", bin, args)

	ctx, cancelFn := context.WithCancel(context.Background())
	command := exec.CommandContext(ctx, bin, args...)
	done := make(chan error)
	go func() {
		output, err := command.CombinedOutput()
		log.Log.V(2).Infof("%v output: %v", bin, string(output))
		done <- err
	}()

	var err error
	select {
	case err = <-done:
		if err != nil {
			log.Log.Errorf("Failed to execute %v due %v", bin, err)
		}
	case <-stop:
		// will stop with cancelFn()
	}
	cancelFn()
	return err
}

func (k *Client) ConnectLocal(device string) {
	// execute local usbredir binary
	address := k.listener.Addr().String()
	k.local = make(chan error)
	go func() {
		defer close(k.done)
		k.local <- k.ClientConnect(device, address, k.stop)
	}()
}

func (k *Client) Run() error {
	var err error

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
	case err = <-k.stream:
	case err = <-k.local:
	case err = <-k.remote:
	case <-k.stop:
	}
	return err
}
