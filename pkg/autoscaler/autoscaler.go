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
	"sync"
	"time"

	log "github.com/inconshreveable/log15"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/api"

	padv1 "github.com/paddleflow/elastictraining/pkg/apis/paddlepaddle/v1alpha1"
	"github.com/paddleflow/elastictraining/pkg/updater"
)

const (
	defaultLoopDur = time.Second * 30
)

// Autoscaler launches and scales the training jobs.
type Autoscaler struct {
	kubeCli        kubernetes.Interface
	jobtracker     *sync.Map
	maxLoadDesired float64
}

// WithMaxLoadDesired init with maxLoadDesired
func WithMaxLoadDesired(maxLoadDesired float64) func(as *Autoscaler) {
	return func(as *Autoscaler) {
		as.maxLoadDesired = maxLoadDesired
	}
}

// NewAutoscaler creates a new Autoscaler.
func NewAutoscaler(kubeClient kubernetes.Interface, jobtracker *sync.Map, options ...func(*Autoscaler)) *Autoscaler {
	c := &Autoscaler{
		kubeCli:        kubeClient,
		jobtracker:     jobtracker,
		maxLoadDesired: 1.0,
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// InquiryResource returns the idle and total resources of the k8s cluster.
func (a *Autoscaler) InquiryResource() (ClusterResource, error) {
	nodes := a.kubeCli.CoreV1().Nodes()
	nodeList, err := nodes.List(metav1.ListOptions{})
	if err != nil {
		return ClusterResource{}, err
	}
	allocatable := make(corev1.ResourceList)
	nodesCPUIdleMilli := make(map[string]int64)
	nodesMemoryFreeMega := make(map[string]int64)

	for _, node := range nodeList.Items {
		nodesCPUIdleMilli[node.GetObjectMeta().GetName()] =
			node.Status.Allocatable.Cpu().ScaledValue(resource.Milli)
		nodesMemoryFreeMega[node.GetObjectMeta().GetName()] =
			node.Status.Allocatable.Memory().ScaledValue(resource.Mega)
		AddResourceList(allocatable, node.Status.Allocatable)
	}

	// Get non-terminated pods from all namespaces.
	namespace := ""

	// FIXME(typhoonzero): scan all pods is not a efficient way.
	// NOTE: pending pods need to be caculated for scale down.
	// NOTE: "terminating" pods' status is still running, do not
	// scale up/down the job if job is still at last scaling
	// process.
	fieldSelector, err := fields.ParseSelector("status.phase!=" + string(api.PodSucceeded) + ",status.phase!=" + string(api.PodFailed))
	if err != nil {
		return ClusterResource{}, err
	}

	allPodsList, err := a.kubeCli.CoreV1().Pods(namespace).List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return ClusterResource{}, err
	}

	allReqs, allLimits, err := getPodsTotalRequestsAndLimits(allPodsList)
	if err != nil {
		return ClusterResource{}, err
	}

	err = updateNodesIdleResource(allPodsList, nodesCPUIdleMilli, nodesMemoryFreeMega)
	if err != nil {
		return ClusterResource{}, err
	}

	res := ClusterResource{
		NodeCount:       len(nodeList.Items),
		GPUTotal:        int(allocatable.NvidiaGPU().Value()),
		CPUTotalMilli:   allocatable.Cpu().ScaledValue(resource.Milli),
		MemoryTotalMega: allocatable.Memory().ScaledValue(resource.Mega),

		GPURequest:        int(allReqs.NvidiaGPU().Value()),
		CPURequestMilli:   allReqs.Cpu().ScaledValue(resource.Milli),
		MemoryRequestMega: allReqs.Memory().ScaledValue(resource.Mega),

		GPULimit:        int(allLimits.NvidiaGPU().Value()),
		CPULimitMilli:   allLimits.Cpu().ScaledValue(resource.Milli),
		MemoryLimitMega: allLimits.Memory().ScaledValue(resource.Mega),

		Nodes: Nodes{
			NodesCPUIdleMilli:   nodesCPUIdleMilli,
			NodesMemoryFreeMega: nodesMemoryFreeMega,
		},
	}
	return res, nil
}

// elastic job filter.
func elastic(j *padv1.TrainingJob) bool {
	return j.Elastic()
}

// gpu job filter
func needGPU(j *padv1.TrainingJob) bool {
	return j.NeedGPU()
}

// sortedJobs return the names of sorted jobs by fulfillment and
// tiebreakers in ascending order.
func sortedJobs(j []*padv1.TrainingJob, filters ...func(*padv1.TrainingJob) bool) []*padv1.TrainingJob {
	var js trainingjobList
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

func searchAssignableNode(r *ClusterResource, j *padv1.TrainingJob) string {
	for name, idle := range r.Nodes.NodesCPUIdleMilli {
		if j.TrainerCPURequestMilli() <= idle &&
			j.TrainerMemRequestMega() <= r.Nodes.NodesMemoryFreeMega[name] {
			return name
		}
	}
	return ""
}

func scaleDryRun(r *ClusterResource, j *padv1.TrainingJob, curDiff int32, maxLoadDesired float64, scaleDown bool) (additional int) {
	cpuRequestMilli := j.TrainerCPURequestMilli()
	memRequestMega := j.TrainerMemRequestMega()
	gpuLimit := j.TrainerGPULimit()
	nodeName := ""
	// Adjust resource upon return.
	defer func() {
		log.Debug("scaleDryRun", "scaledown", scaleDown, "namespace", j.Namespace, "jobname", j.Name, "additional", additional)
		r.GPULimit += gpuLimit * additional
		r.CPURequestMilli += cpuRequestMilli * int64(additional)
		r.MemoryRequestMega += memRequestMega * int64(additional)
		if nodeName != "" {
			r.Nodes.NodesCPUIdleMilli[nodeName] -= cpuRequestMilli * int64(additional)
			r.Nodes.NodesMemoryFreeMega[nodeName] -= memRequestMega * int64(additional)
		}
	}()

	// TODO(helin): j.TrainerJob.Spec.Parallelism may not reflect
	// the actual pod running for the trainer job. We need to
	// count the pod manually. And calculate the additional value
	// based on the running pod count,
	// j.TrainerJob.Spec.Parallelism, and curDiff.
	plannedInstance := int(*j.Spec.Trainer.ReplicaSpec.Spec.Parallelism) + int(curDiff)
	instanceMax := j.Spec.Trainer.MaxInstance
	instanceMin := j.Spec.Trainer.MinInstance
	log.Debug("scaleDryRun instance num", "min", instanceMin, "max", instanceMax, "planned", plannedInstance)

	// TODO(typhoonzero): refine below code to remove direction
	// ======================= scaleDown ======================
	if scaleDown {
		if plannedInstance > instanceMax {
			additional = -1
			return
		}
		gpuThreshold := int(float64(r.GPUTotal) * maxLoadDesired)
		gpuCondition := r.GPULimit > gpuThreshold
		cpuThreshold := int64(float64(r.CPUTotalMilli) * maxLoadDesired)
		cpuCondition := r.CPURequestMilli > cpuThreshold
		memThreshold := int64(float64(r.MemoryTotalMega) * maxLoadDesired)
		memCondition := r.MemoryRequestMega > memThreshold
		log.Debug("scaleDryRun", "gpuRequest", r.GPULimit, "threshold", gpuThreshold)
		log.Debug("scaleDryRun", "cpuRequest", r.CPURequestMilli, "threshold", cpuThreshold)
		log.Debug("scaleDryRun", "memRequest", r.MemoryRequestMega, "threshold", memThreshold)
		log.Debug("scaleDryRun conditions", "gpuCondition", gpuCondition, "cpuCondition", cpuCondition, "memCondition", memCondition)
		if gpuCondition || cpuCondition || memCondition {
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

	// TODO(m3ngyang) this node may not be the one that pod is assigned to in fact.
	if nodeName = searchAssignableNode(r, j); nodeName == "" {
		additional = 0
		return
	}

	// NOTE: do not scale up to use full cluster resource of CPU
	//       but we do scale up for GPU.
	additionalCPUInstance := 0
	if int64(float64(r.CPUTotalMilli)*maxLoadDesired)-r.CPURequestMilli >= cpuRequestMilli {
		additionalCPUInstance = 1
	}

	additionalGPUInstance := 0
	needGPU := gpuLimit > 0
	if needGPU && r.GPUTotal-r.GPULimit >= gpuLimit {
		additionalGPUInstance = 1
	}

	if needGPU && additionalGPUInstance < additionalCPUInstance {
		additional = additionalGPUInstance
	} else {
		additional = additionalCPUInstance
	}

	return
}

func (a *Autoscaler) setAdditional(diff map[string]int32) {
	a.jobtracker.Range(func(k, v interface{}) bool {
		key := k.(string)
		up := v.(*updater.JobUpdater)
		additional := diff[key]
		up.Additional = additional
		log.Debug("setAdditional", "jobname", key, "additional", additional)
		a.jobtracker.Store(k, up)
		return true
	})
}

// scaleAllJobsDryRun pretends to rescale all jobs in order to find
// out the number of pods should be added/deleted for each job, or
// say, delta.  It returns a map from job name to the delta.
func scaleAllJobsDryRun(jobs []*padv1.TrainingJob, r ClusterResource, maxLoadDesired float64) map[string]int32 {
	// Iteratively calculate scaling diff until nothing changes.
	diff := make(map[string]int32)
	for {
		noChange := true
		sorted := sortedJobs(jobs, elastic)
		dryRun := func(j *padv1.TrainingJob, isScaleDown bool) {
			name := j.Namespace + "/" + j.Name
			additional := scaleDryRun(&r, j, diff[name], maxLoadDesired, isScaleDown)
			diff[name] += int32(additional)

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

// Run monitors the cluster resources and training jobs in a loop,
// scales the training jobs according to the cluster resource.
func (a *Autoscaler) Run() {
	ticker := time.NewTicker(defaultLoopDur)
	defer ticker.Stop()
	log.Info("start Autoscaler")
	for {
		<-ticker.C
		r, err := a.InquiryResource()
		if err != nil {
			log.Error("InquiryResource error", err.Error())
			continue
		}
		log.Info("Cluster.InquiryResource done", "resource", r)

		havePending := a.findPendingJob()
		log.Debug("pending job", "exist", havePending)
		diff := scaleAllJobsDryRun(
			a.findTrainingJobsMightBeRescheduled(havePending),
			r,
			a.maxLoadDesired)
		log.Info("Calculated info", "diff", diff)
		a.setAdditional(diff)
	}
}

func (a *Autoscaler) findPendingJob() bool {
	havePending := false
	// TODO(m3ngyang) add a pending status for TrainingJob
	// how to define a pending job? no trainer pod is scheduled?
	traverseFunc := func(k, v interface{}) bool {
		log.Debug("Find pendingJob check", "jobname", k)
		total := 0
		pending := 0
		up, ok := v.(*updater.JobUpdater)
		if !ok {
			log.Debug("Find pendingJob conversion error", "jobname", k)
		}
		job := up.Job

		var labelKey string

		if job.Spec.FaultTolerant {
			labelKey = "paddle-job-master"
		} else {
			labelKey = "paddle-job-pserver"
		}

		pods, err := a.kubeCli.CoreV1().Pods(job.Namespace).List(metav1.ListOptions{LabelSelector: labelKey + "=" + job.Name})
		if err != nil {
			log.Error("Find pendingJob failed to list job pods", "error", err)
			return true
		}
		for _, pod := range pods.Items {
			total++
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodScheduled && cond.Reason == corev1.PodReasonUnschedulable {
					pending++
					break
				}
			}
		}
		if total == pending && total != 0 {
			log.Debug("Find pendingJob", "jobName", job.Name, "totalNum", total)
			havePending = true
			return false
		}
		return true
	}
	a.jobtracker.Range(traverseFunc)
	return havePending
}

func (a *Autoscaler) findTrainingJobsMightBeRescheduled(havePending bool) (js trainingjobList) {
	traverseFunc := func(k, v interface{}) bool {
		jn := k.(string)
		log.Debug("findTrainingJobsMightBeRescheduled", "jobname", jn)

		up, ok := v.(*updater.JobUpdater)
		if !ok {
			return false
		}
		job := up.Job
		if havePending || job.Status.Phase == padv1.TrainingJobPhaseRunning {
			js = append(js, job)
			log.Debug("job might need rescheduling", "jobname", jn)
		}
		return true
	}
	a.jobtracker.Range(traverseFunc)
	return
}
