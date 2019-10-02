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

// Holds a map of experiment: result
var chaosResultMap map[string]string

// Holds a lookup of result: numericValue
var numericStatus = map[string]float64{
	"not-executed": 0,
	"running":      1,
	"fail":         2,
	"pass":         3,
}

// ChaosExpResult...
type ChaosExpResult struct {
	TotalExpCount  float64
	TotalPassedExp float64
	TotalFailedExp float64
	StatusMap      map[string]float64
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
	setChaosResultValue(clientSet, chaosExperimentList, exporterSpec)

	chaosResult := calculateChaosResult(chaosResultMap)
	fmt.Printf("%+v\n", chaosResult.StatusMap)
	totalExpCount := float64(len(engine.Spec.Experiments))

	return totalExpCount, chaosResult.TotalPassedExp, chaosResult.TotalFailedExp, chaosResult.StatusMap, nil
}

func setChaosResultValue(clientSet *clientV1alpha1.Clientset, chaosExperimentList []string, exporterSpec ExporterSpec) {
	for _, test := range chaosExperimentList {
		chaosResultName := fmt.Sprintf("%s-%s", exporterSpec.ChaosEngine, test)
		testResultDump, err := clientSet.LitmuschaosV1alpha1().ChaosResults(exporterSpec.AppNS).Get(chaosResultName, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// lack of result cr indicates experiment not executed
				chaosResultMap[test] = "not-executed"
			}
			continue
		}
		result := testResultDump.Spec.ExperimentStatus.Verdict
		chaosResultMap[test] = result
	}
}

// calculateChaosResult will calculate the number of pass and failed experiments
func calculateChaosResult(chaosResult map[string]string) ChaosExpResult {
	var cr ChaosExpResult

	for index, status := range chaosResult {
		if status == "pass" {
			cr.TotalPassedExp++
		} else if status == "fail" {
			cr.TotalFailedExp++
		}
		cr.StatusMap[index] = statusConversion(status)
	}

	return cr
}
