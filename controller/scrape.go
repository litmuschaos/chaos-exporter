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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
)

// Holds list of experiments in a chaosengine
var chaosExperimentList []string

// Holds the chaosresult of the running experiment
var experimentStatusMap = make(map[string]bool)

// GetLitmusChaosMetrics returns chaos metrics for a given chaosengine
func GetLitmusChaosMetrics(clientSet *clientV1alpha1.Clientset) error {
	chaosEngineList, err := clientSet.LitmuschaosV1alpha1().ChaosEngines("").List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	filteredChaosEngineList := filterMonitoringEnabledEngines(chaosEngineList)
	if err != nil {
		return err
	}
	var total float64 = 0
	var pass float64 = 0
	var fail float64 = 0

	for _, chaosEngine := range filteredChaosEngineList.Items {
		totalEngine, passedEngine, failedEngine, awaitedEngine := getExperimentMetricsFromEngine(&chaosEngine)
		klog.V(0).Infof("ChaosEngineMetrics: EngineName: %v, EngineNamespace: %v, TotalExp: %v, PassedExp: %v, FailedExp: %v, TotalRunningExp: %v", chaosEngine.Name, chaosEngine.Namespace, totalEngine, passedEngine, failedEngine, len(experimentStatusMap))
		var engineDetails ChaosEngineDetail
		engineDetails.Name = chaosEngine.Name
		engineDetails.Namespace = chaosEngine.Namespace
		engineDetails.TotalExp = totalEngine
		engineDetails.PassedExp = passedEngine
		engineDetails.FailedExp = failedEngine
		engineDetails.AwaitedExp = awaitedEngine
		total += totalEngine
		pass += passedEngine
		fail += failedEngine
		setEngineChaosMetrics(engineDetails, &chaosEngine)
		setAwsEngineChaosMetrics(engineDetails, &chaosEngine)
	}

	setClusterChaosMetrics(total, pass, fail)
	setAwsClusterChaosMetrics(total, pass, fail)
	return nil
}

func setClusterChaosMetrics(total float64, pass float64, fail float64) {
	ClusterPassedExperiments.WithLabelValues().Set(pass)
	ClusterFailedExperiments.WithLabelValues().Set(fail)
	ClusterTotalExperiments.WithLabelValues().Set(total)
}
func setEngineChaosMetrics(engineDetails ChaosEngineDetail, chaosEngine *litmuschaosv1alpha1.ChaosEngine) {
	EngineRunningExperiment.WithLabelValues(engineDetails.Namespace, engineDetails.Name, fmt.Sprintf("%s-%s", chaosEngine.Name, chaosEngine.Namespace)).Set(float64(len(experimentStatusMap)))
	EngineTotalExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.TotalExp)
	EnginePassedExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.PassedExp)
	EngineFailedExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.FailedExp)
	EngineWaitingExperiments.WithLabelValues(engineDetails.Namespace, engineDetails.Name).Set(engineDetails.AwaitedExp)
}
func setAwsEngineChaosMetrics(engineDetails ChaosEngineDetail, chaosEngine *litmuschaosv1alpha1.ChaosEngine) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	putAwsMetricData(sess, "chaosengine_passed_experiments", "Count", engineDetails.PassedExp)
	putAwsMetricData(sess, "chaosengine_failed_experiments", "Count", engineDetails.FailedExp)
	putAwsMetricData(sess, "chaosengine_experiments_count", "Count", engineDetails.TotalExp)
	putAwsMetricData(sess, "chaosengine_waiting_experiments", "Count", engineDetails.AwaitedExp)
}

func setAwsClusterChaosMetrics(total float64, pass float64, fail float64) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	putAwsMetricData(sess, "cluster_passed_experiments", "Count", pass)
	putAwsMetricData(sess, "cluster_failed_experiments", "Count", fail)
	putAwsMetricData(sess, "cluster_experiments_count", "Count", total)
}

func getExperimentMetricsFromEngine(chaosEngine *litmuschaosv1alpha1.ChaosEngine) (float64, float64, float64, float64) {
	var total, passed, failed, waiting float64
	expStatusList := chaosEngine.Status.Experiments
	total = float64(len(expStatusList))

	for i := 0; i < len(expStatusList); i++ {
		verdict := strings.ToLower(expStatusList[i].Verdict)
		fmt.Println(verdict)
		switch verdict {
		case "pass":
			passed++
			delete(experimentStatusMap, fmt.Sprintf("%s-%s", chaosEngine.Name, chaosEngine.Namespace))

		case "fail":
			failed++
			delete(experimentStatusMap, fmt.Sprintf("%s-%s", chaosEngine.Name, chaosEngine.Namespace))

		case "waiting":
			waiting++

		case "awaited":
			// Check the unique chaosresult name in hashmap.
			if experimentStatusMap[fmt.Sprintf("%s-%s", chaosEngine.Name, chaosEngine.Namespace)] == false {
				// Set the chaosresult name to true, if it's unique.
				experimentStatusMap[fmt.Sprintf("%s-%s", chaosEngine.Name, chaosEngine.Namespace)] = true
			}
		}
	}
	return total, passed, failed, waiting
}

func filterMonitoringEnabledEngines(engineList *litmuschaosv1alpha1.ChaosEngineList) *litmuschaosv1alpha1.ChaosEngineList {
	var filteredEngineList litmuschaosv1alpha1.ChaosEngineList
	for i := len(engineList.Items) - 1; i >= 0; i-- {
		// Condition to decide if current element has to be deleted:
		if engineList.Items[i].Spec.Monitoring {
			filteredEngineList.Items = append(filteredEngineList.Items, engineList.Items[i])
		}
	}
	return &filteredEngineList
}

func putAwsMetricData(sess *session.Session, metricName string, unit string, value float64) error {
	// Create new Amazon CloudWatch client
	// snippet-start:[cloudwatch.go.create_custom_metric.call]
	dimension1 := "ClusterName"
	dimension2 := "Service"
	svc := cloudwatch.New(sess)
	namespace := os.Getenv("AWS_CLOUDWATCH_METRIC_NAMESPACE")
	clusterName := os.Getenv("CLUSTER_NAME")
	serviceName := os.Getenv("APP_NAME")

	if namespace == "" || serviceName == "" || clusterName == "" {
		fmt.Println("You must supply a namespace, clusterName and serviceName values")
	}

	klog.V(0).Infof("Putting new AWS metric: Namespace %v, Metric %v", namespace, metricName)

	_, err := svc.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: &namespace,
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: &metricName,
				Unit:       &unit,
				Value:      &value,
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  &dimension1,
						Value: &clusterName,
					},
					&cloudwatch.Dimension{
						Name:  &dimension2,
						Value: &serviceName,
					},
				},
			},
		},
	})
	// snippet-end:[cloudwatch.go.create_custom_metric.call]
	if err != nil {
		klog.V(0).Infof("Error during putting metrics to CloudWatch: %v", err)
		return err
	}

	return nil
}
