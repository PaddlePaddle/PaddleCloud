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

package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientgoapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
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
	Image   string      `json:"image"`
	Trainer TrainerSpec `json:"trainer"`
	Pserver PserverSpec `json:"pserver"`
	Master  MasterSpec  `json:"master,omitempty"`
}

// TrainerSpec defination
// +k8s:deepcopy-gen=true
type TrainerSpec struct {
	Entrypoint  string    `json:"entrypoint"`
	Workspace   string    `json:"workspace"`
	MinInstance int       `json:"min-instance"`
	MaxInstance int       `json:"max-instance"`
	Resources   Resources `json:"resources"`
}

// PserverSpec defination
// +k8s:deepcopy-gen=true
type PserverSpec struct {
	MinInstance int       `json:"min-instance"`
	MaxInstance int       `json:"max-instance"`
	Resources   Resources `json:"resources"`
}

// MasterSpec defination
// +k8s:deepcopy-gen=true
type MasterSpec struct {
	EtcdEndpoint string    `json:"etcd-endpoint"`
	Resources    Resources `json:"resources"`
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

// Resources defination
// +k8s:deepcopy-gen=true
type Resources struct {
	Limits struct {
		GPU int
		CPU float64
		Mem float64
	}
	Requests struct {
		GPU int
		CPU float64
		Mem float64
	}
}

// NeedGPU returns true if the job need GPU resource to run.
func (s *TrainingJob) NeedGPU() bool {
	return s.Spec.Trainer.Resources.Limits.GPU > 0
}

// Elastic returns true if the job can scale to more workers.
func (s *TrainingJob) Elastic() bool {
	return s.Spec.Trainer.MinInstance < s.Spec.Trainer.MaxInstance
}

// ConfigureClient will setup required field that the k8s rest client needs.
func ConfigureClient(config *rest.Config) {
	groupversion := schema.GroupVersion{
		Group:   "paddlepaddle.org",
		Version: "v1",
	}

	config.GroupVersion = &groupversion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: clientgoapi.Codecs}

	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				groupversion,
				&TrainingJob{},
				&TrainingJobList{},
				&v1.ListOptions{},
				&v1.DeleteOptions{},
			)
			return nil
		})
	schemeBuilder.AddToScheme(clientgoapi.Scheme)
}
