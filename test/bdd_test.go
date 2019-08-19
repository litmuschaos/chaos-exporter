package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	clientV1alpha1 "github.com/litmuschaos/chaos-exporter/pkg/clientset/v1alpha1"
	v1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
)

func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

var _ = Describe("BDD on chaos-exporter", func() {
	Context("Chaos Engine Liveliness test", func() {

		It("should be a chaosEngine", func() {

			kubeconfig := flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "kubeconfig file")
			flag.Parse()
			config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
			if err != nil {
				fmt.Println("KubeConfig Path is wrong", err)
				os.Exit(1)
			}

			v1alpha1.AddToScheme(scheme.Scheme)
			clientSet, err := clientV1alpha1.NewForConfig(config)
			if err != nil {
				fmt.Println(err)
			}
			engine, err := clientSet.ChaosEngines("litmus").List(metav1.ListOptions{})
			fmt.Println(engine.Items[0].Name)

			Expect(engine.Items[0].Name).To(Equal("engine-nginx"))

		})
	})
})
