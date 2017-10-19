# Auto-scaling Experiment Design

## Environment Requirement

- Kubernetes cluster with 1.6.x installed.
- PaddleCloud with latest develop branch installed.
- At least 4 kubernetes nodes, each node should have 2 GPU cards at least.
- Dataset prepared to multiple files with the RecordIO format.

## Experiment Metric

- Computing resource utils(requests / total).
- Training time for each job.
- The number of running trainer.

## Before Starting The Experiment

- We will use [recognize_digits](https://github.com/PaddlePaddle/cloud/tree/develop/demo/recognize_digits) as the training job for the demo.
- We have 240 CPU cores and 80 GPU cards totally.

## Test Case

### Without auto-scaling Job 

- Start a Deployment to simulate the online serving(10 pods).
- Start a training job(jobA) with 2~100 trainer instances(2 pservers, 1 master), the trainers will be scaled immediately to use the maximum free resources in the cluster.
- Start another training job(jobB) with 50~100 trainer instances(2 pservers, 1 master), there is no enough resource, the job will wait for the adequacy of the resource.

### With auto-scaling Job

- Start a Deployment to simulate the online serving(10 pods).
- Start a training job(jobA) with 2~100 trainer instances(2 pservers, 1 master), the trainers will be scaled immediately to use the maximum free resources in the cluster.
- Start another training job(jobB) with 50~100 trainer instances(2 pservers, 1 master), jobA will scale down, jobA and jobB will run in the cluster at the same time.

### With GPU

The same with above