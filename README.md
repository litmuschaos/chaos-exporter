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

  - AggregateMetrics: These metrics are derived from the all the chaosresults present inside `WATCH_NAMESPACE`. If `WATCH_NAMESPACE` is not defined then it derived metrics from all namespaces. It exposes total_passed_experiment, total_failed_experiment, total_awaited_experiment, experiment_run_count, experiment_installed_count metrices.

  - ExperimentScoped: Individual experiment run status. It exposes passed_experiment, failed_experiment, awaited_experiment, probe_success_percentage, startTime, endTime, totalDuration, chaosInjectTime metrices.

### ExperimentScoped Metrics

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_passed_experiments</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains total number of passed experiments</td>
</tr>
<tr>
  <th>Source</th>
  <td>ChaosResult</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_passed_experiments{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 1</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_passed_experiments</code> contains the commulative sum of passed runs for the given ChaosResult.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_failed_experiments</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains total number of failed experiments</td>
</tr>
<tr>
  <th>Source</th>
  <td>ChaosResult</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_failed_experiments{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 0</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_failed_experiments</code> contains the commulative sum of failed runs for the given ChaosResult.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_awaited_experiments</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains total number of awaited experiments</td>
</tr>
<tr>
  <th>Source</th>
  <td>ChaosResult</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_awaited_experiments{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 1</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_awaited_experiments</code> denotes the queued experiments for each ChaosResult. It contains the value as 1 if the ChaosResult's verdict is Awaited otherwise it's value is 0.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_probe_success_percentage</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the ProbeSuccessPercentage for the experiment</td>
</tr>
<tr>
  <th>Source</th>
  <td>ChaosResult</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_probe_success_percentage{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 100</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_probe_success_percentage</code> defines the percentage of passed probes out of total probes defined inside the ChaosEngine.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_experiment_start_time</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the start time of the experiment</td>
</tr>
<tr>
  <th>Source</th>
  <td><code>ExperimentDependencyCheck</code> event inside the ChaosEngine</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_experiment_start_time{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 1.618425155e+09</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_experiment_start_time</code> denotes the start time of the experiment, which calculated based on the ExperimentDependencyCheck event(created by the chaos-runner just before launching experiment pod).</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_experiment_end_time</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the end time of the experiment</td>
</tr>
<tr>
  <th>Source</th>
  <td><code>Summary</code> event inside the ChaosEngine</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_experiment_end_time{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 1.618425219e+09</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_experiment_end_time</code> denotes the end time of the experiment, which calculated based on the Summary event(created by experiment pod in the end of experiment).</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_experiment_chaos_injected_time</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the chaos injection time of the experiment</td>
</tr>
<tr>
  <th>Source</th>
  <td><code>ChaosInject</code> event inside the ChaosEngine</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_experiment_chaos_injected_time{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 1.618425199e+09</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_experiment_chaos_injected_time</code> defines the time duration when chaos is actually injected, which calculated based on the ChaosInject event(created by the experiment/helper pod just before chaos injection).</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_experiment_total_duration</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total chaos duration of the experiment</td>
</tr>
<tr>
  <th>Source</th>
  <td>It is time difference b/w startTime and endTime</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_experiment_total_duration{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 64</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_experiment_total_duration</code> defines the total chaos duration of the experiment. It is time interval betweeen start time and the end time.</td>
</tr>
</table>
<hr>

### NamespacedScoped Metrics

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_namespace_scoped_passed_experiments"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total passed experiments count in the WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Source</th>
  <td>Aggregated sum of all the <code>litmuschaos_passed_experiments</code> metrics derived from the ChaosResult present inside WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_namespace_scoped_passed_experiments 2</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_namespace_scoped_passed_experiments</code> defines the total number of passed experiments in the WATCH_NAMESPACE. It is the summation of <code>litmuschaos_passed_experiments</code> metrics for every ChaosResult present inside the WATCH_NAMESPACE.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_namespace_scoped_failed_experiments"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total failed experiments count in the WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Source</th>
  <td>Aggregated sum of all the <code>litmuschaos_failed_experiments</code> metrics derived from the ChaosResult present inside WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_namespace_scoped_failed_experiments 0</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_namespace_scoped_failed_experiments</code> defines the total number of failed experiments in the WATCH_NAMESPACE. It is the summation of <code>litmuschaos_failed_experiments</code> metrics for every ChaosResult present inside the WATCH_NAMESPACE.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_namespace_scoped_awaited_experiments"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total awaited experiments count in the WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Source</th>
  <td>Aggregated sum of all the <code>litmuschaos_awaited_experiments</code> metrics derived from the ChaosResult present inside WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_namespace_scoped_awaited_experiments 0</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_namespace_scoped_awaited_experiments</code> defines the total number of awaited/queued experiments in the WATCH_NAMESPACE. It is the summation of <code>litmuschaos_awaited_experiments</code> metrics for every ChaosResult present inside the WATCH_NAMESPACE.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_namespace_scoped_experiments_run_count"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total experiments run count in the WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Source</th>
  <td>Aggregated sum of all the experiments runs in the WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_namespace_scoped_experiments_run_count 2</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_namespace_scoped_experiments_run_count</code> defines the total experiment runs in the WATCH_NAMESPACE. It is summation of  <code>litmuschaos_passed_experiments + litmuschaos_failed_experiments + litmuschaos_awaited_experiments</code> for every ChaosResult present present inside the WATCH_NAMESPACE.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_namespace_scoped_experiments_installed_count"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total unique experiments installed/run in the WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Source</th>
  <td>It contains total unique experiments count in the WATCH_NAMESPACE</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_namespace_scoped_experiments_installed_count 1</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_namespace_scoped_experiments_installed_count</code> defines the total unique experiments installed/run in the WATCH_NAMESPACE. It is equal to the total number of ChaosResult present inside the WATCH_NAMESPACE.</td>
</tr>
</table>
<hr>

### ClusterScoped Metrics

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_cluster_scoped_passed_experiments"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total passed experiments count in all the namespaces</td>
</tr>
<tr>
  <th>Source</th>
  <td>Aggregated sum of all the <code>litmuschaos_passed_experiments</code> metrics derived from the ChaosResult present inside all the namespaces</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_cluster_scoped_passed_experiments 2</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_cluster_scoped_passed_experiments</code> defines the total number of passed experiments across the cluster. It is the summation of <code>litmuschaos_passed_experiments</code> metrics for every ChaosResult in all the namespaces.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_cluster_scoped_failed_experiments"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total failed experiments count in all the namespaces</td>
</tr>
<tr>
  <th>Source</th>
  <td>Aggregated sum of all the <code>litmuschaos_failed_experiments</code> metrics derived from the ChaosResult present inside all the namespaces</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_cluster_scoped_failed_experiments 0</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_cluster_scoped_failed_experiments</code> defines the total number of failed experiments across the cluster. It is the summation of <code>litmuschaos_failed_experiments</code> metrics for every ChaosResult in all the namespaces.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_cluster_scoped_awaited_experiments"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total awaited experiments count in all the namespaces</td>
</tr>
<tr>
  <th>Source</th>
  <td>Aggregated sum of all the <code>litmuschaos_awaited_experiments</code> metrics derived from the ChaosResult present inside all the namespaces</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_cluster_scoped_awaited_experiments 0</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_cluster_scoped_awaited_experiments</code> defines the total number of awaited/queued experiments across the cluster. It is the summation of <code>litmuschaos_awaited_experiments</code> metrics for every ChaosResult in all the namespaces.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_cluster_scoped_experiments_run_count"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total experiments run count in all the namespaces</td>
</tr>
<tr>
  <th>Source</th>
  <td>Aggregated sum of all the experiments runs in all the namespaces</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_cluster_scoped_experiments_run_count 2</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_cluster_scoped_experiments_run_count</code> defines the total experiment runs across the cluster. It is summation of  <code>litmuschaos_passed_experiments + litmuschaos_failed_experiments + litmuschaos_awaited_experiments</code> for every ChaosResult present inside all the namespaces.</td>
</tr>
</table>

<table>
<tr>
  <th>Metrics Name</th>
  <td><code>litmuschaos_cluster_scoped_experiments_installed_count"</code></td>
</tr>
<tr>
  <th>Description</th>
  <td>It contains the total unique experiments installed/run in all the namespaces</td>
</tr>
<tr>
  <th>Source</th>
  <td>It contains total unique experiments count in all the namespaces</td>
</tr>
<tr>
  <th>Sample Metrics</th>
  <td><code>litmuschaos_cluster_scoped_experiments_installed_count 1</code></td>
</tr>
<tr>
  <th>Notes</th>
  <td>The <code>litmuschaos_cluster_scoped_experiments_installed_count</code> defines the total unique experiments installed/run across the cluster. It is equal to the total number of ChaosResult present inside all the namespaces.</td>
</tr>
</table>

## Steps to build & deploy: 

### Running Litmus Chaos Experiments in order to generate metrics

- Follow the steps described [here](https://docs.litmuschaos.io/docs/getstarted/) 
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
litmuschaos_awaited_experiments{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 0
# HELP litmuschaos_cluster_scoped_awaited_experiments Total number of awaited experiments in all namespaces
# TYPE litmuschaos_cluster_scoped_awaited_experiments gauge
litmuschaos_cluster_scoped_awaited_experiments 0
# HELP litmuschaos_cluster_scoped_experiments_installed_count Total number of experiments in all namespaces
# TYPE litmuschaos_cluster_scoped_experiments_installed_count gauge
litmuschaos_cluster_scoped_experiments_installed_count 1
# HELP litmuschaos_cluster_scoped_experiments_run_count Total experiments run in all namespaces
# TYPE litmuschaos_cluster_scoped_experiments_run_count gauge
litmuschaos_cluster_scoped_experiments_run_count 2
# HELP litmuschaos_cluster_scoped_failed_experiments Total number of failed experiments in all namespaces
# TYPE litmuschaos_cluster_scoped_failed_experiments gauge
litmuschaos_cluster_scoped_failed_experiments 0
# HELP litmuschaos_cluster_scoped_passed_experiments Total number of passed experiments in all namespaces
# TYPE litmuschaos_cluster_scoped_passed_experiments gauge
litmuschaos_cluster_scoped_passed_experiments 2
# HELP litmuschaos_experiment_chaos_injected_time chaos injected time of the experiments
# TYPE litmuschaos_experiment_chaos_injected_time gauge
litmuschaos_experiment_chaos_injected_time{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 1.618426086e+09
# HELP litmuschaos_experiment_end_time end time of the experiments
# TYPE litmuschaos_experiment_end_time gauge
litmuschaos_experiment_end_time{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 1.618426108e+09
# HELP litmuschaos_experiment_start_time start time of the experiments
# TYPE litmuschaos_experiment_start_time gauge
litmuschaos_experiment_start_time{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 1.618426056e+09
# HELP litmuschaos_failed_experiments Total number of failed experiments
# TYPE litmuschaos_failed_experiments gauge
litmuschaos_failed_experiments{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 0
# HELP litmuschaos_passed_experiments Total number of passed experiments
# TYPE litmuschaos_passed_experiments gauge
litmuschaos_passed_experiments{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 2
# HELP litmuschaos_probe_success_percentage ProbeSuccesPercentage for the experiments
# TYPE litmuschaos_probe_success_percentage gauge
litmuschaos_probe_success_percentage{chaosengine_context="test",chaosengine_name="helloservice-pod-delete",chaosresult_name="helloservice-pod-delete-pod-delete",chaosresult_namespace="litmus"} 100
```


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Flitmuschaos%2Fchaos-exporter?ref=badge_large)
