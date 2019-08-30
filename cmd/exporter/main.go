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
  "github.com/litmuschaos/chaos-exporter/pkg/chaosmetrics"
  log "github.com/Sirupsen/logrus"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
  "k8s.io/client-go/tools/clientcmd"
  "k8s.io/client-go/rest"
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

// contains checks if the a string is already part of a list of strings
func contains(l []string, e string) bool {
     for _, i := range l {
         if i == e {
             return true
         }
     }
     return false
}

// getEnv checks whether an ENV variable has been set, else sets a default value
func getEnv(key, fallback string)(string){
        if value, ok := os.LookupEnv(key); ok {
            return value
        }
        return fallback
}


// exporter continuously collects the chaos metrics for a given chaosengine 
func exporter(cfg *rest.Config, chaosEngine string, appUUID string, appNS string){

   for {
            // Get the chaos metrics for the specified chaosengine 
            expTotal, passTotal, failTotal, expMap, err := chaosmetrics.GetLitmusChaosMetrics(cfg, chaosEngine, appNS)
            if err != nil {
                //panic(err.Error())
                log.Fatal("Unable to get metrics: ", err.Error())
            }

            // Define, register & set the dynamically obtained chaos metrics (experiment state)
            for index, verdict := range expMap{
                sanitizedExpName := strings.Replace(index, "-", "_", -1)
                var (
                    tmpExp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
                        Namespace: "c",
                        Subsystem: "exp",
                        Name:      sanitizedExpName,
                        Help: "",
                    },
                    []string{"app_uid", "engine_name"},
                    )
                )

                if contains(registeredResultMetrics, sanitizedExpName) {
                   prometheus.Unregister(tmpExp); prometheus.MustRegister(tmpExp)
                   tmpExp.WithLabelValues(appUUID, chaosEngine).Set(verdict)
                } else {
                   prometheus.MustRegister(tmpExp)
                   tmpExp.WithLabelValues(appUUID, chaosEngine).Set(verdict)
                   registeredResultMetrics = append(registeredResultMetrics, sanitizedExpName)
                }

                // Set the fixed chaos metrics
                experimentsTotal.WithLabelValues(appUUID, chaosEngine).Set(expTotal)
                passedExperiments.WithLabelValues(appUUID, chaosEngine).Set(passTotal)
                failedExperiments.WithLabelValues(appUUID, chaosEngine).Set(failTotal)
            }

            time.Sleep(1000 * time.Millisecond)
   }
}

func main(){

    // Get app details & chaoengine name from ENV 
    // Add checks for default
    applicationUUID := os.Getenv("APP_UUID")
    chaosEngine := os.Getenv("CHAOSENGINE")
    //appNS := os.Getenv("APP_NAMESPACE")
    appNamespace := getEnv("APP_NAMESPACE", "default")

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

    // Register the fixed (count) chaos metrics
    prometheus.MustRegister(experimentsTotal)
    prometheus.MustRegister(passedExperiments)
    prometheus.MustRegister(failedExperiments)

    // Trigger the chaos metrics collection
    go exporter(config, chaosEngine, applicationUUID, appNamespace)

    //This section will start the HTTP server and expose
    //any metrics on the /metrics endpoint.
    http.Handle("/metrics", promhttp.Handler())
    log.Info("Beginning to serve on port :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
