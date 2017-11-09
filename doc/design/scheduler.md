# Scheduler for TrainingJob

## Background

We are going to define PaddlePaddle cluster job as a Kubernetes
[TPR](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-third-party-resource/) or
[CRD](https://kubernetes.io/docs/concepts/api-extension/custom-resources/).
Each job is described using a `yaml` representation called
[TrainingJob](../autoscale/README.md). When a `TrainingJob` resource is
submitted to Kubernetes cluster, our customized `controller` program will
receive an event informing the resource creation/deletion.

The `controller` program should contain the following core functions:

- Parser to parse `TrainingJob` resource to corresponding job components,
  including:
    - `ReplicaSet` of master process
    - `ReplicaSet` or `StatefulSet` for etcd cluster
    - `ReplicaSet` of `pserver` process
    - `Job` of  `trainer` process
- Queue to sort `TrainingJob` resource for schedule
- Scheduler to determine which job to run or to scale by:
    - Job static priority
    - Job resource request (GPU > CPU > Memory)
    - Job total running time of every pod
    - Cluster free resource
- Autoscaler to dynamically tunning job resources.

Cases that need to be considered during the implementation:

1. GPU is much more expensive than CPUs, jobs require GPU resource should
have higher priority to run on GPU machines than CPU only jobs. Also a
`TrainingJob` requires GPU must require enough CPU resource for it so that
CPU used for launch CUDA kernels and memory copies is not blocking the
performance of GPU accelerated training jobs.
1. Jobs have priorities. Some offline jobs have the higher priority that can be
able to acquire enough resource so that the job can complete at the desired time,
then job result can be updated to the production service. Other jobs like an experiment and one-shot jobs have lower priority, they can be scaled up when the cluster is free and can be scaled down when the cluster is busy.
1. Otherwise, jobs should share the cluster resource fairly, which means, if
a job is waiting enough long, it can finally be scheduled to the cluster, no
matter it may have very low priority (except that the cluster is full of
production service).
1. A cluster may run both online service and offline batch jobs. The online
services have high priority and are not interruptible. But `TrainingJobs` can
re-use the cluster resource when the online service came to the certain time of
day that is not that active.
1. About quota, users quota should be considered so that scheduled job is not
exceeding it.

## Scheduler design

Here we define the core scheduler interfaces and algorithms.

### Interface

Scheduler deals with atomic scheduling unit named `Unit`. The `TraniningJob`
resource is the member of `Unit`, we can get it by calling `unit.Obj()`.

```go
type PrioLevel int

const (
  Experiement PrioLevel = 10
  Offline = 100
  Normal = 1000
  Production = 10000
)

type Unit interface {
  // GetPrio returns the current priority level.
  GetPrio() PrioLevel
  // SetPrio set the unit priority level directly.
  SetPrio(prio PrioLevel)

  // MaxInstances returns the desired max parallelism of the job.
  MaxInstances() int
  // MinInstances returns the minimal parallelism the job can be running.
  MinInstances() int
  // ResourceScore returns resource score of a single pod. It's
  // caculated by sum(weight*ResourceValue).
  ResourceScore() int64

  // Expected returns expected parallelism (how much pods) to run for
  // current scheduling step.
  ExpectedCount() int64
  // Running returns the current parrallelism of the unit.
  // If Running == 0 means the job is waiting for resources.
  RunningCount() int64

  // Obj returns inner scheduling unit.
  Obj() interface{}
}
```

Currently, we only support 4 levels of priority. Note that the priority is not
continuous, so that we can extend more levels later.

Then we define the scheduler interface:

```go
type GpuPriorityCFS interface {
  // AddUnit insert a new Unit object to the scheduler.
  AddUnit(unit *Unit) error
  // DelUnit remove the completed unit from scheduler.
  DelUnit(unit *Unit) error
  // GetLeftMost return the smallest valued unit in the scheduler's tree.
  GetLeftMost() *Unit
  // GetRightMost return the maximum valued unit in the scheduler's tree.
  GetRightMost() *Unit
  // Len return number of units in the scheduler.
  Len() int

  // Traverse go thought every unit in the scheduler.
    Tranverse(callback ...func(*Unit)) error
}
```

### Scheduling algorithm

We use an implementation similar to
[CFS](https://en.wikipedia.org/wiki/Completely_Fair_Scheduler) as the
default scheduler for `TrainingJobs`. Other jobs or services submitted using
`kubectl` will not be controlled by this scheduler, but the resource
consumption will be considered.

Scheduler stores all units in a red-black tree, sorted by score
`GetPrio() * ResourceScore() * sum(RunningCount() * pod.RunningTime())`
(weighted total running time). In order to make the jobs
"fair", `Unit`'s `ExpectedCount()` is calculated traversing every unit by order,
and increase/decrease one by one in each "dry run", try to make the score
even across cluster:

1. The left most child is selected. This is the job that spent least running
time on the cluster.
2. If the job is not running yet (newly submitted job), try to increase job
parallelism to the initial parallelism value. If the resource is sufficient, 
parse the `TrainingJob` and create the job instance, if the resource is not
sufficient, go to step 1.
3. If the job is already running try to scale it up by 1 to use more free
resources.
4. If the job has completed, stop the pserver and master, then remove it from
the tree.
5. Go to step 1 to run another step until all the units are traversed.
6. Accumulate the diff for each unit.
7. If above steps get no diff for every job, then use the same strategy to scale
down some jobs to achieve fairness, call `GetRightMost()` to get right most
unit, and try scale down jobs if the score is far away from even.

- NOTE: we should make scale up/down operations less frequently, because
cluster job is not like processes, frequent interruptting may cause significant
job performance issue.

### Queues

We **don't** put jobs into several queues, like "RUNNING", "TODO", "DONE". Only
**one** queue is used for indexing jobs. Jobs that have not started, consumes
no resource.

### Scheduling Intervals And Freezing Window

Scheduling operations are triggered:

- Every 5 seconds
- Or by controller event: adding, updating, deleting Of `TrainingJob`

Speaking of "fair", if we do scaling operations very fast, every jobs' trainer
the count will be constantly in flux, and that's what we don't want. We introduce
configurable `FreezingWindow` for every `TrainingJob`, in that time window,
the job should not take any scaling operations to minimize the cost introduced
by scaling the job.

## References

- https://en.wikipedia.org/wiki/Completely_Fair_Scheduler
- https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-third-party-resource/#what-is-thirdpartyresource
- https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/
- https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/