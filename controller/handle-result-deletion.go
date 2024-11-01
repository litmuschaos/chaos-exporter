package controller

import (
	"fmt"
	"os"
	"strconv"
	"time"

	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
)

// unsetDeletedChaosResults unset the metrics correspond to deleted chaosresults
func (gaugeMetrics *GaugeMetrics) unsetDeletedChaosResults(oldChaosResults, newChaosResults []*litmuschaosv1alpha1.ChaosResult) {
	for _, oldResult := range oldChaosResults {
		found := false
		for _, newResult := range newChaosResults {
			if oldResult.UID == newResult.UID {
				found = true
				break
			}
		}

		if !found {
			for _, value := range resultStore[string(oldResult.UID)] {

				probeSuccesPercentage, _ := getProbeSuccessPercentage(oldResult)
				resultDetails := initialiseResult().
					setName(oldResult.Name).
					setNamespace(oldResult.Namespace).
					setProbeSuccessPercentage(probeSuccesPercentage).
					setVerdict(value.Verdict).
					setAppLabel(value.AppLabel).
					setAppNs(value.AppNs).
					setAppKind(value.AppKind).
					setChaosEngineName(oldResult.Spec.EngineName).
					setChaosEngineContext(value.ChaosEngineContext).
					setWorkflowName(value.WorkFlowName).
					setFaultName(value.FaultName)

				gaugeMetrics.unsetResultChaosMetrics(resultDetails)
			}
			// delete the corresponding entry from the map
			delete(resultStore, string(oldResult.UID))
		}
	}
}

// unsetOutdatedMetrics unset the metrics when chaosresult verdict changes
// if same chaosresult is continuously repeated more than scrape interval then it sets the metrics value to 0
func (gaugeMetrics *GaugeMetrics) unsetOutdatedMetrics(resultDetails ChaosResultDetails) (float64, *time.Duration) {
	scrapeTime, _ := strconv.Atoi(getEnv("TSDB_SCRAPE_INTERVAL", "10"))
	result, ok := matchVerdict[string(resultDetails.UID)]
	reset := false
	var needRequeue *time.Duration

	scrapeDuration := time.Duration(scrapeTime) * time.Second

	switch ok {
	case true:
		switch {
		// if verdict is different then delete the older metrics having outdated verdict
		case result.Verdict != resultDetails.Verdict:
			gaugeMetrics.ResultVerdict.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName,
				resultDetails.ChaosEngineContext, result.Verdict, fmt.Sprintf("%f", result.ProbeSuccessPercentage), resultDetails.AppLabel,
				resultDetails.AppNs, resultDetails.AppKind, resultDetails.WorkflowName, resultDetails.FaultName)
			result.Timer = time.Now()
			needRequeue = &scrapeDuration
		default:
			// if time passed scrape time then reset the value to 0
			if time.Since(result.Timer) >= scrapeDuration {
				reset = true
			}
		}
	default:
		result = initialiseResultData().
			setTimer(time.Now()).
			setVerdictReset(false)
		needRequeue = &scrapeDuration
	}

	// update the values inside matchVerdict
	matchVerdict[string(resultDetails.UID)] = result.setVerdict(resultDetails.Verdict).
		setProbeSuccesPercentage(resultDetails.ProbeSuccessPercentage).
		setVerdictReset(reset)

	if reset {
		return float64(0), needRequeue
	}
	return float64(1), needRequeue
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
// can be used while handling chaosresult deletion
func (resultDetails *ChaosResultDetails) setResultData() {
	resultData := initialiseResultData().
		setContext(resultDetails.ChaosEngineContext).
		setWorkflowName(resultDetails.WorkflowName).
		setAppKind(resultDetails.AppKind).
		setNs(resultDetails.AppNs).
		setAppLabel(resultDetails.AppLabel).
		setVerdict(resultDetails.Verdict).
		setFaultName(resultDetails.FaultName).
		setTimer(time.Now()).
		setVerdictReset(false).
		setProbeSuccesPercentage(resultDetails.ProbeSuccessPercentage)

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

// setContext sets the engine context inside resultData struct
func (resultData *ResultData) setContext(context string) *ResultData {
	resultData.ChaosEngineContext = context
	return resultData
}

// setWorkflowName sets the workflow name inside resultData struct
func (resultData *ResultData) setWorkflowName(workflowName string) *ResultData {
	resultData.WorkFlowName = workflowName
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

// setFaultName sets the fault name inside resultData struct
func (resultData *ResultData) setFaultName(fault string) *ResultData {
	resultData.FaultName = fault
	return resultData
}

// setCount sets the count inside resultData struct
func (resultData *ResultData) setTimer(timer time.Time) *ResultData {
	resultData.Timer = timer
	return resultData
}

// setVerdictReset sets the VerdictReset inside resultData struct
func (resultData *ResultData) setVerdictReset(verdictreset bool) *ResultData {
	resultData.VerdictReset = verdictreset
	return resultData
}

// setProbeSuccesPercentage sets the probeSuccessPercentage inside resultData struct
func (resultData *ResultData) setProbeSuccesPercentage(probeSuccessPercentage float64) *ResultData {
	resultData.ProbeSuccessPercentage = float64(probeSuccessPercentage)
	return resultData
}
