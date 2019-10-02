package controller

import (
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/rest"
	clientV1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"

	"github.com/litmuschaos/chaos-exporter/pkg/version"
)

// Exporter continuously collects the chaos metrics for a given chaosengine
func Exporter(config *rest.Config, exporterSpec ExporterSpec) {
	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Info("Unable to create the kubernetes ClientSet")
	}
	litmusClientSet, err := clientV1alpha1.NewForConfig(config)
	if err != nil {
		log.Info("Unable to create the litmus ClientSet")
	}
	// Register the fixed (count) chaos metrics
	prometheus.MustRegister(ExperimentsTotal)
	prometheus.MustRegister(PassedExperiments)
	prometheus.MustRegister(FailedExperiments)

	// This function gets the kubernetes version
	kubernetesVersion, err := version.GetKubernetesVersion(k8sClientSet)
	if err != nil {
		log.Info("Unable to get Kubernetes Version : ", err)
	}
	// This function gets the openebs version
	openebsVersion, err := version.GetOpenebsVersion(k8sClientSet, exporterSpec.OpenebsNamespace)
	if err != nil {
		log.Info("Unable to get OpenEBS Version : ", err)
	}

	for {
		// Get the chaos metrics for the specified chaosengine
		expTotal, passTotal, failTotal, expMap, err := GetLitmusChaosMetrics(litmusClientSet, exporterSpec)
		if err != nil {
			log.Error("Unable to get metrics: ", err.Error())
		}

		// Define, register & set the dynamically obtained chaos metrics (experiment state)
		for index, verdict := range expMap {
			sanitizedExpName := strings.Replace(index, "-", "_", -1)
			var (
				tmpExp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Namespace: "c",
					Subsystem: "exp",
					Name:      sanitizedExpName,
					Help:      "",
				},
					[]string{"app_uid", "engine_name", "kubernetes_version", "openebs_version"},
				)
			)

			if contains(registeredResultMetrics, sanitizedExpName) {
				prometheus.Unregister(tmpExp)
				prometheus.MustRegister(tmpExp)
				tmpExp.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, kubernetesVersion, openebsVersion).Set(verdict)
			} else {
				prometheus.MustRegister(tmpExp)
				tmpExp.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, kubernetesVersion, openebsVersion).Set(verdict)
				registeredResultMetrics = append(registeredResultMetrics, sanitizedExpName)
			}

			// Set the fixed chaos metrics
			ExperimentsTotal.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, kubernetesVersion, openebsVersion).Set(expTotal)
			PassedExperiments.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, kubernetesVersion, openebsVersion).Set(passTotal)
			FailedExperiments.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, kubernetesVersion, openebsVersion).Set(failTotal)
		}

		time.Sleep(1000 * time.Millisecond)
	}
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
