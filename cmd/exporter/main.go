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

package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/litmuschaos/chaos-exporter/controller"
)

// Declare general variables (cluster ops, error handling, misc)
var kubeconfig *string
var config *rest.Config
var err error

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
	klog.InitFlags(nil)
	// Setting up kubeconfig
	config, err := getKubeConfig()
	if err != nil {
		panic(err.Error())
	}
	// Trigger the chaos metrics collection
	go controller.Exporter(config)
	//This section will start the HTTP server and expose metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	klog.Info("Beginning to serve on port :8080")
	klog.Fatal(http.ListenAndServe(":8080", nil))
}
