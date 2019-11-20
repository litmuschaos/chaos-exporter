/*
Copyright 2019 LitmusChaos Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	chaosEngineV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/litmuschaos/chaos-exporter/controller"
	"github.com/litmuschaos/chaos-exporter/pkg/version"
)

var kubeconfig = os.Getenv("HOME") + "/.kube/config"
var config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
var appUUID = os.Getenv("APP_UUID")

var exporterSpec = controller.ExporterSpec{
	AppNS:       os.Getenv("APP_NAMESPACE"),
	ChaosEngine: os.Getenv("CHAOSENGINE"),
}

func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

var _ = BeforeSuite(func() {

	clientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}
	By("Creating ChaosEngine")
	chaosEngine := &chaosEngineV1alpha1.ChaosEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "engine-nginx",
		},
		Spec: chaosEngineV1alpha1.ChaosEngineSpec{
			Appinfo: chaosEngineV1alpha1.ApplicationParams{
				Appns:    "default",
				Applabel: "app=nginx",
			},
			Experiments: []chaosEngineV1alpha1.ExperimentList{
				{
					Name: "container-kill",
				},
				{
					Name: "pod-kill",
				},
			},
		},
	}
	response, err := clientSet.LitmuschaosV1alpha1().ChaosEngines("litmus").Create(chaosEngine)
	if err != nil {
		fmt.Println("Error while creating ChaosEngine, err: ", err)
	}
	fmt.Println("\nDeployed ChaosEngine:", response)
	Expect(err).To(BeNil())
})

var _ = Describe("BDD on chaos-exporter", func() {

	// BDD case 1
	Context("Chaos Engine failed experiments", func() {

		It("should be a zero failed experiments", func() {

			if err != nil {
				Fail(err.Error())
			}

			By("Checking experiments metrics")
			clientSet, err := clientV1alpha1.NewForConfig(config)
			if err != nil {
				fmt.Println(err)
			}
			expTotal, passTotal, failTotal, expMap, err := controller.GetLitmusChaosMetrics(clientSet, exporterSpec)
			if err != nil {
				Fail(err.Error()) // Unable to get metrics:
			}

			fmt.Println(expTotal, failTotal, passTotal, expMap)

			//failed experiments should be 0
			Expect(failTotal).To(Equal(float64(0)))
			// passed experiments should be 0
			Expect(passTotal).To(Equal(float64(0)))
			// total experiment is 2 because we have mentioned it in the chaosengine spec
			Expect(expTotal).To(Equal(float64(2)))

		})
	})

	// BDD case 2
	Context("Curl the prometheus metrics", func() {
		It("Should return prometheus metrics", func() {

			By("Running Exporter and Sending get request to metrics")
			cmd := exec.Command("go", "run", "../../cmd/exporter/main.go", "-kubeconfig="+os.Getenv("HOME")+"/.kube/config")
			if err := cmd.Start(); err != nil {
				log.Fatalf("Failed to start exporter: %v", err)
			}
			Expect(err).To(BeNil())
			// wait for execution of exporter
			log.Println("\nSleeping for 60 second and wait for exporter to start")
			time.Sleep(60 * time.Second)
			response, err := http.Get("http://127.0.0.1:8080/metrics")
			Expect(err).To(BeNil())
			if err != nil {
				fmt.Printf("%s", err)
				os.Exit(1)
			} else {
				defer response.Body.Close()
				metrics, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Printf("%s", err)
					os.Exit(1)
				}
				fmt.Printf("%s\n", string(metrics))
				var k8sVersion, openEBSVersion string
				k8sClientSet, err := kubernetes.NewForConfig(config)
				if err != nil {
					fmt.Printf("unable to generate kubernetes clientSet %s: ", err)
				}
				k8sVersion, _ = version.GetKubernetesVersion(k8sClientSet)             // getting kubernetes version
				openEBSVersion, _ = version.GetOpenebsVersion(k8sClientSet, "openebs") // getting openEBS Version

				var tmpStr = "{app_uid=\"" + appUUID + "\",engine_name=\"engine-nginx\",kubernetes_version=\"" + k8sVersion + "\",openebs_version=\"" + openEBSVersion + "\"}"

				By("Should be matched with total_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring("c_engine_experiment_count" + tmpStr + " 2"))

				By("Should be matched with failed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring("c_engine_failed_experiments" + tmpStr + " 0"))

				By("Should be matched with passed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring("c_engine_passed_experiments" + tmpStr + " 0"))

				By("Should be matched with container_kill experiment regx")
				Expect(string(metrics)).Should(ContainSubstring("c_exp_container_kill" + tmpStr + " 0"))

				By("Should be matched with pod_kill experiment experiments regx")
				Expect(string(metrics)).Should(ContainSubstring("c_exp_pod_kill" + tmpStr + " 0"))
			}
		})
	})
})

// deleting all unused resources
var _ = AfterSuite(func() {

	By("Deleting chaosengine CRD")
	ceDeleteCRDs := exec.Command("kubectl", "delete", "crds", "chaosengines.litmuschaos.io").Run()
	Expect(ceDeleteCRDs).To(BeNil())

	By("Deleting namespace litmus")
	deleteNS := exec.Command("kubectl", "delete", "ns", "litmus").Run()
	Expect(deleteNS).To(BeNil())

})
