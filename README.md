# Litmus Chaos Exporter
[![BUILD STATUS](https://travis-ci.org/litmuschaos/chaos-exporter.svg?branch=master)](https://travis-ci.org/litmuschaos/chaos-exporter)
[![BCH compliance](https://bettercodehub.com/edge/badge/litmuschaos/chaos-exporter?branch=master)](https://bettercodehub.com/)
[![Go Report Card](https://goreportcard.com/badge/github.com/litmuschaos/chaos-exporter)](https://goreportcard.com/report/github.com/litmuschaos/chaos-exporter)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter?ref=badge_shield)

- This is a custom prometheus exporter to expose Litmus Chaos metrics. 
  To learn more about Litmus Chaos Experiments & the Litmus Chaos Operator, 
  visit this link: [Litmus Docs](https://docs.litmuschaos.io/) 

- The exporter is tied to a Chaos-operator deployment resource, which, 
  in-turn is associated with all chaosengines in the cluster.

- The exporter is typically deployed as a sidecar to the Litmus Chaos Operator Deployment, but can be launched as a
  separate deployment as well. 

- Two types of metrics are exposed: 

  - Fixed: ClusterTotalExperimentCount, ClusterTotalPassedCount, ClusterTotalFailedCount, EngineTotalExperimentCount, EnginePassedExperimentCount, EngineFailedExperimentCount, EngineWaitingExperimentCount  which are derived 
    from the ChaosEngine specification upfront

  - Dymanic: Individual Experiment Run Status. The list of experiments may 
    vary across ChaosEngines (or newer tests may be patched into it. 
    The exporter reports experiment status as per list in the chaosengine

- The metrics are of type Gauge, w/ each of the status metrics mapped to a 
  numeric value(not-executed:0, fail:1, running:2, pass:3)

## Steps to build & deploy: 

### Local Machine 

- Set the application deployment (assuming a live K8s cluster w/ app and chaos-operator deployed), and give it ClusterRole authorisation similar to chaos-operator deployment 

- Run the exporter container (litmuschaos/chaos-exporter:ci) on host network. It is necessary to mount the kubeconfig
  & override entrypoint w/ `./exporter -kubeconfig <path>`

- Execute `curl 127.0.0.1:8080/metrics` to view metrics

### On Kubernetes Cluster

- Install the RBAC (serviceaccount, role, rolebinding) as per deploy/rbac.md

- Deploy the chaos-exporter.yaml 

- From a cluster node, execute `curl <exporter-service-ip>:8080/metrics`  or `curl <pod-ip>:8080.metrics`

### Example Metrics

```
# HELP c_exp_RunningExperiment Running Experiment with ChaosEngine Details
# TYPE c_exp_RunningExperiment gauge
c_exp_RunningExperiment{engine_name="engine3",engine_namespace="litmus",experiment_name="pod-delete",result_name="engine3-pod-delete"} 1
# HELP chaosEngine_engine_engine_awaited_experiments Total number of waiting experiments by the chaos engine
# TYPE chaosEngine_engine_engine_awaited_experiments gauge
chaosEngine_engine_engine_awaited_experiments{engine_name="engine3",engine_namespace="litmus"} 1
# HELP chaosEngine_engine_engine_experiment_count Total number of experiments executed by the chaos engine
# TYPE chaosEngine_engine_engine_experiment_count gauge
chaosEngine_engine_engine_experiment_count{engine_name="engine3",engine_namespace="litmus"} 2
# HELP chaosEngine_engine_engine_failed_experiments Total number of failed experiments by the chaos engine
# TYPE chaosEngine_engine_engine_failed_experiments gauge
chaosEngine_engine_engine_failed_experiments{engine_name="engine3",engine_namespace="litmus"} 0
# HELP chaosEngine_engine_engine_passed_experiments Total number of passed experiments by the chaos engine
# TYPE chaosEngine_engine_engine_passed_experiments gauge
chaosEngine_engine_engine_passed_experiments{engine_name="engine3",engine_namespace="litmus"} 0
# HELP cluster_overall_cluster_experiment_count Total number of experiments executed in the Cluster
# TYPE cluster_overall_cluster_experiment_count gauge
cluster_overall_cluster_experiment_count 2
# HELP cluster_overall_cluster_failed_experiments Total number of failed experiments in the Cluster
# TYPE cluster_overall_cluster_failed_experiments gauge
cluster_overall_cluster_failed_experiments 0
# HELP cluster_overall_cluster_passed_experiments Total number of passed experiments in the Cluster
# TYPE cluster_overall_cluster_passed_experiments gauge
cluster_overall_cluster_passed_experiments 0
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 1.1785e-05
go_gc_duration_seconds{quantile="0.25"} 1.1785e-05
go_gc_duration_seconds{quantile="0.5"} 1.4254e-05
go_gc_duration_seconds{quantile="0.75"} 1.9929e-05
go_gc_duration_seconds{quantile="1"} 1.9929e-05
...
```


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter?ref=badge_large)