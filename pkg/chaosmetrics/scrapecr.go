package chaosmetrics

import (
	"fmt"
	"strings"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	// auth for gcp: optional
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	clientV1alpha1 "github.com/litmuschaos/chaos-exporter/pkg/clientset/v1alpha1"
	v1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Holds list of experiments in a chaosengine
var chaosexperimentlist []string

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

// Holds Error type
var err error

// Utility fn to return numeric value for a result
func statusConv(expstatus string) (numeric float64) {
	if numeric, ok := numericstatus[expstatus]; ok {
		return numeric
		//fmt.Printf("%v", numeric)
	}
	//return 127
	return 0
}

/* Exported function to gather chaos metrics

   - TODO: Update the chaosresult to carry verdict alone. Status & Verdict are redundant
*/

// GetLitmusChaosMetrics returns chaos metrics for a given chaosengine
func GetLitmusChaosMetrics(cfg *rest.Config, cEngine string, ns string) (totalExpCount, totalPassedExp, totalFailedExp float64, rMap map[string]float64, err error) {

	v1alpha1.AddToScheme(scheme.Scheme)
	clientSet, err := clientV1alpha1.NewForConfig(cfg)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	engine, err := clientSet.ChaosEngines(ns).Get(cEngine, metav1.GetOptions{})
	if err != nil {
		return 0, 0, 0, nil, err
	}

	/////////////////////////////////////////////////////////
	/*METRIC*/
	totalExpCount = float64(len(engine.Spec.Experiments)) //
	/////////////////////////////////////////////////////////

	for _, element := range engine.Spec.Experiments {
		chaosexperimentlist = append(chaosexperimentlist, element.Name)
	}

	// Initialize the chaosresult map before entering loop
	chaosresultmap := make(map[string]string)

	// Set default values on the chaosresult map before populating w/ actual values
	//for _, test:= range chaosexperimentlist{

	for _, test := range chaosexperimentlist {
		chaosresultname := fmt.Sprintf("%s-%s", cEngine, test)
		testresultdump, err := clientSet.ChaosResults(ns).Get(chaosresultname, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// lack of result cr indicates experiment not executed
				//chaosresultmap[chaosresultname] = "not-executed"
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

	/////////////////////////////////////////////////
	/*METRIC*/                       //
	totalPassedExp = float64(pcount) //
	totalFailedExp = float64(fcount) //
	/////////////////////////////////////////////////
	//fmt.Printf("%+v %+v %+v\n", totalExpCount, totalPassedExp, totalFailedExp)

	//Map verdict to numerical values {0-notstarted, 1-running, 2-fail, 3-pass}
	statusmap := make(map[string]float64)
	for index, status := range chaosresultmap {
		val := statusConv(status)
		statusmap[index] = val
	}
	fmt.Printf("%+v\n", statusmap)

	return totalExpCount, totalPassedExp, totalFailedExp, statusmap, nil
}
