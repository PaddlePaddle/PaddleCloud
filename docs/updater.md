# Overview

# `trainingjob_updater.go`

## type JobUpdater
```go
type JobUpdater struct {
    Job            *paddlev1.TrainingJob
    kubeCli        kubernetes.Interface
    trainingJobCli trainingJobClient.Interface
    status         paddlev1.TrainingJobStatus
    recorder       record.EventRecorder
    autoclean      bool
    Additional     int32
    restartLimit   int
    outter         bool
}
```
TOFILL
- `Job` is the training job to update.
- `kubeCli` is the client for kubernetes built-in api.
- `trainingJobCli` is the client for custom training job api.
- `status` holds the latest status of `Job`.
- `autoclean` decides whether to clean all the pods when the job terminated.


## func NewJobUpdater
```go
func NewJobUpdater(job *paddlev1.TrainingJob, kubeCli kubernetes.Interface, jobCli trainingJobClient.Interface,
                   auto bool, restartLimit int, outter bool) *JobUpdater
```
`NewJobUpdater` constructs a [`JobUpdater`](#type-jobupdater) using the arguments provided. The fields of the constructed `JobUpdater` are filled directly with the arguments, except `status` and `recorder`. `NewJobUpdater` would store a deep copy of the status of `job` in `status` field. `NewJobUpdater` would construct an [`EventRecorder`](https://godoc.org/k8s.io/client-go/tools/record#EventRecorder) with [`EventBroadcaster.NewRecorder`](https://godoc.org/k8s.io/client-go/tools/record#EventBroadcaster) and store it in `recorder` field.



## func (*JobUpdater) Update
```go
func (j *JobUpdater) Update(job *paddlev1.TrainingJob)
```
`Update` replaces `j.Job` with argument `job`.

## func (*JobUpdater) Delete
```go
func (j *JobUpdater) Delete() error
```
`Delete` is the exported version of [deleteTrainingJob](#func-(*JobUpdater)-deleteTrainingJob)



## func (*JobUpdater) Reconcile
```go
func (j *JobUpdater) Reconcile() error
```
TOFILL



## func (*JobUpdater) GetStatus
```go
func (j *JobUpdater) GetStatus() (paddlev1.TrainingJobPhase, string, error)
```
TOFILL


## func (*JobUpdater) updateCRDStatus
```go
func (j *JobUpdater) updateCRDStatus(released bool) error
```
NOTE: `released` feels redundant, since `TrainingJobStatus` already has field `Released`.<br>
`updateCRDStatus` will check if `j.status` has differed from `j.Job.Status` and if `j.IsReleased()` has differed from `released`. If anything differs, `updateCRDStatus` will update `j.Job.Status` to `j.status` and report the update to `j.trainingJobCli`.


## func (*JobUpdater) setup
```go
func (j *JobUpdater) setup() error
```
NOTE: what exactly is outter?
TOFILL


## func (*JobUpdater) createTrainingJob
```go
func (j *JobUpdater) createTrainingJob() error
```
`createTrainingJob` creates all the needed objects in the cluster.

## func (*JobUpdater) createResource
```go
func (j *JobUpdater) createResource(rt paddlev1.TrainingResourceType) error
```
`createResource` creates resource `rt` as a replicaset in the cluster via `j.kubeCli` if the resource is not already present.
It is a helper function for [createTrainingJob](#func-(*JobUpdater)-createTrainingJob).

## func (*JobUpdater) createTrainer
```go
func (j *JobUpdater) createTrainer() error
```
`createTrainer` creates trainer as a job in the cluster via `j.kubeCli` if the resource is not already present.
It is a helper function for [createTrainingJob](#func-(*JobUpdater)-createTrainingJob).



## func (*JobUpdater) deleteTrainingJob
```go
func (j *JobUpdater) deleteTrainingJob() error
```
`deleteTrainingJob` releases resources and deletes objects in the cluster.

## func (j *JobUpdater) deleteResource
```go
func (j *JobUpdater) deleteResource(rt paddlev1.TrainingResourceType) error
```
`deleteResource` deletes the `rt` replicaset in the cluster via `j.kubeCli` if the replicaset exists.
It is a helper function for [createTrainingJob](#func-(*JobUpdater)-createTrainingJob).

## func (j *JobUpdater) deleteTrainer
```go
func (j *JobUpdater) deleteTrainer() error
```
`deleteTrainer` deletes the trainer job in the cluster via `j.kubeCli` if the job exists.
It is a helper function for [createTrainingJob](#func-(*JobUpdater)-createTrainingJob).



## func (*JobUpdater) releaseMasterRoles
```go
func (j *JobUpdater) releaseMasterRoles() error
```
`releaseMasterRoles` releases the resources of pserver and master.

## func (*JobUpdater) releaseResource
```go
func (j *JobUpdater) releaseResource(rt paddlev1.TrainingResourceType) error
```
`releaseResource` sets `rt` replica to 0 and deletes all its pods.

## func (*JobUpdater) releaseTrainer
```go
func (j *JobUpdater) releaseTrainer() error
```
`releaseTrainer` sets the trainer parallelism to 0 and deletes all the trainer pods.



## func (*JobUpdater) jobTotalRunning
```go
func (j *JobUpdater) jobTotalRunning() (bool, error)
```
Return true if the status of `j.Job` matches its spec.

## func (*JobUpdater) masterRoleTotalRunning
```go
func (j *JobUpdater) masterRoleTotalRunning(rt paddlev1.TrainingResourceType) (bool, error)
```
Return true if for `rt`, the number of `ReadyReplicas` is larger than or equal to the number of `Replicas`.
It is a helper function for [jobTotalRunning](#func-(*JobUpdater)-jobTotalRunning).

## func (*JobUpdater) trainerTotalRunning
```go
func (j *JobUpdater) trainerTotalRunning() (bool, error)
```
Return true if the number of running or succeeded pods is larger than or equal to the `Parallelism` of the trainer job.
It is a helper function for [jobTotalRunning](#func-(*JobUpdater)-jobTotalRunning).


## func (*JobUpdater) findFailedTrainerPods
```go
func (j *JobUpdater) findFailedTrainerPods() ([]*corev1.Pod, error)
```
`findFailedTrainerPods` finds and returns failed trainer pods in the cluster.


## func (*JobUpdater) scale
```go
func (j *JobUpdater) scale() (err error)
```
NOTE: new version of kubernetes [improves backoff policy in JobController](https://github.com/kubernetes/kubernetes/pull/60202). BackoffLimit changes may not be needed here.<br>
`scale` adds `j.Additional` to parallelism and reports the change to `j.kubeCli`. If `j.Additional` is negative, `scale` increases the job backoffLimit by the abosolute value of `j.Additional`.


## func (*JobUpdater) initLabelOfPods
```go
func (j *JobUpdater) initLabelOfPods()
```
NOTE: Seems `IndexSucceed` should be in `status` instead of `spec`.<br>
If the labels have not been initialized, `initLabelOfPods` initializes the labels of all the trainer pods and pserver pods if applicable and sets the initialized flags.

## func (*JobUpdater) addLabelToPods
```go
func (j *JobUpdater) addLabelToPods(podType PodType) (bool, error)
```
NOTE: returned bool is redundant. Single error return if enough.<br>
FIXME: See FIXME in code.<br>
`addLabelToPods` adds an index field to the labels of the pods with `podType` through `j.kubeCli`. If the number of pods exceeds the desired number by more than 1, `addLabelToPods` will ignore the rest of the pods.
It is a helper function for [initLabelOfPods](#func-(*JobUpdater)-initLabelOfPods).



## func (*JobUpdater) traceLabelOfPods
```go
func (j *JobUpdater) traceLabelOfPods()
```
`traceLabelOfPods` tries to use unindexed pods to fill in the places of missing index pods for trainer pods and pserver pods if applicable.

## func (*JobUpdater) traceAddLabelToPods
```go
func (j *JobUpdater) traceAddLabelToPods(podType PodType) error
```
`traceAddLabelToPods` first finds all the unindexed pods of `podType` and logs their information. Then it uses the unindexed pods to fill in the places of missing indexed pods and logs all the indexed pods.
It is a helper function for [traceLabelOfPods](#func-(*JobUpdater)-traceLabelOfPods).




# `jobparser.go`


## func parseToMaster
```go
func parseToMaster(job *paddlev1.TrainingJob) *v1beta1.ReplicaSet
```
`parseToMaster` uses `job` information to construct a `ReplicaSet` for master.

# `labels.go`
