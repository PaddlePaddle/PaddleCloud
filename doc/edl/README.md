# Design Doc: Elastic Deep Learning

## Background

A PaddlePaddle training job contains several trainer instances,
several parameter server instances, and one master instance. 
We would like to atomically and easily manage the lifestyle of a PaddlePaddle training job. 
We would also like to automatically scale the number of training instances and the
number of parameter server instances to fully utilize the cluster's
computation resources. We call this Elastic Deep Learning. 

K8s provide [custom resource](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#custom-resources) 
that user can define an kind of resource with an extension of the k8s API.

K8s also provide [custom controller](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#custom-controllers) 
that user can deploy and update on a running cluster, independent of the cluster’s own lifecycle. Custom controllers
 can work with any kind of resource, 
but they are especially effective when combined with custom resources. 

With the reasons mentioned above, we need to develop a controller for lifecycle management and instances auto scaling 
of a PaddlePaddle distributed training job.

## Solution

We will build a Controller that can manage the lifestyle and auto scale the instance of a PaddlePaddle trainning job
. To be more precise, we will create a custom 
Kubernetes Resource 
and a custom Kubernetes Controller. Kubernetes is designed to be flexible so adding custom Resources and custom Controllers into our cluster does not require modifying the Kubernetes source code.
Currently, the lifecycle and auto scale were detached. We will merge them in the near future.

### Terminology
#### Training Job Resource

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
  name: paddlejob
  namespace: testspace
spec:
  image: "paddlepaddle/paddlecloud-job"
  port: 7164
  ports_num: 1
  ports_num_for_sparse: 1
  fault_tolerant: false
  mountPath: "/home/work/namespace/"
  master:
    resources:
      limits:
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
  pserver:
    min-instance: 2
    max-instance: 2
    resources:
      limits:
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
  trainer:
    entrypoint: "python train.py"
    workspace: "/home/job-1/"
    passes: 10
    min-instance: 2
    max-instance: 6
    resources:
      limits:
        cpu: "200m"
        memory: "200Mi"
      requests:
        cpu: "200m"
        memory: "200Mi"
```

For lifecycle management, the controller will create a PaddlePaddle cluster job with a master, 2 pserver, 2 trainer and 
update state 
of the training job.

For auto scaling, the controller will create and continuously scale the number of trainers and the number of pservers
 between the corresponding `min-instance` and `max-instance`.
 
Currently, we will only support trainer autoscaling. Parameter server autoscaling will be supported in the near 
 future. This design doc considers both of them.

Since the `master` server is required only when the trainer is using
`paddle.v2.reader.creator.cloud_reader`, The `master` spec is
optional: the master server will be created only when configured.

The training job custom resource can be created with: `kubectl create
-f training_job.yaml`.

#### TrainingJobUpdater

For lifecycle management. We need an object which named TrainingJobUpdater to manages a specific TrainingJob. 
Including job spec, current status and events of this job. 
Here is the struct of TrainingJobUpdater.
 
 ```go
 type trainingJobEventType string
 
 const (
 	trainingJobEventDelete trainingJobEventType = "Delete"
 	trainingJobEventModify trainingJobEventType = "Modify"
 )
 
 type trainingJobEvent struct {
 	// pet is the TrainingJobEventType of TrainingJob
 	pet trainingJobEventType
 	// The job transfer the information fo job
 	job *v1.TrainingJob
 }
 
 // TrainingJobUpdater is to manager a specific TrainingJob
 type TrainingJobUpdater struct {
 	// job is the job the TrainingJobUpdater manager.
 	job *v1.TrainingJob
 
 	// kubeCli is standard kubernetes client.
 	kubeCli kubernetes.Interface
 
 	// trainingJobClient is the client of TrainingJob.
 	trainingJobClient trainingJobClient.Interface
 
 	// status is the status in memory, update when TrainingJob status changed and update the CRD resource status.
 	status v1.TrainingJobStatus
 
 	// eventCh is the channel received by Controller, include Modify and Delete.
 	// When trainingJobEvent is Delete it will delete all resources
 	// The maximum is 1000.
 	eventCh chan *trainingJobEvent
 }
 
 ```
 
 When user submit a TrainingJob, the controller will start a TrainingJobUpdater to manage the TrainingJob. 
 - It will parser the config to TrainingJob spec including PSERVER, MASTER and TRAINER.
 - It will create resource orderly. 
 - It will sync the status of the job at regular time. While the status changed, it will update the job's status to k8s. 
 - It will release all the resource while the job succeeded or failed.
 
#### Controller

For lifecycle management, the controller manages distributed PaddlePaddle jobs by creating a series of 
TrainingJobUpdaters. Here is 
the struct of Controller.
```go
type Controller struct {
	// KubeCli is a standard kubernetes clientset
	KubeCli kubernetes.Interface
	// ApiCli is the extension kubernetes clientset
	ApiCli apiextensionsclient.Interface
	// PaddleCli is a clientset for our own API group
	PaddleCli paddleclientset.Interface

	trainingjobLister paddlelisters.TrainingJobLister
	trainingjobSynced cache.InformerSynced

	jobtracker map[string]*updater.TrainingJobUpdater

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

```

- It start to confirm that there are only one TrainingJob controller in cluster.
- It will register TrainingJob CRD to cluster if there is no TrainingJob resource.
- It will create a TrainingJobUpdater while PaddlePaddle job was submitted and delete the TrainingJobUpdater while job was 
deleted.

For auto scaling, the controller will runs as a Pod. It has the global view of
the computation resources. It watches the training job resources and
schedules and scales the training jobs using the Kuberenetes API.

The pseudo controller declaration (`autoscale_controller.yaml`) is
as follows:

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: autoscale-controller
spec:
  replicas: 1
  template:
    metadata:
      labels:
        name: autoscale-controller
    spec:
      containers:
      - name: autoscale-controller
        image: paddlepaddle/training-job-controller
```

The training job controller can be started by the cluster
administrator with command: `kubectl create -f autoscale_controller.yaml`

You can use `go/cmd/autoscaler/Dockerfile` to build a new controller image 
or download an existing one.

Currently, `Autoscaler` is not a k8s controller actually, we will merge it to controller in a near feature.

## Implementation

### Training Job Resource

The training job resource is a custom resource. There are two ways of implementing custom resources:

- [Custom Resource Definition (CRD)](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/), since Kubernetes v1.7.
- [Third Party Resource (TPR)](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-third-party-resource/), since Kubernetes v1.2, fully deprecated in v1.8, will be removed in v1.9.

We will support TPR first, because some of our clusters is using Kubernetes v1.6.

In the near feature, we will use CRD to replace TPR to define our training job resource,
then the resource defined under `go/edl/resource` will be deprecated.
If you want to use TPR in Kubernetes with version <= 1.7,
please checkout the `unreleased-tpr` tag.

Currently, implementation using CRD is still under development,
you can use the following command to verify and generate dependent codes:
```
# check freshness of generated codes
$ scripts/verify-codegen.sh

# update dependent codes
$ scripts/update-codegen.sh
```
For more details, please refer to [article](https://blog.openshift.com/kubernetes-deep-dive-code-generation-customresources/).
When the definition of resource is stable, we will commit generated dependent codes.
Just skip `go/apis` when running `go test` during your development.

By the way, if you want to generate these codes by yourself, due to this [issue](https://github.com/kubernetes/code-generator/issues/20),
you have to deal with the import problem.

### Controller
#### Lifecycle Overall

The whole life cycle of TrainingJob is managed through the two layer control of Lifecycle Controller and TrainingJobUpdater. As 
shown in the following figure:

![](pictures/lifecycle_overall.jpg)

As shown above, when the job is submitted, `Controller` will create a `TrainingJobUpdater` by `NewUpdater` and start 
the lifecycle manager. `TrainingJobUpdater` will start `InitResource` with a goroutine and start a `Ticker` to sync the
state of the trainingjob. When the state is changed to `failed` or `succeeded`, resources of PSERVER and MASTER will 
be released.

When the job is killed, resources of PSERVER and MASTER will be released and the metadata will be deleted.


#### State Machine

The struct of TrainingJob Status as follows.

```go
type TrainingJobStatus struct {
	// Phase is phase of TrainingJob
	Phase TrainingJobPhase `json:"phase"`
	// Reason is the reason of job phase failed
	Reason string `json:"reason"`
	// ScaleStatus is autoscale status of trainer jobs
	ScaleStatus TrainerJobScaleStatus `json:"scale_status"`
	// ReplicaStatuses is detail status of resources
	ReplicaStatuses []*TrainingResourceStatus `json:"replica_statuses"`
}
```

We define five phases to describe the traingingjob:

- TrainingJobPhaseNone: `""`
- TrainingJobPhaseCreating: `creating`
- TrainingJobPhaseRunning: `running`
- TrainingJobPhaseSucceeded: `succeeded`
- TrainingJobPhaseFailed: `failed`

When a job was submitted to cluster. Controller will start a Updater to manager the lifecycle. Here is the state 
machine of a TrainingJob:

![](pictures/state_machine.jpg)

As shown above, when the job is submitted, controller will start a Updater and the state of the job is none. While 
the job config is valid and through parser, the state will convert to `creating`. While all the resources are set up,
 the state will convert to `running`. While all trainer are succeeded or any trainer succeeded (fault tolerant mode),
  the state will convert to `succeeded`. Otherwise, the state will convert to `failed`.

### Autoscaler

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
`1.0`). A GPU resource will be taken away from the most fulfilled job
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
score of `1.0`). The CPU and Mem resource will be taken away from the
most fulfilled job when there is another job's `min-instance` is not
satisfied (unless the most fulfilled job's `cur-instance` equals to
`min-instance`). When the most fulfilled job's `cur-instance` equals
to `min-instance`, no training job will be scaled down, but the job
will still be scheduled optimistically.

## References

- [Writing a custom controller: Extending the functionality of your cluster](https://resources.coreos.com/youtube-coreos-fest-2017/writing-a-custom-controller-extending-the-functionality-of-your-cluster)
- [Introducing Operators: Putting Operational Knowledge into Software](https://coreos.com/blog/introducing-operators.html)
- [TPR Is Dead! Kubernetes 1.7 Turns to CRD](https://coreos.com/blog/custom-resource-kubernetes-v17)
- [Writing Controllers](https://github.com/kubernetes/community/blob/master/contributors/devel/controllers.md)
