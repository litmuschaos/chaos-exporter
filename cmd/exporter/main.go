/*
Copyright 2019 LitmusChaos Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"

	"github.com/litmuschaos/chaos-exporter/controller"
)

// Declare general variables (cluster ops, error handling, misc)
var kubeconfig *string
var config *rest.Config
var err error

// getNamespaceEnv checks whether an ENV variable has been set, else sets a default value
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

func getExporterSpecs() (controller.ExporterSpec, error) {
	// Get app details & chaosengine name from ENV
	// Add checks for default
	applicationUUID := os.Getenv("APP_UUID")
	chaosEngine := os.Getenv("CHAOSENGINE")

	// Validate availability of mandatory ENV
	if chaosEngine == "" || applicationUUID == "" {
		return controller.ExporterSpec{}, fmt.Errorf("please specify correct APP_UUID & CHAOSENGINE ENVs")
	}

	exporterSpec := controller.ExporterSpec{
		ChaosEngine:      chaosEngine,
		AppUUID:          applicationUUID,
		AppNS:            getNamespaceEnv("APP_NAMESPACE", "default"),
		OpenebsNamespace: getOpenebsEnv("OPENEBS_NAMESPACE", "openebs"),
	}
	return exporterSpec, nil
}

// getKubeConfig setup the config for access cluster resource
func getKubeConfig() (*rest.Config, error) {
	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
	// Use in-cluster config if kubeconfig path is specified
	if *kubeconfig == "" {
		config, err = rest.InClusterConfig()
		if err != nil {
			return config, err
		}
	}
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return config, err
	}
	return config, err
}

func main() {
	// Setting up kubeconfig
	config, err := getKubeConfig()
	if err != nil {
		panic(err.Error())
	}
	exporterSpec, err := getExporterSpecs()
	if err != nil {
		log.Fatal("Error: ", err)
	}
	// Trigger the chaos metrics collection
	go controller.Exporter(config, exporterSpec)
	//This section will start the HTTP server and expose metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	log.Info("Beginning to serve on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
