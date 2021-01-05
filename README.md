# Litmus Chaos Exporter
[![BUILD STATUS](https://travis-ci.org/litmuschaos/chaos-exporter.svg?branch=master)](https://travis-ci.org/litmuschaos/chaos-exporter)
[![BCH compliance](https://bettercodehub.com/edge/badge/litmuschaos/chaos-exporter?branch=master)](https://bettercodehub.com/)
[![Go Report Card](https://goreportcard.com/badge/github.com/litmuschaos/chaos-exporter)](https://goreportcard.com/report/github.com/litmuschaos/chaos-exporter)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter?ref=badge_shield)

- This is a custom Prometheus and CloudWatch exporter to expose Litmus Chaos metrics. 
  To learn more about Litmus Chaos Experiments & the Litmus Chaos Operator, 
  visit this link: [Litmus Docs](https://docs.litmuschaos.io/) 

- Typically deployed along with the chaos-operator deployment, which, 
  in-turn is associated with all chaosresults in the cluster.

- Two types of metrics are exposed: 

  - NamespacedScoped: These metrics are derived from the all the chaosresults present inside `WATCH_NAMESPACE`. If `WATCH_NAMESPACE` is not defined then it derived metrics from all namespaces. It exposes total_passed_experiment, total_failed_experiment, total_awaited_experiment, experiment_run_count, experiment_installed_count metrices.

  - ExperimentScoped: Individual experiment run status. It exposes passed_experiment, failed_experiment, awaited_experiment, probe_success_percentage, startTime, endTime, totalDuration, chaosInjectTime metrices.

- The metrics are of type Gauge, w/ each of the status metrics mapped to a 
  numeric value(not-executed:0, fail:1, running:2, pass:3)

- The CloudWatch metrics are of type Count, w/ each of the status metrics mapped to a 
  numeric value(not-executed:0, fail:1, running:2, pass:3)

## Steps to build & deploy: 

### Running Litmus Chaos Experiments in order to generate metrics

- Follow the steps described [here](https://github.com/litmuschaos/chaos-operator/blob/master/deploy/README.md) 
  to start running litmus chaos experiments ans storing chaos results. The chaos custom resources are used by the 
  exporter to generate metrics. 
  
### Running Chaos Exporter on the local Machine 

- Run the exporter container (litmuschaos/chaos-exporter:ci) on host network. It is necessary to mount the kubeconfig
  & override entrypoint w/ `./exporter -kubeconfig <path>`

- Execute `curl 127.0.0.1:8080/metrics` to view metrics

### Running Chaos Exporter as a deployment on the Kubernetes Cluster

- Install the RBAC (serviceaccount, role, rolebinding) as per deploy/rbac.md

- Deploy the chaos-exporter.yaml 

- From a cluster node, execute `curl <exporter-service-ip>:8080/metrics` 

### Example Metrics

```
# HELP litmuschaos_awaited_experiments Total number of awaited experiments
# TYPE litmuschaos_awaited_experiments gauge
litmuschaos_awaited_experiments{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 0
# HELP litmuschaos_experiment_chaos_injected_time chaos injected time of the experiments
# TYPE litmuschaos_experiment_chaos_injected_time gauge
litmuschaos_experiment_chaos_injected_time{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 1.609783037e+09
# HELP litmuschaos_experiment_end_time end time of the experiments
# TYPE litmuschaos_experiment_end_time gauge
litmuschaos_experiment_end_time{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 1.609783055e+09
# HELP litmuschaos_experiment_start_time start time of the experiments
# TYPE litmuschaos_experiment_start_time gauge
litmuschaos_experiment_start_time{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 1.609783003e+09
# HELP litmuschaos_failed_experiments Total number of failed experiments
# TYPE litmuschaos_failed_experiments gauge
litmuschaos_failed_experiments{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 0
# HELP litmuschaos_overall_awaited_experiments Total number of awaited experiments
# TYPE litmuschaos_overall_awaited_experiments gauge
litmuschaos_overall_awaited_experiments{chaosresult_namespace=""} 0
# HELP litmuschaos_overall_experiments_installed_count Total number of experiments
# TYPE litmuschaos_overall_experiments_installed_count gauge
litmuschaos_overall_experiments_installed_count{chaosresult_namespace=""} 1
# HELP litmuschaos_overall_experiments_run_count Total experiments run
# TYPE litmuschaos_overall_experiments_run_count gauge
litmuschaos_overall_experiments_run_count{chaosresult_namespace=""} 4
# HELP litmuschaos_overall_failed_experiments Total number of failed experiments
# TYPE litmuschaos_overall_failed_experiments gauge
litmuschaos_overall_failed_experiments{chaosresult_namespace=""} 0
# HELP litmuschaos_overall_passed_experiments Total number of passed experiments
# TYPE litmuschaos_overall_passed_experiments gauge
litmuschaos_overall_passed_experiments{chaosresult_namespace=""} 4
# HELP litmuschaos_passed_experiments Total number of passed experiments
# TYPE litmuschaos_passed_experiments gauge
litmuschaos_passed_experiments{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 4
# HELP litmuschaos_probe_success_percentage ProbeSuccesPercentage for the experiments
# TYPE litmuschaos_probe_success_percentage gauge
litmuschaos_probe_success_percentage{chaosresult_name="engine-nginx-pod-delete",chaosresult_namespace="litmus"} 100
```


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter?ref=badge_large)
