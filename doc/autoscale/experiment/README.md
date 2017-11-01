# Auto-scaling Experiment Design

## Purpose

To verify the effectiveness of PaddlePaddle's fault-tolerance and auto-scaling mechanism.

## Metrics

How the effectiveness is measured.

1. Cluster computing resource overall utilization.
    - The higher the better.
    - Higher utilization means less resource is idle. Autoscaling intended to maximize the overall cluster resource(CPU, GPU, memory) usage by ensuring resource for production level jobs/services, then fairly scale jobs that are scalable to use the resource left in the cluster.
1. Task average pending time.
    - The less the better.
    - The less pending time the earlier developers and researchers can start seeing the training cost curve, and the better they can verify the training algorithm effectiveness.
    - This is a common pain point of researchers with the internal cloud.
1. Task average execution time.
    - The less the better in general.
    - However, the average execution time is bound to increase due to prioritizing production jobs/services. In this case, we would say the less the average job running time increases, the better the scaler performances.
    - Average execution time is also the way of measuring the effectiveness of fault-tolerance. If the fault-tolerance is not working properly, the training job will simply fail or finish with significantly longer duration.
1. Quality of service with general purpose cluster
    - Check if the Machine learning process will yield resources to more important online services when the load is getting intensive.

## Our setup

- Kubernetes cluster with 1.6.x installed.
- PaddleCloud with latest develop branch installed.
- 133 physical nodes.
- Use [recognize_digits](https://github.com/PaddlePaddle/cloud/tree/develop/demo/recognize_digits) as benchmark training job.

## Test Cases

### Autoscaling on the Special Purpose Cluster

All the job in the cluster will be training jobs (hence the name
special purpose cluster). This case is a very typical scenario for research institutes.

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
serving together. This case is a very typical scenario for large enterprises and Internet companies.

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

- Prepare
    1. Configure kubectl and paddlectl on your host.
    1. Submit the TrainingJob controller with the YAML file.
    ```bash
    > git clone https://github.com/PaddlePaddle/cloud.git && cd cloud
    > kubectl create -f k8s/controller/trainingjob_resource.yaml
    > kubectl create -f k8s/controller/controller.yaml
    ```
- Run the Test Case
    1. Run the TestCase1 or TestCase2 for serval passes with the bash script `./run.sh`:
        For example, run TestCase1 for 10 passes and 10 jobs:
        ```bash
        > cd cloud/doc/autoscale/experiment
        > AUTO_SCALING=OFF PASSES=5 JOB_COUNT=40 ./run.sh start case1
        ```
        Or submit an auto-scaline training job
        > cd cloud/doc/autoscale/experiment
        ```bash
        > AUTO_SCALING=ON PASSES=5 JOB_COUNT=40 ./run.sh start case1
        ```
        Or run the TestCase2 with 5 jobs:
        ```bash
        > JOB_COUNT=5 ./run.sh start case2
        ```
	1. Get the time series data.
		The time serise data will be appended in the file `./out/mnist-case[1|2]-pass[0-9].log`
		as the following format:


		```
		0,2.11,0,3,0,0,0,0,0|0|0,0.00|0.00|0.00
		2,2.11,0,3,0,0,0,0,0|0|0,0.00|0.00|0.00
		4,2.11,0,3,0,0,0,0,0|0|0,0.00|0.00|0.00
		5,2.11,0,2,1,0,0,0,0|0|0,0.00|0.00|0.00
		7,5.30,7,2,0,1,0,0,7|0|0,3.19|0.00|0.00
		9,7.90,19,2,0,1,0,0,19|0|0,5.79|0.00|0.00
		10,8.11,20,2,0,1,0,0,20|0|0,6.01|0.00|0.00
		```
		The meaning of each column is:

		timestamp|total cpu util|# of running trainer|# of not exist jobs|# of pending jobs|# of running jobs|# of done jobs|# of nginx pods|running trainers for each job |cpu utils for each job
		--|--|--|--|--|--|--|--|--|--

	1. Calculate the average wainting time, and the average running time from time series data.
		The statistical data will be generated in the file: `./out/mnist-case[1|2]-result.csv`
		as the following format:
		```
		PASS|AVG RUNNING TIME|AVG PENDING TIME|JOB RUNNING TIME|AVG CLUSTER CPU UTILS
		0|240|37|306,288,218,208,204,158,268,253,250,228,214,207,212,268,173,277,330,332,257,164|55.99
		AVG|240|37|N/A|55.99
		```

	1. Plot from the time series data.
	    TODO

    1. Gernerate Experiment Report
        After all the passes are finished, the report will generated at './out' folder.

## Conclusions

### Resource utilization

TBD

### Average Pending time

TBD

### Average execution time

TBD

### Improved the service quality with general purpose cluster

As shown in test case two, PaddlePaddle yields resource to more important online services when the load is getting intensive.
