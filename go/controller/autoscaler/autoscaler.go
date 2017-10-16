package autoscaler

import (
	"sort"
	"time"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLoopDur = time.Second * 5
)

// Cluster represents the cluster managment system such as Kubernetes.
type Cluster interface {
	// Available free resources, must reflect the resources
	// consumed by the jobs created by SubmitJob that are still
	// pending, or the resource release by the job deletion by
	// DeleteJob that are still pending.
	FreeGPU() int
	FreeCPU() float64
	FreeMem() int64 // in Gi bytes

	Scale(*paddlejob.TrainingJob) error
	// SyncResource will sync resource values with the cluster.
	// should call this function in every tick.
	SyncResource() error
	// GetTrainerJobParallelism get current parallelism for the trainer.
	GetTrainerJobParallelism(job *paddlejob.TrainingJob) int32
}

type job struct {
	Config      *paddlejob.TrainingJob
	CurInstance int
}

func (j job) Fulfillment() float64 {
	minInstance := j.Config.Spec.Trainer.MinInstance
	maxInstance := j.Config.Spec.Trainer.MaxInstance

	if minInstance == maxInstance {
		return 1
	}

	curInstance := j.CurInstance
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
		// FIXME: use Quantity type, should refine these code.
		resALimitsGPU := resA.Limits[paddlejob.GPUResourceName]
		resBLimitsGPU := resB.Limits[paddlejob.GPUResourceName]
		if resALimitsGPU.Cmp(resALimitsGPU) == 0 {
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
	a.eventCh <- event{Type: add, Job: trainingjob}
}

// OnDel notifies the autoscaler that a job has been deleted.
func (a *Autoscaler) OnDel(trainingjob *paddlejob.TrainingJob) {
	a.eventCh <- event{Type: del, Job: trainingjob}
}

// sortedElasticJobs return the names of sorted jobs by fulfillment
// and tiebreakers in ascending order.
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

func (a *Autoscaler) dynamicScaling() {
	err := a.cluster.SyncResource()
	if err != nil {
		log.Errorln("Unable to SyncResource: ", err)
	}
	a.sortedJobs()
	// FIXME: need to determin the order/priority to scale jobs.
	// Currently: resource asc order to scale, GPU first
	log.Infoln("before scaling job: ", len(a.jobs), a)
	for jobname, j := range a.jobs {
		log.Infoln("try scaling job: ", jobname)
		a.cluster.Scale(j.Config)
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
				log.Debugln("AddJob to autoscaler: ", e.Job.ObjectMeta.Name)
				j := job{Config: e.Job, CurInstance: 0}
				// TODO: use int(a.cluster.GetTrainerJobParallelism(trainingjob))
				log.Debugln("AddJob to autoscaler1: ", e.Job.ObjectMeta.Name)
				a.jobs[e.Job.ObjectMeta.Name] = j
				log.Debugln("AddJob to autoscaler2: ", a)
			case del:
				log.Debugln("DelJob to autoscaler: ", e.Job.ObjectMeta.Name)
				delete(a.jobs, e.Job.ObjectMeta.Name)
			default:
				log.Errorln("Unrecognized event: %v.", e)
			}
		}
		a.dynamicScaling()
	}
}
