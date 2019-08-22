# Litmus Chaos Exporter

- This is a custom prometheus exporter to expose Litmus Chaos metrics. 
  To learn more about Litmus Chaos Experiments & the Litmus Chaos Operator, 
  visit this link: [Litmus Docs](https://docs.litmuschaos.io/) 

- The exporter is tied to a Chaosengine custom resource, which, 
  in-turn is associated with a given application deployment.

- The exporter is typically deployed as a to to the Litmus Experiment
  Runner container in the engine-runner pod, but can be launched as a
  separate deployment as well. 

- Two types of metrics are exposed: 

  - Fixed: TotalExperimentCount, TotalPassedTests, TotalFailedTests which are derived 
    from the ChaosEngine specification upfront

  - Dymanic: Individual Experiment Run Status. The list of experiments may 
    vary across ChaosEngines (or newer tests may be patched into it. 
    The exporter reports experiment status as per list in the chaosengine

- The metrics are of type Gauge, w/ each of the status metrics mapped to a 
  numeric value(not-executed:0, running:1, fail:2, pass:3)

- The metrics carry the application_uuid as label (this has to be passed as ENV)
## Steps to build & deploy: 

### Local Machine 

#### Pre-requisites:

- A Working Local Kubernetes Cluster (Eg: Minikube or Vagrant)
  - Set each of these Custom Resource Definition in your Kubernetes Cluster
  - For ChaosEngine : https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosengine_crd.yaml
  - For ChaosResult : https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosresults_crd.yaml
  - For ChaosExperiment: https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosexperiment_crd.yaml
  For information on these Custom Resources, please check this link : https://docs.litmuschaos.io/docs/next/co-components.html
- Kube-config path of your local Kubernetes Cluster.
- `$GOPATH` set to your working directory.
- Set an `APP_UUID` parameter with a valid string in the environment variables,as `APP_UUID` is derived from the app to be added as a metric label for Prometheus Exporter, as same for the ChaosEngine.

### Further Steps: 

The following steps are required to create sample chaos-related custom resources in order to visualize the metrics gathered by the chaos exporter.

- Clone this repo into your $GOPATH/litmuschaos"
  `git clone https://github.com/litmuschaos/chaos-exporter`
- Now, start your Local Cluster, (this guide helps in `minikube` but can be used for other offline clusters as well)
- Create Kubernetes CR's(Custom Resources) for the litmus operator, link down below:
  - For ChaosEngine : https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosengine_crd.yaml
  - For ChaosResult : https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosresults_crd.yaml
  - For ChaosExperiment: https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosexperiment_crd.yaml
- Now, as you have created the CustomResourceDefinition, Now it time to create the CustomResources for these definition above:
  - For Default ChaosEngine : https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosengine.yaml
    NOTE THAT THIS CHAOSENGINE COMES WITH A DEFAULT NAME ASSIGNED WITH IT WHICH IS : `engine-nginx`  you would need this afterwards
  - For Default ChaosResult : https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosresult.yaml
  - For the Default Experiments (Pod Delete Experiment) : https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosexperiment.yaml
- As you have created the ChaosEngine, make another environment variable namely `CHAOSENGINE` with the name of the ChaosEngine, which is needed by the Promethus Exporter as a metric.
- Run the command `make build` in the root directory.
- Find your kube-config file for your local cluster.
  - For minikube it is located in the directory `/home/user_name/.kube/config`, keep this path handy with you as it will needed at the time of execution of the chaos-exporter.
- After building the file execute this command `sudo ./main -kubeconfig=path_for_the_kubeconfig`
- Execute `curl 127.0.0.1:8080/metrics | less` to view metrics

### On Kubernetes Cluster

- Install the RBAC (serviceaccount, role, rolebinding) as per deploy/rbac.md

- Deploy the chaos-exporter.yaml 

- From a cluster node, execute `curl <exporter-service-ip>:8080/metrics` 

### Example Metrics

```
c_engine_experiment_count{app_uid="3f2092f8-6400-11e9-905f-42010a800131"} 2
# HELP c_engine_failed_experiments Total number of failed experiments
# TYPE c_engine_failed_experiments gauge
c_engine_failed_experiments{app_uid="3f2092f8-6400-11e9-905f-42010a800131"} 1
# HELP c_engine_passed_experiments Total number of passed experiments
# TYPE c_engine_passed_experiments gauge
c_engine_passed_experiments{app_uid="3f2092f8-6400-11e9-905f-42010a800131"} 1
# HELP c_exp_engine_nginx_container_kill 
# TYPE c_exp_engine_nginx_container_kill gauge
c_exp_engine_nginx_container_kill{app_uid="3f2092f8-6400-11e9-905f-42010a800131"} 2
# HELP c_exp_engine_nginx_pod_failure 
# TYPE c_exp_engine_nginx_pod_failure gauge
c_exp_engine_nginx_pod_failure{app_uid="3f2092f8-6400-11e9-905f-42010a800131"} 3
```

