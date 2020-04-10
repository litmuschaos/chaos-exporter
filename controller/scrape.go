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
	allChaosEngineList, err := clientSet.LitmuschaosV1alpha1().ChaosEngines("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	chaosEngineList := filterMonitoringEnabledEngines(allChaosEngineList)
	if err != nil {
		return err
	}
	var total, pass, fail float64
	total = 0
	pass = 0
	fail = 0
	for _, chaosEngine := range chaosEngineList.Items {
		totalEngine, passedEngine, failedEngine := getExperimentMetricsFromEngine(&chaosEngine)
		klog.V(2).Infof("ChaosEngineMetrics: EngineName: %v, EngineNamespace: %v, TotalExp: %v, PassedExp: %v, FailedExp: %v", chaosEngine.Name, chaosEngine.Namespace, totalEngine, passedEngine, failedEngine)
		var engineDetails ChaosEngineDetail
		engineDetails.Name = chaosEngine.Name
		engineDetails.Namespace = chaosEngine.Namespace
		engineDetails.PassedExp = passedEngine
		engineDetails.FailedExp = failedEngine
		engineDetails.TotalExp = totalEngine
		total += totalEngine
		pass += passedEngine
		fail += failedEngine
		setEngineChaosMetrics(engineDetails)
	}
	setClusterChaosMetrics(total, pass, fail)
	time.Sleep(1000 * time.Millisecond)
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
}

func getExperimentMetricsFromEngine(chaosEngine *litmuschaosv1alpha1.ChaosEngine) (float64, float64, float64) {
	var total, passed, failed float64
	passed = 0
	failed = 0
	expStatusList := chaosEngine.Status.Experiments
	total = float64(len(expStatusList))
	for i, v := range expStatusList {
		verdictFloat := getValueFromVerdict(v.Verdict)
		if verdictFloat == 3 {
			passed++
		} else if verdictFloat == 2 {
			failed++
		} else if verdictFloat == 1 {
			defineRunningExperimentMetric(chaosEngine.Name, chaosEngine.Namespace, chaosEngine.Spec.Experiments[i].Name)
		}
	}
	return total, passed, failed

}
func defineRunningExperimentMetric(engineName string, engineNamespace string, experimentName string) {
	klog.V(2).Infof("Running Experiment Metrics: EnginaName: %v, EngineNamespace: %v, ExperimentName: %v", engineName, engineNamespace, experimentName)
	RunningExperiment.WithLabelValues(engineNamespace, engineName, experimentName, fmt.Sprintf("%s-%s", engineName, experimentName)).Set(float64(1))

}

func getValueFromVerdict(verdict string) float64 {
	if verdict == "Pass" || verdict == "pass" || verdict == "passed" || verdict == "Passed" {
		return 3
	} else if verdict == "Fail" || verdict == "Failed" || verdict == "failed" || verdict == "fail" {
		return 2
	} else if verdict == "Awaited" || verdict == "awaited" {
		return 1
	} else {
		return 0
	}
}

func filterMonitoringEnabledEngines(allEngineList *litmuschaosv1alpha1.ChaosEngineList) *litmuschaosv1alpha1.ChaosEngineList {
	//var filtedEngine *litmuschaosv1alpha1.ChaosEngineList
	engineList := allEngineList.Items
	for i, v := range engineList {
		if v.Spec.Monitoring != true {
			engineList = append(engineList[:i], engineList[i+1:]...)
		}
	}
	allEngineList.Items = engineList
	return allEngineList
}
