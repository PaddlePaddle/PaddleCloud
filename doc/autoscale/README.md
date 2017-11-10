# Design Doc: Elastic Deep Learning

## Background

A PaddlePaddle training job contains several trainer instances,
several parameter server instances, and one master instance. We would
like to automatically scale the number of training instances and the
number of parameter server instances to fully utilize the cluster's
computation resources. We call this Elastic Deep Learning.

Currently, we will only support trainer autoscaling. Parameter server
autoscaling will be supported in the near future. This design doc
considers both of them.

[Horizontal Pod Autoscaling (HPA)](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) is
an autoscaling solution provided by Kuberentes, but it's not suitable
for the training job autoscaling for the following reasons:

- The goal of autoscaling is to fairly distribute the computation
  resources to different training jobs in a way that optimizes the
  computation resource utilization of the **entire** cluster. The HPA
  is trying to improve the quality of service of a **single**
  service. The training job autoscaling requires the controller to
  have a global view of all the available computation resources and
  all the training jobs, but HPA does not have the global view.

- HPA is designed to automatically scale a homogeneous set of Pods,
  but we need to scale a heterogeneous set of Pods (the trainer Pods
  and the parameter server Pods): because the required number of
  parameter servers is correlated to the required number of trainers,
  we need to scale them together.

We need to develop our own solution for autoscaling.

## Solution

We will build
an [Operator](https://coreos.com/blog/introducing-operators.html) that
do the autoscaling. To be more precise, we will create
a
[custom Kubernetes Resource](https://kubernetes.io/docs/concepts/api-extension/custom-resources/) and
a
[custom Kuberentes Controller](https://resources.coreos.com/youtube-coreos-fest-2017/writing-a-custom-controller-extending-the-functionality-of-your-cluster). Kuberentes
is designed to be flexible so adding custom Resources and custom
Controllers into our cluster does not require modifying the Kubernetes
source code.

### Training Job Resource

Just
like
[Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) is
a resource that describes a deployment. We will have a training
job
[Custom Resource](https://kubernetes.io/docs/concepts/api-extension/custom-resources/) that
describes the training job.

A pseudo resource declaration (`training_job.yaml`) is as follows:

```yaml
apiVersion: paddlepaddle.org/v1
kind: TrainingJob
metadata:
  name: job-1
spec:
  trainer:
    entrypoint: "python train.py"
    workspace: "/home/job-1/"
    min-instance: 3
    max-instance: 6
    resources:
      limits:
        alpha.kubernetes.io/nvidia-gpu: 1
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
  pserver:
    min-instance: 3
    max-instance: 3
    resources:
      limits:
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
  master:
    resources:
      limits:
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
```

The training job controller will create and continuously scale the
number of trainers and the number of pservers between the
corresponding `min-instance` and `max-instance`.

Since the `master` server is necessary only when the trainer is using
`paddle.v2.reader.creator.cloud_reader`, The `master` spec is
optional: the master server will be created only when configured.

The training job custom resource can be created with: `kubectl create
-f training_job.yaml`.

The custom resource will only be saved on Kuberentes, we will need a
custom controller that operate on it.

### Training Job Controller

The training job controller run as a Pod. It has the global view of
the computation resources. It watches the training job resources and
schedules and scales the training jobs using the Kuberenetes API.

The pseudo controller declaration (`training_job_controller.yaml`) is
as follows:

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: training-job-controller
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: training-job-controller
    spec:
      containers:
      - name: training-job-controller
        image: paddlepaddle/training-job-controller
```

The training job controller can be started by the cluster
administrator with command: `kubectl create -f
training_job_controller.yaml`

## Implementation

### Training Job Resource

The training job resource is a custom resource, there are two ways of
implementing custom resources:

- [Custom Resource Definition (CRD)](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/),
  since Kuberentes v1.7.
- [Third Party Resource (TPR)](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-third-party-resource/),
  since Kubernetes v1.2, fully deprecated in v1.8, will be removed in
  v1.9.

We will support TPR first, because some of our clusters is using
Kubernetes v1.6. CRD will be supported in the future.


### Training Job Controller

Currently, we will run a single training job controller instance and
assume that there is no training job controller running concurrently
(the assumption could be false
when
[split-brain](https://en.wikipedia.org/wiki/Split-brain_(computing))
happens). In the future, we will run multiple instances and use leader
election to choose a leader.

The pseudo-logic is as follows:

```go
registerThirdPartyResource()
for {
  quota := getTotalComputationResourceQuota()
  current := getCurrentJobStates()
  desired := getDesiredJobStates()
  dynamicScaling(quota, current, desired)
}
```

#### Scaling Algorithm

##### Elastic Job

A job is elastic only when it's trainer and pserver's `min-instance`
equals to the `max-instance` respectively. We will only scale elastic
jobs.

Currently, we will not scale the parameter server instances.

##### Fulfillment Score

When there are available computation resources, the algorithm needs to
decide which jobs to assign the resources to. When there are no more
available computation resources but the newly submitted job needs it,
the algorithm needs to decide which job to take the resource away
from. We will introduce the *fulfillment score* to answer these
questions:

```go
func (j Job) Score() float64 {
  minInstance := j.spec.trainer.minInstance
  maxInstance := j.spec.trainer.maxInstance
  curInstance := j.trainer.currentInstance()
  return float64(curInstance - minInstance) / float64(maxInstance - minInstance)
}
```

##### Scaling GPU Jobs

The controller knows the total number of available GPUs in a cluster
and will try to assign all of them to the training jobs.

All elastic GPU jobs will be sorted according to their fulfillment
score. The number of GPU per instance, CPU requests value, Mem
requests value will be used as tiebreakers in decreasing importance.

An available GPU resource will be assigned to the least fulfilled job
unless that job is already fulfilled (with a fulfillment score of
`1.0`). A GPU resource will be take away from the most fulfilled job
when there is another GPU job's `min-instance` is not satisfied
(unless the most fulfilled job's `cur-instance` equals to
`min-instance`). When the most fulfilled job's `cur-instance` equals
to `min-instance`, no training job will be scaled down, the new job
cannot be scheduled and will wait for more resources.


##### Scaling CPU Jobs

The controller knows the total CPU capacity, Mem capacity of the
cluster, and the total CPU limits, Mem limits of all training jobs. We
define the available CPU and Mem as the difference of the capacity and
the
[limits](https://kubernetes.io/docs/concepts/policy/resource-quotas/#requests-vs-limits) (not
the
[requests](https://kubernetes.io/docs/concepts/policy/resource-quotas/#requests-vs-limits))
respectively.

All elastic CPU jobs will be sorted according to their fulfillment
score. The CPU requests value, Mem requests value will be used as
tiebreakers in decreasing importance.

The available CPU and Mem resource will be assigned to the least
fulfilled job unless that job is already fulfilled (with a fulfillment
score of `1.0`). The CPU and Mem resource will be take away from the
most fulfilled job when there is another job's `min-instance` is not
satisfied (unless the most fulfilled job's `cur-instance` equals to
`min-instance`). When the most fulfilled job's `cur-instance` equals
to `min-instance`, no training job will be scaled down, but the job
will be still scheduled optimistically.

## References

- [Writing a custom controller: Extending the functionality of your cluster](https://resources.coreos.com/youtube-coreos-fest-2017/writing-a-custom-controller-extending-the-functionality-of-your-cluster)
- [Introducing Operators: Putting Operational Knowledge into Software](https://coreos.com/blog/introducing-operators.html)
- [TPR Is Dead! Kubernetes 1.7 Turns to CRD](https://coreos.com/blog/custom-resource-kubernetes-v17)
- [Writing Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md)
