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

// UpdateTrainerJob updates the trainer job spec.
func (c Cluster) UpdateTrainerJob(job *batchv1.Job) error {
	_, err := c.clientset.BatchV1().Jobs(job.ObjectMeta.Namespace).Update(job)
	return err
}

func podRequestsAndLimits(pod *v1.Pod) (reqs map[v1.ResourceName]resource.Quantity, limits map[v1.ResourceName]resource.Quantity, err error) {
	reqs, limits = map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}
	for _, container := range pod.Spec.Containers {
		for name, quantity := range container.Resources.Requests {
			if value, ok := reqs[name]; !ok {
				reqs[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				reqs[name] = value
			}
		}
		for name, quantity := range container.Resources.Limits {
			if value, ok := limits[name]; !ok {
				limits[name] = *quantity.Copy()
			} else {
				value.Add(quantity)
				limits[name] = value
			}
		}
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		for name, quantity := range container.Resources.Requests {
			value, ok := reqs[name]
			if !ok {
				reqs[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				reqs[name] = *quantity.Copy()
			}
		}
		for name, quantity := range container.Resources.Limits {
			value, ok := limits[name]
			if !ok {
				limits[name] = *quantity.Copy()
				continue
			}
			if quantity.Cmp(value) > 0 {
				limits[name] = *quantity.Copy()
			}
		}
	}
	return
}

func getPodsTotalRequestsAndLimits(podList *v1.PodList) (reqs map[v1.ResourceName]resource.Quantity, limits map[v1.ResourceName]resource.Quantity, err error) {
	reqs, limits = map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}
	for _, pod := range podList.Items {
		podReqs, podLimits, err := podRequestsAndLimits(&pod)
		if err != nil {
			return nil, nil, err
		}
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = *podReqValue.Copy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = *podLimitValue.Copy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
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

	allReqs := make(map[v1.ResourceName]resource.Quantity)
	allLimits := make(map[v1.ResourceName]resource.Quantity)
	allocatable := make(v1.ResourceList)
	for _, node := range nodeList.Items {
		a := node.Status.Capacity
		if len(node.Status.Allocatable) > 0 {
			a = node.Status.Allocatable
		}

		for key, item := range a {
			a := allocatable[key]
			a.Add(item)
			allocatable[key] = a
		}

		namespace := "" // get pods from all namespaces.
		var fieldSelector fields.Selector
		fieldSelector, err := fields.ParseSelector("spec.nodeName=" + node.Name + ",status.phase!=" + string(api.PodSucceeded) + ",status.phase!=" + string(api.PodFailed))
		if err != nil {
			return autoscaler.ClusterResource{}, err
		}
		nodeNonTerminatedPodsList, err := c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
		if err != nil {
			return autoscaler.ClusterResource{}, err
		}

		reqs, limits, err := getPodsTotalRequestsAndLimits(nodeNonTerminatedPodsList)
		if err != nil {
			return autoscaler.ClusterResource{}, err
		}

		for key, item := range reqs {
			a := allReqs[key]
			a.Add(item)
			allReqs[key] = a
		}

		for key, item := range limits {
			a := allLimits[key]
			a.Add(item)
			allLimits[key] = a
		}
	}

	gpuReq, gpuLimit := allReqs[v1.ResourceNvidiaGPU], allLimits[v1.ResourceNvidiaGPU]
	cpuReq, cpuLimit := allReqs[v1.ResourceCPU], allLimits[v1.ResourceCPU]
	memoryReq, memoryLimit := allReqs[v1.ResourceMemory], allLimits[v1.ResourceMemory]
	res = autoscaler.ClusterResource{
		NodeCount:         len(nodeList.Items),
		GPURequest:        int(gpuReq.Value()),
		GPULimit:          int(gpuLimit.Value()),
		GPUTotal:          int(allocatable.NvidiaGPU().Value()),
		CPURequestKilo:    float64(cpuReq.ScaledValue(resource.Kilo)),
		CPULimitKilo:      float64(cpuLimit.ScaledValue(resource.Kilo)),
		CPUTotalKilo:      float64(allocatable.Cpu().ScaledValue(resource.Kilo)),
		MemoryRequestMega: memoryReq.ScaledValue(resource.Mega),
		MemoryLimitMega:   memoryLimit.ScaledValue(resource.Mega),
		MemoryTotalMega:   allocatable.Memory().ScaledValue(resource.Mega),
	}

	return
}
