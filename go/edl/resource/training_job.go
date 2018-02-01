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

package resource

import (
	"encoding/json"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoapi "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

// TrainingJobs string for registration
const TrainingJobs = "TrainingJobs"

// TrainingJob is a Kubernetes resource type defined by EDL.  It
// describes a PaddlePaddle training job.  As a Kubernetes resource,
//
//  - Its content must follow the Kubernetes resource definition convention.
//  - It must be a Go struct with JSON tags.
//  - It must implement the deepcopy interface.
//
// An example TrainingJob instance:
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
	Image             string           `json:"image,omitempty"`
	Port              int              `json:"port,omitempty"`
	PortsNum          int              `json:"ports_num,omitempty"`
	PortsNumForSparse int              `json:"ports_num_for_sparse,omitempty"`
	FaultTolerant     bool             `json:"fault_tolerant,omitempty"`
	Passes            int              `json:"passes,omitempty"`
	Volumes           []v1.Volume      `json:"volumes"`
	VolumeMounts      []v1.VolumeMount `json:"VolumeMounts"`
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
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TrainingJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []TrainingJob `json:"items"`
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

// NeedGPU returns true if the job need GPU resource to run.
func (s *TrainingJob) NeedGPU() bool {
	return s.GPU() > 0
}

func (s *TrainingJob) String() string {
	b, _ := json.MarshalIndent(s, "", "   ")
	return string(b[:])
}

// RegisterResource registers a resource type and the corresponding
// resource list type to the local Kubernetes runtime under group
// version "paddlepaddle.org", so the runtime could encode/decode this
// Go type.  It also change config.GroupVersion to "paddlepaddle.org".
func RegisterResource(config *rest.Config, resourceType, resourceListType runtime.Object) *rest.Config {
	groupversion := schema.GroupVersion{
		Group:   "paddlepaddle.org",
		Version: "v1",
	}

	config.GroupVersion = &groupversion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: clientgoapi.Codecs}

	clientgoapi.Scheme.AddKnownTypes(
		groupversion,
		resourceType,
		resourceListType,
		&v1.ListOptions{},
		&v1.DeleteOptions{},
	)

	return config
}
