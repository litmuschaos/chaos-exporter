package controller

import (
	"github.com/prometheus/client_golang/prometheus"
)

var registeredResultMetrics []string

// Declare the fixed chaos metrics. Dynamic (testStatus) metrics are defined in metrics()
var (
	ExperimentsTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "c",
		Subsystem: "engine",
		Name:      "experiment_count",
		Help:      "Total number of experiments executed by the chaos engine",
	},
		[]string{"app_uid", "engine_name", "kubernetes_version", "openebs_version"},
	)

	PassedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "c",
		Subsystem: "engine",
		Name:      "passed_experiments",
		Help:      "Total number of passed experiments",
	},
		[]string{"app_uid", "engine_name", "kubernetes_version", "openebs_version"},
	)

	FailedExperiments = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "c",
		Subsystem: "engine",
		Name:      "failed_experiments",
		Help:      "Total number of failed experiments",
	},
		[]string{"app_uid", "engine_name", "kubernetes_version", "openebs_version"},
	)
)

// ExporterSpec contains the exporter related specs
type ExporterSpec struct {
	ChaosEngine      string
	AppUUID          string
	AppNS            string
	OpenebsNamespace string
}

// Version contains the version related information
type Version struct {
	KubernetesVersion string
	OpenebsVersion    string
}

// ExporterConfig contains the config for exporter function
type ExporterConfig struct {
	Spec ExporterSpec
	version Version
}