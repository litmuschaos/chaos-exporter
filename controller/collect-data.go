package controller

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"github.com/litmuschaos/chaos-exporter/pkg/log"
	"github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientTypes "k8s.io/apimachinery/pkg/types"
)

//go:generate mockgen -destination=mocks/mock_collect-data.go -package=mocks github.com/litmuschaos/chaos-exporter/controller ResultCollector

// ResultCollector interface for the both functions GetResultList and getExperimentMetricsFromResult
type ResultCollector interface {
	GetResultList(clients clients.ClientSets, chaosNamespace string, monitoringEnabled *MonitoringEnabled) (litmuschaosv1alpha1.ChaosResultList, error)
	GetExperimentMetricsFromResult(chaosResult *litmuschaosv1alpha1.ChaosResult, clients clients.ClientSets) (bool, error)
}
type ResultDetails struct {
	resultDetails ChaosResultDetails
}

// GetResultList return the result list correspond to the monitoring enabled chaosengine
func (r *ResultDetails) GetResultList(clients clients.ClientSets, chaosNamespace string, monitoringEnabled *MonitoringEnabled) (litmuschaosv1alpha1.ChaosResultList, error) {

	chaosResultList, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosResults(chaosNamespace).List(context.Background(), metav1.ListOptions{})
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

// GetExperimentMetricsFromResult derive all the metrics data from the chaosresult and set into resultDetails struct
func (r *ResultDetails) GetExperimentMetricsFromResult(chaosResult *litmuschaosv1alpha1.ChaosResult, clients clients.ClientSets) (bool, error) {
	verdict := strings.ToLower(string(chaosResult.Status.ExperimentStatus.Verdict))
	probeSuccesPercentage, err := getProbeSuccessPercentage(chaosResult)
	if err != nil {
		return false, err
	}
	engine, err := clients.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(chaosResult.Namespace).Get(context.Background(), chaosResult.Spec.EngineName, metav1.GetOptions{})
	if err != nil {
		// k8serrors.IsNotFound(err) checking k8s resource is found or not,
		// It will skip this result if k8s resource is not found.
		if k8serrors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}

	// deriving all the events present inside specific chaosengine
	events, err := getEventsForSpecificInvolvedResource(clients, engine.UID, chaosResult.Namespace)
	if err != nil {
		return false, err
	}
	// setting all the values inside resultdetails struct
	r.resultDetails.setName(chaosResult.Name).
		setUID(chaosResult.UID).
		setNamespace(chaosResult.Namespace).
		setProbeSuccessPercentage(probeSuccesPercentage).
		setVerdict(string(chaosResult.Status.ExperimentStatus.Verdict)).
		setStartTime(events).
		setEndTime(events).
		setChaosInjectTime(events).
		setChaosEngineName(chaosResult.Spec.EngineName).
		setChaosEngineContext(engine.Labels[EngineContext]).
		setWorkflowName(engine.Labels[WorkFlowName]).
		setAppLabel(engine.Spec.Appinfo.Applabel).
		setAppNs(engine.Spec.Appinfo.Appns).
		setAppKind(engine.Spec.Appinfo.AppKind).
		setTotalDuration().
		setVerdictCount(verdict, chaosResult).
		setFaultName(engine.Spec.Experiments[0].Name).
		setResultData()

	// it won't export/override the metrics if chaosengine is in completed state and
	// experiment's final verdict[passed,failed,stopped] is already exported/overridden
	if engine.Status.EngineStatus == v1alpha1.EngineStatusCompleted {
		result, ok := matchVerdict[string(r.resultDetails.UID)]
		if !ok || (ok && result.Verdict == r.resultDetails.Verdict) {
			return true, nil
		}
	}

	return false, nil
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

// setProbeSuccessPercentage sets ProbeSuccessPercentage inside resultDetails struct
func (resultDetails *ChaosResultDetails) setProbeSuccessPercentage(probeSuccessPercentage float64) *ChaosResultDetails {
	resultDetails.ProbeSuccessPercentage = probeSuccessPercentage
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

// setChaosEngineContext sets the chaosEngine context inside resultDetails struct
func (resultDetails *ChaosResultDetails) setChaosEngineContext(engineLabel string) *ChaosResultDetails {
	resultDetails.ChaosEngineContext = engineLabel
	return resultDetails
}

// setWorkflowName sets the workflow name inside resultDetails struct
func (resultDetails *ChaosResultDetails) setWorkflowName(workflowName string) *ChaosResultDetails {
	resultDetails.WorkflowName = workflowName
	return resultDetails
}

// setFaultName sets the fault name inside resultDetails struct
func (resultDetails *ChaosResultDetails) setFaultName(faultName string) *ChaosResultDetails {
	resultDetails.FaultName = faultName
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
	resultDetails.InjectionTime = chaosInjectTime
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
	eventsList, err := clients.KubeClient.CoreV1().Events(chaosNamespace).List(context.Background(), metav1.ListOptions{})
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
