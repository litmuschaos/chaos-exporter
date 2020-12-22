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
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"github.com/litmuschaos/chaos-exporter/pkg/log"
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var err error

// GetLitmusChaosMetrics derive and send the chaos metrics
func GetLitmusChaosMetrics(clients clients.ClientSets) error {
	// initialising the parameters for the namespaced scope metrics
	namespacedScopeMetrics := NamespacedScopeMetrics{
		PassedExperiments:  0,
		FailedExperiments:  0,
		AwaitedExperiments: 0,
		ExperimentRunCount: 0,
	}
	// getting all the data required for aws configuration
	awsConfig := AWSConfig{
		Namespace:   os.Getenv("AWS_CLOUDWATCH_METRIC_NAMESPACE"),
		ClusterName: os.Getenv("CLUSTER_NAME"),
		Service:     os.Getenv("APP_NAME"),
	}
	watchNamespace := os.Getenv("WATCH_NAMESPACE")
	// Getting list of all the chaosresults for the monitoring
	resultList, err := GetResultList(clients, watchNamespace)
	if err != nil {
		return err
	}

	// iterating over all chaosresults and derive all the metrics data it generates metrics per chaosresult
	// and aggregate metrics of all results present inside chaos namespace, if chaos namespace is defined
	// otherwise it derive metrics for all chaosresults present inside cluster
	for _, chaosresult := range resultList.Items {
		resultDetails := ChaosResultDetails{
			PassedExperiments:  0,
			FailedExperiments:  0,
			AwaitedExperiments: 0,
		}

		// deriving metrics data from the chaosresult
		err = resultDetails.getExperimentMetricsFromResult(&chaosresult, clients)
		if err != nil {
			log.Errorf("err: %v", err)
		}

		//DISPLAY THE METRICS INFORMATION
		log.InfoWithValues("The chaos metrics are as follows", logrus.Fields{
			"ResultName":             resultDetails.Name,
			"ResultNamespace":        resultDetails.Namespace,
			"PassedExperiments":      resultDetails.PassedExperiments,
			"FailedExperiments":      resultDetails.FailedExperiments,
			"AwaitedExperiments":     resultDetails.AwaitedExperiments,
			"ProbeSuccessPercentage": resultDetails.ProbeSuccesPercentage,
			"StartTime":              resultDetails.StartTime,
			"EndTime":                resultDetails.EndTime,
			"ChaosInjectTime":        resultDetails.InjectionTime,
			"TotalDuration":          resultDetails.TotalDuration,
		})

		// generating the aggeregate metrics from per chaosresult metric
		namespacedScopeMetrics.AwaitedExperiments += resultDetails.AwaitedExperiments
		namespacedScopeMetrics.PassedExperiments += resultDetails.PassedExperiments
		namespacedScopeMetrics.FailedExperiments += resultDetails.FailedExperiments
		namespacedScopeMetrics.ExperimentsInstalledCount++
		namespacedScopeMetrics.ExperimentRunCount = resultDetails.AwaitedExperiments + resultDetails.PassedExperiments + resultDetails.FailedExperiments
		// setting chaosresult metrics for the given chaosresult
		setResultChaosMetrics(resultDetails)

		// setting chaosresult aws metrics for the given chaosresult, which can be used for cloudwatch
		if awsConfig.Namespace != "" && awsConfig.ClusterName != "" && awsConfig.Service != "" {
			awsConfig.setAwsResultChaosMetrics(resultDetails)
		}
	}
	//setting aggregate metrics from the all chaosresults
	setNamespacedChaosMetrics(namespacedScopeMetrics, watchNamespace)
	//setting aggregate aws metrics from the all chaosresults, which can be used for cloudwatch
	if awsConfig.Namespace != "" && awsConfig.ClusterName != "" && awsConfig.Service != "" {
		awsConfig.setAwsNamespacedChaosMetrics(namespacedScopeMetrics)
	}
	return nil
}

// setNamespacedChaosMetrics sets metrics for the all chaosresults
func setNamespacedChaosMetrics(namespacedScopeMetrics NamespacedScopeMetrics, watchNamespace string) {
	TotalAwaitedExperiments.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.AwaitedExperiments)
	TotalPassedExperiments.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.PassedExperiments)
	TotalFailedExperiments.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.FailedExperiments)
	ExperimentsRunCount.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.ExperimentRunCount)
	ExperimentsInstalledCount.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.ExperimentsInstalledCount)
}

// setResultChaosMetrics sets metrics for the given chaosresult
func setResultChaosMetrics(resultDetails ChaosResultDetails) {
	ResultAwaitedExperiments.WithLabelValues(resultDetails.Namespace, resultDetails.Name).Set(resultDetails.AwaitedExperiments)
	ResultPassedExperiments.WithLabelValues(resultDetails.Namespace, resultDetails.Name).Set(resultDetails.PassedExperiments)
	ResultFailedExperiments.WithLabelValues(resultDetails.Namespace, resultDetails.Name).Set(resultDetails.FailedExperiments)
	ResultProbeSuccessPercentage.WithLabelValues(resultDetails.Namespace, resultDetails.Name).Set(resultDetails.ProbeSuccesPercentage)
	ExperimentStartTime.WithLabelValues(resultDetails.Namespace, resultDetails.Name).Set(resultDetails.StartTime)
	ExperimentEndTime.WithLabelValues(resultDetails.Namespace, resultDetails.Name).Set(resultDetails.EndTime)
	ExperimentChaosInjectedTime.WithLabelValues(resultDetails.Namespace, resultDetails.Name).Set(resultDetails.InjectionTime)
	ExperimentTotalDuration.WithLabelValues(resultDetails.Namespace, resultDetails.Name).Set(resultDetails.TotalDuration)
}

// setAwsResultChaosMetrics sets aws metrics for the given chaosresult
func (awsConfig *AWSConfig) setAwsResultChaosMetrics(resultDetails ChaosResultDetails) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	awsConfig.putAwsMetricData(sess, "chaosresult_passed_experiments", "Count", resultDetails.PassedExperiments)
	awsConfig.putAwsMetricData(sess, "chaosresult_failed_experiments", "Count", resultDetails.FailedExperiments)
	awsConfig.putAwsMetricData(sess, "chaosresult_awaited_experiments", "Count", resultDetails.AwaitedExperiments)
	awsConfig.putAwsMetricData(sess, "chaosresult_probe_success_percentage", "Count", resultDetails.ProbeSuccesPercentage)
	awsConfig.putAwsMetricData(sess, "chaosresult_start_time", "Count", resultDetails.StartTime)
	awsConfig.putAwsMetricData(sess, "chaosresult_end_time", "Count", resultDetails.EndTime)
	awsConfig.putAwsMetricData(sess, "chaosresult_inject_time", "Count", resultDetails.InjectionTime)
	awsConfig.putAwsMetricData(sess, "chaosresult_total_duration", "Count", resultDetails.TotalDuration)
}

// setAwsNamespacedChaosMetrics sets aws metrics for all chaosresults
func (awsConfig *AWSConfig) setAwsNamespacedChaosMetrics(namespacedScopeMetrics NamespacedScopeMetrics) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	awsConfig.putAwsMetricData(sess, "total_passed_experiments", "Count", namespacedScopeMetrics.PassedExperiments)
	awsConfig.putAwsMetricData(sess, "total_failed_experiments", "Count", namespacedScopeMetrics.FailedExperiments)
	awsConfig.putAwsMetricData(sess, "total_awaited_experiments", "Count", namespacedScopeMetrics.AwaitedExperiments)
	awsConfig.putAwsMetricData(sess, "experiments_run_count", "Count", namespacedScopeMetrics.ExperimentRunCount)
	awsConfig.putAwsMetricData(sess, "experiments_installed_count", "Count", namespacedScopeMetrics.ExperimentsInstalledCount)
}

// filterMonitoringEnabledEngines filters the monitoring enabled engines from the given list
func filterMonitoringEnabledEngines(engineList *litmuschaosv1alpha1.ChaosEngineList) *litmuschaosv1alpha1.ChaosEngineList {
	var filteredEngineList litmuschaosv1alpha1.ChaosEngineList
	for i := range engineList.Items {
		// Condition to decide whether current element need to be picked for monitoring
		if engineList.Items[i].Spec.Monitoring {
			filteredEngineList.Items = append(filteredEngineList.Items, engineList.Items[i])
		}
	}
	return &filteredEngineList
}

// putAwsMetricData put the metrics data in cloudwatch service
func (awsConfig *AWSConfig) putAwsMetricData(sess *session.Session, metricName string, unit string, value float64) error {
	dimension1 := "ClusterName"
	dimension2 := "Service"
	// Create new Amazon CloudWatch client
	svc := cloudwatch.New(sess)

	if awsConfig.Namespace == "" || awsConfig.ClusterName == "" || awsConfig.Service == "" {
		return errors.Errorf("You must supply a namespace, clusterName and serviceName values")
	}

	log.Infof("Putting new AWS metric: Namespace %v, Metric %v", awsConfig.Namespace, metricName)

	_, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: &awsConfig.Namespace,
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: &metricName,
				Unit:       &unit,
				Value:      &value,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  &dimension1,
						Value: &awsConfig.ClusterName,
					},
					{
						Name:  &dimension2,
						Value: &awsConfig.Service,
					},
				},
			},
		},
	})
	if err != nil {
		log.Errorf("Error during putting metrics to CloudWatch: %v", err)
		return err
	}

	return nil
}

// GetResultList return the result list correspond to the monitoring enabled chaosengine
func GetResultList(clients clients.ClientSets, chaosNamespace string) (litmuschaosv1alpha1.ChaosResultList, error) {

	finalChaosResultList := litmuschaosv1alpha1.ChaosResultList{}
	chaosEngineList, err := clients.LitmusClient.ChaosEngines(chaosNamespace).List(metav1.ListOptions{})
	if err != nil {
		return litmuschaosv1alpha1.ChaosResultList{}, err
	}
	// filter the chaosengines based on monitoring enabled
	filteredChaosEngineList := filterMonitoringEnabledEngines(chaosEngineList)
	if len(filteredChaosEngineList.Items) == 0 {
		log.Warnf("No chaosengine found with monitoring enabled")
		return litmuschaosv1alpha1.ChaosResultList{}, nil
	}

	chaosResultList, err := clients.LitmusClient.ChaosResults(chaosNamespace).List(metav1.ListOptions{})
	if err != nil {
		return litmuschaosv1alpha1.ChaosResultList{}, err
	}

	// pick only those chaosresults, which correspond to the filtered chaosengines
	for _, chaosresult := range chaosResultList.Items {
		for _, chaosengine := range filteredChaosEngineList.Items {
			if chaosengine.Name == chaosresult.Spec.EngineName {
				finalChaosResultList.Items = append(finalChaosResultList.Items, chaosresult)
			}
		}
	}
	return finalChaosResultList, nil
}

// getExperimentMetricsFromResult derive all the metrics data from the chaosresult and set into resultDetails struct
func (resultDetails *ChaosResultDetails) getExperimentMetricsFromResult(chaosResult *litmuschaosv1alpha1.ChaosResult, clients clients.ClientSets) error {
	probeSuccesPercentage := float64(0)
	verdict := strings.ToLower(chaosResult.Status.ExperimentStatus.Verdict)
	if chaosResult.Status.ExperimentStatus.ProbeSuccessPercentage != "Awaited" {
		probeSuccesPercentage, err = strconv.ParseFloat(chaosResult.Status.ExperimentStatus.ProbeSuccessPercentage, 64)
		if err != nil {
			return err
		}
	}
	// deriving all the events present inside specific chaosengine
	events, err := getEventsForSpecificInvolvedResource(clients, chaosResult.Spec.EngineName, chaosResult.Namespace)
	if err != nil {
		return err
	}

	// setting all the values inside resultdetails struct
	resultDetails.setName(chaosResult.Name).
		setNamespace(chaosResult.Namespace).
		setProbeSuccesPercentage(probeSuccesPercentage).
		setStartTime(events).
		setEndTime(events).
		setChaosInjectTime(events).
		setTotalDuration()

	if err = resultDetails.setVerdictCount(verdict, clients); err != nil {
		return err
	}

	return nil
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

// setVerdict increase the metric count based on given verdict/events
func (resultDetails *ChaosResultDetails) setVerdictCount(verdict string, clients clients.ClientSets) error {

	// count the passed events counts
	passedEvents, err := resultDetails.getPassedEventsInResult(clients)
	if err != nil {
		return err
	}
	resultDetails.PassedExperiments = float64(passedEvents)

	// count the failed events counts
	failedEvents, err := resultDetails.getFailedEventsInResult(clients)
	if err != nil {
		return err
	}
	resultDetails.FailedExperiments = float64(failedEvents)

	// count the chaosresult as awaited if verdict is awaited
	switch verdict {
	case "awaited":
		resultDetails.AwaitedExperiments++
	}
	return nil
}

// setProbeSuccesPercentage sets ProbeSuccesPercentage inside resultDetails struct
func (resultDetails *ChaosResultDetails) setProbeSuccesPercentage(probeSuccesPercentage float64) *ChaosResultDetails {
	resultDetails.ProbeSuccesPercentage = probeSuccesPercentage
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

// getEventsForSpecificInvolvedResource derive all the events correspond to the specific resource
func getEventsForSpecificInvolvedResource(clients clients.ClientSets, resourceName, chaosNamespace string) (corev1.EventList, error) {
	finalEventList := corev1.EventList{}
	eventsList, err := clients.KubeClient.CoreV1().Events(chaosNamespace).List(metav1.ListOptions{})
	if err != nil {
		return corev1.EventList{}, err
	}

	for _, event := range eventsList.Items {
		if event.InvolvedObject.Name == resourceName {
			finalEventList.Items = append(finalEventList.Items, event)
		}
	}
	return finalEventList, nil
}

// getPassedEventsInResult count the passed events inside given chaosresult
func (resultDetails ChaosResultDetails) getPassedEventsInResult(clients clients.ClientSets) (int, error) {
	passedEventCount := 0
	eventsList, err := getEventsForSpecificInvolvedResource(clients, resultDetails.Name, resultDetails.Namespace)
	if err != nil {
		return passedEventCount, err
	}
	for _, event := range eventsList.Items {
		if event.Reason == "Pass" {
			passedEventCount++
		}
	}
	return passedEventCount, nil
}

// getFailedEventsInResult count the passed events inside given chaosresult
func (resultDetails ChaosResultDetails) getFailedEventsInResult(clients clients.ClientSets) (int, error) {
	failedEventCount := 0
	eventsList, err := getEventsForSpecificInvolvedResource(clients, resultDetails.Name, resultDetails.Namespace)
	if err != nil {
		return failedEventCount, err
	}
	for _, event := range eventsList.Items {
		if event.Reason == "Fail" {
			failedEventCount++
		}
	}
	return failedEventCount, nil
}

// Maximum returns the maximum value
func maximum(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
