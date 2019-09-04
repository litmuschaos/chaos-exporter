package bdd

import (
	"fmt"
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
			By("Checking Total failed experiments")
			expTotal, passTotal, failTotal, expMap, err := chaosmetrics.GetLitmusChaosMetrics(config, chaosengine, appNS)
			if err != nil {
				Fail(err.Error()) // Unable to get metrics:
			}

			fmt.Println(expTotal, failTotal, passTotal, expMap)

			Expect(failTotal).To(Equal(float64(0)))

		})
	})

	Context("Curl the prometheus metrics", func() {
		It("Should return prometheus metrics", func() {
			fmt.Println(os.Getenv("CHAOSENGINE"))
			fmt.Println(os.Getenv("APP_NAMESPACE"))
			fmt.Println(os.Getenv("APP_UUID"))

			resp, err := http.Get("http://127.0.0.1:8080/metrics")
			Expect(err).To(BeNil())
			fmt.Println(resp.Body)
			defer resp.Body.Close()
		})
	})
})

var _ = AfterSuite(func() {

	By("Deleting chaosengine CRD")
	ceDeleteCRDs := exec.Command("kubectl", "delete", "crds", "chaosengines.litmuschaos.io").Run()
	Expect(ceDeleteCRDs).To(BeNil())

	By("Deleting namespace litmus")
	deleteNS := exec.Command("kubectl", "delete", "ns", "litmus").Run()
	Expect(deleteNS).To(BeNil())

})
