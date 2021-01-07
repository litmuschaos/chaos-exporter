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
)

var registeredResultMetrics []string

// Declare the fixed chaos metrics. Dynamic (testStatus) metrics are defined in metrics()
var (
	ResultPassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name"},
	)

	ResultFailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name"},
	)

	ResultAwaitedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "awaited_experiments",
		Help:      "Total number of awaited experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name"},
	)

	ResultProbeSuccessPercentage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "probe_success_percentage",
		Help:      "ProbeSuccesPercentage for the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name"},
	)

	ExperimentStartTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "experiment_start_time",
		Help:      "start time of the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name"},
	)

	ExperimentEndTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "experiment_end_time",
		Help:      "end time of the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name"},
	)

	ExperimentChaosInjectedTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "experiment_chaos_injected_time",
		Help:      "chaos injected time of the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name"},
	)

	ExperimentTotalDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "",
		Name:      "experiment_total_duration",
		Help:      "total duration of the experiments",
	},
		[]string{"chaosresult_namespace", "chaosresult_name"},
	)

	NamespaceScopedTotalPassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	NamespaceScopedTotalFailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	NamespaceScopedTotalAwaitedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "awaited_experiments",
		Help:      "Total number of awaited experiments in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	NamespaceScopedExperimentsRunCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "experiments_run_count",
		Help:      "Total experiments run in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	NamespaceScopedExperimentsInstalledCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "namespace_scoped",
		Name:      "experiments_installed_count",
		Help:      "Total number of experiments in watch namespace",
	},
		[]string{"chaosresult_namespace"},
	)

	ClusterScopedTotalPassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments in all namespaces",
	},
		[]string{},
	)

	ClusterScopedTotalFailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments in all namespaces",
	},
		[]string{},
	)

	ClusterScopedTotalAwaitedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "awaited_experiments",
		Help:      "Total number of awaited experiments in all namespaces",
	},
		[]string{},
	)

	ClusterScopedExperimentsRunCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "experiments_run_count",
		Help:      "Total experiments run in all namespaces",
	},
		[]string{},
	)

	ClusterScopedExperimentsInstalledCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "cluster_scoped",
		Name:      "experiments_installed_count",
		Help:      "Total number of experiments in all namespaces",
	},
		[]string{},
	)
)

// ChaosResultDetails contains chaosresult details
type ChaosResultDetails struct {
	Name                  string
	Namespace             string
	PassedExperiments     float64
	FailedExperiments     float64
	AwaitedExperiments    float64
	ProbeSuccesPercentage float64
	StartTime             float64
	EndTime               float64
	InjectionTime         float64
	TotalDuration         float64
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
