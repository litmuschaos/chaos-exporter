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
	"time"

	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"github.com/litmuschaos/chaos-exporter/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter continuously collects the chaos metrics for a given chaosengine
func Exporter(clients clients.ClientSets) {
	log.Info("Started creating Metrics")
	// Register the fixed (count) chaos metrics
	log.Info("Registering Fixed Metrics")
	registerFixedMetrics()

	for {
		if err := GetLitmusChaosMetrics(clients); err != nil {
			log.Errorf("err: %v", err)
		}
		time.Sleep(1000 * time.Millisecond)
	}
}

func registerFixedMetrics() {
	prometheus.MustRegister(ResultPassedExperiments)
	prometheus.MustRegister(ResultFailedExperiments)
	prometheus.MustRegister(ResultAwaitedExperiments)
	prometheus.MustRegister(ResultProbeSuccessPercentage)
	prometheus.MustRegister(ExperimentStartTime)
	prometheus.MustRegister(ExperimentEndTime)
	prometheus.MustRegister(ExperimentChaosInjectedTime)
	prometheus.MustRegister(TotalPassedExperiments)
	prometheus.MustRegister(TotalFailedExperiments)
	prometheus.MustRegister(TotalAwaitedExperiments)
	prometheus.MustRegister(ExperimentsRunCount)
	prometheus.MustRegister(ExperimentsInstalledCount)
}
