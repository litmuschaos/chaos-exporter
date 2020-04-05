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

package controller

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Exporter continuously collects the chaos metrics for a given chaosengine
func Exporter(config *rest.Config) {
	_, litmusClientSet, err := generateClientSets(config)
	log.Printf("Started creating Metrics")
	if err != nil {
		log.Error(err)
	}

	// Register the fixed (count) chaos metrics
	log.Printf("Registering Fixed Metrics")
	registerFixedMetrics()

	log.Printf("Going into for loop")
	for {
		GetLitmusChaosMetrics(litmusClientSet)
	}
}

// generateClientSets will generate clientSet for kubernetes and litmus
func generateClientSets(config *rest.Config) (*kubernetes.Clientset, *clientV1alpha1.Clientset, error) {
	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate kubernetes clientSet %s: ", err)
	}
	litmusClientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate litmus clientSet %s: ", err)
	}
	return k8sClientSet, litmusClientSet, nil
}

func generatePrometheusGaugeVec(index string) (string, *prometheus.GaugeVec) {
	sanitizedExpName := strings.Replace(index, "-", "_", -1)
	var (
		tmpExp = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: "c", Subsystem: "exp", Name: sanitizedExpName, Help: ""},
			[]string{"app_uid", "engine_name", "kubernetes_version", "openebs_version"},
		)
	)
	return sanitizedExpName, tmpExp
}

func registerFixedMetrics() {
	prometheus.MustRegister(EngineTotalExperiments)
	prometheus.MustRegister(EnginePassedExperiments)
	prometheus.MustRegister(EngineFailedExperiments)
	prometheus.MustRegister(RunningExperiment)
	prometheus.MustRegister(ClusterTotalExperiments)
	prometheus.MustRegister(ClusterFailedExperiments)
	prometheus.MustRegister(ClusterPassedExperiments)
}
