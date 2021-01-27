/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const KIND = "PaddleJob"

// PaddleRole defines the role of pod of a job
type PaddleRole string

const (
	PaddleRoleServer PaddleRole = "server"
	PaddleRoleWorker PaddleRole = "worker"
)

// PaddleJobMode defines the avaiable mode of a job
type PaddleJobMode string

const (
	PaddleJobModePS PaddleJobMode = "PS"

	PaddleJobModeCollective PaddleJobMode = "Collective"

	PaddleJobModeSingle PaddleJobMode = "Single"
)

// ElasticStatus defines the status of elastic process
type ElasticStatus string

const (
	ElasticStatusNone ElasticStatus = "NONE"

	ElasticStatusING ElasticStatus = "ING"

	ElasticStatusDone ElasticStatus = "DONE"

	ElasticStatusERR ElasticStatus = "ERROR"
)

// PaddleJobSpec defines the desired state of PaddleJob
type PaddleJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Server describes the spec of server base on pod template
	// PS mode is auto enabled when server is set
	// Single/Collective is enabled if server is missing
	Server RoleSpec `json:"server,omitempty"`

	// Worker describes the spec of worker base on pod template
	Worker RoleSpec `json:"worker"`
}

type RoleSpec struct {
	// Requests set the minimal replicas of server to be run
	Requests int `json:"requests"`

	// Requests set the maximal replicas of server to be run, elastic is auto enbale if limits is set larger than 0
	Limits int `json:"limits,omitempty"`

	// Template specifies the podspec of a server
	Template corev1.PodTemplateSpec `json:"template"`
}

// PaddleJobStatus defines the observed state of PaddleJob
type PaddleJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//Phase // pod phase ?

	// Mode indicates in which the PaddleJob run with : PS/Collective/Single
	Mode PaddleJobMode `json:"mode,omitempty"`

	Elastic ElasticStatus `json:"elastic,omitempty"`

	Pods []corev1.ObjectReference `json:"pods,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PaddleJob is the Schema for the paddlejobs API
type PaddleJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PaddleJobSpec   `json:"spec,omitempty"`
	Status PaddleJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PaddleJobList contains a list of PaddleJob
type PaddleJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PaddleJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PaddleJob{}, &PaddleJobList{})
}
