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
 * Copyright The KubeVirt Authors.
 *
 */

package hooks

import (
	"context"
	_ "embed"
	"encoding/xml"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/grpc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/util/rand"

	v1 "kubevirt.io/api/core/v1"
	hooksInfo "kubevirt.io/kubevirt/pkg/hooks/info"
	hooksV1alpha3 "kubevirt.io/kubevirt/pkg/hooks/v1alpha3"
	virtwrapApi "kubevirt.io/kubevirt/pkg/virt-launcher/virtwrap/api"
)

type infoServer struct {
	hooksInfo.InfoResult
}

func (s infoServer) Info(
	_ context.Context,
	_ *hooksInfo.InfoParams,
) (*hooksInfo.InfoResult, error) {
	GinkgoWriter.Println("Hook's Info method has been called")
	return &s.InfoResult, nil
}

type callbackServer struct {
	done chan struct{}
}

func (s callbackServer) OnDefineDomain(
	_ context.Context,
	params *hooksV1alpha3.OnDefineDomainParams,
) (*hooksV1alpha3.OnDefineDomainResult, error) {
	GinkgoWriter.Println("Hook's OnDefineDomain method has been called")

	return &hooksV1alpha3.OnDefineDomainResult{
		DomainXML: params.GetDomainXML(),
	}, nil
}

func (s callbackServer) PreCloudInitIso(
	_ context.Context,
	params *hooksV1alpha3.PreCloudInitIsoParams,
) (*hooksV1alpha3.PreCloudInitIsoResult, error) {
	GinkgoWriter.Println("Hook's PreCloudInitIso method has been called")
	return &hooksV1alpha3.PreCloudInitIsoResult{
		CloudInitData: params.GetCloudInitData(),
	}, nil
}

func (s callbackServer) Shutdown(
	_ context.Context,
	_ *hooksV1alpha3.ShutdownParams,
) (*hooksV1alpha3.ShutdownResult, error) {
	GinkgoWriter.Println("Hook's Shutdown method has been called")
	close(s.done)
	return &hooksV1alpha3.ShutdownResult{}, nil
}

type testCase struct {
	socketPath string
	info       infoServer
	callback   callbackServer

	// error from the Run(), will be read on Stop()
	errch  chan error
	ctx    context.Context
	cancel context.CancelFunc
}

// Create boilerplate for the test
func newTestCase(socketDir, name string) *testCase {
	hookPath := filepath.Join(socketDir, "hook-sidecar-"+rand.String(5))
	os.MkdirAll(hookPath, os.ModePerm)
	socketPath := filepath.Join(hookPath, fmt.Sprintf("%s.sock", name))

	ctx, cancel := context.WithCancel(context.Background())
	return &testCase{
		socketPath: socketPath,
		info: infoServer{
			hooksInfo.InfoResult{
				Name:       name,
				Versions:   []string{hooksV1alpha3.Version},
				HookPoints: []*hooksInfo.HookPoint{},
			},
		},
		callback: callbackServer{
			done: make(chan struct{}),
		},
		errch:  make(chan error),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (t *testCase) Run() {
	socket, err := net.Listen("unix", t.socketPath)
	Expect(err).ToNot(HaveOccurred())
	defer socket.Close()

	server := grpc.NewServer([]grpc.ServerOption{}...)

	hooksInfo.RegisterInfoServer(server, t.info)
	hooksV1alpha3.RegisterCallbacksServer(server, t.callback)

	grpcDone := make(chan error)
	go func() {
		GinkgoWriter.Printf("Starting hook server exposing 'info' services on socket %s\n", t.socketPath)
		grpcDone <- server.Serve(socket)
	}()

	select {
	case err = <-grpcDone:
	case <-t.ctx.Done():
		err = t.ctx.Err()
	}

	// Wait Stop() read it
	t.errch <- err
}

func (t *testCase) Stop() error {
	defer os.Remove(t.socketPath)

	// Check if error already
	select {
	case err := <-t.errch:
		return err
	default:
	}

	t.cancel()
	return <-t.errch
}

//go:embed testdata/domain.xml
var domainXML []byte

var _ = Describe("HooksManager", func() {
	Context("With existing sockets", func() {
		var socketDir string

		BeforeEach(func() {
			var err error
			socketDir, err = os.MkdirTemp("", "hooksocketdir")
			Expect(err).ToNot(HaveOccurred())
			os.MkdirAll(socketDir, os.ModePerm)
		})

		It("Should find sidecar", func() {
			hookPointName := hooksInfo.OnDefineDomainHookPointName

			t := newTestCase(socketDir, "hook1")
			t.info.HookPoints = append(t.info.HookPoints, &hooksInfo.HookPoint{
				Name:     hookPointName,
				Priority: 0,
			})
			go t.Run()
			defer t.Stop()

			manager := newManager(socketDir)
			err := manager.Collect(1, 10*time.Second)
			Expect(err).ToNot(HaveOccurred())

			callbackMaps := manager.CallbacksPerHookPoint
			Expect(callbackMaps).Should(HaveKey(hookPointName))
			Expect(callbackMaps[hookPointName]).Should(HaveLen(1))
		})

		It("Should find multiple sidecars on the same hook point", func() {
			hookPointName := hooksInfo.OnDefineDomainHookPointName
			hookNames := []string{"hook1", "hook2"}

			for _, hookName := range hookNames {
				t := newTestCase(socketDir, hookName)
				go t.Run()
				t.info.HookPoints = append(t.info.HookPoints, &hooksInfo.HookPoint{
					Name:     hookPointName,
					Priority: 0,
				})
				defer t.Stop()
			}

			manager := newManager(socketDir)
			err := manager.Collect(uint(len(hookNames)), 10*time.Second)
			Expect(err).ToNot(HaveOccurred())

			callbackMaps := manager.CallbacksPerHookPoint
			Expect(callbackMaps).Should(HaveKey(hookPointName))
			Expect(callbackMaps[hookPointName]).Should(HaveLen(len(hookNames)))
		})

		It("Should find multiple sidecars on different hook points", func() {
			hookNameList := []struct {
				hookName      string
				hookPointName string
			}{
				{"hook1", hooksInfo.OnDefineDomainHookPointName},
				{"hook2", hooksInfo.PreCloudInitIsoHookPointName},
			}
			for _, hook := range hookNameList {
				t := newTestCase(socketDir, hook.hookName)
				go t.Run()
				t.info.HookPoints = append(t.info.HookPoints, &hooksInfo.HookPoint{
					Name:     hook.hookPointName,
					Priority: 0,
				})
				defer t.Stop()
			}

			manager := newManager(socketDir)
			err := manager.Collect(uint(len(hookNameList)), 10*time.Second)
			Expect(err).ToNot(HaveOccurred())

			callbackMaps := manager.CallbacksPerHookPoint

			for _, hook := range hookNameList {
				Expect(callbackMaps).Should(HaveKey(hook.hookPointName))
				Expect(callbackMaps[hook.hookPointName]).Should(HaveLen(1))
			}
		})

		Context("on calling the methods", func() {
			It("should call OnDefineDomain", func() {
				hookPointName := hooksInfo.OnDefineDomainHookPointName

				t := newTestCase(socketDir, "hook1")
				t.info.HookPoints = append(t.info.HookPoints, &hooksInfo.HookPoint{
					Name:     hookPointName,
					Priority: 0,
				})
				go t.Run()
				defer t.Stop()

				manager := newManager(socketDir)
				err := manager.Collect(1, 10*time.Second)
				Expect(err).ToNot(HaveOccurred())

				callbackMaps := manager.CallbacksPerHookPoint
				Expect(callbackMaps).Should(HaveKey(hookPointName))
				Expect(callbackMaps[hookPointName]).Should(HaveLen(1))

				domainSpec := &virtwrapApi.DomainSpec{}
				err = xml.Unmarshal(domainXML, domainSpec)
				Expect(err).ToNot(HaveOccurred())
				Expect(domainSpec).ToNot(BeNil())

				vmi := &v1.VirtualMachineInstance{
					Spec: v1.VirtualMachineInstanceSpec{
						Domain: v1.DomainSpec{
							Devices: v1.Devices{},
						},
					},
				}

				resultXML, err := manager.OnDefineDomain(domainSpec, vmi)
				Expect(err).ToNot(HaveOccurred())

				resultSpec := &virtwrapApi.DomainSpec{}
				err = xml.Unmarshal([]byte(resultXML), resultSpec)
				Expect(err).ToNot(HaveOccurred())
				Expect(domainSpec).To(Equal(resultSpec))
			})
		})

		AfterEach(func() {
			os.RemoveAll(socketDir)
		})
	})
})
