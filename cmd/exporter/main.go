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
  "github.com/litmuschaos/chaos-exporter/pkg/util"
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
var app_uuid, chaosengine string
var s_expResultMetricList  []string
var registeredResultMetrics []string

// Declare the fixed chaos metrics. Dynamic (testStatus) metrics are defined in metrics()
var (
    experimentsTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "c",
        Subsystem: "engine",
        Name:      "experiment_count",
        Help:      "Total number of experiments executed by the chaos engine",
    },
    []string{"app_uid"},
    )

    passedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "c",
        Subsystem: "engine",
        Name:      "passed_experiments",
        Help:      "Total number of passed experiments",
    },
    []string{"app_uid"},
    )

    failedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
        Namespace: "c",
        Subsystem: "engine",
        Name:      "failed_experiments",
        Help:      "Total number of failed experiments",
    },
    []string{"app_uid"},
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


func metrics(cfg *rest.Config, c_engine string, a_uuid string){

   for {
            // Get the chaos metrics for the specified chaosengine 
            expTotal, passTotal, failTotal, expMap, err := util.GetChaosMetrics(cfg, c_engine)
            if err != nil {
                //panic(err.Error())
                log.Fatal("Unable to get metrics: ", err.Error())
            }

            // Define, register & set the dynamically obtained chaos metrics (experiment state)
            for index, verdict := range expMap{
                sanitized_exp_name := strings.Replace(index, "-", "_", -1)
                var (
                    tmpExp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
                        Namespace: "c",
                        Subsystem: "exp",
                        Name:      sanitized_exp_name,
                        Help: "",
                    },
                    []string{"app_uid"},
                    )
                )

                if contains(registeredResultMetrics, sanitized_exp_name) {
                   prometheus.Unregister(tmpExp); prometheus.MustRegister(tmpExp)
                   tmpExp.WithLabelValues(a_uuid).Set(verdict)
                } else {
                   prometheus.MustRegister(tmpExp)
                   tmpExp.WithLabelValues(a_uuid).Set(verdict)
                   registeredResultMetrics = append(registeredResultMetrics, sanitized_exp_name)
                }

                // Set the fixed chaos metrics
                experimentsTotal.WithLabelValues(a_uuid).Set(expTotal)
                passedExperiments.WithLabelValues(a_uuid).Set(passTotal)
                failedExperiments.WithLabelValues(a_uuid).Set(failTotal)
            }

            time.Sleep(1000 * time.Millisecond)
   }
}

func main(){

    // Get app details & chaoengine name from ENV 
    app_uuid := os.Getenv("APP_UUID")
    chaosengine := os.Getenv("CHAOSENGINE")

    flag.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig file")
    flag.Parse()

    // Use in-cluster config if kubeconfig file not available
    if kubeconfig == "" {
        log.Info("using the in-cluster config")
        config, err = rest.InClusterConfig()
    } else {
        log.Info("using configuration from '%s'", kubeconfig)
        config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
    }

    if err != nil {
        panic(err.Error())
    }

    // Validate availability of mandatory ENV
    if chaosengine == "" || app_uuid == "" {
        log.Fatal("ERROR: please specify correct APP_UUID & CHAOSENGINE ENVs")
        os.Exit(1)
    }

    // Register the fixed (count) chaos metrics
    prometheus.MustRegister(experimentsTotal)
    prometheus.MustRegister(passedExperiments)
    prometheus.MustRegister(failedExperiments)

    go metrics(config, chaosengine, app_uuid)

    //This section will start the HTTP server and expose
    //any metrics on the /metrics endpoint.
    http.Handle("/metrics", promhttp.Handler())
    log.Info("Beginning to serve on port :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
