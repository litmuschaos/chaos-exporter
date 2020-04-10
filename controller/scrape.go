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

	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	// auth for gcp: optional
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// Holds list of experiments in a chaosengine
var chaosExperimentList []string

// Holds a lookup of result: numericValue
var numericStatus = map[string]float64{
	"not-executed": 0,
	"running":      1,
	"fail":         2,
	"pass":         3,
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
	var total, pass, fail float64
	total = 0
	pass = 0
	fail = 0
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
		setEngineChaosMetrics(engineDetails)
	}
	setClusterChaosMetrics(total, pass, fail)
	return nil
}

func setClusterChaosMetrics(total float64, pass float64, fail float64) {
	ClusterPassedExperiments.WithLabelValues().Set(pass)
	ClusterFailedExperiments.WithLabelValues().Set(fail)
	ClusterTotalExperiments.WithLabelValues().Set(total)
}
func setEngineChaosMetrics(engineDetails ChaosEngineDetail) {
	EngineTotalExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.TotalExp)
	EnginePassedExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.PassedExp)
	EngineFailedExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.FailedExp)
	EngineAwaitedExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.AwaitedExp)
}

func getExperimentMetricsFromEngine(chaosEngine *litmuschaosv1alpha1.ChaosEngine) (float64, float64, float64, float64) {
	var total, passed, failed, awaited float64
	passed = 0
	failed = 0
	awaited = 0
	expStatusList := chaosEngine.Status.Experiments
	total = float64(len(expStatusList))
	for i, v := range expStatusList {
		verdictFloat := getValueFromVerdict(strings.ToLower(v.Verdict))
		if verdictFloat == 1 {
			awaited++
		} else if verdictFloat == 4 {
			passed++
		} else if verdictFloat == 3 {
			failed++
		} else if verdictFloat == 2 {
			defineRunningExperimentMetric(chaosEngine.Name, chaosEngine.Namespace, chaosEngine.Spec.Experiments[i].Name)
		}
	}
	return total, passed, failed, awaited

}
func defineRunningExperimentMetric(engineName string, engineNamespace string, experimentName string) {
	klog.V(2).Infof("Running Experiment Metrics: EnginaName: %v, EngineNamespace: %v, ExperimentName: %v", engineName, engineNamespace, experimentName)
	RunningExperiment.WithLabelValues(engineNamespace, engineName, experimentName, fmt.Sprintf("%s-%s", engineName, experimentName)).Set(float64(2))

}

func getValueFromVerdict(verdict string) float64 {

	switch verdict {
	case "pass":
		return 4
	case "fail":
		return 3
	case "awaited":
		return 2
	case "waiting":
		return 1
	default:
		return 0
	}
}

func filterMonitoringEnabledEngines(engineList *litmuschaosv1alpha1.ChaosEngineList) *litmuschaosv1alpha1.ChaosEngineList {
	//var filtedEngine *litmuschaosv1alpha1.ChaosEngineList
	engineListItems := engineList.Items
	for i, v := range engineListItems {
		if v.Spec.Monitoring != true {
			engineListItems = append(engineListItems[:i], engineListItems[i+1:]...)
		}
	}
	engineList.Items = engineListItems
	return engineList
}
