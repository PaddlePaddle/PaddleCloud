package v1alpha1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

const (
	CRDKind       = "TraingingJob"
	CRDKindPlural = "traingingjobs"
	CRDGroup      = "paddlepaddle.org"
	CRDVersion    = "v1alpha1"
)

// CRDName returns name of crd
func CRDName() string {
	return fmt.Sprintf("%s.%s", CRDKindPlural, CRDGroup)
}

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=trainingjob

// TrainingJob is a specification for a TrainingJob resource
type TrainingJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TrainingJobSpec   `json:"spec"`
	Status            TrainingJobStatus `json:"status"`
}

// TrainingJobSpec is the spec for a TrainingJob resource
type TrainingJobSpec struct {
	// General job attributes.
	Image             string              `json:"image,omitempty"`
	Port              int                 `json:"port,omitempty"`
	PortsNum          int                 `json:"ports_num,omitempty"`
	PortsNumForSparse int                 `json:"ports_num_for_sparse,omitempty"`
	FaultTolerant     bool                `json:"fault_tolerant,omitempty"`
	Passes            int                 `json:"passes,omitempty"`
	Volumes           []apiv1.Volume      `json:"volumes"`
	VolumeMounts      []apiv1.VolumeMount `json:"VolumeMounts"`
	//TrainingJob components.
	Master  MasterSpec  `json:"master"`
	Pserver PserverSpec `json:"pserver"`
	Trainer TrainerSpec `json:"trainer"`
}

// MasterSpec is the spec for a master in the paddle job
type MasterSpec struct {
	EtcdEndpoint string                     `json:"etcd-endpoint"`
	Resources    apiv1.ResourceRequirements `json:"resources"`
}

// PserverSpec is the spec for pservers in the paddle job
type PserverSpec struct {
	MinInstance int                        `json:"min-instance"`
	MaxInstance int                        `json:"max-instance"`
	Resources   apiv1.ResourceRequirements `json:"resources"`
}

// TrainerSpec is the spec for trainers in the paddle job
type TrainerSpec struct {
	EtcdEndpoint string                     `json:"etcd-endpoint"`
	Entrypoint   string                     `json:"entrypoint"`
	Workspace    string                     `json:"workspace"`
	MinInstance  int                        `json:"min-instance"`
	MaxInstance  int                        `json:"max-instance"`
	Resources    apiv1.ResourceRequirements `json:"resources"`
}

// TrainingJobStatus is the status for a TrainingJob resource
type TrainingJobStatus struct {
	// TODO define the status of paddle training job
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +resource:path=trainingjobs

// TrainingJobList is a list of TrainingJob resources
type TrainingJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	// Items means the list of paddle job/TrainingJob
	Items []TrainingJob `json:"items"`
}
