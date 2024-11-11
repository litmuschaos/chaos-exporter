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
	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"github.com/litmuschaos/chaos-exporter/pkg/log"
	litmuschaosv1alpha1 "github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/util/workqueue"
)

// Exporter continuously collects the chaos metrics for a given chaosengine
func Exporter(clientSet clients.ClientSets, wq workqueue.RateLimitingInterface) {
	log.Info("Started creating Metrics")
	// Register the fixed (count) chaos metrics
	log.Info("Registering Fixed Metrics")

	r := MetricesCollecter{
		ResultCollector: &ResultDetails{},
	}
	//gaugeMetrics := GaugeMetrics{}
	overallChaosResults := []*litmuschaosv1alpha1.ChaosResult{}

	r.GaugeMetrics.InitializeGaugeMetrics().
		RegisterFixedMetrics()

	monitoringEnabled := MonitoringEnabled{
		IsChaosResultsAvailable: true,
		IsChaosEnginesAvailable: true,
	}

	// refresh metrics whenever there's a change in chaosengine or chaosresult
	// or every informer resync duration, whichever is earlier
	for _, done := wq.Get(); !done; _, done = wq.Get() {
		needRequeue, err := r.GetLitmusChaosMetrics(clientSet, &overallChaosResults, &monitoringEnabled)
		if err != nil {
			log.Errorf("err: %v", err)
		}
		wq.Done(clients.ProcessKey)
		// Add after
		if needRequeue != nil {
			wq.AddAfter(clients.ProcessKey, *needRequeue)
		}
	}
}

// RegisterFixedMetrics register the prometheus metrics
func (gaugeMetrics *GaugeMetrics) RegisterFixedMetrics() {
	prometheus.MustRegister(gaugeMetrics.ResultPassedExperiments)
	prometheus.MustRegister(gaugeMetrics.ResultFailedExperiments)
	prometheus.MustRegister(gaugeMetrics.ResultAwaitedExperiments)
	prometheus.MustRegister(gaugeMetrics.ResultProbeSuccessPercentage)
	prometheus.MustRegister(gaugeMetrics.ResultVerdict)
	prometheus.MustRegister(gaugeMetrics.ExperimentStartTime)
	prometheus.MustRegister(gaugeMetrics.ExperimentEndTime)
	prometheus.MustRegister(gaugeMetrics.ExperimentChaosInjectedTime)
	prometheus.MustRegister(gaugeMetrics.ExperimentTotalDuration)
	prometheus.MustRegister(gaugeMetrics.ClusterScopedTotalPassedExperiments)
	prometheus.MustRegister(gaugeMetrics.ClusterScopedTotalFailedExperiments)
	prometheus.MustRegister(gaugeMetrics.ClusterScopedTotalAwaitedExperiments)
	prometheus.MustRegister(gaugeMetrics.ClusterScopedExperimentsRunCount)
	prometheus.MustRegister(gaugeMetrics.ClusterScopedExperimentsInstalledCount)
	prometheus.MustRegister(gaugeMetrics.NamespaceScopedTotalPassedExperiments)
	prometheus.MustRegister(gaugeMetrics.NamespaceScopedTotalFailedExperiments)
	prometheus.MustRegister(gaugeMetrics.NamespaceScopedTotalAwaitedExperiments)
	prometheus.MustRegister(gaugeMetrics.NamespaceScopedExperimentsRunCount)
	prometheus.MustRegister(gaugeMetrics.NamespaceScopedExperimentsInstalledCount)
}
