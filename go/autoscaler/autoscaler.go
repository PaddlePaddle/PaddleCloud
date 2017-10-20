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
	log "github.com/sirupsen/logrus"
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

	// JobRunning check if all the pods are in "Running" status.
	JobRunning(job *paddlejob.TrainingJob) (int, int, error)
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
	ticker  *time.Ticker
	cluster Cluster
	jobs    map[string]job
	eventCh chan event
}

// New creates a new Autoscaler.
func New(cluster Cluster, options ...func(*Autoscaler)) *Autoscaler {
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

func scaleDryRun(r *ClusterResource, j job, curDiff int) (additional int) {
	additionalGPUInstance := 0
	additionalCPUInstance := 0
	cpuRequestMilli := j.TrainerCPURequestMilli()
	memRequestMega := j.TrainerMemRequestMega()
	gpuLimit := j.TrainerGPULimit()

	// Adjust resource upon return.
	defer func() {
		r.GPULimit += gpuLimit * additional
		r.CPURequestMilli += cpuRequestMilli * int64(additional)
		r.MemoryRequestMega += memRequestMega * int64(additional)
	}()

	// if we are going to scale down
	if r.GPULimit > r.GPUTotal || r.CPURequestMilli > r.CPUTotalMilli {
		return -1
	}

	// TODO(helin): j.TrainerJob.Spec.Parallelism may not reflect
	// the actual pod running for the trainer job. We need to
	// count the pod manually. And calculate the additional value
	// based on the running pod count,
	// j.TrainerJob.Spec.Parallelism, and curDiff.
	plannedInstance := int(*j.TrainerJob.Spec.Parallelism) + curDiff
	instanceMax := j.Config.Spec.Trainer.MaxInstance

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

	if r.CPUTotalMilli-r.CPURequestMilli >= cpuRequestMilli {
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

func scaleAllDryRun(jobs []job, r ClusterResource) map[string]int {
	// Iteratively calculate scaling diff until nothing changes.
	diff := make(map[string]int)
	for {
		noChange := true
		sorted := sortedJobs(jobs, elastic)
		// TODO(typhoonzero): implement GPU priority CFS scheduler from here.
		for _, j := range sorted {
			name := j.Config.Name
			additional := scaleDryRun(&r, j, diff[name])
			log.Debugf("Dry run scale job %s: current %d, additional %d, remaining resource: %#v", name, diff[name], additional, r)
			diff[name] += additional

			if additional != 0 {
				noChange = false
			}
		}

		if noChange {
			break
		}
	}

	return diff
}

func (a *Autoscaler) scaleAll(diff map[string]int) {
	for name := range diff {
		if diff[name] != 0 {
			log.Infof("Scaling job %s, diff: %d.", name, diff[name])
			target := *a.jobs[name].TrainerJob.Spec.Parallelism + int32(diff[name])

			var err error
			for retry := 0; retry < 3; retry++ {
				tj, err := a.cluster.GetTrainerJob(a.jobs[name].Config)
				if err != nil {
					log.Warnf("sync trainer job error: %v, retry remaining: %d ", err, retry)
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
					log.Errorf("Error updating trainer job %v, retry remaining: %d.", err, retry)
					continue
				}

				break
			}

			if err != nil {
				log.Warnf("Error updating trainer job %v.", err)
			}
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
			log.Debugf("Event: %#v", e)
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
				delete(a.jobs, e.Job.ObjectMeta.Name)
			default:
				log.Errorln("Unrecognized event: %v.", e)
			}
		}

		r, err := a.cluster.SyncResource()
		if err != nil {
			log.Errorln("error sync resource: %v", err)
			continue
		}

		log.Infof("Latest cluster resource: %#v", r)
		var js []job
		for _, j := range a.jobs {
			// Scale jobs only when it's running (defined
			// by all pods are in the "running"
			// status). Pods are
			// pending/starting/terminating if the job is
			// just submited or just scaled up/down.
			total, running, err := a.cluster.JobRunning(j.Config)
			if err != nil {
				log.Errorln("Get if job is running failed:", err)
				continue
			}

			if total == running {
				js = append(js, j)
			}
		}
		diff := scaleAllDryRun(js, r)
		log.Infof("Scaling plan: %v", diff)
		a.scaleAll(diff)
	}
}
