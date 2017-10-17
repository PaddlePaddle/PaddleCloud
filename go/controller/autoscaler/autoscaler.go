package autoscaler

import (
	"sort"
	"time"

	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLoopDur = time.Second * 5
)

// ClusterResurce is the resource of a cluster
type ClusterResource struct {
	NodeCount int
	GPUTotal  int
	CPUTotal  float64
	GPUFree   int
	CPUFree   float64

	MemoryTotalGi int64
	MemoryFreeGi  int64
}

// Cluster represents the cluster managment system such as Kubernetes.
type Cluster interface {
	// SyncResource will sync resource values with the cluster.
	// should call this function in every tick.
	SyncResource() (ClusterResource, error)

	// GetTrainerJob gets the trainer job spec.
	GetTrainerJob(job *paddlejob.TrainingJob) (*batchv1.Job, error)

	// UpdateTrainerJob updates the trainer job spec.
	UpdateTrainerJob(job *batchv1.Job) error
}

type job struct {
	Config     *paddlejob.TrainingJob
	TrainerJob *batchv1.Job
}

func (j job) GPULimit() int {
	var jobGPURequest int
	for _, container := range j.TrainerJob.Spec.Template.Spec.Containers {
		qLim := container.Resources.Limits.NvidiaGPU()
		jobGPURequest += int(qLim.Value())
	}
	return jobGPURequest
}

func (j job) CPURequest() float64 {
	var jobCPURequest float64
	for _, container := range j.TrainerJob.Spec.Template.Spec.Containers {
		q := container.Resources.Requests.Cpu()
		jobCPURequest += float64(q.MilliValue()) / 1000
	}
	return jobCPURequest
}

func (j job) Fulfillment() float64 {
	minInstance := j.Config.Spec.Trainer.MinInstance
	maxInstance := j.Config.Spec.Trainer.MaxInstance

	if minInstance == maxInstance {
		return 1
	}

	curInstance := int(*j.TrainerJob.Spec.Parallelism)
	return float64(curInstance-minInstance) / float64(maxInstance-minInstance)
}

// Autoscaler launches and scales the training jobs.
type Autoscaler struct {
	ticker  *time.Ticker
	cluster Cluster
	jobs    map[string]job
	eventCh chan event
}

// NewAutoscaler creates a new Autoscaler.
func NewAutoscaler(cluster Cluster, options ...func(*Autoscaler)) *Autoscaler {
	c := &Autoscaler{
		cluster: cluster,
		ticker:  time.NewTicker(defaultLoopDur),
		jobs:    make(map[string]job),
		eventCh: make(chan event),
	}
	for _, option := range options {
		option(c)
	}
	return c
}

type jobs []job

func (j jobs) Len() int {
	return len(j)
}

func (j jobs) Less(a int, b int) bool {
	scoreA := j[a].Fulfillment()
	scoreB := j[b].Fulfillment()

	if scoreA == scoreB {
		resA := j[a].Config.Spec.Trainer.Resources
		resB := j[b].Config.Spec.Trainer.Resources
		resALimitsGPU := resA.Limits[paddlejob.GPUResourceName]
		resBLimitsGPU := resB.Limits[paddlejob.GPUResourceName]
		if resALimitsGPU.Cmp(resBLimitsGPU) == 0 {
			resARequestsCPU := resA.Requests["cpu"]
			resBRequestsCPU := resB.Requests["cpu"]
			if resARequestsCPU.Cmp(resBRequestsCPU) == 0 {
				resARequestsMem := resA.Requests["memory"]
				resBRequestsMem := resB.Requests["memory"]
				return resARequestsMem.Cmp(resBRequestsMem) == -1
			}
			return resARequestsCPU.Cmp(resBRequestsCPU) == -1
		}
		return resALimitsGPU.Cmp(resBLimitsGPU) == -1
	}
	return scoreA < scoreB
}

func (j jobs) Swap(a int, b int) {
	j[a], j[b] = j[b], j[a]
}

// elastic job filter.
func elastic(j job) bool {
	return j.Config.Elastic()
}

// gpu job filter.
func gpu(j job) bool {
	return j.Config.NeedGPU()
}

type eventType int

const (
	add eventType = iota
	del
)

type event struct {
	Type eventType
	Job  *paddlejob.TrainingJob
}

// OnAdd notifies the autoscaler that a job has been added.
func (a *Autoscaler) OnAdd(trainingjob *paddlejob.TrainingJob) {
	log.Debugln("OnAdd, adding job to event channel...", a.eventCh)
	a.eventCh <- event{Type: add, Job: trainingjob}
}

// OnDel notifies the autoscaler that a job has been deleted.
func (a *Autoscaler) OnDel(trainingjob *paddlejob.TrainingJob) {
	a.eventCh <- event{Type: del, Job: trainingjob}
}

// sortedJobs return the names of sorted jobs by fulfillment and
// tiebreakers in ascending order.
func (a *Autoscaler) sortedJobs(filters ...func(job) bool) []string {
	var js jobs
nextJob:
	for _, v := range a.jobs {
		for _, f := range filters {
			if !f(v) {
				continue nextJob
			}
		}
		js = append(js, v)
	}

	sort.Sort(js)
	var result []string
	for _, v := range js {
		result = append(result, v.Config.ObjectMeta.Name)
	}
	return result
}

func scaleDryRun(r *ClusterResource, j job, curDiff int) int {
	additionalGPUInstance := 0
	additionalCPUInstance := 0

	if r.GPUFree > j.GPULimit()*(curDiff+1) {
		additionalCPUInstance = 1
	}

	if r.CPUFree > j.CPURequest()*float64(curDiff+1) {
		additionalCPUInstance = 1
	}

	var additional int
	if additionalCPUInstance < additionalGPUInstance {
		additional = additionalCPUInstance
	} else {
		additional = additionalGPUInstance
	}

	if additional > 0 {
		// TODO(helin): consider memory request as well.
		r.GPUFree -= j.GPULimit() * additional
		r.CPUFree -= j.CPURequest() * float64(additional)
		return additional
	}

	return 0
}

func (a *Autoscaler) dynamicScaling(r ClusterResource) {
	_, err := a.cluster.SyncResource()
	if err != nil {
		log.Errorln("Unable to SyncResource: ", err)
	}

	// Iteratively calculate scaling diff until nothing changes.
	diff := make(map[string]int)
	lastDiff := make(map[string]int)
	for {
		order := a.sortedJobs(elastic)
		for _, name := range order {
			additional := scaleDryRun(&r, a.jobs[name], diff[name])
			log.Debugln("%s scale: %d, remaining resource: %v", name, additional, r)
			diff[name] += additional
		}

		noChange := true
		for key := range diff {
			if diff[key] != lastDiff[key] {
				noChange = false
			}
		}

		if noChange {
			break
		}

		for key := range diff {
			lastDiff[key] = diff[key]
		}
	}

	for name := range diff {
		if diff[name] != 0 {
			log.Infoln("Scaling job %s, diff: %d.", name, diff[name])
			target := *a.jobs[name].TrainerJob.Spec.Parallelism + int32(diff[name])
			*a.jobs[name].TrainerJob.Spec.Parallelism = target
			*a.jobs[name].TrainerJob.Spec.Completions = target
			err := a.cluster.UpdateTrainerJob(a.jobs[name].TrainerJob)
			if err != nil {
				log.Errorln("Error updating trainer job: %v", err)
			}
		}
	}
}

// Monitor monitors the cluster free resource in a loop. Do
// scale/shrink according to the cluster resource.
func (a *Autoscaler) Monitor() {
	for {
		select {
		case <-a.ticker.C:
		case e := <-a.eventCh:
			switch e.Type {
			case add:
				// TODO(helin): schedule the training
				// k8s Job. Currently we don't
				// schedule the trainer job, but only
				// scale it.
				log.Debugln("AddJob to autoscaler: ", e.Job.ObjectMeta.Name)
				var tj *batchv1.Job
				var err error
				for {
					tj, err = a.cluster.GetTrainerJob(e.Job)
					if err == nil {
						break
					}

					log.Errorln(
						"Error getting the trainer job for %s, retrying soon.",
						e.Job.ObjectMeta.Name,
					)
					time.Sleep(3 * time.Second)
				}

				j := job{
					Config:     e.Job,
					TrainerJob: tj,
				}
				a.jobs[e.Job.ObjectMeta.Name] = j
			case del:
				// TODO(helin): delete all created
				// resources (e.g., trainer Job,
				// pserver Replica Set) when we
				// schedules the resources.
				log.Debugln("DelJob to autoscaler: ", e.Job.ObjectMeta.Name)
				delete(a.jobs, e.Job.ObjectMeta.Name)
			default:
				log.Errorln("Unrecognized event: %v.", e)
			}
		}

		r, err := a.cluster.SyncResource()
		if err != nil {
			log.Errorln("error sync resource: %v", err)
		}

		a.dynamicScaling(r)
	}
}
