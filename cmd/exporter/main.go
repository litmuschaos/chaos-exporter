/* The chaos exporter collects and exposes the following type of metrics:

   Fixed (always captured):
     - Total number of chaos experiments
     - Total number of passed experiments
     - Total Number of failed experiments

   Dynamic (experiment list may vary based on c.engine):
     - States of individual chaos experiments
     - {not-executed:0, running:1, fail:2, pass:3}
       TODO: Improve representaion of test state

   Common experiments include:

     - pod_failure
     - container_kill
     - container_network_delay
     - container_packet_loss
*/

package main

import (
	"os"
	"time"

	//"fmt"
	"flag"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/litmuschaos/chaos-exporter/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Declare general variables (cluster ops, error handling, misc)
var kubeconfig string
var config *rest.Config
var err error
var registeredResultMetrics []string

// Declare the fixed chaos metrics. Dynamic (testStatus) metrics are defined in metrics()
var (
	experimentsTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "c",
		Subsystem: "engine",
		Name:      "experiment_count",
		Help:      "Total number of experiments executed by the chaos engine",
	},
		[]string{"app_uid", "engine_name"},
	)

	passedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "c",
		Subsystem: "engine",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments",
	},
		[]string{"app_uid", "engine_name"},
	)

	failedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "c",
		Subsystem: "engine",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments",
	},
		[]string{"app_uid", "engine_name"},
	)
)

func contains(l []string, e string) bool {
	for _, i := range l {
		if i == e {
			return true
		}
	}
	return false
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func metrics(cfg *rest.Config, cEngine string, aUUID string, aNS string) {

	for {
		// Get the chaos metrics for the specified chaosengine
		expTotal, passTotal, failTotal, expMap, err := util.GetChaosMetrics(cfg, cEngine, aNS)
		if err != nil {
			//panic(err.Error())
			log.Fatal("Unable to get metrics: ", err.Error())
		}

		// Define, register & set the dynamically obtained chaos metrics (experiment state)
		for index, verdict := range expMap {
			sanitizedExpName := strings.Replace(index, "-", "_", -1)
			var (
				tmpExp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Namespace: "c",
					Subsystem: "exp",
					Name:      sanitizedExpName,
					Help:      "",
				},
					[]string{"app_uid", "engine_name"},
				)
			)

			if contains(registeredResultMetrics, sanitizedExpName) {
				prometheus.Unregister(tmpExp)
				prometheus.MustRegister(tmpExp)
				tmpExp.WithLabelValues(aUUID, cEngine).Set(verdict)
			} else {
				prometheus.MustRegister(tmpExp)
				tmpExp.WithLabelValues(aUUID, cEngine).Set(verdict)
				registeredResultMetrics = append(registeredResultMetrics, sanitizedExpName)
			}

			// Set the fixed chaos metrics
			experimentsTotal.WithLabelValues(aUUID, cEngine).Set(expTotal)
			passedExperiments.WithLabelValues(aUUID, cEngine).Set(passTotal)
			failedExperiments.WithLabelValues(aUUID, cEngine).Set(failTotal)
		}

		time.Sleep(1000 * time.Millisecond)
	}
}

func main() {

	// Get app details & chaoengine name from ENV
	// Add checks for default
	appUUID := os.Getenv("APP_UUID")
	// appUUID := "1234"

	// chaosengine := "engine-nginx"
	chaosengine := os.Getenv("CHAOSENGINE")
	//appNS := os.Getenv("APP_NAMESPACE")
	appNS := getEnv("APP_NAMESPACE", "litmus")

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
	if chaosengine == "" || appUUID == "" {
		log.Fatal("ERROR: please specify correct APP_UUID & CHAOSENGINE ENVs")
		os.Exit(1)
	}

	// Register the fixed (count) chaos metrics
	prometheus.MustRegister(experimentsTotal)
	prometheus.MustRegister(passedExperiments)
	prometheus.MustRegister(failedExperiments)

	go metrics(config, chaosengine, appUUID, appNS)

	//This section will start the HTTP server and expose
	//any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
