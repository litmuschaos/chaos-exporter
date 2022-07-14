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
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	v1alpha1 "github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	"github.com/litmuschaos/litmus-go/pkg/utils/retry"
	"github.com/pkg/errors"

	chaosClient "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"github.com/litmuschaos/chaos-exporter/pkg/log"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	client     clients.ClientSets
	kubeconfig string
)

func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", os.Getenv("HOME")+"/.kube/config", "path to kubeconfig to invoke kubernetes API calls")
}

var _ = BeforeSuite(func() {

	// Getting kubeconfig and generate clientSets
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	Expect(err).To(BeNil(), "failed to get config")

	client.KubeClient, err = kubernetes.NewForConfig(config)
	Expect(err).To(BeNil(), "failed to generate k8sClientSet")

	client.LitmusClient, err = chaosClient.NewForConfig(config)
	Expect(err).To(BeNil(), "failed to generate litmusClientSet")

	By("Installing Litmus")
	err = exec.Command("kubectl", "apply", "-f", "https://litmuschaos.github.io/litmus/litmus-operator-ci.yaml").Run()
	Expect(err).To(BeNil(), "unable to install litmus")

	err = retry.
		Times(uint(180 / 2)).
		Wait(time.Duration(2) * time.Second).
		Try(func(attempt uint) error {
			podSpec, err := client.KubeClient.CoreV1().Pods("litmus").List(context.Background(), metav1.ListOptions{LabelSelector: "name=chaos-operator"})
			if err != nil || len(podSpec.Items) == 0 {
				return errors.Errorf("unable to list chaos-operator, err: %v", err)
			}
			for _, v := range podSpec.Items {
				if v.Status.Phase != "Running" {
					return errors.Errorf("chaos-operator is not in running state, phase: %v", v.Status.Phase)
				}
			}
			return nil
		})

	Expect(err).To(BeNil(), "the chaos-operator is not in running state")
	log.Info("litmus installed successfully")

	By("Installing RBAC")
	err = exec.Command("kubectl", "apply", "-f", "../manifest/pod-delete-rbac.yaml", "-n", "litmus").Run()
	Expect(err).To(BeNil(), "unable to create RBAC Permissions")
	log.Info("RBAC created")

	By("Installing Generic Experiments")
	err = exec.Command("kubectl", "apply", "-f", "https://hub.litmuschaos.io/api/chaos/master?file=charts/generic/experiments.yaml", "-n", "litmus").Run()
	Expect(err).To(BeNil(), "unable to install experiments")
	log.Info("generic experiments created")

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
	_, err = client.KubeClient.AppsV1().Deployments("litmus").Create(context.Background(), deployment, metav1.CreateOptions{})
	Expect(err).To(
		BeNil(),
		"while creating nginx deployment in namespace litmus",
	)
	log.Info("nginx deployment created")

	cmd := exec.Command("go", "run", "../../cmd/exporter/main.go", "-kubeconfig="+kubeconfig)
	err = cmd.Start()
	Expect(err).To(
		BeNil(),
		"failed while started chaos-exporter",
	)

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
			EngineState:      "active",
			Experiments: []v1alpha1.ExperimentList{
				{
					Name: "pod-delete",
					Spec: v1alpha1.ExperimentAttributes{
						Components: v1alpha1.ExperimentComponents{
							ExperimentImage: "litmuschaos/go-runner:ci",
						},
					},
				},
			},
		},
	}

	_, err = client.LitmusClient.ChaosEngines("litmus").Create(context.Background(), chaosEngine, metav1.CreateOptions{})
	Expect(err).To(
		BeNil(),
		"while building ChaosEngine engine-nginx in namespace litmus",
	)

	log.Info("chaos engine created")
})

var _ = Describe("BDD on chaos-exporter", func() {

	// BDD case 1
	Context("Check availabiity of chaos-runner", func() {
		It("chaos-runner should be present", func() {
			err := retry.
				Times(uint(180 / 2)).
				Wait(time.Duration(2) * time.Second).
				Try(func(attempt uint) error {
					pod, err := client.KubeClient.CoreV1().Pods("litmus").Get(context.Background(), "engine-nginx-runner", metav1.GetOptions{})
					if err != nil {
						return errors.Errorf("unable to get chaos-runner pod, err: %v", err)
					}
					if pod.Status.Phase != v1.PodRunning && pod.Status.Phase != v1.PodSucceeded {
						return errors.Errorf("chaos runner is not in running state, phase: %v", pod.Status.Phase)
					}
					return nil
				})

			if err != nil {
				log.Errorf("The chaos-runner is not in running state, err: %v", err)
			}
			log.Info("runner pod created")
		})
	})

	// BDD case 2
	Context("Curl the prometheus metrics", func() {
		It("Should return prometheus metrics", func() {
			By("Running Exporter and Sending get request to metrics")
			// wait for execution of exporter
			log.Info("Sleeping for 120 second and wait for exporter to start")
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
				Expect(string(metrics)).Should(ContainSubstring("litmuschaos_cluster_scoped_experiments_run_count 1"))

				By("Should be matched with failed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring("litmuschaos_cluster_scoped_failed_experiments 0"))

				By("Should be matched with passed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring("litmuschaos_cluster_scoped_passed_experiments 1"))

				By("Should be matched with engine_failed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring(`litmuschaos_failed_experiments{chaosengine_context="",chaosengine_name="engine-nginx",chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 0`))

				By("Should be matched with engine_passed_experiments regx")
				Expect(string(metrics)).Should(ContainSubstring(`litmuschaos_passed_experiments{chaosengine_context="",chaosengine_name="engine-nginx",chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 1`))

			}
		})
	})
})

// deleting all unused resources
var _ = AfterSuite(func() {

	By("Deleting chaosengines")
	err := exec.Command("kubectl", "delete", "chaosengines", "--all", "-n", "litmus").Run()
	Expect(err).To(BeNil())

	By("Uninstalling litmus")
	deleteNS := exec.Command("kubectl", "delete", "-f", "https://litmuschaos.github.io/litmus/litmus-operator-ci.yaml").Run()
	Expect(deleteNS).To(BeNil())

})
