/* The chaos exporter collects and exposes the following type of metrics:

   Fixed (always captured):
     - Total number of chaos experiments
     - Total number of passed experiments
     - Total Number of failed experiments

   Dynamic (experiment list may vary based on c.engine):
     - States of individual chaos experiments
     - {not-executed:0, running:1, fail:2, pass:3}
       Improve representation of test state

   Common experiments include:

     - pod_failure
     - container_kill
     - container_network_delay
     - container_packet_loss
*/

package main

import (
	"flag"
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"

	"github.com/litmuschaos/chaos-exporter/pkg/version"
	"github.com/litmuschaos/chaos-exporter/controller"
)

// Declare general variables (cluster ops, error handling, misc)
var kubeconfig string
var config *rest.Config
var err error

// getnamespaceEnv checks whether an ENV variable has been set, else sets a default value
func getNamespaceEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// get OpenEBS related environments
func getOpenebsEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {

	// Get app details & chaoengine name from ENV
	// Add checks for default
	applicationUUID := os.Getenv("APP_UUID")
	chaosEngine := os.Getenv("CHAOSENGINE")
	appNamespace := getNamespaceEnv("APP_NAMESPACE", "default")
	//openEBS installation namespace
	openebsNamespace := getOpenebsEnv("OPENEBS_NAMESPACE", "openebs")

	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
	flag.Parse()

	// Use in-cluster config if kubeconfig file not available
	if kubeconfig == "" {
		log.Info("using the in-cluster config")
		config, err = rest.InClusterConfig()
	} else {
		log.Info("using configuration from: ", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		panic(err.Error())
	}

	// Validate availability of mandatory ENV
	if chaosEngine == "" || applicationUUID == "" {
		log.Fatal("ERROR: please specify correct APP_UUID & CHAOSENGINE ENVs")
		os.Exit(1)
	}
	// This function gets the kubernetes version
	kubernetesVersion, err := version.GetKubernetesVersion(config)
	if err != nil {
		log.Info("Unable to get Kubernetes Version : ", err)
		//kubernetesVersion = "N/A"
	}
	// This function gets the openebs version
	openebsVersion, err := version.GetOpenebsVersion(config, openebsNamespace)
	if err != nil {
		log.Info("Unable to get OpenEBS Version : ", err)
		//openebsVersion = "N/A"
	}
	// Register the fixed (count) chaos metrics
	prometheus.MustRegister(controller.ExperimentsTotal)
	prometheus.MustRegister(controller.PassedExperiments)
	prometheus.MustRegister(controller.FailedExperiments)

	exporterSpec := controller.ExporterSpec{
		ChaosEngine: chaosEngine,
		AppUUID: applicationUUID,
		AppNS: appNamespace,
		KubernetesVersion: kubernetesVersion,
		OpenebsVersion: openebsVersion,
	}
	// Trigger the chaos metrics collection
	go controller.Exporter(config, exporterSpec)

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
