# Litmus Chaos Monitor

- This is a custom prometheus exporter to expose Litmus Chaos metrics. 
  To learn more about Litmus Chaos Experiments & the Litmus Chaos Operator, 
  visit this link: [Litmus Docs](https://docs.litmuschaos.io/) 

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

#### Pre-requisites:

- A Working Local Kubernetes Cluster (Eg: Minikube or Vagrant)
  - Set each of these Custom Resource Definition in your Kubernetes Cluster
  - For ChaosEngine : https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosengine_crd.yaml
  - For ChaosResult : https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosresults_crd.yaml
  - For ChaosExperiment: https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosexperiment_crd.yaml
  For information on these Custom Resources, please check this link : https://docs.litmuschaos.io/docs/next/co-components.html
- Kube-config path of your local Kubernetes Cluster
- `$GOPATH` set to your working directory

### Further Steps: 

The following steps are required to create sample chaos-related custom resources in order to visualize the metrics gathered by the chaos exporter

- Clone this repo into your $GOPATH/litmuschaos"
  `git clone https://github.com/litmuschaos/chaos-exporter`
- Now, start your Local Cluster, (this guide helps in `minikube` but can be used for other offline clusters as well)
- Create Kubernetes CR's(Custom Resources) for the litmus operator, link down below:
- Now, as you have created the CustomResourceDefinition, Now it time to run some chaos experiments with the help of chaos operator(github.com/litmuschaos/chaos-operator):
- Try to run chaos-monitor with the chaos-operator, or in other namespace, but will a ClusterRole similar to that of `https://github.com/litmuschaos/chaos-operator/blob/master/deploy/rbac.yaml`
- Run the command `make build` in the root directory.
- Find your kube-config file for your local cluster.
  - For minikube it is located in the directory `/home/user_name/.kube/config`, keep this path handy with you
- After building the file execute this command `sudo ./main -kubeconfig=path_for_the_kubeconfig`
- Execute `curl 127.0.0.1:8080/metrics | less` to view metrics

### On Kubernetes Cluster

- Install the RBAC (serviceaccount, role, rolebinding) as per deploy/rbac.md

- Deploy the chaos-exporter.yaml 

- From a cluster node, execute `curl <exporter-service-ip>:8080/metrics` 

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

