package autoscaler

import (
	padv1 "github.com/paddleflow/paddle-operator/pkg/apis/paddlepaddle/v1alpha1"
)

type trainingjobList []*padv1.TrainingJob

func (ts trainingjobList) Len() int {
	return len(ts)
}

func (ts trainingjobList) Swap(a, b int) {
	ts[a], ts[b] = ts[b], ts[a]
}

func (ts trainingjobList) Less(a, b int) bool {
	scoreA := ts[a].Fulfillment()
	scoreB := ts[b].Fulfillment()

	if scoreA != scoreB {
		return scoreA < scoreB
	}

	resA := ts[a].Spec.Trainer.Resources
	resB := ts[b].Spec.Trainer.Resources

	resARequestsCPU := *resA.Requests.Cpu()
	resBRequestsCPU := *resB.Requests.Cpu()
	if cmpCPU := resARequestsCPU.Cmp(resBRequestsCPU); cmpCPU != 0 {
		return cmpCPU == -1
	}

	resARequestsMem := *resA.Requests.Memory()
	resBRequestsMem := *resB.Requests.Memory()
	return resARequestsMem.Cmp(resBRequestsMem) == -1
}
