package autoscaler

import (
	log "github.com/inconshreveable/log15"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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

// getPodsTotalRequestsAndLimits accumulate resource requests and
// limits from all pods containers.
func getPodsTotalRequestsAndLimits(podList *v1.PodList) (reqs v1.ResourceList, limits v1.ResourceList, err error) {
	reqs, limits = v1.ResourceList{}, v1.ResourceList{}
	for _, pod := range podList.Items {
		podname := pod.Namespace + "/" + pod.Name
		log.Debug("getPodsTotalRequestsAndLimits", "podName", podname)
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
		podname := pod.Namespace + "/" + pod.Name
		log.Debug("updateNodesIdleResource", "podName", podname, "phase", pod.Status.Phase)
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
