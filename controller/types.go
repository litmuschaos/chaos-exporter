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

	TotalPassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "overall",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments",
	},
		[]string{"chaosresult_namespace"},
	)

	TotalFailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "overall",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments",
	},
		[]string{"chaosresult_namespace"},
	)

	TotalAwaitedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "overall",
		Name:      "awaited_experiments",
		Help:      "Total number of awaited experiments",
	},
		[]string{"chaosresult_namespace"},
	)

	ExperimentsRunCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "overall",
		Name:      "experiments_run_count",
		Help:      "Total experiments run",
	},
		[]string{"chaosresult_namespace"},
	)

	ExperimentsInstalledCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "litmuschaos",
		Subsystem: "overall",
		Name:      "experiments_installed_count",
		Help:      "Total number of experiments",
	},
		[]string{"chaosresult_namespace"},
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
