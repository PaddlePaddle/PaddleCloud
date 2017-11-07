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

package controller

import (
	"fmt"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"github.com/PaddlePaddle/cloud/go/autoscaler"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/kubernetes/pkg/api"
)

// Cluster resprensents a Kubernetes cluster.
type Cluster struct {
	clientset *kubernetes.Clientset
}

// NewCluster create a new instance of K8sCluster.
func NewCluster(clientset *kubernetes.Clientset) *Cluster {
	return &Cluster{
		clientset: clientset,
	}
}

// GetTrainerJob gets the trainer job spec.
func (c Cluster) GetTrainerJob(job *paddlejob.TrainingJob) (*batchv1.Job, error) {
	namespace := job.ObjectMeta.Namespace
	jobname := job.ObjectMeta.Name
	return c.clientset.
		BatchV1().
		Jobs(namespace).
		Get(fmt.Sprintf("%s-trainer", jobname), metav1.GetOptions{})
}

// UpdateTrainerJob updates the trainer job spec
// this will do the actual scale up/down.
func (c Cluster) UpdateTrainerJob(job *batchv1.Job) error {
	_, err := c.clientset.BatchV1().Jobs(job.ObjectMeta.Namespace).Update(job)
	return err
}

// JobPods returns the number total desired pods and the number of
// running pods of a job.
func (c Cluster) JobPods(job *paddlejob.TrainingJob) (total, running, pending int, err error) {
	if err != nil {
		return
	}
	// get pods of the job
	jobPods, err := c.clientset.CoreV1().
		Pods(job.ObjectMeta.Namespace).
		List(metav1.ListOptions{LabelSelector: "paddle-job=" + job.ObjectMeta.Name})
	for _, pod := range jobPods.Items {
		total++
		// pod.ObjectMeta.DeletionTimestamp means pod is terminating
		if pod.ObjectMeta.DeletionTimestamp == nil && pod.Status.Phase == v1.PodRunning {
			running++
		}
		if pod.ObjectMeta.DeletionTimestamp == nil && pod.Status.Phase == v1.PodPending {
			pending++
		}
	}
	return
}

// getPodsTotalRequestsAndLimits accumulate resource requests and
// limits from all pods containers.
func getPodsTotalRequestsAndLimits(podList *v1.PodList) (reqs v1.ResourceList, limits v1.ResourceList, err error) {
	reqs, limits = v1.ResourceList{}, v1.ResourceList{}
	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			AddResourceList(reqs, container.Resources.Requests)
			AddResourceList(limits, container.Resources.Limits)
		}

		for _, container := range pod.Spec.InitContainers {
			AddResourceList(reqs, container.Resources.Requests)
			AddResourceList(limits, container.Resources.Limits)
		}
	}
	return
}

func updateNodesIdleResource(podList *v1.PodList, nodesCPUIdleMilli map[string]int64, nodesMemoryFreeMega map[string]int64) (err error) {
	for _, pod := range podList.Items {
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			continue
		}
		for _, container := range pod.Spec.Containers {
			nodesCPUIdleMilli[nodeName] -= container.Resources.Requests.Cpu().ScaledValue(resource.Milli)
			nodesMemoryFreeMega[nodeName] -= container.Resources.Requests.Memory().ScaledValue(resource.Mega)
		}

		for _, container := range pod.Spec.InitContainers {
			nodesCPUIdleMilli[nodeName] -= container.Resources.Requests.Cpu().ScaledValue(resource.Milli)
			nodesMemoryFreeMega[nodeName] -= container.Resources.Requests.Memory().ScaledValue(resource.Mega)
		}
	}
	return
}

// SyncResource will update free and total resources in k8s cluster.
func (c *Cluster) SyncResource() (res autoscaler.ClusterResource, err error) {
	nodes := c.clientset.CoreV1().Nodes()
	nodeList, err := nodes.List(metav1.ListOptions{})
	if err != nil {
		return autoscaler.ClusterResource{}, err
	}
	allocatable := make(v1.ResourceList)
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
		return autoscaler.ClusterResource{}, err
	}

	allPodsList, err := c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return autoscaler.ClusterResource{}, err
	}

	allReqs, allLimits, err := getPodsTotalRequestsAndLimits(allPodsList)
	if err != nil {
		return autoscaler.ClusterResource{}, err
	}

	err = updateNodesIdleResource(allPodsList, nodesCPUIdleMilli, nodesMemoryFreeMega)
	if err != nil {
		return autoscaler.ClusterResource{}, err
	}

	res = autoscaler.ClusterResource{
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

		NodeInfos: autoscaler.NodeInfos{
			NodesCPUIdleMilli:   nodesCPUIdleMilli,
			NodesMemoryFreeMega: nodesMemoryFreeMega,
		},
	}
	return
}
