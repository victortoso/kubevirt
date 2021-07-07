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
 * Copyright 2021 Red Hat, Inc.
 *
 */

package tests_test

import (
	"context"
	"io"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "kubevirt.io/client-go/api/v1"
	"kubevirt.io/client-go/kubecli"
	"kubevirt.io/client-go/log"
	"kubevirt.io/kubevirt/tests"
	"kubevirt.io/kubevirt/tests/util"
)

// Capabilities from client side
var helloMessageLocal = []byte{
	0x00, 0x00, 0x00, 0x00, 0x44, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x75, 0x73, 0x62, 0x72,
	0x65, 0x64, 0x69, 0x72, 0x20, 0x30, 0x2e, 0x31, 0x30, 0x2e, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
}

// Expected capabilities from QEMU's usbredir
var helloMessageRemote = []byte{
	0x00, 0x00, 0x00, 0x00, 0x44, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x71, 0x65, 0x6d, 0x75,
	0x20, 0x75, 0x73, 0x62, 0x2d, 0x72, 0x65, 0x64, 0x69, 0x72, 0x20, 0x67, 0x75, 0x65, 0x73, 0x74,
	0x20, 0x35, 0x2e, 0x32, 0x2e, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x00, 0x00, 0x00,
}

var _ = Describe("[Serial][crit:medium][vendor:cnv-qe@redhat.com][level:component][sig-compute] USB Redirection", func() {

	var err error
	var virtClient kubecli.KubevirtClient
	var vmi *v1.VirtualMachineInstance

	Describe("[crit:medium][vendor:cnv-qe@redhat.com][level:component] A VirtualMachineInstance without usbredir support", func() {
		tests.BeforeAll(func() {
			virtClient, err = kubecli.GetKubevirtClient()
			util.PanicOnError(err)

			tests.BeforeTestCleanup()
			vmi = tests.NewRandomVMI()
			Expect(virtClient.RestClient().Post().Resource("virtualmachineinstances").Namespace(util.NamespaceTestDefault).Body(vmi).Do(context.Background()).Error()).To(Succeed())
			tests.WaitForSuccessfulVMIStart(vmi)
		})

		FIt("should fail to connect to VMI's usbredir socket", func() {
			usbredirVMI, err := virtClient.VirtualMachineInstance(vmi.ObjectMeta.Namespace).USBRedir(vmi.ObjectMeta.Name)
			Expect(err).To(HaveOccurred())
			Expect(usbredirVMI).To(BeNil())
		})
	})

	Describe("[crit:medium][vendor:cnv-qe@redhat.com][level:component] A VirtualMachineInstance with usbredir support", func() {
		tests.BeforeAll(func() {
			virtClient, err = kubecli.GetKubevirtClient()
			util.PanicOnError(err)

			tests.BeforeTestCleanup()
			vmi = tests.NewRandomVMI()
			vmi.Spec.Domain.Devices.ClientPassthrough = &v1.ClientPassthroughDevices{}
			Expect(virtClient.RestClient().Post().Resource("virtualmachineinstances").Namespace(util.NamespaceTestDefault).Body(vmi).Do(context.Background()).Error()).To(Succeed())
			tests.WaitForSuccessfulVMIStart(vmi)
		})

		Context("with an usbredir connection", func() {

			usbredirConnect := func() {
				pipeInReader, pipeInWriter := io.Pipe()
				pipeOutReader, pipeOutWriter := io.Pipe()
				defer pipeInReader.Close()
				defer pipeOutReader.Close()

				k8ResChan := make(chan error)
				readStop := make(chan []byte)

				By("Stablishing communication with usbredir socket from VMI")
				go func() {
					defer GinkgoRecover()
					usbredirVMI, err := virtClient.VirtualMachineInstance(vmi.ObjectMeta.Namespace).USBRedir(vmi.ObjectMeta.Name)
					if err != nil {
						k8ResChan <- err
						return
					}

					k8ResChan <- usbredirVMI.Stream(kubecli.StreamOptions{
						In:  pipeInReader,
						Out: pipeOutWriter,
					})
				}()

				By("Exchanging hello message between client and QEMU's usbredir")
				go func() {
					defer GinkgoRecover()
					buf := make([]byte, 1024, 1024)

					// write hello message to remote (VMI)
					nw, err := pipeInWriter.Write(helloMessageLocal)
					Expect(nw).To(Equal(len(helloMessageLocal)))
					Expect(err).ToNot(HaveOccurred())

					// reading hello message from remote (VMI)
					nr, err := pipeOutReader.Read(buf)
					if err != nil && err != io.EOF {
						Expect(err).ToNot(HaveOccurred())
						return
					}
					if nr == 0 && err == io.EOF {
						log.Log.Info("zero bytes read from usbredir socket.")
						return
					}
					readStop <- buf[0:nr]
				}()

				select {
				case response := <-readStop:
					By("Checking the response from VNC server")
					Expect(response).To(Equal(helloMessageRemote))
				case err = <-k8ResChan:
					Expect(err).ToNot(HaveOccurred())
				case <-time.After(45 * time.Second):
					Fail("Timout reached while waiting for valid VNC server response")
				}
			}

			FIt("Should work several times", func() {

				for i := 0; i < 10; i++ {
					usbredirConnect()
				}
			})
		})
	})
})
