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
	"time"

	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
)

// Exporter continuously collects the chaos metrics for a given chaosengine
func Exporter(config *rest.Config) {
	_, litmusClientSet, err := generateClientSets(config)
	klog.V(0).Infof("Started creating Metrics")
	if err != nil {
		klog.Error(err)
	}

	// Register the fixed (count) chaos metrics
	klog.V(0).Infof("Registering Fixed Metrics")
	registerFixedMetrics()

	for {
		GetLitmusChaosMetrics(litmusClientSet)
		time.Sleep(1000 * time.Millisecond)
	}
}

// GetLitmusChaosMetrics returns chaos metrics for a given chaosengine
func GetLitmusChaosMetrics(clientSet *clientV1alpha1.Clientset) error {
	chaosEngineList, err := clientSet.LitmuschaosV1alpha1().ChaosEngines("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	filteredChaosEngineList := filterMonitoringEnabledEngines(chaosEngineList)
	if err != nil {
		return err
	}
	var total float64 = 0
	var pass float64 = 0
	var fail float64 = 0
	for _, chaosEngine := range filteredChaosEngineList.Items {
		totalEngine, passedEngine, failedEngine, awaitedEngine := getExperimentMetricsFromEngine(&chaosEngine)
		klog.V(2).Infof("ChaosEngineMetrics: EngineName: %v, EngineNamespace: %v, TotalExp: %v, PassedExp: %v, FailedExp: %v", chaosEngine.Name, chaosEngine.Namespace, totalEngine, passedEngine, failedEngine)
		var engineDetails ChaosEngineDetail
		engineDetails.Name = chaosEngine.Name
		engineDetails.Namespace = chaosEngine.Namespace
		engineDetails.TotalExp = totalEngine
		engineDetails.PassedExp = passedEngine
		engineDetails.FailedExp = failedEngine
		engineDetails.AwaitedExp = awaitedEngine
		total += totalEngine
		pass += passedEngine
		fail += failedEngine
		SetEngineChaosMetrics(engineDetails)
	}
	SetClusterChaosMetrics(total, pass, fail)
	return nil
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

func registerFixedMetrics() {
	prometheus.MustRegister(EngineTotalExperiments)
	prometheus.MustRegister(EnginePassedExperiments)
	prometheus.MustRegister(EngineFailedExperiments)
	prometheus.MustRegister(EngineWaitingExperiments)
	prometheus.MustRegister(RunningExperiment)
	prometheus.MustRegister(ClusterTotalExperiments)
	prometheus.MustRegister(ClusterFailedExperiments)
	prometheus.MustRegister(ClusterPassedExperiments)
}
