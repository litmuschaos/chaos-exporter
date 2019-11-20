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

	// auth for gcp: optional
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	//"github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// Utility fn to return numeric value for a result
func statusConversion(expStatus string) (numeric float64) {
	if numeric, ok := numericStatus[expStatus]; ok {
		return numeric
	}
	return numericStatus["not-executed"]
}

// GetLitmusChaosMetrics returns chaos metrics for a given chaosengine
func GetLitmusChaosMetrics(clientSet *clientV1alpha1.Clientset, exporterSpec ExporterSpec) (float64, float64, float64, map[string]float64, error) {
	engine, err := clientSet.LitmuschaosV1alpha1().ChaosEngines(exporterSpec.AppNS).Get(exporterSpec.ChaosEngine, metav1.GetOptions{})
	if err != nil {
		return 0, 0, 0, nil, err
	}

	for _, element := range engine.Spec.Experiments {
		chaosExperimentList = append(chaosExperimentList, element.Name)
	}

	// Set default values on the chaosResult map before populating w/ actual values
	spec := ChaosResultSpec{ExporterSpec: exporterSpec, ChaosExperimentList: chaosExperimentList}
	chaosResultMap := setChaosResultValue(clientSet, spec)

	chaosResult, StatusMap := calculateChaosResult(chaosResultMap)
	fmt.Printf("%+v\n", StatusMap)
	totalExpCount := float64(len(engine.Spec.Experiments))

	return totalExpCount, chaosResult.TotalPassedExp, chaosResult.TotalFailedExp, StatusMap, nil
}

//setChaosResultValue will populate the default value of chaos result
func setChaosResultValue(clientSet *clientV1alpha1.Clientset, chaosResultSpec ChaosResultSpec) map[string]string {
	chaosResultMap := make(map[string]string)
	for _, test := range chaosResultSpec.ChaosExperimentList {
		chaosResultName := fmt.Sprintf("%s-%s", chaosResultSpec.ExporterSpec.ChaosEngine, test)
		testResultDump, err := clientSet.LitmuschaosV1alpha1().ChaosResults(chaosResultSpec.ExporterSpec.AppNS).Get(chaosResultName, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// lack of result cr indicates experiment not executed
				chaosResultMap[test] = "not-executed"
			}
			continue
		}
		chaosResultMap[test] = testResultDump.Spec.ExperimentStatus.Verdict
	}
	return chaosResultMap
}

// calculateChaosResult will calculate the number of pass and failed experiments
func calculateChaosResult(chaosResult map[string]string) (ChaosExpResult, map[string]float64) {
	var cr ChaosExpResult
	StatusMap := make(map[string]float64)
	for index, status := range chaosResult {
		if status == "pass" {
			cr.TotalPassedExp++
		} else if status == "fail" {
			cr.TotalFailedExp++
		}
		StatusMap[index] = statusConversion(status)
	}

	return cr, StatusMap
}
