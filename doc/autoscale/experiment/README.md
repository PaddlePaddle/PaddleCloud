# Auto-scaling Experiment Design

## Environment Requirement

- Kubernetes cluster with 1.6.x installed.
- PaddleCloud with latest develop branch installed.
- At least 4 kubernetes nodes, each node should have 2 GPU cards at least.
- Dataset prepared to multiple files with the RecordIO format.

## Experiment Metric

- Computing resource utils(requests / total) for the cluster.
- A total number of the pods for all training job.

## Before Starting The Experiment

- All the demos in [book](https://github.com/PaddlePaddle/book) should be tested.
- We will use [recognize_digits](https://github.com/PaddlePaddle/cloud/tree/develop/demo/recognize_digits) as the training job for the demo.
- We have 240 CPU cores and 80 GPU cards totally.

## Test Cases

### Compare The Fault-tolerant Training Job with Auto-scaling Or Not

- Submit the fault-tolerant training job 
    1. Submit 10 jobs, which have 20 parallelisms(requests 10 CPU per Pod), 10 pservers and 1 master.
    1. The resource which the jobs requested need greater than the resource of the cluster.
    1. Collect the time series data and the statistical result.
- Submit the fault-tolerant training job with auto-scaling
    1. Submit 10 jobs, which have 5 parallelisms(request 10 CPU per Pod), 10 pservers and 1 master.
    1. Submit the TrainingJob, which have min-instance is 2 and max-instance is 20.
    1. Collect the time series data and the statistical result.

- Experiment metrics
    1. Compare the **CPU utils** with auto-scaling training job and general training job.
    1. Compare the **running time** for each job.
    1. Compare the **average pending time** for the jobs. 
    1. Compare the **average running time** for the jobs 

- Experiment result example:

PASS|AVG_RUNNINT_TIME|AVG_PENDING_TIME|JOB_RUNNING_TIME|CPU_UTILS
--- | --- | --- | --- | ---
0|124|21|135,125,120,120,115,115,205,100,105,105|56.33
1|134|26|130,130,125,125,120,120,225,125,125,115|56.60
2|126|23|135,130,125,120,115,110,185,110,110,120|56.04
3|160|10|175,210,185,130,125,125,125,220,190,120|42.59
4|155|23|160,160,150,150,145,140,220,130,135,165|52.49
AVG|139|20|N/A|52.81

### Hybrid Deployment with Online Serving and Offline Training Job

In the general cluster, we will deploy some online serving such as Nginx cluster, Dataset serving such as MySQL and some offline training Job. we will deploy some Nginx Pods to simulate the production environment. 

- Submit Nginx Deployment and training job
    1. Deploy Nginx Pods in the cluster with Deployment.
    1. Submit 5 fault-tolerant training job with auto-scaling, which have 5 parallelisms.
    1. Submit the TrainingJob, which have min-instance is 2 and max-instance is 20.
    1. Make the resource of Nginx and the training job full fill the cluster.
    1. Scale up Nginx Deployment, and the parallelism of training job will be scaled down.

- Experiment metrics
    1. CPU utils for the cluster(requests / total).
    1. Trainer Pods count.
    1. Nginx Pods count.
- Experiment result example

TIME|RUNNING_TRAINERS|NGINX_PODS|CPU_UTIL
-- | -- | -- | --
0|200|50|80
100|150|120|85
150|100|200|85

## Reproduce the experiment

- Configure kubectl on your host
- Prepare
    1. Configure kubectl 
    1. Configure paddlectl
    1. Submit the TrainingJob controller with YAML file
    ```bash
    > git clone https://github.com/PaddlePaddle/cloud.git && cd cloud
    > kubectl create -f k8s/controller/trainingjob_resource.yaml
    > kubectl create -f k8s/controller/controller.yaml
    ```
- Test Case1
    1. Run the TestCase1 for serval passes with bash scripts`./control_case.1.sh`:
        For example, run TestCase1 for 10 passes and 10 jobs:
        ```bash
        > cd cloud/doc/autoscale/experiment
        > PASSES=5 JOB_COUNT=10 ./control_case1.sh start
        ```
        Or Submit an auto-scaline training job
        > cd cloud/doc/autoscale/experiment
        ```bash
        > AUTO_SCALING=ON PASSES=5 JOB_COUNT=10 ./control_case1.sh start
        ```
    1. Gernerate Experiment Report
        After all the passes are finished, the report will be generated at './out' folder.
