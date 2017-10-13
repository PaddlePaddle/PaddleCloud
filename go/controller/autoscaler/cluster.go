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

// K8sCluster is an implementation of Cluster interface to monitor
// kubernetes cluster resources.
type K8sCluster struct {
	NodeCount int
	GPUTotal  int
	CPUTotal  float64
	GPUFree   int
	CPUFree   float64

	MemoryTotalGi int64
	MemoryFreeGi  int64

	clientset *kubernetes.Clientset
}

// NewK8sCluster create a new instance of K8sCluster.
func NewK8sCluster(clientset *kubernetes.Clientset) K8sCluster {
	return K8sCluster{
		NodeCount: 0,
		GPUTotal:  0,
		CPUTotal:  0,
		clientset: clientset,
	}
}

// FreeGPU returns cluster total freeGPU card count.
func (c K8sCluster) FreeGPU() int {
	return c.GPUFree
}

// FreeCPU returns cluster total free CPU resource.
func (c K8sCluster) FreeCPU() float64 {
	return c.CPUFree
}

// FreeMem returns cluster total free memory in Gi bytes.
func (c K8sCluster) FreeMem() int64 {
	return c.MemoryFreeGi
}

// Scale one job if there's enough resource.
func (c K8sCluster) Scale(job *paddlejob.TrainingJob) error {
	namespace := job.ObjectMeta.Namespace
	jobname := job.ObjectMeta.Name
	// TODO(typhoonzero): ignore namespace quota for now, scale
	// will return error if quota is not sufficient.
	// 1. get current replicas for the job
	trainerJob, err := c.clientset.
		BatchV1().
		Jobs(namespace).
		Get(fmt.Sprintf("%s-trainer", jobname), metav1.GetOptions{})
	if err != nil {
		log.Errorln("get trainer job error: ", err)
	}

	newSize := c.scaleSizeTrainer(job, trainerJob)
	if newSize == trainerJob.Spec.Parallelism {
		log.Infoln("no need to scale: ", jobname)
		return nil
	}
	// Paddle job scale need to update both parallelism and completions
	*trainerJob.Spec.Parallelism = newSize
	*trainerJob.Spec.Completions = newSize
	// TODO: Sync status before scale?
	c.clientset.BatchV1().Jobs(namespace).Update(trainerJob)
	// TODO: wait the scale to finish
	return nil
}

func (c K8sCluster) scaleSizeTrainer(job *paddlejob.TrainingJob,
	trainerJob batchv1.Job) int32 {
	// FIXME: use static parallelism or the active pod?
	//        Some pod may already completed?
	current := int(*trainerJob.Spec.Parallelism)
	min := job.Spec.Trainer.MinInstance
	max := job.Spec.Trainer.MaxInstance
	if current < min {
		return int32(min)
	}
	if current >= max {
		return int32(max)
	}
	// job requested GPU/CPU for each pod
	jobGPURequest := 0
	jobCPURequest := 0
	for _, container := range trainerJob.Spec.Template.Spec.Containers {
		qReq := container.Resources.Requests.NvidiaGPU()
		qLim := container.Resources.Limits.NvidiaGPU()
		if qReq.IsZero() {
			jobGPURequest += int(qLim.Value())
		} else {
			jobGPURequest += int(qReq.Value())
		}
		q := container.Resources.Requests.Cpu()
		jobCPURequest += int(q.Value())
	}
	// scale the job by GPU first
	if jobGPURequest > 0 {
		if c.FreeGPU() > jobGPURequest*(max-current) {
			// can scale to max
			return int32(max)
		}
		// can add at least one pod
		if c.FreeGPU() > jobGPURequest {
			return int32(c.FreeGPU()/jobGPURequest + current)
		}
	}
	// TODO(typhoonzer): scale by CPU

	return int32(current)
}

// SyncResource will update free and total resources in k8s cluster.
func (c K8sCluster) SyncResource() error {
	nodes := c.clientset.CoreV1().Nodes()
	nodeList, err := nodes.List(metav1.ListOptions{})
	if err != nil {
		log.Errorln("Fetching node list error: ", err)
		return err
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
				totalCPU += float64(q.Value()) + float64(q.MilliValue())/1000.0
			}
			if resname == v1.ResourceNvidiaGPU {
				totalGPU += int(q.Value())
			}
			if resname == v1.ResourceMemory {
				totalMemory.Add(q)
			}
		}
	}
	c.NodeCount = readyNodeCount
	// namespace == "" will get pods from all namespaces.
	podList, err := c.clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.Errorln("Fetching pods error: ", err)
		return err
	}
	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			q := container.Resources.Requests.Cpu()
			requestedCPU += float64(q.Value()) + float64(q.MilliValue())/1000.0

			qGPU := container.Resources.Requests.NvidiaGPU()
			if qGPU.IsZero() {
				qGPU = container.Resources.Limits.NvidiaGPU()
			}
			requestedGPU += int(qGPU.Value())
			requestedMemory.Add(*container.Resources.Requests.Memory())
		}
	}
	c.CPUFree = totalCPU - requestedCPU
	c.GPUFree = totalGPU - requestedGPU
	c.MemoryTotalGi = totalMemory.ScaledValue(resource.Giga)
	totalMemory.Sub(*requestedMemory)
	c.MemoryFreeGi = totalMemory.ScaledValue(resource.Giga)
	log.Debugf("GPU: %d/%d, CPU: %f/%f, Mem: %d/%d Gi", c.GPUFree, c.GPUTotal,
		c.CPUFree, c.CPUTotal, c.MemoryFreeGi, c.MemoryTotalGi)

	return nil
}
