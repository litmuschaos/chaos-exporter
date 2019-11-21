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
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/litmuschaos/chaos-exporter/pkg/version"
)

// Exporter continuously collects the chaos metrics for a given chaosengine
func Exporter(config *rest.Config, exporterSpec ExporterSpec) {
	k8sClientSet, litmusClientSet, err := generateClientSets(config)
	if err != nil {
		log.Error(err)
	}

	versions, err := getVersion(k8sClientSet, exporterSpec)
	if err != nil {
		log.Error(err)
	}
	// Register the fixed (count) chaos metrics
	registerFixedMetrics()
	spec := ExporterConfig{Spec: exporterSpec, version: versions}
	for {
		generateChaosMetrics(spec, litmusClientSet)
	}
}

func getVersion(clientSet *kubernetes.Clientset, exporterSpec ExporterSpec) (Version, error) {
	var v Version
	var err error
	v.KubernetesVersion, err = version.GetKubernetesVersion(clientSet)
	if err != nil {
		return v, fmt.Errorf("unable to get Kubernetes Version %s: ", err)
	}
	// This function gets the openebs version
	v.OpenebsVersion, err = version.GetOpenebsVersion(clientSet, exporterSpec.OpenebsNamespace)
	if err != nil {
		return v, fmt.Errorf("unable to get OpenEBS Version %s: ", err)
	}
	return v, nil
}

// generateClientSets will generate clientSet for kubernetes and litmus
func generateClientSets(config *rest.Config) (*kubernetes.Clientset, *clientV1alpha1.Clientset, error) {
	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate kubernetes clientSet %s: ", err)
	}
	litmusClientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to generate litmus clientSet %s: ", err)
	}
	return k8sClientSet, litmusClientSet, nil
}

// contains checks if the a string is already part of a list of strings
func contains(l []string, e string) bool {
	for _, i := range l {
		if i == e {
			return true
		}
	}
	return false
}

func generateChaosMetrics(exporterConfig ExporterConfig, litmusClientSet *clientV1alpha1.Clientset) {

	// Get the chaos metrics for the specified chaosengine
	expTotal, passTotal, failTotal, expMap, err := GetLitmusChaosMetrics(litmusClientSet, exporterConfig.Spec)
	if err != nil {
		log.Error("Unable to get metrics: ", err.Error())
	}

	// Define, register & set the dynamically obtained chaos metrics (experiment state)
	chaosMetricsSpec := ChaosMetricsSpec{ExpTotal: expTotal, PassTotal: passTotal, FailTotal: failTotal, ExperimentList: expMap}
	defineChaosMetrics(chaosMetricsSpec, exporterConfig)
	time.Sleep(1000 * time.Millisecond)
}

func defineChaosMetrics(chaosMetricsSpec ChaosMetricsSpec, exporterConfig ExporterConfig) {
	for index, verdict := range chaosMetricsSpec.ExperimentList {
		sanitizedExpName, tmpExp := generatePrometheusGaugeVec(index)
		if contains(registeredResultMetrics, sanitizedExpName) {
			prometheus.Unregister(tmpExp)
			prometheus.MustRegister(tmpExp)
			tmpExp.WithLabelValues(exporterConfig.Spec.AppUUID, exporterConfig.Spec.ChaosEngine, exporterConfig.version.KubernetesVersion, exporterConfig.version.OpenebsVersion).Set(verdict)
		} else {
			prometheus.MustRegister(tmpExp)
			tmpExp.WithLabelValues(exporterConfig.Spec.AppUUID, exporterConfig.Spec.ChaosEngine, exporterConfig.version.KubernetesVersion, exporterConfig.version.OpenebsVersion).Set(verdict)
			registeredResultMetrics = append(registeredResultMetrics, sanitizedExpName)
		}

		// Set the fixed chaos metrics
		setFixedChaosMetrics(chaosMetricsSpec, exporterConfig)
	}
}

func generatePrometheusGaugeVec(index string) (string, *prometheus.GaugeVec) {
	sanitizedExpName := strings.Replace(index, "-", "_", -1)
	var (
		tmpExp = prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: "c", Subsystem: "exp", Name: sanitizedExpName, Help: ""},
			[]string{"app_uid", "engine_name", "kubernetes_version", "openebs_version"},
		)
	)
	return sanitizedExpName, tmpExp
}

func setFixedChaosMetrics(chaosMetricsSpec ChaosMetricsSpec, exporterConfig ExporterConfig) {
	ExperimentsTotal.WithLabelValues(exporterConfig.Spec.AppUUID, exporterConfig.Spec.ChaosEngine, exporterConfig.version.KubernetesVersion, exporterConfig.version.OpenebsVersion).Set(chaosMetricsSpec.ExpTotal)
	PassedExperiments.WithLabelValues(exporterConfig.Spec.AppUUID, exporterConfig.Spec.ChaosEngine, exporterConfig.version.KubernetesVersion, exporterConfig.version.OpenebsVersion).Set(chaosMetricsSpec.PassTotal)
	FailedExperiments.WithLabelValues(exporterConfig.Spec.AppUUID, exporterConfig.Spec.ChaosEngine, exporterConfig.version.KubernetesVersion, exporterConfig.version.OpenebsVersion).Set(chaosMetricsSpec.FailTotal)
}

func registerFixedMetrics() {
	prometheus.MustRegister(ExperimentsTotal)
	prometheus.MustRegister(PassedExperiments)
	prometheus.MustRegister(FailedExperiments)
}
