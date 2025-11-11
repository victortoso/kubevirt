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
	"strconv"
	"time"

	"google.golang.org/grpc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
	p := hooksInfo.InfoResult(s)
	return &p, nil
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

func hookListenAndServe(socketPath string, hookName string, hookPointName string, hookPointPriority int32) (net.Listener, error) {
	socket, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer([]grpc.ServerOption{}...)
	hooksInfo.RegisterInfoServer(server, infoServer{
		Name:     hookName,
		Versions: []string{hooksV1alpha3.Version},
		HookPoints: []*hooksInfo.HookPoint{
			{
				Name:     hookPointName,
				Priority: hookPointPriority,
			},
		},
	})
	hooksV1alpha3.RegisterCallbacksServer(server, callbackServer{done: make(chan struct{})})
	go func() {
		GinkgoWriter.Printf("Starting hook server exposing 'info' services on socket %s\n", socketPath)
		server.Serve(socket)
	}()
	return socket, nil
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

			hookPath := filepath.Join(socketDir, "hook-sidecar-0")
			os.MkdirAll(hookPath, os.ModePerm)
			socketPath := filepath.Join(hookPath, "hook1.sock")
			socket, err := hookListenAndServe(socketPath, "hook1", hookPointName, 0)
			Expect(err).ToNot(HaveOccurred())
			defer socket.Close()
			defer os.Remove(socketPath)

			manager := newManager(socketDir)
			err = manager.Collect(1, 10*time.Second)
			Expect(err).ToNot(HaveOccurred())

			callbackMaps := manager.CallbacksPerHookPoint
			Expect(callbackMaps).Should(HaveKey(hookPointName))
			Expect(callbackMaps[hookPointName]).Should(HaveLen(1))
		})

		It("Should find multiple sidecars on the same hook point", func() {
			hookPointName := hooksInfo.OnDefineDomainHookPointName
			hookNames := []string{"hook1", "hook2"}

			for i, hookName := range hookNames {
				hookPath := filepath.Join(socketDir, "hook-sidecar-"+strconv.Itoa(i))
				os.MkdirAll(hookPath, os.ModePerm)
				socketPath := filepath.Join(hookPath, fmt.Sprintf("%s.sock", hookName))
				socket, err := hookListenAndServe(socketPath, hookName, hookPointName, 0)
				Expect(err).ToNot(HaveOccurred())
				defer socket.Close()
				defer os.Remove(socketPath)
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
			for i, hook := range hookNameList {
				hookPath := filepath.Join(socketDir, "hook-sidecar-"+strconv.Itoa(i))
				os.MkdirAll(hookPath, os.ModePerm)
				socketPath := filepath.Join(hookPath, fmt.Sprintf("%s.sock", hook.hookName))
				socket, err := hookListenAndServe(socketPath, hook.hookName, hook.hookPointName, 0)
				Expect(err).ToNot(HaveOccurred())
				defer socket.Close()
				defer os.Remove(socketPath)
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

				hookPath := filepath.Join(socketDir, "hook-sidecar-0")
				os.MkdirAll(hookPath, os.ModePerm)
				socketPath := filepath.Join(hookPath, "hook1.sock")
				socket, err := hookListenAndServe(socketPath, "hook1", hookPointName, 0)
				Expect(err).ToNot(HaveOccurred())
				defer socket.Close()
				defer os.Remove(socketPath)

				manager := newManager(socketDir)
				err = manager.Collect(1, 10*time.Second)
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
