package controller

import (
	"math"
	"strconv"
	"strings"

	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"github.com/litmuschaos/chaos-exporter/pkg/log"
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientTypes "k8s.io/apimachinery/pkg/types"
)

// GetResultList return the result list correspond to the monitoring enabled chaosengine
func GetResultList(clients clients.ClientSets, chaosNamespace string, monitoringEnabled *MonitoringEnabled) (litmuschaosv1alpha1.ChaosResultList, error) {

	chaosResultList, err := clients.LitmusClient.ChaosResults(chaosNamespace).List(metav1.ListOptions{})
	if err != nil {
		return litmuschaosv1alpha1.ChaosResultList{}, err
	}
	// waiting until any chaosresult found
	if len(chaosResultList.Items) == 0 {
		if monitoringEnabled.IsChaosResultsAvailable {
			monitoringEnabled.IsChaosResultsAvailable = false
			log.Warnf("No chaosresult found!")
			log.Info("[Wait]: Waiting for the chaosresult ... ")
		}
		return litmuschaosv1alpha1.ChaosResultList{}, nil
	}

	if !monitoringEnabled.IsChaosResultsAvailable {
		log.Info("[Wait]: Cheers! Wait is over, found desired chaosresult")
		monitoringEnabled.IsChaosResultsAvailable = true
	}

	return *chaosResultList, nil
}

// getExperimentMetricsFromResult derive all the metrics data from the chaosresult and set into resultDetails struct
func (resultDetails *ChaosResultDetails) getExperimentMetricsFromResult(chaosResult *litmuschaosv1alpha1.ChaosResult, clients clients.ClientSets) error {
	verdict := strings.ToLower(chaosResult.Status.ExperimentStatus.Verdict)
	probeSuccesPercentage, err := getProbeSuccessPercentage(chaosResult)
	if err != nil {
		return err
	}

	engine, err := clients.LitmusClient.ChaosEngines(chaosResult.Namespace).Get(chaosResult.Spec.EngineName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// deriving all the events present inside specific chaosengine
	events, err := getEventsForSpecificInvolvedResource(clients, engine.UID, chaosResult.Namespace)
	if err != nil {
		return err
	}

	// setting all the values inside resultdetails struct
	resultDetails.setName(chaosResult.Name).
		setUID(chaosResult.UID).
		setNamespace(chaosResult.Namespace).
		setProbeSuccesPercentage(probeSuccesPercentage).
		setVerdict(chaosResult.Status.ExperimentStatus.Verdict).
		setStartTime(events).
		setEndTime(events).
		setChaosInjectTime(events).
		setChaosEngineName(chaosResult.Spec.EngineName).
		setChaosEngineLabel(engine.Labels[EngineLabelKey]).
		setAppLabel(engine.Spec.Appinfo.Applabel).
		setAppNs(engine.Spec.Appinfo.Appns).
		setAppKind(engine.Spec.Appinfo.AppKind).
		setTotalDuration().
		setVerdictCount(verdict, chaosResult).
		setResultData()

	return nil
}

// initialiseResult create the new instance of the ChaosResultDetails struct
func initialiseResult() *ChaosResultDetails {
	return &ChaosResultDetails{}

}

// setName sets name inside resultDetails struct
func (resultDetails *ChaosResultDetails) setName(name string) *ChaosResultDetails {
	resultDetails.Name = name
	return resultDetails
}

// setNamespace sets namespace inside resultDetails struct
func (resultDetails *ChaosResultDetails) setNamespace(namespace string) *ChaosResultDetails {
	resultDetails.Namespace = namespace
	return resultDetails
}

// setUID sets result uid inside the resultDetails struct
func (resultDetails *ChaosResultDetails) setUID(uid clientTypes.UID) *ChaosResultDetails {
	resultDetails.UID = uid
	return resultDetails
}

// setVerdict sets result verdict inside the resultDetails struct
func (resultDetails *ChaosResultDetails) setVerdict(verdict string) *ChaosResultDetails {
	resultDetails.Verdict = verdict
	return resultDetails
}

// setVerdict increase the metric count based on given verdict/events
func (resultDetails *ChaosResultDetails) setVerdictCount(verdict string, chaosResult *litmuschaosv1alpha1.ChaosResult) *ChaosResultDetails {

	// count the chaosresult as awaited if verdict is awaited
	switch verdict {
	case "awaited":
		resultDetails.AwaitedExperiments++
	}
	resultDetails.PassedExperiments = float64(chaosResult.Status.History.PassedRuns)
	resultDetails.FailedExperiments = float64(chaosResult.Status.History.FailedRuns)
	return resultDetails
}

// setProbeSuccesPercentage sets ProbeSuccesPercentage inside resultDetails struct
func (resultDetails *ChaosResultDetails) setProbeSuccesPercentage(probeSuccesPercentage float64) *ChaosResultDetails {
	resultDetails.ProbeSuccesPercentage = probeSuccesPercentage
	return resultDetails
}

// setChaosEngineName sets the chaosEngine name inside resultDetails struct
func (resultDetails *ChaosResultDetails) setChaosEngineName(chaosEngineName string) *ChaosResultDetails {
	resultDetails.ChaosEngineName = chaosEngineName
	return resultDetails
}

// setAppLabel sets the target application labels inside resultDetails struct
func (resultDetails *ChaosResultDetails) setAppLabel(appLabel string) *ChaosResultDetails {
	resultDetails.AppLabel = appLabel
	return resultDetails
}

// setAppLabel sets the target application namespace inside resultDetails struct
func (resultDetails *ChaosResultDetails) setAppNs(appNs string) *ChaosResultDetails {
	resultDetails.AppNs = appNs
	return resultDetails
}

// setAppLabel sets the target application kind inside resultDetails struct
func (resultDetails *ChaosResultDetails) setAppKind(appKind string) *ChaosResultDetails {
	resultDetails.AppKind = appKind
	return resultDetails
}

// setChaosEngineLabel sets the chaosEngine label inside resultDetails struct
func (resultDetails *ChaosResultDetails) setChaosEngineLabel(engineLabel string) *ChaosResultDetails {
	resultDetails.ChaosEngineLabel = engineLabel
	return resultDetails
}

// setStartTime sets start time of experiment run
func (resultDetails *ChaosResultDetails) setStartTime(events corev1.EventList) *ChaosResultDetails {
	startTime := int64(0)
	for _, event := range events.Items {
		// job create event by runner
		if event.Reason == "ExperimentDependencyCheck" {
			startTime = maximum(startTime, event.LastTimestamp.Unix())
		}
	}
	resultDetails.StartTime = float64(startTime)
	return resultDetails
}

// setEndTime sets end time of the experiment run
func (resultDetails *ChaosResultDetails) setEndTime(events corev1.EventList) *ChaosResultDetails {
	endTime := int64(0)
	for _, event := range events.Items {
		if event.Reason == "Summary" {
			endTime = maximum(endTime, event.LastTimestamp.Unix())
		}
	}
	resultDetails.EndTime = float64(endTime)
	return resultDetails
}

// setChaosInjectTime sets the chaos injection time
func (resultDetails *ChaosResultDetails) setChaosInjectTime(events corev1.EventList) *ChaosResultDetails {
	chaosInjectTime := int64(0)
	for _, event := range events.Items {
		if event.Reason == "ChaosInject" {
			chaosInjectTime = maximum(chaosInjectTime, event.LastTimestamp.Unix())
		}
	}
	resultDetails.InjectionTime = float64(chaosInjectTime)
	return resultDetails
}

// setTotalDuration sets total chaos duration for the experiment run
func (resultDetails *ChaosResultDetails) setTotalDuration() *ChaosResultDetails {
	resultDetails.TotalDuration = math.Max(0, resultDetails.EndTime-resultDetails.StartTime)
	return resultDetails
}

// getProbeSuccessPercentage derive the probeSucessPercentage from the chaosresult
func getProbeSuccessPercentage(chaosResult *litmuschaosv1alpha1.ChaosResult) (float64, error) {
	probeSuccesPercentage := float64(0)
	if chaosResult.Status.ExperimentStatus.ProbeSuccessPercentage != "Awaited" && chaosResult.Status.ExperimentStatus.ProbeSuccessPercentage != "" {
		probeSuccesPercentage, err = strconv.ParseFloat(chaosResult.Status.ExperimentStatus.ProbeSuccessPercentage, 64)
		if err != nil {
			return 0, err
		}
	}
	return probeSuccesPercentage, nil
}

// getEventsForSpecificInvolvedResource derive all the events correspond to the specific resource
func getEventsForSpecificInvolvedResource(clients clients.ClientSets, resourceUID clientTypes.UID, chaosNamespace string) (corev1.EventList, error) {
	finalEventList := corev1.EventList{}
	eventsList, err := clients.KubeClient.CoreV1().Events(chaosNamespace).List(metav1.ListOptions{})
	if err != nil {
		return corev1.EventList{}, err
	}

	for _, event := range eventsList.Items {
		if event.InvolvedObject.UID == resourceUID {
			finalEventList.Items = append(finalEventList.Items, event)
		}
	}
	return finalEventList, nil
}

// Maximum returns the maximum value
func maximum(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
