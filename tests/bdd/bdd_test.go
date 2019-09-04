package bdd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"

	"github.com/litmuschaos/chaos-exporter/pkg/chaosmetrics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
	os.Setenv("APP_UUID", "1234")

	// Below 3 commands are creating custom resource defination and custom resource chaosEngine under litmus namespace
	crdcmd := exec.Command("kubectl", "create", "-f", "\"https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/crds/chaosengine_crd.yaml\"").Run()
	nscmd := exec.Command("kubectl", "create", "ns", "litmus").Run()
	chaosenginecmd := exec.Command("kubectl", "create", "-f", "\"https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/crds/chaosengine.yaml\"").Run()

	buildCMD := exec.Command("nohup", "go", "run", "../../cmd/exporter/main.go", "-kubeconfig=/home/rajdas/.kube/config", "&").Run()
	if crdcmd != nil {
		log.Fatal(crdcmd)
	}

	if nscmd != nil {
		log.Fatal(nscmd)
	}

	if chaosenginecmd != nil {
		log.Fatal(chaosenginecmd)
	}

	if buildCMD != nil {
		log.Fatal(buildCMD)
	}
})

var _ = Describe("BDD on chaos-exporter", func() {

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
			expTotal, passTotal, failTotal, expMap, err := chaosmetrics.GetLitmusChaosMetrics(config, chaosengine, appNS)
			if err != nil {
				Fail(err.Error()) // Unable to get metrics:
			}

			fmt.Println(expTotal, failTotal, passTotal, expMap)

			// check if failed experiments is 0
			Expect(failTotal).To(Equal(float64(0)))

		})
	})

	Context("Curl the prometheus metrics", func() {
		It("Should return prometheus metrics", func() {

			resp, err := http.Get("127.0.0.1:8080/metrics")
			Expect(err).To(BeNil())
			defer resp.Body.Close()
		})
	})
})

var _ = AfterSuite(func() {
	// command for delete the custom resource definition
	deletecrdcmd := exec.Command("kubectl", "delete", "crd", "--all").Run()
	Expect(deletecrdcmd).To(BeNil())

	// command for delete the namespace
	deletenscmd := exec.Command("kubectl", "delete", "ns", "litmus").Run()
	Expect(deletenscmd).To(BeNil())

})
