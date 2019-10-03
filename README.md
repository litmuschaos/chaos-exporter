# Litmus Chaos Exporter
[![BCH compliance](https://bettercodehub.com/edge/badge/litmuschaos/chaos-exporter?branch=master)](https://bettercodehub.com/)

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

- Set the application deployment (assuming a live K8s cluster w/ app) UUID as ENV (APP_UUID)

- Set the ChaosEngine CR name as ENV (CHAOSENGINE) 
  - For CR spec, see: https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosengine.yaml

- If the experiments are not executed, apply the ChaosResult CRs manually 
  - For CR spec, see: https://github.com/litmuschaos/chaos-operator/blob/master/deploy/crds/chaosresult.yaml

- Run the exporter container (litmuschaos/chaos-exporter:ci) on host network. It is necessary to mount the kubeconfig
  & override entrypoint w/ `./exporter -kubeconfig <path>`

- Execute `curl 127.0.0.1:8080/metrics` to view metrics

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

