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

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"

	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

// Holds list of experiments in a chaosengine
var chaosExperimentList []string

func SetClusterChaosMetrics(total float64, pass float64, fail float64) {
	ClusterPassedExperiments.WithLabelValues().Set(pass)
	ClusterFailedExperiments.WithLabelValues().Set(fail)
	ClusterTotalExperiments.WithLabelValues().Set(total)
}
func SetEngineChaosMetrics(engineDetails ChaosEngineDetail) {
	EngineTotalExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.TotalExp)
	EnginePassedExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.PassedExp)
	EngineFailedExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.FailedExp)
	EngineWaitingExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.AwaitedExp)
}

func getExperimentMetricsFromEngine(chaosEngine *litmuschaosv1alpha1.ChaosEngine) (float64, float64, float64, float64) {
	var total, passed, failed, waiting float64
	expStatusList := chaosEngine.Status.Experiments
	total = float64(len(expStatusList))
	for i, v := range expStatusList {
		verdict := strings.ToLower(v.Verdict)
		switch verdict {
		case "pass":
			passed++
		case "fail":
			failed++
		case "waiting":
			waiting++
		case "awaited":
			defineRunningExperimentMetric(chaosEngine.Name, chaosEngine.Namespace, chaosEngine.Spec.Experiments[i].Name)
		}
	}
	return total, passed, failed, waiting

}
func defineRunningExperimentMetric(engineName string, engineNamespace string, experimentName string) {
	klog.V(2).Infof("Running Experiment Metrics: EnginaName: %v, EngineNamespace: %v, ExperimentName: %v", engineName, engineNamespace, experimentName)
	RunningExperiment.WithLabelValues(engineNamespace, engineName, experimentName, fmt.Sprintf("%s-%s", engineName, experimentName)).Set(float64(2))

}

func filterMonitoringEnabledEngines(engineList *litmuschaosv1alpha1.ChaosEngineList) *litmuschaosv1alpha1.ChaosEngineList {
	var filteredEngineList litmuschaosv1alpha1.ChaosEngineList
	for i := len(engineList.Items) - 1; i >= 0; i-- {
		// Condition to decide if current element has to be deleted:
		if engineList.Items[i].Spec.Monitoring {
			filteredEngineList.Items = append(filteredEngineList.Items, engineList.Items[i])
		}
	}
	return &filteredEngineList
}
