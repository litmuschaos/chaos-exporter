package controller

import (
	"fmt"
	"os"
	"strconv"

	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
)

// unsetDeletedChaosResults unset the metrics correspond to deleted chaosresults
func (gaugeMetrics *GaugeMetrics) unsetDeletedChaosResults(oldChaosResults, newChaosResults *litmuschaosv1alpha1.ChaosResultList) {
	for _, oldResult := range oldChaosResults.Items {
		found := false
		for _, newResult := range newChaosResults.Items {
			if oldResult.UID == newResult.UID {
				found = true
				break
			}
		}

		if !found {
			for _, value := range resultStore[string(oldResult.UID)] {

				probeSuccesPercentage, _ := getProbeSuccessPercentage(&oldResult)
				resultDetails := initialiseResult().
					setName(oldResult.Name).
					setNamespace(oldResult.Namespace).
					setProbeSuccesPercentage(probeSuccesPercentage).
					setVerdict(value.Verdict).
					setAppLabel(value.AppLabel).
					setAppNs(value.AppNs).
					setAppKind(value.AppKind).
					setChaosEngineName(oldResult.Spec.EngineName).
					setChaosEngineLabel(value.Label)

				gaugeMetrics.unsetResultChaosMetrics(resultDetails)
			}
			// delete the corresponding entry from the map
			delete(resultStore, string(oldResult.UID))
		}
	}
}

// unsetVerdictMetrics unset the metrics when chaosresult verdict changes
// if same chaosresult is continuously repeated more than scrape interval then it sets the metrics value to 0
func (gaugeMetrics *GaugeMetrics) unsetVerdictMetrics(resultDetails ChaosResultDetails) float64 {
	scrapeTime, _ := strconv.Atoi(getEnv("TSDB_SCRAPE_INTERVAL", "10"))
	result, ok := matchVerdict[string(resultDetails.UID)]
	if ok {
		// if verdict is different then delete the older metrics having outdated verdict
		if result.Verdict != resultDetails.Verdict {
			gaugeMetrics.ResultVerdict.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineLabel, result.Verdict,
				fmt.Sprintf("%f", result.ProbeSuccessPercentage), resultDetails.AppLabel, resultDetails.AppNs, resultDetails.AppKind)

			// update the values inside matchVerdict
			matchVerdict[string(resultDetails.UID)] = result.setCount(1).
				setVerdict(resultDetails.Verdict).
				setProbeSuccesPercentage(resultDetails.ProbeSuccesPercentage)
			return float64(1)
		} else {
			result.Count++
			matchVerdict[string(resultDetails.UID)] = result
			if result.Count >= scrapeTime {
				return float64(0)
			}
			return float64(1)
		}
	}

	// update the values inside matchVerdict
	matchVerdict[string(resultDetails.UID)] = initialiseResultData().
		setCount(1).
		setVerdict(resultDetails.Verdict).
		setProbeSuccesPercentage(resultDetails.ProbeSuccesPercentage)
	return float64(0)
}

// getEnv derived the ENVs and sets the default value if env contains empty value
func getEnv(key, defaultValue string) string {
	scrapeTime := os.Getenv(key)
	if scrapeTime == "" {
		scrapeTime = defaultValue
	}
	return scrapeTime
}

// setResultData sets the result data into resultStore so that the data
//can be used while handling chaosresult deletion
func (resultDetails *ChaosResultDetails) setResultData() {
	resultData := initialiseResultData().
		setLabel(resultDetails.ChaosEngineLabel).
		setAppKind(resultDetails.AppKind).
		setNs(resultDetails.AppNs).
		setAppLabel(resultDetails.AppLabel).
		setVerdict(resultDetails.Verdict).
		setCount(0).
		setProbeSuccesPercentage(resultDetails.ProbeSuccesPercentage)

	if resultStore[string(resultDetails.UID)] != nil {
		resultStore[string(resultDetails.UID)] = append(resultStore[string(resultDetails.UID)], *resultData)
	} else {
		resultStore[string(resultDetails.UID)] = []ResultData{*resultData}
	}
}

// initialiseResultData creates the instance of ResultData struct
func initialiseResultData() *ResultData {
	return &ResultData{}
}

// setLabel sets the label inside resultData struct
func (resultData *ResultData) setLabel(label string) *ResultData {
	resultData.Label = label
	return resultData
}

// setAppKind sets the appkind inside resultData struct
func (resultData *ResultData) setAppKind(appKind string) *ResultData {
	resultData.AppKind = appKind
	return resultData
}

// setNs sets the appNs inside resultData struct
func (resultData *ResultData) setNs(appNs string) *ResultData {
	resultData.AppNs = appNs
	return resultData
}

// setAppLabel sets the appLabel inside resultData struct
func (resultData *ResultData) setAppLabel(appLabel string) *ResultData {
	resultData.AppLabel = appLabel
	return resultData
}

// setVerdict sets the verdict inside resultData struct
func (resultData *ResultData) setVerdict(verdict string) *ResultData {
	resultData.Verdict = verdict
	return resultData
}

// setCount sets the count inside resultData struct
func (resultData *ResultData) setCount(count int) *ResultData {
	resultData.Count = count
	return resultData
}

// setProbeSuccesPercentage sets the probeSuccessPercentage inside resultData struct
func (resultData *ResultData) setProbeSuccesPercentage(probeSuccessPercentage float64) *ResultData {
	resultData.ProbeSuccessPercentage = float64(probeSuccessPercentage)
	return resultData
}
