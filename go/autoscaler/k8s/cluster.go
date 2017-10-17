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
package k8s

import (
	"fmt"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"github.com/PaddlePaddle/cloud/go/autoscaler"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
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

// SyncResource will update free and total resources in k8s cluster.
func (c *Cluster) SyncResource() (autoscaler.ClusterResource, error) {
	nodes := c.clientset.CoreV1().Nodes()
	nodeList, err := nodes.List(metav1.ListOptions{})
	if err != nil {
		log.Errorln("Fetching node list error: ", err)
		return autoscaler.ClusterResource{}, err
	}
	readyNodeCount := 0
	totalCPU := 0.0
	totalGPU := 0
	totalMemory := resource.NewQuantity(0, resource.BinarySI)
	requestedCPU := 0.0
	requestedGPU := 0
	requestedMemory := resource.NewQuantity(0, resource.BinarySI)
	for _, item := range nodeList.Items {
		for _, cond := range item.Status.Conditions {
			if cond.Type == v1.NodeReady && cond.Status == v1.ConditionTrue {
				readyNodeCount++
			}
		}
		for resname, q := range item.Status.Allocatable {
			if resname == v1.ResourceCPU {
				totalCPU += float64(q.MilliValue()) / 1000
			}
			if resname == v1.ResourceNvidiaGPU {
				totalGPU += int(q.Value())
			}
			if resname == v1.ResourceMemory {
				totalMemory.Add(q)
			}
		}
	}

	namespace := "" // get pods from all namespaces.
	podList, err := c.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorln("Fetching pods error: ", err)
		return autoscaler.ClusterResource{}, err
	}

	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			q := container.Resources.Requests.Cpu()
			requestedCPU += float64(q.MilliValue()) / 1000

			qGPU := container.Resources.Requests.NvidiaGPU()
			if qGPU.IsZero() {
				qGPU = container.Resources.Limits.NvidiaGPU()
			}
			requestedGPU += int(qGPU.Value())
			requestedMemory.Add(*container.Resources.Requests.Memory())
		}
	}

	totalMem := totalMemory.ScaledValue(resource.Giga)
	totalMemory.Sub(*requestedMemory)
	freeMem := totalMemory.ScaledValue(resource.Giga)
	r := autoscaler.ClusterResource{
		CPUTotal:      totalCPU,
		GPUTotal:      totalGPU,
		NodeCount:     readyNodeCount,
		CPUFree:       totalCPU - requestedCPU,
		GPUFree:       totalGPU - requestedGPU,
		MemoryTotalGi: totalMem,
		MemoryFreeGi:  freeMem,
	}

	log.Debugf("GPU: %d/%d, CPU: %f/%f, Mem: %d/%d Gi", r.GPUFree, r.GPUTotal,
		r.CPUFree, r.CPUTotal, r.MemoryFreeGi, r.MemoryTotalGi)
	return r, nil
}
