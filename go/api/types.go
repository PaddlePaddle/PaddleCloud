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

// sample resource:
/*
apiVersion: paddlepaddle.org/v1
kind: TrainingJob
metadata:
	name: job-1
spec:
	trainer:
		entrypoint: "python train.py"
		workspace: "/home/job-1/"
		min-instance: 3
		max-instance: 6
		resources:
			limits:
				alpha.kubernetes.io/nvidia-gpu: 1
				cpu: "800m"
				memory: "1Gi"
			requests:
				cpu: "500m"
				memory: "600Mi"
	pserver:
		min-instance: 3
		max-instance: 3
		resources:
			limits:
				cpu: "800m"
				memory: "1Gi"
			requests:
				cpu: "500m"
				memory: "600Mi"
	master:
		resources:
			limits:
				cpu: "800m"
				memory: "1Gi"
			requests:
				cpu: "500m"
				memory: "600Mi"
*/

package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

// TrainingJobs string for registration
const TrainingJobs = "TrainingJobs"

// TrainingJob defination
// +k8s:deepcopy-gen=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TrainingJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TrainingJobSpec   `json:"spec"`
	Status            TrainingJobStatus `json:"status,omitempty"`
}

// TrainingJobSpec defination
// +k8s:deepcopy-gen=true
type TrainingJobSpec struct {
	// General job attributes.
	Image             string `json:"image,omitempty"`
	Port              int    `json:"port,omitempty"`
	PortsNum          int    `json:"ports_num,omitempty"`
	PortsNumForSparse int    `json:"ports_num_for_sparse,omitempty"`
	FaultTolerant     bool   `json:"fault_tolerant,omitempty"`
	Passes            int    `json:"passes,omitempty"`
	// Job components.
	Trainer TrainerSpec `json:"trainer"`
	Pserver PserverSpec `json:"pserver"`
	Master  MasterSpec  `json:"master,omitempty"`
}

// TrainerSpec defination
// +k8s:deepcopy-gen=true
type TrainerSpec struct {
	Entrypoint  string                  `json:"entrypoint"`
	Workspace   string                  `json:"workspace"`
	MinInstance int                     `json:"min-instance"`
	MaxInstance int                     `json:"max-instance"`
	Resources   v1.ResourceRequirements `json:"resources"`
}

// PserverSpec defination
// +k8s:deepcopy-gen=true
type PserverSpec struct {
	MinInstance int                     `json:"min-instance"`
	MaxInstance int                     `json:"max-instance"`
	Resources   v1.ResourceRequirements `json:"resources"`
}

// MasterSpec defination
// +k8s:deepcopy-gen=true
type MasterSpec struct {
	EtcdEndpoint string                  `json:"etcd-endpoint"`
	Resources    v1.ResourceRequirements `json:"resources"`
}

// TrainingJobStatus defination
// +k8s:deepcopy-gen=true
type TrainingJobStatus struct {
	State   TrainingJobState `json:"state,omitempty"`
	Message string           `json:"message,omitempty"`
}

// TrainingJobState defination
type TrainingJobState string

// TrainingJobState consts
const (
	StateCreated TrainingJobState = "Created"
	StateRunning TrainingJobState = "Running"
	StateFailed  TrainingJobState = "Failed"
	StateSucceed TrainingJobState = "Succeed"
)

// TrainingJobList defination
// +k8s:deepcopy-gen=true
type TrainingJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []TrainingJob `json:"items"`
}

// NeedGPU returns true if the job need GPU resource to run.
func (s *TrainingJob) NeedGPU() bool {
	q := s.Spec.Trainer.Resources.Limits.NvidiaGPU()
	return q.CmpInt64(0) == 1
}

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
