package autoscaler

import (
	padv1 "github.com/paddleflow/elastictraining/pkg/apis/paddlepaddle/v1alpha1"
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

	if scoreA == scoreB {
		resA := ts[a].Spec.Trainer.Resources
		resB := ts[b].Spec.Trainer.Resources
		resALimitsGPU := *resA.Limits.NvidiaGPU()
		resBLimitsGPU := *resB.Limits.NvidiaGPU()
		if resALimitsGPU.Cmp(resBLimitsGPU) == 0 {
			resARequestsCPU := *resA.Requests.Cpu()
			resBRequestsCPU := *resB.Requests.Cpu()
			if resARequestsCPU.Cmp(resBRequestsCPU) == 0 {
				resARequestsMem := *resA.Requests.Memory()
				resBRequestsMem := *resB.Requests.Memory()
				return resARequestsMem.Cmp(resBRequestsMem) == -1
			}
			return resARequestsCPU.Cmp(resBRequestsCPU) == -1
		}
		return resALimitsGPU.Cmp(resBLimitsGPU) == -1
	}
	return scoreA < scoreB
}
