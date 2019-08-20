package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	clientV1alpha1 "github.com/litmuschaos/chaos-exporter/pkg/clientset/v1alpha1"
	"github.com/litmuschaos/chaos-exporter/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/node/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

// Allocating all resources and env before the test suite
var _ = BeforeSuite(func() {
	os.Setenv("CHAOSENGINE", "engine-nginx") // set env chaosengine to rngine-nginx
	os.Setenv("APP_NAMESPACE", "litmus")     // set enc ns to litmus

	// Below 3 commands are creating custom resource defination and custom resource chaosEngine under litmus namespace
	crdcmd := exec.Command("kubectl", "create", "-f", "\"https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/crds/chaosengine_crd.yaml\"").Run()
	nscmd := exec.Command("kubectl", "create", "ns", "litmus").Run()
	chaosenginecmd := exec.Command("kubectl", "create", "-f", "\"https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/crds/chaosengine.yaml\"").Run()

	if crdcmd != nil {
		log.Fatal(crdcmd)
	}

	if nscmd != nil {
		log.Fatal(nscmd)
	}

	if chaosenginecmd != nil {
		log.Fatal(chaosenginecmd)
	}
})

var _ = Describe("BDD on chaos-exporter", func() {

	// BDD TEST CASE 1

	Context("Chaos Engine Liveliness test", func() {

		It("should be a engine-nginx(chaosEngine)", func() {

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
				Fail(err.Error())
			}
			engine, err := clientSet.ChaosEngines("litmus").List(metav1.ListOptions{})
			fmt.Println(engine.Items[0].Name)

			// check if chaosEngine is engine-nginx or not. Failed when it is unmatched
			Expect(engine.Items[0].Name).To(Equal("engine-nginx"))

		})
	})

	// BDD TEST CASE 2
	Context("Chaos Engine failed experiments", func() {

		It("should be a zero failed experiments", func() {
			chaosengine := os.Getenv("CHAOSENGINE")
			appNS := os.Getenv("APP_NAMESPACE")

			var kubeconfig string = string(os.Getenv("HOME") + "/.kube/config")
			var config *rest.Config
			var err error

			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)

			if err != nil {
				Fail(err.Error())
			}
			expTotal, passTotal, failTotal, expMap, err := util.GetChaosMetrics(config, chaosengine, appNS)
			if err != nil {
				Fail(err.Error()) // Unable to get metrics:
			}

			fmt.Println(expTotal, failTotal, passTotal, expMap)

			// check if failed experiments is 0
			Expect(failTotal).To(Equal(float64(0)))

		})
	})
})

// deleting all unused resources and env after the test suite
var _ = AfterSuite(func() {
	// command for delete the custom resource definition
	deletecrdcmd := exec.Command("kubectl", "delete", "crd", "chaosengines.litmuschaos.io").Run()

	if deletecrdcmd != nil {
		log.Fatal(deletecrdcmd)
	}
	// command for delete the namespace
	deletenscmd := exec.Command("kubectl", "delete", "ns", "litmus").Run()

	if deletenscmd != nil {
		log.Fatal(deletenscmd)
	}
})
