# Scheduler for TrainingJob

## Background

We are going to define PaddlePaddle cluster job as a Kubernetes [TPR]() or
[CRD](). Each job is described using a `yaml` representation called
[TrainingJob](../autoscale/README.md). When a `TrainingJob` resource is
submited to Kubernetes cluster, our customized `controller` program will
receive an event informing the resource creation/deletion.

The `controller` program should contain the following core functions:

- Parser to parse `TrainingJob` resource to corresponding job components,
  including:
    - `ReplicaSet` of master process
    - `ReplicaSet` or `StatefulSet` for etcd cluster
    - `ReplicaSet` of `pserver` process
    - `Job` of `trainer` process
- Queue to sort `TrainingJob` resource for schedule
- Scheduler to determine which job to run or to scale by:
    - Job static priority
    - Job dynamic priority
    - Job resource request (GPU > CPU)
    - Cluster free resource
- Autoscaler to dynamically tunning job resources.

Cases that need to be considerd during the implementaion:

1. GPU is much more expensive than CPUs, jobs require GPU resource should
have higher priority to run on GPU machines than CPU only jobs. Also a
`TrainingJob` requires GPU must require enough CPU resource for it so that
CPU used for laugch CUDA kernels and memory copys is not blocking the
performance of GPU accelerated training jobs.
1. Jobs have priorities. Some offline jobs have higher priority that can be
able to aquire enough resource so that the job can complete at desired time,
then job result can be updated to the production service. Other jobs like
experiement and one-shot jobs have lower priority, they can be scaled up when
cluster is free and can be scaled down when cluster is busy.
1. Otherwise, jobs should share the cluster resource fairly, which means, if
a job is waiting enough long, it can finally be scheduled to the cluster, no
matter it may have very low priority (except that the cluster is full of
production service).
1. A cluster may run both online service and offline batch jobs. The online
services have high priority and is not interuptable. But trainingjobs can
re-use the cluster resource when the online service came to certain time of
day that is not that active.
1. About quota, users quota should be considered so that scheduled job is not
exceeding it.

## Scheduler design

Here we define the core scheduler interfaces and algorighm.

### Interface

Scheduler deal with atomic scheduling unit named `Node`. The `TraniningJob`
resource is member of `Node`, we can get it by calling `node.Obj()`.

```go
type PrioLevel int

const (
  Experiement PrioLevel = 0
  Offline = 3
  Normal = 7
  Production = 11
)

type Node interface {
  // GetPrio returns the current priority level.
  GetPrio() PrioLevel
  // SetPrio set the node priority level directly.
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
  Expected() int64
  // Running returns the current parrallelism of the node.
  // If Running == 0 means the job is waiting for resources.
  Running() int64

  // Obj returns inner scheduling unit.
  Obj() *interface{}
}
```

Currently we only support 4 levels of priority. Note that the priority is not
continuous, so that we can extend more levels later.

Then we define the scheduler interface:

```go
type GpuPriorityCFS interface {
  // AddNode insert new node to the scheduler.
  AddNode(node *Node) error
  // DelNode remove the completed node from scheduler.
  DelNode(node *Node) error
  // GetLeftMost return the smallest valued node in the scheduler's tree.
  GetLeftMost() *Node
  // GetRightMost return the maximum valued node in the scheduler's tree.
  GetRightMost() *Node
  // Len return number of nodes in the scheduler.
  Len() int

  // Tranverse go thought every nodes in the scheduler.
	Tranverse(callback ...func(*Node)) error
}
```

### Scheduling algorithm

We use an implementation simmilar to
[CFS](https://en.wikipedia.org/wiki/Completely_Fair_Scheduler) as the
default sheduler for `TrainingJobs`. Other jobs or services submited using
`kubectl` will not be controlled by this scheduler, but the resource
consumption will be considered.

Scheduler stores all nodes in a red-black tree, sorted by score
`sum(Prio() * ResourceScore() * Running() * running time)`
(total running time scores with resource request). In order to make the jobs
"fair", `Node`'s `Expected()` is caculated tranversing every node by order,
and increase/decrease one by one in each "dry run", try to make the score
even across cluster:

1. The left most child is selected. This is the job that spent least running
time on the cluster.
2. If the job is not running yet (newly submited job), try increase job
parallelism to the initial parallelism value. If resource is sufficient, 
parse the `TrainingJob` and create the job instance, if resource is not
sufficient, go to step 1.
3. If the job is already running try to scale it up by 1 to use more free
resources.
4. If the job has completed, stop the pserver and master, then remove it from
the tree.
5. Go to step 1 to run another step util all the nodes are tranversed.
6. Accumulate the diff for each node.
7. If above steps get no diff for every job, then use same strategy to scale
down some jobs to achive fairness, call `GetRightMost()` to get right most
node, and try scale down jobs if the score is far away from even.

- NOTE: we should make scale up/down operations less frequently, because
cluster job is not like processes, frequent interrupt may cause significant
job performance issue.

### Queues

We **don't** put jobs into sevaral queues, like "RUNNING", "TODO", "DONE". Only
**one** queue is used for indexing jobs. Jobs that have not started, consumes
no resource.


## References

https://en.wikipedia.org/wiki/Completely_Fair_Scheduler
https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-third-party-resource/#what-is-thirdpartyresource
https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/
https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/