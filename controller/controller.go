package controller

import (
	log "github.com/Sirupsen/logrus"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/rest"
	"strings"
	"time"
	)

// Exporter continuously collects the chaos metrics for a given chaosengine
func Exporter(config *rest.Config, exporterSpec ExporterSpec) {

	for {
		// Get the chaos metrics for the specified chaosengine
		expTotal, passTotal, failTotal, expMap, err := GetLitmusChaosMetrics(config, exporterSpec)
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
				tmpExp.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, exporterSpec.KubernetesVersion, exporterSpec.OpenebsVersion).Set(verdict)
			} else {
				prometheus.MustRegister(tmpExp)
				tmpExp.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, exporterSpec.KubernetesVersion, exporterSpec.OpenebsVersion).Set(verdict)
				registeredResultMetrics = append(registeredResultMetrics, sanitizedExpName)
			}

			// Set the fixed chaos metrics
			ExperimentsTotal.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, exporterSpec.KubernetesVersion, exporterSpec.OpenebsVersion).Set(expTotal)
			PassedExperiments.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, exporterSpec.KubernetesVersion, exporterSpec.OpenebsVersion).Set(passTotal)
			FailedExperiments.WithLabelValues(exporterSpec.AppUUID, exporterSpec.ChaosEngine, exporterSpec.KubernetesVersion, exporterSpec.OpenebsVersion).Set(failTotal)
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