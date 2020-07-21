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
 * Copyright 2018 Red Hat, Inc.
 *
 */

package tests_test

import (
	"fmt"
	"strings"
	"time"

	"kubevirt.io/kubevirt/pkg/hooks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sv1 "k8s.io/api/core/v1"

	v1 "kubevirt.io/client-go/api/v1"
	"kubevirt.io/client-go/kubecli"
	hooksv1alpha1 "kubevirt.io/kubevirt/pkg/hooks/v1alpha1"
	hooksv1alpha2 "kubevirt.io/kubevirt/pkg/hooks/v1alpha2"
	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
	"kubevirt.io/kubevirt/tests"
	"kubevirt.io/kubevirt/tests/flags"
)

const hookSidecarImage = "example-hook-sidecar"

var _ = Describe("HookSidecars", func() {

	var err error
	var virtClient kubecli.KubevirtClient

	var vmi *v1.VirtualMachineInstance

	BeforeEach(func() {
		virtClient, err = kubecli.GetKubevirtClient()
		tests.PanicOnError(err)

		tests.BeforeTestCleanup()
		vmi = tests.NewRandomVMIWithEphemeralDisk(tests.ContainerDiskFor(tests.ContainerDiskAlpine))
		vmi.ObjectMeta.Annotations = RenderSidecar(hooksv1alpha1.Version)
	})

	Describe("[rfe_id:2667][crit:medium][vendor:cnv-qe@redhat.com][level:component] VMI definition", func() {

		Context("with SM BIOS hook sidecar", func() {
			It("[test_id:3155]should successfully start with hook sidecar annotation", func() {
				By("Starting a VMI")
				vmi, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
				Expect(err).ToNot(HaveOccurred())
				tests.WaitForSuccessfulVMIStart(vmi)
			}, 300)

			It("[test_id:3156]should successfully start with hook sidecar annotation for v1alpha2", func() {
				By("Starting a VMI")
				vmi, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
				vmi.ObjectMeta.Annotations = RenderSidecar(hooksv1alpha2.Version)
				Expect(err).ToNot(HaveOccurred())
				tests.WaitForSuccessfulVMIStart(vmi)
			}, 300)

			It("[test_id:3157]should call Collect and OnDefineDomain on the hook sidecar", func() {
				By("Getting hook-sidecar logs")
				vmi, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
				Expect(err).ToNot(HaveOccurred())
				logs := func() string { return getHookSidecarLogs(virtClient, vmi) }
				tests.WaitForSuccessfulVMIStart(vmi)
				Eventually(logs,
					11*time.Second,
					500*time.Millisecond).
					Should(ContainSubstring("Hook's Info method has been called"))
				Eventually(logs,
					11*time.Second,
					500*time.Millisecond).
					Should(ContainSubstring("Hook's OnDefineDomain callback method has been called"))
			}, 300)

			It("[test_id:3158]should update domain XML with SM BIOS properties", func() {
				By("Reading domain XML using virsh")
				tests.SkipIfNoCmd("kubectl")
				vmi, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
				tests.WaitForSuccessfulVMIStart(vmi)
				domainXml := getVmDomainXml(virtClient, vmi)
				Expect(domainXml).Should(ContainSubstring("<sysinfo type='smbios'>"))
				Expect(domainXml).Should(ContainSubstring("<smbios mode='sysinfo'/>"))
				Expect(domainXml).Should(ContainSubstring("<entry name='manufacturer'>Radical Edward</entry>"))
			}, 300)
		})

		Context("with sidecar feature gate disabled", func() {
			BeforeEach(func() {
				tests.DisableFeatureGate(virtconfig.SidecarGate)
			})

			AfterEach(func() {
				tests.EnableFeatureGate(virtconfig.SidecarGate)
			})

			It("[test_id:2666]should not start with hook sidecar annotation", func() {
				By("Starting a VMI")
				vmi, err = virtClient.VirtualMachineInstance(tests.NamespaceTestDefault).Create(vmi)
				Expect(err).To(HaveOccurred(), "should not create a VMI without sidecar feature gate")
				Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("invalid entry metadata.annotations.%s", hooks.HookSidecarListAnnotationName)))
			}, 30)
		})
	})
})

func getHookSidecarLogs(virtCli kubecli.KubevirtClient, vmi *v1.VirtualMachineInstance) string {
	namespace := vmi.GetObjectMeta().GetNamespace()
	podName := tests.GetVmPodName(virtCli, vmi)

	var tailLines int64 = 100
	logsRaw, err := virtCli.CoreV1().
		Pods(namespace).
		GetLogs(podName, &k8sv1.PodLogOptions{
			TailLines: &tailLines,
			Container: "hook-sidecar-0",
		}).
		DoRaw()
	Expect(err).To(BeNil())

	return string(logsRaw)
}

func getVmDomainXml(virtCli kubecli.KubevirtClient, vmi *v1.VirtualMachineInstance) string {
	podName := tests.GetVmPodName(virtCli, vmi)

	// passing an empty namespace allows to position --namespace argument correctly
	vmNameListRaw, _, err := tests.RunCommandWithNS("", "kubectl", "exec", "-ti", "--namespace", vmi.GetObjectMeta().GetNamespace(), podName, "--container", "compute", "--", "virsh", "list", "--name")
	Expect(err).ToNot(HaveOccurred())

	vmName := strings.Split(vmNameListRaw, "\n")[0]
	// passing an empty namespace allows to position --namespace argument correctly
	vmDomainXML, _, err := tests.RunCommandWithNS("", "kubectl", "exec", "-ti", "--namespace", vmi.GetObjectMeta().GetNamespace(), podName, "--container", "compute", "--", "virsh", "dumpxml", vmName)
	Expect(err).ToNot(HaveOccurred())

	return vmDomainXML
}

func RenderSidecar(version string) map[string]string {
	return map[string]string{
		"hooks.kubevirt.io/hookSidecars":              fmt.Sprintf(`[{"args": ["--version", "%s"],"image": "%s/%s:%s", "imagePullPolicy": "IfNotPresent"}]`, version, flags.KubeVirtUtilityRepoPrefix, hookSidecarImage, flags.KubeVirtUtilityVersionTag),
		"smbios.vm.kubevirt.io/baseBoardManufacturer": "Radical Edward",
	}
}
