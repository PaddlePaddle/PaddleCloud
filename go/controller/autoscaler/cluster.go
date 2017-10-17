package autoscaler

import (
	"fmt"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
)

// K8sCluster resprensents a Kubernetes cluster.
type K8sCluster struct {
	clientset *kubernetes.Clientset
}

// NewK8sCluster create a new instance of K8sCluster.
func NewK8sCluster(clientset *kubernetes.Clientset) *K8sCluster {
	return &K8sCluster{
		clientset: clientset,
	}
}

// GetTrainerJob gets the trainer job spec.
func (c K8sCluster) GetTrainerJob(job *paddlejob.TrainingJob) (*batchv1.Job, error) {
	namespace := job.ObjectMeta.Namespace
	jobname := job.ObjectMeta.Name
	return c.clientset.
		BatchV1().
		Jobs(namespace).
		Get(fmt.Sprintf("%s-trainer", jobname), metav1.GetOptions{})
}

// UpdateTrainerJob updates the trainer job spec.
func (c K8sCluster) UpdateTrainerJob(job *batchv1.Job) error {
	_, err := c.clientset.BatchV1().Jobs(job.ObjectMeta.Namespace).Update(job)
	return err
}

// SyncResource will update free and total resources in k8s cluster.
func (c *K8sCluster) SyncResource() (ClusterResource, error) {
	nodes := c.clientset.CoreV1().Nodes()
	nodeList, err := nodes.List(metav1.ListOptions{})
	if err != nil {
		log.Errorln("Fetching node list error: ", err)
		return ClusterResource{}, err
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
		return ClusterResource{}, err
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
	r := ClusterResource{
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
