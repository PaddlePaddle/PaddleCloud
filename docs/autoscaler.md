# Overview
Package autoscaler provides an auto-scaler that monitors cluster resources status and decides how to scale the training jobs accordingly.

# `autoscaler.go`

## type Autoscaler
```go
type Autoscaler struct {
	// kubeCli is the client for k8s standard API resources.
	kubeCli        kubernetes.Interface
	// jobtracker is a map from job name to its jobUpdater
	jobtracker     *sync.Map
	// maxLoadDesired is the max cluster load desired.
	// 1 is full load, 0 is no load.
	maxLoadDesired float64
}
```

## func WithMaxLoadDesired
```go
func WithMaxLoadDesired(maxLoadDesired float64) func(as *Autoscaler)
```
`WithMaxLoadDesired` returns a `func(as *Autoscaler)` that sets `maxLoadDesired` field for `as`.


## func NewAutoscaler
```go
func NewAutoscaler(kubeClient kubernetes.Interface, jobtracker *sync.Map, options ...func(*Autoscaler)) *Autoscaler
```
`NewAutoscaler` constructs an auto-scaler and applies all `options` to it.


## func (*Autoscaler) InquiryResource
```go
func (a *Autoscaler) InquiryResource() (ClusterResource, error)
```
`InquiryResource` returns the resource status of the cluster.


## func elastic
```go
func elastic(j *padv1.TrainingJob) bool
```
`elastic` returns if training job `j` is elastic.

## func needGPU
```go
func needGPU(j *padv1.TrainingJob) bool
```
`needGPU` returns if training job `j` needs GPU.

## func sortedJobs
```go
func sortedJobs(j []*padv1.TrainingJob, filters ...func(*padv1.TrainingJob) bool) []*padv1.TrainingJob
```
`sortedJobs` sorts the training jobs that pass all the `filters` by fulfillment and resources demand in ascending order.


## func searchAssignableNode
```go
func searchAssignableNode(r *ClusterResource, j *padv1.TrainingJob) string
```
`searchAssignableNode` returns the id of a node that has enough idle CPU and free memory to hold training job `j`. It returns an empty string if no such node is available.


## func scaleDryRun
```go
func scaleDryRun(r *ClusterResource, j *padv1.TrainingJob, curDiff int32, maxLoadDesired float64, scaleDown bool) (additional int)
```
`scaleDryRun` proposes to scale training job `j` by `additional` instance in addition to `curDiff`. It will modify ClusterResource `r` according to its proposal.
When scaling up, `scaleDryRun` will propose a positive number if additional resource is available and max instance number isn't violated.
When scaling up, `scaleDryRun` will propose a negative number if total requested resource exceeds total cluster resource and min instance number isn't violated.


## func scaleAllJobsDryRun
```go
func scaleAllJobsDryRun(jobs []*padv1.TrainingJob, r ClusterResource, maxLoadDesired float64) map[string]int32
```
`scaleAllJobsDryRun` repeatedly tries to scale all elastic training jobs in `jobs` until equilibrium is reached. It returns a proposal to scale the jobs as a map.


## func (*Autoscaler) setAdditional
```go
func (a *Autoscaler) setAdditional(diff map[string]int32)
```
`setAdditional` loops through `a.jobtracker` and sets `additional` field for the jobUpdater according to `diff`.


## func (*Autoscaler) Run
```go
func (a *Autoscaler) Run()
```
`Run` periodically monitors cluster resources status and sets `Additional` field for all jobUpdaters through `a.jobtracker` accordingly.


## func (*Autoscaler) findPendingJob
```go
func (a *Autoscaler) findPendingJob() bool
```
NOTE: what exactly is pending? why only master pods or pserver pods, but not both if present? why not trainer pods?<br>
NOTE: why is `cond.Type == corev1.PodScheduled && cond.Reason == corev1.PodReasonUnschedulable`?<br>
`findPendingJob` returns true if there is any training job whose master pods or pserver pods are all unschedulable.


## func (*Autoscaler) findTrainingJobsMightBeRescheduled
```go
func (a *Autoscaler) findTrainingJobsMightBeRescheduled(havePending bool)
```
`findTrainingJobsMightBeRescheduled` returns a trainingjob list. If `havePending` is true, it returns all jobs in `a.jobtracker`, else it returns 


# resource.go

## type ClusterResource
```go
type ClusterResource struct {
	NodeCount int // The total number of nodes in the cluster.

	// Each Kubernetes job could require some number of GPUs in
	// the range of [request, limit].
	GPURequest int // sum_job num_gpu_request(job)
	GPULimit   int // sum_job num_gpu_limit(job)
	GPUTotal   int // The total number of GPUs in the cluster

	// Each Kubernetes job could require some CPU timeslices in
	// the unit of *milli*.
	CPURequestMilli int64 // sum_job cpu_request_in_milli(job)
	CPULimitMilli   int64 // sum_job cpu_request_in_milli(job)
	CPUTotalMilli   int64 // The total amount of CPUs in the cluster in milli.

	// Each Kubernetes job could require some amount of memory in
	// the unit of *mega*.
	MemoryRequestMega int64 // sum_job memory_request_in_mega(job)
	MemoryLimitMega   int64 // sum_job memory_limit_in_mega(job)
	MemoryTotalMega   int64 // The total amount of memory in the cluster in mega.

	Nodes Nodes
}
```
`ClusterResource` records the cluster resource status.


## type Nodes
```go
type Nodes struct {
	NodesCPUIdleMilli   map[string]int64 // node id -> idle CPU
	NodesMemoryFreeMega map[string]int64 // node id -> free memory
}
```
`Nodes` records the amount of idle CPU and free memory of each node in the cluster.


## func getPodsTotalRequestsAndLimits
```go
func getPodsTotalRequestsAndLimits(podList *v1.PodList) (reqs v1.ResourceList, limits v1.ResourceList, err error)
```
`getPodsTotalRequestsAndLimits` accumulates resource requests and limits from all pods containers.


## func updateNodesIdleResource
```go
func updateNodesIdleResource(podList *v1.PodList, nodesCPUIdleMilli map[string]int64, nodesMemoryFreeMega map[string]int64) (err error)
```
NOTE: maybe merge this function with getPodsTotalRequestsAndLimits so that we dont need to loop the same things twice.<br>
`updateNodesIdleResource` iterates all the pods in `podlist` and substracts requested resources from `nodesCPUIdleMilli` and `nodesMemoryFreeMega`.


# `utils.go`

## func AddResourceList
```go
func AddResourceList(a v1.ResourceList, b v1.ResourceList)
```
`AddResourceList` adds ResourceList `b` to `a`.



# `trainingjob_list.go`
`trainingjob_list.go` defines `trainingjobList`, which implements [`sort.Interface`](https://golang.org/pkg/sort/#Interface).

## type trainingjobList
```go
type trainingjobList []*padv1.TrainingJob
```
`trainingjobList` is defined as `[]*padv1.TrainingJob`.

## func (trainingjobList) Len
```go
func (ts trainingjobList) Len() int
```
`Len` returns the number of elements in the trainingjobList.


## func (trainingjobList) Swap
```go
func (ts trainingjobList) Swap(a, b int)
```
`Swap` swaps the elements with indexes `a` and `b`.


## func (trainingjobList) Less
```go
func (ts trainingjobList) Less(a, b int) bool
```
NOTE: why lower demand -> needs scaling more?<br>
NOTE: why limit for GPU but request for CPU and Mem?<br>
`Less` reports whether the element with index `a` needs scaling up more than the element with index `b`.
