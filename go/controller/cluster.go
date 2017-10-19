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

// IsJobAllRunning check if all the pods are in "Running" status.
func (c Cluster) IsJobAllRunning(job *paddlejob.TrainingJob) bool {
	k8sjob, err := c.GetTrainerJob(job)
	if err != nil {
		return false
	}
	if k8sjob.Status.Active == *k8sjob.Spec.Parallelism {
		return true
	}
	return false
}

// getPodsTotalRequestsAndLimits accumulate resource requests and limits from all pods containers.
func getPodsTotalRequestsAndLimits(podList *v1.PodList) (reqs v1.ResourceList, limits v1.ResourceList, err error) {
	reqs, limits = v1.ResourceList{}, v1.ResourceList{}
	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			AddResourceList(&reqs, container.Resources.Requests)
			AddResourceList(&limits, container.Resources.Limits)
		}
	}
	// NOTE: Currently paddle trainer do *not* use "InitContainers", add if needed.
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
	for _, node := range nodeList.Items {
		AddResourceList(&allocatable, node.Status.Allocatable)
	}

	// get non-terminated pods from all namespaces all nodes.
	// FIXME(typhoonzero): scan all pods is not a efficient way.
	namespace := ""
	fieldSelector, err := fields.ParseSelector("status.phase!=" + string(api.PodSucceeded) + ",status.phase!=" + string(api.PodFailed))
	if err != nil {
		return autoscaler.ClusterResource{}, err
	}
	// FIXME(typhoonzero): allPodList may be large
	allPodsList, err := c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
	if err != nil {
		return autoscaler.ClusterResource{}, err
	}
	allReqs, allLimits, err := getPodsTotalRequestsAndLimits(allPodsList)
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
	}

	return
}
