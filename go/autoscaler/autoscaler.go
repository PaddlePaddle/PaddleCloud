/* Copyright (c) 2016 PaddlePaddle Authors All Rights Reserve.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
	 limitations under the License. */

package autoscaler

import (
	"fmt"
	"sort"
	"time"

	// TODO(typhoonzero): this package still depends on k8s API, try to remove this.
	"k8s.io/apimachinery/pkg/api/resource"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	log "github.com/inconshreveable/log15"
)

const (
	defaultLoopDur = time.Second * 5
)

// ClusterResource is the resource of a cluster
type ClusterResource struct {
	NodeCount int

	GPURequest int
	GPULimit   int
	GPUTotal   int

	CPURequestMilli int64
	CPULimitMilli   int64
	CPUTotalMilli   int64

	MemoryRequestMega int64
	MemoryLimitMega   int64
	MemoryTotalMega   int64

	NodeInfos NodeInfos
}

// NodeInfos is the information of all nodes.
type NodeInfos struct {
	NodesCPUIdleMilli   map[string]int64
	NodesMemoryFreeMega map[string]int64
}

// String returns the string that represents NodeInfo when printed.
func (n NodeInfos) String() string {
	return fmt.Sprintf("NodeInfo(%d nodes)", len(n.NodesCPUIdleMilli))
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

	// JobPods returns the number total desired pods and the
	// number of running pods of a job.
	JobPods(job *paddlejob.TrainingJob) (total, running, pending int, err error)
}

type job struct {
	Config     *paddlejob.TrainingJob
	TrainerJob *batchv1.Job
}

func (j job) TrainerGPULimit() int {
	q := j.Config.Spec.Trainer.Resources.Limits.NvidiaGPU()
	return int(q.Value())
}

func (j job) TrainerCPURequestMilli() int64 {
	q := j.Config.Spec.Trainer.Resources.Requests.Cpu()
	return q.ScaledValue(resource.Milli)
}

func (j job) TrainerMemRequestMega() int64 {
	q := j.Config.Spec.Trainer.Resources.Requests.Memory()
	return q.ScaledValue(resource.Mega)
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
	ticker         *time.Ticker
	cluster        Cluster
	jobs           map[string]job
	eventCh        chan event
	maxLoadDesired float64
}

// WithMaxLoadDesired init with maxLoadDesired
func WithMaxLoadDesired(maxLoadDesired float64) func(as *Autoscaler) {
	return func(as *Autoscaler) {
		as.maxLoadDesired = maxLoadDesired
	}
}

// New creates a new Autoscaler.
func New(cluster Cluster, options ...func(*Autoscaler)) *Autoscaler {
	c := &Autoscaler{
		cluster:        cluster,
		ticker:         time.NewTicker(defaultLoopDur),
		jobs:           make(map[string]job),
		eventCh:        make(chan event),
		maxLoadDesired: 1.0,
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
		resALimitsGPU := *resA.Limits.NvidiaGPU()
		resBLimitsGPU := *resB.Limits.NvidiaGPU()
		if resALimitsGPU.Cmp(resBLimitsGPU) == 0 {
			resARequestsCPU := *resA.Requests.Cpu()
			resBRequestsCPU := *resB.Requests.Cpu()
			if resARequestsCPU.Cmp(resBRequestsCPU) == 0 {
				resARequestsMem := *resA.Requests.Memory()
				resBRequestsMem := *resB.Requests.Memory()
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
	update
)

type event struct {
	Type eventType
	Job  *paddlejob.TrainingJob
}

func (e event) GoString() string {
	return fmt.Sprintf("event{type: %d, job name: %s}", e.Type, e.Job.Name)
}

// OnAdd notifies the autoscaler that a job has been added.
func (a *Autoscaler) OnAdd(trainingjob *paddlejob.TrainingJob) {
	a.eventCh <- event{Type: add, Job: trainingjob}
}

// OnDel notifies the autoscaler that a job has been deleted.
func (a *Autoscaler) OnDel(trainingjob *paddlejob.TrainingJob) {
	a.eventCh <- event{Type: del, Job: trainingjob}
}

// OnUpdate notifies the autoscaler that a job has been deleted.
func (a *Autoscaler) OnUpdate(trainingjob *paddlejob.TrainingJob) {
	a.eventCh <- event{Type: update, Job: trainingjob}
}

// sortedJobs return the names of sorted jobs by fulfillment and
// tiebreakers in ascending order.
func sortedJobs(j []job, filters ...func(job) bool) []job {
	var js jobs
nextJob:
	for _, v := range j {
		for _, f := range filters {
			if !f(v) {
				continue nextJob
			}
		}
		js = append(js, v)
	}

	sort.Sort(js)
	return js
}

func searchAssignableNode(r *ClusterResource, j job) (nodeName string) {
	for name, idle := range r.NodeInfos.NodesCPUIdleMilli {
		if j.TrainerCPURequestMilli() <= idle &&
			j.TrainerMemRequestMega() <= r.NodeInfos.NodesMemoryFreeMega[name] {
			return name
		}
	}
	return ""
}

func scaleDryRun(r *ClusterResource, j job, curDiff int, maxLoadDesired float64, scaleDown bool) (additional int) {
	additionalGPUInstance := 0
	additionalCPUInstance := 0
	cpuRequestMilli := j.TrainerCPURequestMilli()
	memRequestMega := j.TrainerMemRequestMega()
	gpuLimit := j.TrainerGPULimit()
	nodeName := ""
	// Adjust resource upon return.
	defer func() {
		r.GPULimit += gpuLimit * additional
		r.CPURequestMilli += cpuRequestMilli * int64(additional)
		r.MemoryRequestMega += memRequestMega * int64(additional)
		if nodeName != "" {
			r.NodeInfos.NodesCPUIdleMilli[nodeName] += cpuRequestMilli * int64(additional)
			r.NodeInfos.NodesMemoryFreeMega[nodeName] += memRequestMega * int64(additional)
		}
	}()

	// TODO(helin): j.TrainerJob.Spec.Parallelism may not reflect
	// the actual pod running for the trainer job. We need to
	// count the pod manually. And calculate the additional value
	// based on the running pod count,
	// j.TrainerJob.Spec.Parallelism, and curDiff.
	plannedInstance := int(*j.TrainerJob.Spec.Parallelism) + curDiff
	instanceMax := j.Config.Spec.Trainer.MaxInstance
	instanceMin := j.Config.Spec.Trainer.MinInstance

	// TODO(typhoonzero): refine below code to remove direction
	// ======================= scaleDown ======================
	if scaleDown {
		if plannedInstance > instanceMax {
			additional = -1
			return
		}
		gpuCondition := r.GPULimit > int(float64(r.GPUTotal)*maxLoadDesired)
		cpuCondition := r.CPURequestMilli > int64(float64(r.CPUTotalMilli)*maxLoadDesired)
		if gpuCondition || cpuCondition {
			if plannedInstance > instanceMin {
				additional = -1
				return
			}

			// can not scale down further
			additional = 0
			return
		}
		// do not try to scale up
		return
	}
	// ======================= scaleUp ==========================

	if plannedInstance >= instanceMax {
		// Do not scale or scale down, don't need to check if
		// there are available free resources.
		additional = instanceMax - plannedInstance
		return
	}

	if r.MemoryTotalMega-r.MemoryRequestMega <= memRequestMega {
		// insufficient memory, do not scale
		additional = 0
		return
	}
	if nodeName = searchAssignableNode(r, j); nodeName == "" {
		additional = 0
		return
	}

	// NOTE: do not scale up to use full cluster resource of CPU
	//       but we do scale up for GPU.
	if int64(float64(r.CPUTotalMilli)*maxLoadDesired)-r.CPURequestMilli >= cpuRequestMilli {
		additionalCPUInstance = 1
	}

	needGPU := gpuLimit > 0
	if needGPU && r.GPUTotal-r.GPULimit >= gpuLimit {
		additionalGPUInstance = 1
	}

	if needGPU {
		if additionalGPUInstance < additionalCPUInstance {
			additional = additionalGPUInstance
		} else {
			additional = additionalCPUInstance
		}
	} else {
		additional = additionalCPUInstance
	}

	return
}

func scaleAllDryRun(jobs []job, r ClusterResource, maxLoadDesired float64) map[string]int {
	// Iteratively calculate scaling diff until nothing changes.
	diff := make(map[string]int)
	for {
		noChange := true
		sorted := sortedJobs(jobs, elastic)
		dryRun := func(j job, isScaleDown bool) {
			name := j.Config.Name
			additional := scaleDryRun(&r, j, diff[name], maxLoadDesired, isScaleDown)
			log.Debug(
				"dry run scale job",
				"name", name, "current scale difference", diff[name],
				"scale up number of instances", additional, "cluster resource", r,
			)
			diff[name] += additional

			if additional != 0 {
				noChange = false
			}
		}

		// TODO(typhoonzero): implement GPU priority CFS scheduler from here.

		// scale up from the ones that need scaling up the
		// most.
		for _, j := range sorted {
			dryRun(j, false)
		}

		// scale down from the ones that need scaling up the
		// least.
		for i := len(sorted) - 1; i >= 0; i-- {
			dryRun(sorted[i], true)
		}

		if noChange {
			break
		}
	}

	return diff
}

func (a *Autoscaler) scaleAll(target map[string]int) {
	for name := range target {
		log.Info("scaling job",
			"name", name, "target", target[name])
		target := int32(target[name])

		var err error
		for retry := 5; retry > 0; retry-- {
			var tj *batchv1.Job
			// don't shadow err
			tj, err = a.cluster.GetTrainerJob(a.jobs[name].Config)
			if err != nil {
				log.Warn("sync trainer job error",
					"error", err, "remaining retry", retry)
				continue
			}
			j := a.jobs[name]
			// NOTE: only update batchv1.job from k8s api-server before updating
			// for efficiency. Update the job resource to get latest k8s
			// resource reversion.
			j.TrainerJob = tj
			a.jobs[name] = j
			*a.jobs[name].TrainerJob.Spec.Parallelism = target
			err = a.cluster.UpdateTrainerJob(a.jobs[name].TrainerJob)
			if err != nil {
				log.Error("error updating trainer job",
					"error", err, "remaining retry", retry)
				continue
			}

			break
		}

		if err != nil {
			log.Warn("Error updating trainer job", "error", err)
		}
	}
}

// Monitor monitors the cluster resources and training jobs in a loop,
// scales the training jobs according to the cluster resource.
func (a *Autoscaler) Monitor() {
	for {
		select {
		case <-a.ticker.C:
		case e := <-a.eventCh:
			log.Debug("monitor received event", "event", e)
			switch e.Type {
			case add:
				fallthrough
			case update:
				// TODO(helin): schedule the training
				// k8s Job. Currently we don't
				// schedule the trainer job, but only
				// scale it.
				var tj *batchv1.Job
				var err error
				tj, err = a.cluster.GetTrainerJob(e.Job)
				if err != nil {
					log.Error(
						"Error getting the trainer k8s job, will sync later.",
						"name", e.Job.ObjectMeta.Name,
						"error", err,
					)
					j := job{
						Config:     e.Job,
						TrainerJob: nil,
					}
					a.jobs[e.Job.ObjectMeta.Name] = j
					continue
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
				delete(a.jobs, e.Job.ObjectMeta.Name)
			default:
				log.Error("unrecognized event", "event", e)
			}
		}

		r, err := a.cluster.SyncResource()
		if err != nil {
			log.Error("error sync resource", "error", err)
			continue
		}
		log.Info("sync cluster resource done", "resource", r)

		havePending := false
		for key, j := range a.jobs {
			// k8s job for trainers may not be created immediently
			// try sync it here
			if j.TrainerJob == nil {
				tj, err := a.cluster.GetTrainerJob(j.Config)
				if err != nil {
					log.Error(
						"Error getting the trainer k8s job, will sync later.",
						"name", j.Config.ObjectMeta.Name,
						"error", err,
					)
					continue
				}
				j.TrainerJob = tj
				a.jobs[key] = j
			}

			total, _, pending, err := a.cluster.JobPods(j.Config)
			if err != nil {
				log.Error("check if job is running failed", "error", err)
				continue
			}

			if total == pending {
				havePending = true
				break
			}
		}

		var js []job
		for key, j := range a.jobs {
			// k8s job for trainers may not be created immediently
			// try sync it here
			if j.TrainerJob == nil {
				tj, err := a.cluster.GetTrainerJob(j.Config)
				if err != nil {
					log.Error(
						"Error getting the trainer k8s job, will sync later.",
						"name", j.Config.ObjectMeta.Name,
						"error", err,
					)
					continue
				}
				j.TrainerJob = tj
				a.jobs[key] = j
			}

			total, running, pending, err := a.cluster.JobPods(j.Config)
			if err != nil {
				log.Error("check if job is running failed", "error", err)
				continue
			}

			log.Info(
				"job info",
				"name", key,
				"running", running,
				"pending", pending,
				"total", total,
			)

			// Scale jobs only when all pods' "Phase" are
			// running, or some job is pending.
			if total == running || havePending {
				js = append(js, j)
			}
		}
		diff := scaleAllDryRun(js, r, a.maxLoadDesired)
		target := make(map[string]int)
		for k, v := range diff {
			target[k] = int(*a.jobs[k].TrainerJob.Spec.Parallelism) + v
		}

		if len(target) > 0 {
			log.Info("calculated scaling plan", "target", target,
				"clusterResource", r)
		}
		a.scaleAll(target)
	}
}
