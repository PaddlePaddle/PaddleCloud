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

package edl

import (
	"fmt"

	edlresource "github.com/PaddlePaddle/cloud/go/edl/resource"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/api"
)

// ClusterResource is the resource of a cluster
type ClusterResource struct {
	NodeCount int // The total number of nodes in the cluster.

	// Each Kubernetes job could require some number of GPUs in
	// the range of [request, limit].
	GPURequest int // \sum_job num_gpu_request(job)
	GPULimit   int // \sum_job num_gpu_limit(job)
	GPUTotal   int // The total number of GPUs in the cluster

	// Each Kubernetes job could require some CPU timeslices in
	// the unit of *milli*.
	CPURequestMilli int64 // \sum_job cpu_request_in_milli(job)
	CPULimitMilli   int64 // \sum_job cpu_request_in_milli(job)
	CPUTotalMilli   int64 // The total amount of CPUs in the cluster in milli.

	// Each Kubernetes job could require some amount of memory in
	// the unit of *mega*.
	MemoryRequestMega int64 // \sum_job memory_request_in_mega(job)
	MemoryLimitMega   int64 // \sum_job memory_limit_in_mega(job)
	MemoryTotalMega   int64 // The total amount of memory in the cluster in mega.

	Nodes Nodes
}

// Nodes records the amount of idle CPU and free memory of each node
// in the cluster.
type Nodes struct {
	NodesCPUIdleMilli   map[string]int64 // node id -> idle CPU
	NodesMemoryFreeMega map[string]int64 // node id -> free memory
}

func (ns *Nodes) String() string {
	if len(ns.NodesCPUIdleMilli) != len(ns.NodesMemoryFreeMega) {
		panic("Inconsistent length in Nodes")
	}

	return fmt.Sprintf("%d Nodes", len(ns.NodesCPUIdleMilli))
}

// Cluster is our interface to the Kubernetes cluster. It can inquiry
// the cluster's overall status and the status of a specific
// PaddlePaddle trainning job.  It can also create training jobs and
// replica.
//
// TODO(yi): The above functionalities are NOT logically related with
// each other.  I am not sure if it is a good idea to group them in
// this source file.
type Cluster struct {
	clientset *kubernetes.Clientset
}

// newCluster create a new instance of K8sCluster.
func newCluster(clientset *kubernetes.Clientset) *Cluster {
	return &Cluster{
		clientset: clientset,
	}
}

// GetTrainerJob gets the trainer job spec.
func (c Cluster) GetTrainerJob(job *edlresource.TrainingJob) (*batchv1.Job, error) {
	namespace := job.ObjectMeta.Namespace
	jobname := job.ObjectMeta.Name
	return c.clientset.
		BatchV1().
		Jobs(namespace).
		Get(fmt.Sprintf("%s-trainer", jobname), metav1.GetOptions{})
}

// GetTrainerJobByName gets the trainer job spec.
func (c Cluster) GetTrainerJobByName(namespace, name string) (*batchv1.Job, error) {
	return c.clientset.
		BatchV1().
		Jobs(namespace).
		Get((name), metav1.GetOptions{})
}

// UpdateTrainerJob updates the trainer job spec
// this will do the actual scale up/down.
func (c Cluster) UpdateTrainerJob(job *batchv1.Job) error {
	_, err := c.clientset.BatchV1().Jobs(job.ObjectMeta.Namespace).Update(job)
	return err
}

// JobPods returns the number total desired pods and the number of
// running pods of a job.
func (c Cluster) JobPods(job *edlresource.TrainingJob) (total, running, pending int, err error) {
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

// InquiryResource returns the idle and total resources of the Kubernetes cluster.
func (c *Cluster) InquiryResource() (res ClusterResource, err error) {
	nodes := c.clientset.CoreV1().Nodes()
	nodeList, err := nodes.List(metav1.ListOptions{})
	if err != nil {
		return ClusterResource{}, err
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
		return ClusterResource{}, err
	}

	allPodsList, err := c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
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

	res = ClusterResource{
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
	return
}

// CreateJob creates a Job.
func (c *Cluster) CreateJob(j *batchv1.Job) (*batchv1.Job, error) {
	return c.clientset.
		BatchV1().
		Jobs(j.ObjectMeta.Namespace).
		Create(j)
}

// CreateReplicaSet creates a ReplicaSet.
func (c *Cluster) CreateReplicaSet(r *v1beta1.ReplicaSet) (*v1beta1.ReplicaSet, error) {
	return c.clientset.
		ExtensionsV1beta1().
		ReplicaSets(r.ObjectMeta.Namespace).
		Create(r)
}

// GetReplicaSet gets a ReplicaSet.
func (c *Cluster) GetReplicaSet(namespace, name string) (*v1beta1.ReplicaSet, error) {
	return c.clientset.
		ExtensionsV1beta1().
		ReplicaSets(namespace).
		Get(name, metav1.GetOptions{})
}

// DeleteTrainerJob deletes a trainerjob and their pods.
// see: https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/
func (c *Cluster) DeleteTrainerJob(namespace, name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	options := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	return c.clientset.
		BatchV1().
		Jobs(namespace).
		Delete(name, &options)
}

// DeleteReplicaSet delete a ReplicaSet and their pods.
func (c *Cluster) DeleteReplicaSet(namespace, name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	options := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	return c.clientset.
		ExtensionsV1beta1().
		ReplicaSets(namespace).
		Delete(name, &options)
}
