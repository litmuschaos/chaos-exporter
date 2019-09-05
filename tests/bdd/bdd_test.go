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

	"github.com/litmuschaos/chaos-exporter/pkg/chaosmetrics"
	version "github.com/litmuschaos/chaos-exporter/pkg/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	clientV1alpha1 "github.com/litmuschaos/chaos-exporter/pkg/clientset/v1alpha1"
	v1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis"
	chaosEngineV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = string(os.Getenv("HOME") + "/.kube/config")
var config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
var chaosengine = os.Getenv("CHAOSENGINE")
var appNS = os.Getenv("APP_NAMESPACE")
var appUUID = os.Getenv("APP_UUID")

func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

var _ = BeforeSuite(func() {

	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		fmt.Println(err)
	}

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
	time.Sleep(4000000000)

	fmt.Println("process id", cmd.Process.Pid)

})

var _ = Describe("BDD on chaos-exporter", func() {

	// BDD case 1
	Context("Chaos Engine failed experiments", func() {

		It("should be a zero failed experiments", func() {

			if err != nil {
				Fail(err.Error())
			}

			By("Checking experiments metrics")
			expTotal, passTotal, failTotal, expMap, err := chaosmetrics.GetLitmusChaosMetrics(config, chaosengine, appNS)
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
	Context("Curl the prometheus metrics--", func() {
		It("Should return prometheus metrics", func() {

			By("Sending get request to metrics")
			response, err := http.Get("http://127.0.0.1:8080/metrics")
			Expect(err).To(BeNil())

			if err != nil {
				fmt.Printf("%s", err)
				os.Exit(1)
			} else {
				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Printf("%s", err)
					os.Exit(1)
				}
				fmt.Printf("%s\n", string(contents))

				var k8sVersion string
				var openEBSVersion string
				k8sVersion, _ = version.GetKubernetesVersion(config)            // getting kubernetes version
				openEBSVersion, _ = version.GetOpenebsVersion(config, "litmus") // getting openEBS Version

				var tmpStr = "{app_uid=\"" + appUUID + "\",engine_name=\"engine-nginx\",kubernetes_version=\"" + k8sVersion + "\",openebs_version=\"" + openEBSVersion + "\"}"

				By("Should be matched with total_experiments regx")
				Expect(string(contents)).Should(ContainSubstring(string("c_engine_experiment_count" + tmpStr + " 2")))

				By("Should be matched with failed_experiments regx")
				Expect(string(contents)).Should(ContainSubstring(string("c_engine_failed_experiments" + tmpStr + " 0")))

				By("Should be matched with passed_experiments regx")
				Expect(string(contents)).Should(ContainSubstring(string("c_engine_passed_experiments" + tmpStr + " 0")))

				By("Should be matched with container_kill experiment regx")
				Expect(string(contents)).Should(ContainSubstring(string("c_exp_container_kill" + tmpStr + " 0")))

				By("Should be matched with pod_kill experiment experiments regx")
				Expect(string(contents)).Should(ContainSubstring(string("c_exp_pod_kill" + tmpStr + " 0")))

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
