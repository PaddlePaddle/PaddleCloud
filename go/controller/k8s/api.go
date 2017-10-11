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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TrainingJobs string for registration
const TrainingJobs = "TrainingJobs"

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

// TrainingJob defination
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TrainingJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TrainingJobSpec   `json:"spec"`
	Status            TrainingJobStatus `json:"status,omitempty"`
}

// TrainingJobSpec defination
type TrainingJobSpec struct {
	Trainer TrainerSpec `json:"trainer"`
	Pserver PserverSpec `json:"pserver"`
	Master  MasterSpec  `json:"master,omitempty"`
}

// TrainerSpec defination
type TrainerSpec struct {
	Entrypoint  string `json:"entrypoint"`
	Workspace   string `json:"workspace"`
	MinInstance int    `json:"min-instance"`
	MaxInstance int    `json:"max-instance"`
	// TODO(typhoonzero): Resource field from k8s API
}

// PserverSpec defination
type PserverSpec struct {
	MinInstance int `json:"min-instance"`
	MaxInstance int `json:"max-instance"`
	// TODO(typhoonzero): Resource field from k8s API
}

// MasterSpec defination
type MasterSpec struct {
	EtcdEndpoint string `json:"etcd-endpoint"`
}

// TrainingJobStatus defination
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
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TrainingJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []TrainingJob `json:"items"`
}
