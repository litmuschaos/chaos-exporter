package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/litmuschaos/chaos-exporter/pkg/chaosmetrics"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	clientV1alpha1 "github.com/litmuschaos/chaos-exporter/pkg/clientset/v1alpha1"
	v1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis"
	chaosEngineV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

var _ = BeforeSuite(func() {
	var kubeconfig string = string(os.Getenv("HOME") + "/.kube/config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	v1alpha1.AddToScheme(scheme.Scheme)
	clientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
	}

	By("Creating ChaosEngine")
	chaosEngine := &chaosEngineV1alpha1.ChaosEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "engine-nginx",
			Namespace: "litmus",
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
			Schedule: chaosEngineV1alpha1.ChaosSchedule{
				Interval:          "half-hourly",
				ExcludedTimes:     "",
				ExcludedDays:      "",
				ConcurrencyPolicy: "",
			},
		},
	}
	response, err := clientSet.ChaosEngines("litmus").Create(chaosEngine)
	Expect(err).To(BeNil())

	fmt.Println(response, err)

	By("Building Exporter")
	cmd := exec.Command("go", "run", "../../cmd/exporter/main.go", "-kubeconfig="+os.Getenv("HOME")+"/.kube/config")
	cmd.Stdout = os.Stdout
	err = cmd.Start()

	if err != nil {
		log.Fatal(err)
	}
	Expect(err).To(BeNil())

	// wait for execution of exporter
	time.Sleep(3000000000)

	fmt.Println("process id", cmd.Process.Pid)

})

var _ = Describe("BDD on chaos-exporter", func() {

	// BDD case 1
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
	// BDD case 2
	Context("Curl the prometheus metrics--", func() {
		It("Should return prometheus metrics", func() {

			resp, err := http.Get("http://127.0.0.1:8080/metrics")
			Expect(err).To(BeNil())
			defer resp.Body.Close()
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
