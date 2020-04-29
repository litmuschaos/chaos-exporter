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
	EngineTotalExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "chaosengine",
		Subsystem: "",
		Name:      "total_experiments",
		Help:      "Total number of experiments executed by the chaos engine",
	},
		[]string{"engine_namespace", "engine_name"},
	)

	EnginePassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "chaosengine",
		Subsystem: "",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments by the chaos engine",
	},
		[]string{"engine_namespace", "engine_name"},
	)

	EngineFailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "chaosengine",
		Subsystem: "",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments by the chaos engine",
	},
		[]string{"engine_namespace", "engine_name"},
	)

	EngineWaitingExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "chaosengine",
		Subsystem: "",
		Name:      "waiting_experiments",
		Help:      "Total number of waiting experiments by the chaos engine",
	},
		[]string{"engine_namespace", "engine_name"},
	)

	ClusterTotalExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cluster",
		Subsystem: "overall",
		Name:      "experiment_count",
		Help:      "Total number of experiments executed in the Cluster",
	},
		[]string{},
	)

	ClusterPassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cluster",
		Subsystem: "overall",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments in the Cluster",
	},
		[]string{},
	)

	ClusterFailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "cluster",
		Subsystem: "overall",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments in the Cluster",
	},
		[]string{},
	)

	RunningExperiment = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: "cluster", Subsystem: "overall", Name: "RunningExperiment", Help: "Running Experiment with ChaosEngine Details"},
		[]string{"engine_namespace", "engine_name", "experiment_name", "result_name"},
	)
)

// ChaosMetricsSpec contains the specs related to chaos metrics
type ChaosMetricsSpec struct {
	ExpTotal   float64
	PassTotal  float64
	FailTotal  float64
	ResultList map[string]float64
}

// ChaosExpResult contains the structure of Chaos Result
type ChaosExpResult struct {
	TotalExpCount  float64
	TotalPassedExp float64
	TotalFailedExp float64
}

type ChaosEngineDetail struct {
	Name       string
	Namespace  string
	TotalExp   float64
	PassedExp  float64
	FailedExp  float64
	AwaitedExp float64
}
