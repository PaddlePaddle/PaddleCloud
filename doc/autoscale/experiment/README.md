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

### Comparing the auto-scaling training job and the general training job

- Submit the general training job 
    1. Submit a job(job-A), which requests 100 trainer Pods(1 CPU cores per Pod), 2 pservers and 1 master.
    1. Submit another job(job-B), which requests 200 trainer Pods(1 CPU cores per Pod), 2 pservers and 1 master.
    1. The job-B will be the pending status only if job-A finished because there are not enough CPU cores for the requests.
- Submit the auto-scaling training job
    1. Submit a job(job-A), which requests 100 trainer Pods(1 CPU core per Pod, min-instances is 50, max-instances is 500), 2 pservers and 1 master, And then job-A will be scaled up to immediately to use the maximum free resources(max 500 trainer Pods).
    1. Submit another job(job-B), which requests 200 trainer Pods(1 CPU core per Pod, min-instances is 50, max-instances is 200), 2 pservers and 1 master.
    1. Job-A will be scaled down and job-A and job-B will run in the cluster at the same time, and they will use the maximum free resources.

- Experiment metrics
    1. Compare the **CPU utils** with auto-scaling training job and general training job.
    1. Compare the **training time** for each job.
    1. Compare the **average waiting time** for each job. 

- Experiment result example:

metrics |  auto-scaling training job| general training job
-- | -- | --
average running time | 6h | 8h
average pending time | 0 | 2h
CPU utils | 100% | 60%

### Hybrid Deployment with Online Serving and Offline Training Job

In the general cluster, we will deploy some online serving such as Nginx cluster, Dataset serving such as MySQL and some offline training Job. we will deploy some Nginx Pods to simulate the production environment. 

- Deploy Nginx Pods in the cluster, configure HPA on Nginx Deployment.
- Submit a training Job, which requests 100 trainer Pods(2 pservers, 1 master, min-instance=2, max-instance=100), the trainers will be scaled immediately to use the maximum free resources in the cluster.
- Increase the QPS of the Nginx serving, the Nginx pods count will be scaled up by HPA, and the training job will be scaled down by TrainingJob controller.
- Experiment metrics
    1. CPU utils for the cluster(requests / total).
    1. Trainer Pods count.
- Experiment result example

metrics | QPS(1w) | QPS(10w) | QPS(50w)
-- | -- | -- | --
Trainer Pods | 100 | 80 | 50
Nginx Pods | 80 | 100 | 150
CPU utils| 100% | 100% | 100%

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
        ```bash
        > cd cloud/doc/autoscale/experiment/python
        > ./control_case1.sh --help
        > usage: control_case1.sh <action>
            action[required]: str[start|stop], will start or stop all the jobs.
          env var:
            JOB_COUNT[optional]:             int, The number of submiting jobs, defualt is 1.
            FAULT_TOLERANT[optional]:   str[ON|OFF], whether a fault-tolerant job,default is OFF.
            PASSES[optional]:           int, The number of run passes.
            DETAILS[optional:           str[ON|OFF], print detail monitor information.
        ```
        For example, run TestCase1 for 10 passes and 10 jobs:
        ```bash
            > PASSES=10 JOB_COUNT=10 ./control_case1.sh start
        ```
    1. Gernerate Experiment Report
        After all the passes are finished, the report will generated at './out' folder.