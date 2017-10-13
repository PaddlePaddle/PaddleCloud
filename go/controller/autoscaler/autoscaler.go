package autoscaler

import (
	"sort"
	"time"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	log "github.com/sirupsen/logrus"
)

const (
	defaultLoopDur = time.Second
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
}

type job struct {
	Config      paddlejob.TrainingJob
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
	a.sortedJobs(nil)
	// FIXME: need to determin the order/priority to scale jobs.
	// Currently: resource asc order to scale, GPU first
	for _, j := range a.jobs {
		a.cluster.Scale(&j.Config)
	}
}

// Monitor monitors the cluster free resource in a loop. Do
// scale/shrink according to the cluster resource.
func (a *Autoscaler) Monitor() {
	for {
		select {
		case <-a.ticker.C:
		}
		a.dynamicScaling()
	}
}
