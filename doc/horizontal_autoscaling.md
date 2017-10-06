# Design Doc: Horizontal Autoscaling

## Background

A PaddlePaddle training job contains several trainer instances,
several parameter server instances, and one master instance. We would
like to automatically scale the number of training instances and the
number of parameter server instances to fully utilize the cluster's
computation resources.

Currently, we will only support horizontal trainer
autoscaling. Parameter server horizontal autoscaling will be supported
in the near future. This design doc considers both of them.

[Horizontal Pod Autoscaling (HPA)](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/) is
an autoscaling solution provided by Kuberentes, but it's not suitable
for the training job autoscaling for the following reasons:

- The training job autoscaling tries to fairly distribute the
  computation resources to different training jobs, in order to
  optimize the computation resource utilization of the **entire**
  cluster. The HPA is trying to improve the quality of service of a
  **single** service. The training job autoscaling requires the
  controller to have a global view of all the available computation
  resources and all the training jobs, but HPA does not have the
  global view.

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
apiVersion: cloud.paddlepaddle.org/v1
kind: TrainingJob
metadata:
  name: job-1
spec:
  trainer:
    entrypoint: "python train.py"
    workspace: "/home/job-1/"
    min-replica: 3
    max-replica: 6
    resources:
      limits:
        alpha.kubernetes.io/nvidia-gpu: 1
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
  pserver:
    min-replica: 3
    max-replica: 3
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
corresponding `min-replica` and `max-replica`.

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

We will be using CRD unless our cluster can not be updated to
Kubernetes v1.7, since TPR will be removed in v1.9.

### Training Job Controller

Currently, we will run a single training job instance and assume that
there is no training job controller running concurrently (the
assumption could be false when split-brain happens). In the future, we
will run multiple instances and use leader election to choose a
leader.

The pseudo-logic is as follows:

```go
registerThirdPartyResource()
for {
  quota := getTotalComputationResourceQuota()
  current := getCurrentJobStates()
  desired := getDesiredJobStates()
  dynamicSchedule(quota, current, desired)
}
```

## References

- [Writing a custom controller: Extending the functionality of your cluster](https://resources.coreos.com/youtube-coreos-fest-2017/writing-a-custom-controller-extending-the-functionality-of-your-cluster)
- [Introducing Operators: Putting Operational Knowledge into Software](https://coreos.com/blog/introducing-operators.html)
- [TPR Is Dead! Kubernetes 1.7 Turns to CRD](https://coreos.com/blog/custom-resource-kubernetes-v17)
- [Writing Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md)
