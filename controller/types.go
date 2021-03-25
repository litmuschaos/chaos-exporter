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
	"github.com/prometheus/client_golang/prometheus"
	clientTypes "k8s.io/apimachinery/pkg/types"
)

//EngineLabelKey is key for ChaosEngineLabel
//matchLabelUID stores UID as a key and EngineLabelKey as a value
var (
	EngineLabelKey = "context"
	matchLabelUID  = map[string]string{}
)

// ChaosResultDetails contains chaosresult details
type ChaosResultDetails struct {
	Name                  string
	UID                   clientTypes.UID
	Namespace             string
	PassedExperiments     float64
	FailedExperiments     float64
	AwaitedExperiments    float64
	ProbeSuccesPercentage float64
	StartTime             float64
	EndTime               float64
	InjectionTime         float64
	TotalDuration         float64
	ChaosEngineName       string
	ChaosEngineLabel      string
}

// NamespacedScopeMetrics contains metrics for the chaos namespace
type NamespacedScopeMetrics struct {
	PassedExperiments         float64
	FailedExperiments         float64
	AwaitedExperiments        float64
	ExperimentRunCount        float64
	ExperimentsInstalledCount float64
}

// AWSConfig contains aws configuration details
type AWSConfig struct {
	Namespace   string
	ClusterName string
	Service     string
}

// InitializeGaugeMetrics ...
func (gaugeMetrics *GaugeMetrics) InitializeGaugeMetrics() *GaugeMetrics {
	gaugeMetrics.ResultPassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name", "chaosengine_name", "chaosengine_context"},
	)

	gaugeMetrics.ResultFailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name", "chaosengine_name", "chaosengine_context"},
	)

	gaugeMetrics.ResultAwaitedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "awaited_experiments",
		Help:      "Total number of awaited experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name", "chaosengine_name", "chaosengine_context"},
	)

	gaugeMetrics.ResultProbeSuccessPercentage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "probe_success_percentage",
		Help:      "ProbeSuccesPercentage for the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name", "chaosengine_name", "chaosengine_context"},
	)

	gaugeMetrics.ExperimentStartTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "experiment_start_time",
		Help:      "start time of the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name", "chaosengine_name", "chaosengine_context"},
	)

	gaugeMetrics.ExperimentEndTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "experiment_end_time",
		Help:      "end time of the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name", "chaosengine_name", "chaosengine_context"},
	)

	gaugeMetrics.ExperimentChaosInjectedTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "experiment_chaos_injected_time",
		Help:      "chaos injected time of the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name", "chaosengine_name", "chaosengine_context"},
	)

	gaugeMetrics.ExperimentTotalDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "experiment_total_duration",
		Help:      "total duration of the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name", "chaosengine_name", "chaosengine_context"},
	)

	gaugeMetrics.NamespaceScopedTotalPassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	gaugeMetrics.NamespaceScopedTotalFailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	gaugeMetrics.NamespaceScopedTotalAwaitedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "awaited_experiments",
		Help:      "Total number of awaited experiments in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	gaugeMetrics.NamespaceScopedExperimentsRunCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "experiments_run_count",
		Help:      "Total experiments run in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	gaugeMetrics.NamespaceScopedExperimentsInstalledCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "experiments_installed_count",
		Help:      "Total number of experiments in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	gaugeMetrics.ClusterScopedTotalPassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments in all namespaces",
	},
		[]string{},
	)

	gaugeMetrics.ClusterScopedTotalFailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments in all namespaces",
	},
		[]string{},
	)

	gaugeMetrics.ClusterScopedTotalAwaitedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "awaited_experiments",
		Help:      "Total number of awaited experiments in all namespaces",
	},
		[]string{},
	)

	gaugeMetrics.ClusterScopedExperimentsRunCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "experiments_run_count",
		Help:      "Total experiments run in all namespaces",
	},
		[]string{},
	)

	gaugeMetrics.ClusterScopedExperimentsInstalledCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "experiments_installed_count",
		Help:      "Total number of experiments in all namespaces",
	},
		[]string{},
	)
	return gaugeMetrics
}

// GaugeMetrics contains all the metrics definition
type GaugeMetrics struct {
	ResultPassedExperiments                  *prometheus.GaugeVec
	ResultFailedExperiments                  *prometheus.GaugeVec
	ResultAwaitedExperiments                 *prometheus.GaugeVec
	ResultProbeSuccessPercentage             *prometheus.GaugeVec
	ExperimentStartTime                      *prometheus.GaugeVec
	ExperimentEndTime                        *prometheus.GaugeVec
	ExperimentTotalDuration                  *prometheus.GaugeVec
	ExperimentChaosInjectedTime              *prometheus.GaugeVec
	NamespaceScopedTotalPassedExperiments    *prometheus.GaugeVec
	NamespaceScopedTotalFailedExperiments    *prometheus.GaugeVec
	NamespaceScopedTotalAwaitedExperiments   *prometheus.GaugeVec
	NamespaceScopedExperimentsInstalledCount *prometheus.GaugeVec
	NamespaceScopedExperimentsRunCount       *prometheus.GaugeVec
	ClusterScopedTotalPassedExperiments      *prometheus.GaugeVec
	ClusterScopedTotalFailedExperiments      *prometheus.GaugeVec
	ClusterScopedTotalAwaitedExperiments     *prometheus.GaugeVec
	ClusterScopedExperimentsInstalledCount   *prometheus.GaugeVec
	ClusterScopedExperimentsRunCount         *prometheus.GaugeVec
}

// MonitoringEnabled contains existance/availability of chaosengines and chaosresults
type MonitoringEnabled struct {
	IsChaosResultsAvailable bool
	IsChaosEnginesAvailable bool
}
