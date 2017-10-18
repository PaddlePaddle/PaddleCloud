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
	GPUTotal  int
	CPUTotal  float64
	GPUFree   int
	CPUFree   float64

	MemoryTotalMi int64
	MemoryFreeMi  int64
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

func (j job) TrainerGPULimit() int {
	q := j.Config.Spec.Trainer.Resources.Limits.NvidiaGPU()
	return int(q.Value())
}

func (j job) TrainerCPURequest() float64 {
	q := j.Config.Spec.Trainer.Resources.Requests.Cpu()
	return float64(q.MilliValue()) / 1000
}

func (j job) TrainerMemRequestMi() int64 {
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
		resALimitsGPU := resA.Limits[paddlejob.GPUResourceName]
		resBLimitsGPU := resB.Limits[paddlejob.GPUResourceName]
		if resALimitsGPU.Cmp(resBLimitsGPU) == 0 {
			resARequestsCPU := *resA.Requests.Cpu()
			resBRequestsCPU := *resB.Requests.Cpu()
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
	cpuRequest := j.TrainerCPURequest()
	memRequest := j.TrainerMemRequestMi()
	gpuLimit := j.TrainerGPULimit()

	// Adjust resource upon return.
	defer func() {
		r.GPUFree -= gpuLimit * additional
		r.CPUFree -= cpuRequest * float64(additional)
	}()

	plannedInstance := int(*j.TrainerJob.Spec.Parallelism) + curDiff
	instanceMax := j.Config.Spec.Trainer.MaxInstance

	if plannedInstance >= instanceMax {
		// Do not scale or scale down, don't need to check if
		// there are available free resources.
		additional = instanceMax - plannedInstance
		return
	}

	if r.MemoryFreeMi <= memRequest {
		// insufficient memory, do not scale
		additional = 0
		return
	}

	if r.CPUFree >= cpuRequest {
		additionalCPUInstance = 1
	}

	needGPU := gpuLimit > 0
	if needGPU && r.GPUFree >= gpuLimit {
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
		for _, j := range sorted {
			name := j.Config.Name
			additional := scaleDryRun(&r, j, diff[name])
			log.Debugf("Dry run scale job %s: current %d, additional %d, remaining resource: %v", name, diff[name], additional, r)
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
			*a.jobs[name].TrainerJob.Spec.Parallelism = target
			// *a.jobs[name].TrainerJob.Spec.Completions = target
			err := a.cluster.UpdateTrainerJob(a.jobs[name].TrainerJob)
			if err != nil {
				log.Errorf("Error updating trainer job: %v", err)
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
			continue
		}

		log.Infof("Lastest cluster resource: %v", r)
		var js []job
		for _, j := range a.jobs {
			js = append(js, j)
		}
		diff := scaleAllDryRun(js, r)
		log.Infof("Scaling plan: %v", diff)
		a.scaleAll(diff)
	}
}
