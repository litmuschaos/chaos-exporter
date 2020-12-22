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

	v1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"

	chaosClient "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//auth for gcp: optional
	"github.com/litmuschaos/chaos-exporter/controller"
	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = os.Getenv("HOME") + "/.kube/config"
var config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
var client = clients.ClientSets{}

func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

var _ = BeforeSuite(func() {
	client.KubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}
	client.LitmusClient, err = chaosClient.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}

	cmd := exec.Command("kubectl", "apply", "-f", "https://litmuschaos.github.io/litmus/litmus-operator-ci.yaml")
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to create operator: %v", err)
	}
	time.Sleep(30 * time.Second)
	podDeleteRbac := exec.Command("kubectl", "apply", "-f", "../manifest/pod-delete-rbac.yaml", "-n", "litmus")
	if err := podDeleteRbac.Start(); err != nil {
		log.Fatalf("Failed to create pod-delete rbac: %v", err)
	}
	experimentCreate := exec.Command("kubectl", "apply", "-f", "https://hub.litmuschaos.io/api/chaos/master?file=charts/generic/experiments.yaml", "-n", "litmus")
	if err := experimentCreate.Start(); err != nil {
		log.Fatalf("Failed to create experiment: %v", err)
	}
	time.Sleep(30 * time.Second)
	By("Creating nginx deployment")
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "litmus",
			Labels: map[string]string{
				"app": "nginx",
			},
			Annotations: map[string]string{
				"litmuschaos.io/chaos": "true",
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: func(i int32) *int32 { return &i }(3),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "litmus",
					Containers: []v1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Ports: []v1.ContainerPort{
								{

									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := client.KubeClient.AppsV1().Deployments("litmus").Create(deployment)
	if err != nil {
		fmt.Println("Deployment is not created and error is ", err)
	}

	time.Sleep(30 * time.Second)
	cmd = exec.Command("go", "run", "../../cmd/exporter/main.go", "-kubeconfig="+os.Getenv("HOME")+"/.kube/config")
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start exporter: %v", err)
	}

	time.Sleep(10 * time.Second)
	//Creating chaosEngine
	By("Creating ChaosEngine")
	chaosEngine := &v1alpha1.ChaosEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "engine-nginx",
			Namespace: "litmus",
		},
		Spec: v1alpha1.ChaosEngineSpec{
			Appinfo: v1alpha1.ApplicationParams{
				Appns:    "litmus",
				Applabel: "app=nginx",
				AppKind:  "deployment",
			},
			ChaosServiceAccount: "pod-delete-sa",
			Components: v1alpha1.ComponentParams{
				Runner: v1alpha1.RunnerInfo{
					Image: "litmuschaos/chaos-runner:ci",
					Type:  "go",
				},
			},
			JobCleanUpPolicy: "retain",
			Monitoring:       true,
			EngineState:      "active",
			Experiments: []v1alpha1.ExperimentList{
				{
					Name: "pod-delete",
				},
			},
		},
	}

	_, err = client.LitmusClient.ChaosEngines("litmus").Create(chaosEngine)
	Expect(err).To(BeNil())

	time.Sleep(30 * time.Second)
})

var _ = Describe("BDD on chaos-exporter", func() {

	// BDD case 1
	Context("Chaos Engine failed experiments", func() {

		It("should be a zero failed experiments", func() {
			By("Checking experiments metrics")

			err := controller.GetLitmusChaosMetrics(client)
			Expect(err).To(BeNil())

		})
	})

	// BDD case 2
	Context("Curl the prometheus metrics", func() {
		It("Should return prometheus metrics", func() {
			By("Running Exporter and Sending get request to metrics")
			// wait for execution of exporter
			log.Println("\nSleeping for 60 second and wait for exporter to start")
			time.Sleep(120 * time.Second)

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

				By("Should be matched with total_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring("litmuschaos_overall_experiments_run_count{chaosresult_namespace=\"\"} 1"))

				By("Should be matched with failed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring("litmuschaos_overall_failed_experiments{chaosresult_namespace=\"\"} 0"))

				By("Should be matched with passed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring("litmuschaos_overall_passed_experiments{chaosresult_namespace=\"\"} 1"))

				By("Should be matched with engine_failed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring(`litmuschaos_failed_experiments{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 0`))

				By("Should be matched with engine_passed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring(`litmuschaos_passed_experiments{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 1`))

				By("Should be matched with engine_waiting_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring(`litmuschaos_awaited_experiments{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 0`))

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
