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
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"github.com/litmuschaos/chaos-exporter/pkg/log"
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
)

var err error

// GetLitmusChaosMetrics derive and send the chaos metrics
func (m *MetricesCollecter) GetLitmusChaosMetrics(clients clients.ClientSets, overallChaosResults []*litmuschaosv1alpha1.ChaosResult, monitoringEnabled *MonitoringEnabled) (*time.Duration, error) {
	engineCount := 0

	// initialising the parameters for the namespaced scope metrics
	namespacedScopeMetrics := NamespacedScopeMetrics{
		PassedExperiments:         0,
		FailedExperiments:         0,
		AwaitedExperiments:        0,
		ExperimentRunCount:        0,
		ExperimentsInstalledCount: 0,
	}
	// getting all the data required for aws configuration
	awsConfig := AWSConfig{
		Namespace:   os.Getenv("AWS_CLOUDWATCH_METRIC_NAMESPACE"),
		ClusterName: os.Getenv("CLUSTER_NAME"),
		Service:     os.Getenv("APP_NAME"),
	}
	watchNamespace := os.Getenv("WATCH_NAMESPACE")
	// Getting list of all the chaosresults for the monitoring
	resultList, err := m.ResultCollector.GetResultList(clients, watchNamespace, monitoringEnabled)
	if err != nil {
		return nil, err
	}
	// unset the metrics correspond to deleted chaosresults
	m.GaugeMetrics.unsetDeletedChaosResults(overallChaosResults, resultList)
	// updating the overall chaosresults items to latest
	overallChaosResults = resultList

	var needRequeue *time.Duration

	// iterating over all chaosresults and derive all the metrics data it generates metrics per chaosresult
	// and aggregate metrics of all results present inside chaos namespace, if chaos namespace is defined
	// otherwise it derive metrics for all chaosresults present inside cluster
	for _, chaosresult := range resultList {

		m.ResultCollector.SetResultDetails()
		// deriving metrics data from the chaosresult
		skip, err := m.ResultCollector.GetExperimentMetricsFromResult(chaosresult, clients)
		resultDetails := m.ResultCollector.GetResultDetails()
		if err != nil {
			return nil, err
		}
		// generating the aggeregate metrics from per chaosresult metric
		namespacedScopeMetrics.AwaitedExperiments += resultDetails.AwaitedExperiments
		namespacedScopeMetrics.PassedExperiments += resultDetails.PassedExperiments
		namespacedScopeMetrics.FailedExperiments += resultDetails.FailedExperiments
		namespacedScopeMetrics.ExperimentsInstalledCount++
		namespacedScopeMetrics.ExperimentRunCount += resultDetails.AwaitedExperiments + resultDetails.PassedExperiments + resultDetails.FailedExperiments
		// skipping exporting metrics for the results, whose chaosengine is either completed or not exist
		if skip {
			continue
		}
		//engineCount is storing count of chaosengines
		//It is helping in keeping track of available chaosengines associated with chaosresults
		engineCount++

		//DISPLAY THE METRICS INFORMATION
		log.InfoWithValues("The chaos metrics are as follows", logrus.Fields{
			"ResultName":             resultDetails.Name,
			"ResultNamespace":        resultDetails.Namespace,
			"PassedExperiments":      resultDetails.PassedExperiments,
			"FailedExperiments":      resultDetails.FailedExperiments,
			"AwaitedExperiments":     resultDetails.AwaitedExperiments,
			"ProbeSuccessPercentage": resultDetails.ProbeSuccessPercentage,
			"StartTime":              resultDetails.StartTime,
			"EndTime":                resultDetails.EndTime,
			"ChaosInjectTime":        resultDetails.InjectionTime,
			"TotalDuration":          resultDetails.TotalDuration,
			"ResultVerdict":          resultDetails.Verdict,
			"FaultName":              resultDetails.FaultName,
		})

		// setting chaosresult metrics for the given chaosresult
		verdictValue, requeue := m.GaugeMetrics.unsetOutdatedMetrics(resultDetails)
		if requeue != nil {
			needRequeue = requeue
		}
		m.GaugeMetrics.setResultChaosMetrics(resultDetails, verdictValue)
		// setting chaosresult aws metrics for the given chaosresult, which can be used for cloudwatch
		if awsConfig.Namespace != "" && awsConfig.ClusterName != "" && awsConfig.Service != "" {
			awsConfig.setAwsResultChaosMetrics(resultDetails)
		}
	}
	if engineCount == 0 {
		if monitoringEnabled.IsChaosEnginesAvailable && monitoringEnabled.IsChaosResultsAvailable {
			monitoringEnabled.IsChaosEnginesAvailable = false
			log.Info("[Wait]: Hold on, no active chaosengine found ... ")
		}
	}
	if !monitoringEnabled.IsChaosEnginesAvailable && engineCount != 0 {
		monitoringEnabled.IsChaosEnginesAvailable = true
	}

	//setting aggregate metrics from the all chaosresults
	m.GaugeMetrics.setNamespacedChaosMetrics(namespacedScopeMetrics, watchNamespace)
	//setting aggregate aws metrics from the all chaosresults, which can be used for cloudwatch
	if awsConfig.Namespace != "" && awsConfig.ClusterName != "" && awsConfig.Service != "" {
		awsConfig.setAwsNamespacedChaosMetrics(namespacedScopeMetrics)
	}
	return needRequeue, nil
}

// setNamespacedChaosMetrics sets metrics for the all chaosresults
func (gaugeMetrics *GaugeMetrics) setNamespacedChaosMetrics(namespacedScopeMetrics NamespacedScopeMetrics, watchNamespace string) {
	switch watchNamespace {
	case "":
		gaugeMetrics.ClusterScopedTotalAwaitedExperiments.WithLabelValues().Set(namespacedScopeMetrics.AwaitedExperiments)
		gaugeMetrics.ClusterScopedTotalPassedExperiments.WithLabelValues().Set(namespacedScopeMetrics.PassedExperiments)
		gaugeMetrics.ClusterScopedTotalFailedExperiments.WithLabelValues().Set(namespacedScopeMetrics.FailedExperiments)
		gaugeMetrics.ClusterScopedExperimentsRunCount.WithLabelValues().Set(namespacedScopeMetrics.ExperimentRunCount)
		gaugeMetrics.ClusterScopedExperimentsInstalledCount.WithLabelValues().Set(namespacedScopeMetrics.ExperimentsInstalledCount)
	default:
		gaugeMetrics.NamespaceScopedTotalAwaitedExperiments.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.AwaitedExperiments)
		gaugeMetrics.NamespaceScopedTotalPassedExperiments.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.PassedExperiments)
		gaugeMetrics.NamespaceScopedTotalFailedExperiments.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.FailedExperiments)
		gaugeMetrics.NamespaceScopedExperimentsRunCount.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.ExperimentRunCount)
		gaugeMetrics.NamespaceScopedExperimentsInstalledCount.WithLabelValues(watchNamespace).Set(namespacedScopeMetrics.ExperimentsInstalledCount)
	}
}

// setResultChaosMetrics sets metrics for the given chaosresult details
func (gaugeMetrics *GaugeMetrics) setResultChaosMetrics(resultDetails ChaosResultDetails, verdictValue float64) {

	gaugeMetrics.ResultAwaitedExperiments.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.WorkflowName, resultDetails.FaultName).Set(resultDetails.AwaitedExperiments)
	gaugeMetrics.ResultPassedExperiments.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName).Set(resultDetails.PassedExperiments)
	gaugeMetrics.ResultFailedExperiments.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName).Set(resultDetails.FailedExperiments)
	gaugeMetrics.ResultProbeSuccessPercentage.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName).Set(resultDetails.ProbeSuccessPercentage)
	switch strings.ToLower(resultDetails.Verdict) {
	case "awaited":
		gaugeMetrics.ResultVerdict.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.Verdict, fmt.Sprintf("%f", resultDetails.ProbeSuccessPercentage),
			resultDetails.AppLabel, resultDetails.AppNs, resultDetails.AppKind, resultDetails.WorkflowName, resultDetails.FaultName).Set(float64(0))
	default:
		gaugeMetrics.ResultVerdict.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.Verdict, fmt.Sprintf("%f", resultDetails.ProbeSuccessPercentage),
			resultDetails.AppLabel, resultDetails.AppNs, resultDetails.AppKind, resultDetails.WorkflowName, resultDetails.FaultName).Set(verdictValue)
	}
	gaugeMetrics.ExperimentStartTime.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName).Set(resultDetails.StartTime)
	gaugeMetrics.ExperimentEndTime.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName).Set(resultDetails.EndTime)
	gaugeMetrics.ExperimentChaosInjectedTime.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName).Set(float64(resultDetails.InjectionTime))
	gaugeMetrics.ExperimentTotalDuration.WithLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName).Set(resultDetails.TotalDuration)
}

// unsetResultChaosMetrics unset metrics for the given chaosresult details
func (gaugeMetrics *GaugeMetrics) unsetResultChaosMetrics(resultDetails *ChaosResultDetails) {
	gaugeMetrics.ResultAwaitedExperiments.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.WorkflowName, resultDetails.FaultName)
	gaugeMetrics.ResultPassedExperiments.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName)
	gaugeMetrics.ResultFailedExperiments.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName)
	gaugeMetrics.ResultProbeSuccessPercentage.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName)
	gaugeMetrics.ResultVerdict.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.Verdict,
		fmt.Sprintf("%f", resultDetails.ProbeSuccessPercentage), resultDetails.AppLabel, resultDetails.AppNs, resultDetails.AppKind, resultDetails.WorkflowName, resultDetails.FaultName)
	gaugeMetrics.ExperimentStartTime.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName)
	gaugeMetrics.ExperimentEndTime.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName)
	gaugeMetrics.ExperimentChaosInjectedTime.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName)
	gaugeMetrics.ExperimentTotalDuration.DeleteLabelValues(resultDetails.Namespace, resultDetails.Name, resultDetails.ChaosEngineName, resultDetails.ChaosEngineContext, resultDetails.FaultName)
}

// setAwsResultChaosMetrics sets aws metrics for the given chaosresult
func (awsConfig *AWSConfig) setAwsResultChaosMetrics(resultDetails ChaosResultDetails) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	awsConfig.putAwsMetricData(sess, "chaosresult_passed_experiments", "Count", resultDetails.PassedExperiments)
	awsConfig.putAwsMetricData(sess, "chaosresult_failed_experiments", "Count", resultDetails.FailedExperiments)
	awsConfig.putAwsMetricData(sess, "chaosresult_awaited_experiments", "Count", resultDetails.AwaitedExperiments)
	awsConfig.putAwsMetricData(sess, "chaosresult_probe_success_percentage", "Count", resultDetails.ProbeSuccessPercentage)
	awsConfig.putAwsMetricData(sess, "chaosresult_start_time", "Count", resultDetails.StartTime)
	awsConfig.putAwsMetricData(sess, "chaosresult_end_time", "Count", resultDetails.EndTime)
	awsConfig.putAwsMetricData(sess, "chaosresult_inject_time", "Count", float64(resultDetails.InjectionTime))
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
