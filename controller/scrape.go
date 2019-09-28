package controller

import (
	"fmt"
	"strings"

	"k8s.io/client-go/rest"
	// auth for gcp: optional
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Holds list of experiments in a chaosengine
var chaosExperimentList []string

// Holds a map of experiment: result
var chaosresultmap map[string]string

// Holds a map of experiment: numeric representation(result)
var statusmap map[string]float64

// Holds a lookup of result: numericValue
var numericstatus = map[string]float64{
	"not-executed": 0,
	"running":      1,
	"fail":         2,
	"pass":         3,
}

// Utility fn to return numeric value for a result
func statusConv(expstatus string) (numeric float64) {
	if numeric, ok := numericstatus[expstatus]; ok {
		return numeric
	}
	//return 127
	return 0
}

// GetLitmusChaosMetrics returns chaos metrics for a given chaosengine
func GetLitmusChaosMetrics(config *rest.Config, exporterSpec ExporterSpec) (float64, float64, float64, map[string]float64, error) {

	clientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	engine, err := clientSet.LitmuschaosV1alpha1().ChaosEngines(exporterSpec.AppNS).Get(exporterSpec.ChaosEngine, metav1.GetOptions{})
	if err != nil {
		return 0, 0, 0, nil, err
	}

	for _, element := range engine.Spec.Experiments {
		chaosExperimentList = append(chaosExperimentList, element.Name)
	}

	// Set default values on the chaosresult map before populating w/ actual values

	for _, test := range chaosExperimentList {
		chaosresultname := fmt.Sprintf("%s-%s", exporterSpec.ChaosEngine, test)
		testresultdump, err := clientSet.LitmuschaosV1alpha1().ChaosResults(exporterSpec.AppNS).Get(chaosresultname, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// lack of result cr indicates experiment not executed
				chaosresultmap[test] = "not-executed"
			}
			//return 0, 0, 0, nil, err
		}
		result := testresultdump.Spec.ExperimentStatus.Verdict
		//chaosresultmap[chaosresultname] = result
		chaosresultmap[test] = result
	}

	pcount, fcount := 0, 0
	for _, verdict := range chaosresultmap {
		if verdict == "pass" {
			pcount++
		} else if verdict == "fail" {
			fcount++
		}
	}

	//Map verdict to numerical values {0-notstarted, 1-running, 2-fail, 3-pass}
	statusmap := make(map[string]float64)
	for index, status := range chaosresultmap {
		val := statusConv(status)
		statusmap[index] = val
	}
	fmt.Printf("%+v\n", statusmap)

	totalPassedExp := float64(pcount)
	totalFailedExp := float64(fcount)
	totalExpCount := float64(len(engine.Spec.Experiments))

	return totalExpCount, totalPassedExp, totalFailedExp, statusmap, nil
}
