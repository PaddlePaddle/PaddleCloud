# Auto-scaling Experiment Design


## Environment Enviroment

- Kubernetes v1.6 cluster with 133 nodes.
- PaddleCloud with latest develop branch installed.
- We will train the [recognize_digits](https://github.com/PaddlePaddle/cloud/tree/develop/demo/recognize_digits) model in the experiment.


## Experiment Metric

- Cluster overall resource utilization.
- Average job running time.
- Average job pending time.


## Test Cases

### Autoscaling on the Special Purpose Cluster

All the job in the cluster will be training jobs (hence the name
special purpose cluster).

#### Variable

- Autoscaling ON/OFF.

#### Invariant

- The number of jobs.
- The configuration of each job.
- The submission time for each job.

#### Experiment Steps

1. With autoscaling turned off, submit the training jobs over
   predefined submission delay between each job.
1. With autoscaling turned on, submit the training jobs over
   predefined submission delay between each job.


#### Experiment Result Example:

- Autoscaling OFF

	PASS|AVG RUNNING TIME|AVG PENDING TIME|JOB RUNNING TIME|CLUSTER CPU UTILS
	--- | --- | --- | --- | ---
	0|379|102|415,365,380,375,350,365,495,365,345,335|63.38
	1|322|85|375,315,395,310,280,330,380,270,285,280|65.05
	AVG|331|86|N/A|63.55

- Autoscaling ON

	PASS|AVG RUNNING TIME|AVG PENDING TIME|JOB RUNNING TIME|CLUSTER CPU UTILS
	--- | --- | --- | --- | ---
	0|379|102|415,365,380,375,350,365,495,365,345,335|63.38
	1|322|85|375,315,395,310,280,330,380,270,285,280|65.05
	AVG|331|86|N/A|63.55


### Autoscaling on the General Purpose Cluster

Hybrid deployment with online serving and offline training Job (hence
the name general purpose cluster). We will deploy PaddlePaddle
training job and [Nginx](https://www.nginx.com/resources/wiki/) web
serving together.

#### Variable

- The number of Nginx instances, changing over time, simulating the
  real world traffic load distribution over time.
- Autoscaling ON/OFF.

#### Invariant

- The number of training jobs.
- The configuration of each training job.
- The configuration of each Nginx job.
- The submission time for each training job.

#### Experiment Steps

1. Start `N` Nginx instances to simulate the number of nginx instances
   required for the peak time load.

1. Start the training jobs.

1. Decrease the Nginx instances count of `N` to `M` over time, to
   simulate the Nginx load decreases, requiring fewer nginx instances.

1. Increase the Nginx instances count of `M` to `N` over time, to
   simulate the fully Nginx load cycle.

#### Experiment Result Example:

- Autoscaling OFF

	PASS|AVG RUNNING TIME|AVG PENDING TIME|JOB RUNNING TIME|CLUSTER CPU UTILS
	--- | --- | --- | --- | ---
	0|379|102|415,365,380,375,350,365,495,365,345,335|63.38
	1|322|85|375,315,395,310,280,330,380,270,285,280|65.05
	AVG|331|86|N/A|63.55

	Time | NGINX COUNT | TRAINER COUNT | CLUSTER CPU UTILS
	-- | -- | -- | --
	0  | 100 | 50 | 100
	5  | 90  | 50 | 90
	10 | 80  | 50 | 80
	15 | 90  | 50 | 90
	20 | 100 | 50 | 100

- Autoscaling ON

	PASS|AVG RUNNING TIME|AVG PENDING TIME|JOB RUNNING TIME|CLUSTER CPU UTILS
	--- | --- | --- | --- | ---
	0|379|102|415,365,380,375,350,365,495,365,345,335|63.38
	1|322|85|375,315,395,310,280,330,380,270,285,280|65.05
	AVG|331|86|N/A|63.55

	Time | NGINX COUNT | TRAINER COUNT | CLUSTER CPU UTILS
	-- | -- | -- | --
	0  | 100 | 50 | 100
	5  | 90  | 55 | 100
	10 | 80  | 60 | 100
	15 | 90  | 55 | 100
	20 | 100 | 50 | 100


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
