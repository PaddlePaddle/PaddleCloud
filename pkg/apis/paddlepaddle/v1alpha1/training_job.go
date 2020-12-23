package v1alpha1

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

// Elastic returns true if the job can scale to more workers.
func (s *TrainingJob) Elastic() bool {
	return s.Spec.Trainer.MinInstance < s.Spec.Trainer.MaxInstance
}

// GPU convert Resource Limit Quantity to int
func (s *TrainingJob) GPU() int {
	q := s.Spec.Trainer.Resources.Limits.NvidiaGPU()
	gpu, ok := q.AsInt64()
	if !ok {
		// FIXME: treat errors
		gpu = 0
	}
	return int(gpu)
}

// NeedGPU returns true if the job need GPU resource to run.
func (s *TrainingJob) NeedGPU() bool {
	return s.GPU() > 0
}

// String returns marshal string of TrainingJob
func (s *TrainingJob) String() string {
	b, _ := json.MarshalIndent(s, "", "   ")
	return fmt.Sprintf("%s", b)
}

// TrainerGPULimit returns gpu limit of each trainer instance
func (s *TrainingJob) TrainerGPULimit() int {
	q := s.Spec.Trainer.Resources.Limits.NvidiaGPU()
	return int(q.Value())
}

// TrainerCPURequestMilli returns cpu request of each trainer instance
func (s *TrainingJob) TrainerCPURequestMilli() int64 {
	q := s.Spec.Trainer.Resources.Requests.Cpu()
	return q.ScaledValue(resource.Milli)
}

// TrainerMemRequestMega returns memory request of each trainer instance
func (s *TrainingJob) TrainerMemRequestMega() int64 {
	q := s.Spec.Trainer.Resources.Requests.Memory()
	return q.ScaledValue(resource.Mega)
}

// Fulfillment returns the fulfillment of a trainingjob
func (s *TrainingJob) Fulfillment() float64 {
	minInstance := s.Spec.Trainer.MinInstance
	maxInstance := s.Spec.Trainer.MaxInstance

	if minInstance == maxInstance {
		return 1
	}

	curInstance := int(*s.Spec.Trainer.ReplicaSpec.Spec.Parallelism)
	return float64(curInstance-minInstance) / float64(maxInstance-minInstance)
}
