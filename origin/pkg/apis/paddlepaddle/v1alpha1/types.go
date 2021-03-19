package v1alpha1

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// CRDKind is the kind of K8s CRD.
	CRDKind = "TrainingJob"
	// CRDKindPlural is the plural of CRDKind.
	CRDKindPlural = "trainingjobs"
	// CRDShortName is the short name of CRD.
	CRDShortName = "tj"
	// CRDGroup is the name of group.
	CRDGroup = "paddlepaddle.org"
	// CRDVersion is the version of CRD.
	CRDVersion = "v1alpha1"
)

// CRDName returns name of crd
func CRDName() string {
	return fmt.Sprintf("%s.%s", CRDKindPlural, CRDGroup)
}

// +genclient
// +genclient:noStatus
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
	Image             string               `json:"image"`
	HostNetwork       bool                 `json:"host_network"`
	Port              int                  `json:"port"`
	PortsNum          int                  `json:"ports_num"`
	PortsNumForSparse int                  `json:"ports_num_for_sparse"`
	TrainerPort       int                  `json:"trainer_port"`
	TrainerPortsNum   int                  `json:"trainer_ports_num"`
	FaultTolerant     bool                 `json:"fault_tolerant"`
	LocalJob          bool                 `json:"local_job"` // LocalJob indicates if the job is local job or cluster job
	Passes            int                  `json:"passes"`
	Volumes           []corev1.Volume      `json:"volumes"`
	VolumeMounts      []corev1.VolumeMount `json:"VolumeMounts"`

	// TODO how to use these two params in matrix
	Mountpath string `json:"mountpath"`
	Nfsmount  string `json:"nfsmount"`

	Annotations Annotations `json:"annotations"`

	//TrainingJob components.
	Master  MasterSpec  `json:"master"`
	Pserver PserverSpec `json:"pserver"`
	Trainer TrainerSpec `json:"trainer"`

	IsNccl    bool       `json:"is_nccl"`
	FrameWork *Framework `json:"frame_work"`
	//Scheduling components.
	SchedulerName string `json:"schedulerName,omitempty"`
	PodGroupName  string `json:"podGroupName,omitempty"`

	// Matrix field indicates whether the backend container is matrix
	Matrix bool `json:"matrix"`
}

// MasterSpec is the spec for a master in the paddle job
type MasterSpec struct {
	EtcdEndpoint string                      `json:"etcd-endpoint"`
	Resources    corev1.ResourceRequirements `json:"resources"`
	ReplicaSpec  *appsv1.ReplicaSet          `json:"replicaSpec"`
	Envs         map[string]string           `json:"envs"`

	//for preStop
	GracePeriodSeconds *int64              `json:"grace_period_seconds"`
	PreStopCmd         []string            `json:"pre_stop_cmd"`
	Tolerations        []corev1.Toleration `json:"tolerations"`
	NodeSelector       map[string]string   `json:"node_selector"`
	LivenessProbe      *corev1.Probe       `json:"liveness_probe"`
	ReadinessProbe     *corev1.Probe       `json:"readiness_probe"`
}

// PserverSpec is the spec for pservers in the paddle job
type PserverSpec struct {
	Entrypoint       string                        `json:"entrypoint"`
	MinInstance      int                           `json:"min-instance"`
	MaxInstance      int                           `json:"max-instance"`
	Resources        corev1.ResourceRequirements   `json:"resources"`
	ReplicaSpec      *appsv1.ReplicaSet            `json:"replicaSpec"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets"`
	Envs             map[string]string             `json:"envs"`
	//for preStop
	GracePeriodSeconds *int64              `json:"grace_period_seconds"`
	PreStopCmd         []string            `json:"pre_stop_cmd"`
	Tolerations        []corev1.Toleration `json:"tolerations"`
	NodeSelector       map[string]string   `json:"node_selector"`
	//IndexSucceed marks if the operator has added labels to pservers successfully in the initial phase
	IndexSucceed   bool          `json:"index_succeed"`
	LivenessProbe  *corev1.Probe `json:"liveness_probe"`
	ReadinessProbe *corev1.Probe `json:"readiness_probe"`
}

// TrainerSpec is the spec for trainers in the paddle job
type TrainerSpec struct {
	EtcdEndpoint     string                        `json:"etcd-endpoint"`
	Entrypoint       string                        `json:"entrypoint"`
	Workspace        string                        `json:"workspace"`
	MinInstance      int                           `json:"min-instance"`
	MaxInstance      int                           `json:"max-instance"`
	Resources        corev1.ResourceRequirements   `json:"resources"`
	ReplicaSpec      *batchv1.Job                  `json:"replicaSpec"`
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets"`
	Envs             map[string]string             `json:"envs"`
	//for preStop
	GracePeriodSeconds *int64              `json:"grace_period_seconds"`
	PreStopCmd         []string            `json:"pre_stop_cmd"`
	Tolerations        []corev1.Toleration `json:"tolerations"`
	NodeSelector       map[string]string   `json:"node_selector"`
	//IndexSucceed marks if the operator has added labels to trainers successfully in the initial phase
	IndexSucceed   bool          `json:"index_succeed"`
	LivenessProbe  *corev1.Probe `json:"liveness_probe"`
	ReadinessProbe *corev1.Probe `json:"readiness_probe"`
}

// TrainingJobPhase is the phase of TrainingJob
type TrainingJobPhase string

const (
	// TrainingJobPhaseNone is empty TrainingJobPhase.
	TrainingJobPhaseNone TrainingJobPhase = ""
	// TrainingJobPhaseCreating is creating TrainingJobPhase.
	TrainingJobPhaseCreating = "Creating"
	// TrainingJobPhaseRunning is running TrainingJobPhase.
	TrainingJobPhaseRunning = "Running"
	// TrainingJobPhaseScaling is scaling TrainingJobPhase.
	TrainingJobPhaseScaling = "Scaling"
	// TrainingJobPhaseSucceeded is succeeded TrainingJobPhase.
	TrainingJobPhaseSucceeded = "Succeed"
	// TrainingJobPhaseFailed is failed TrainingJobPhase.
	TrainingJobPhaseFailed = "Failed"
	// TrainingJobPhaseTimeout is failed TrainingJobPhase.
	TrainingJobPhaseTimeout = "Timeout"
)

// ScaleResults is the result of scale
type ScaleResults string

const (
	// ScaleTrue means scale succeed.
	ScaleTrue ScaleResults = "True"
	// ScaleFalse means scale failed.
	ScaleFalse ScaleResults = "False"
	// ScaleUnknown means kubernetes can't decide if a scale succeed or not.
	ScaleUnknown ScaleResults = "Unknown"
)

// TrainerJobScaleRecord is record of trainer jobs.
type TrainerJobScaleRecord struct {
	// ScaleTimestamp is the time to scale a TrainingJob
	ScaleTimestamp metav1.Time `json:"scaleTimestamp"`
	// Additional is the additional the job to scale
	Additional int32 `json:"additional"`
	// Status is the result of the scale。
	Status ScaleResults `json:"status"`
	// reason is the reason for the scale failed.
	// +optional
	Reason string `json:"reason,omitempty"`
}

// TrainerJobScaleRecords is records of trainer jobs.
type TrainerJobScaleRecords struct {
	ScaleRecords []*TrainerJobScaleRecord `json:"scaleRecords"`
}

// TrainingResourceType the type of TrainingJob resource, include MASTER PSERVER and TRAINER
type TrainingResourceType string

const (
	// MASTER is the master name of TrainingResourceType.
	MASTER TrainingResourceType = "master"
	// PSERVER is the pserver name of TrainingResourceType.
	PSERVER TrainingResourceType = "pserver"
	// TRAINER is the trainer name of TrainingResourceType.
	TRAINER TrainingResourceType = "trainer"
)

// ResourceState is the state of a type of resource
type ResourceState string

const (
	// ResourceStateNone is the initial state of training job
	ResourceStateNone ResourceState = ""
	// ResourceStateStarting is the starting state of ResourceState.
	ResourceStateStarting = "starting"
	// ResourceStateRunning is the  running state of ResourceState.
	ResourceStateRunning = "running"
	// ResourceStateFailed is the failed state of ResourceState.
	ResourceStateFailed = "failed"
	// ResourceStateSucceeded is the succeeded state of ResourceState
	ResourceStateSucceeded = "succeeded"
)

// TrainingResourceStatus is the status of every resource
type TrainingResourceStatus struct {
	// TrainingResourceType the type of TrainingJob resource, include MASTER PSERVER and TRAINER
	TrainingResourceType `json:"training_resource_type"`
	// State is the state of a type of resource
	State ResourceState `json:"state"`
	// ResourceStates is the number of resource in different state
	ResourceStates map[ResourceState]int `json:"resource_states"`
}

// TrainingJobStatus is the status for a TrainingJob resource.
type TrainingJobStatus struct {
	// Phase is phase of TrainingJob
	Phase TrainingJobPhase `json:"phase"`
	// Reason is the reason of job phase failed
	Reason string `json:"reason"`
	// ScaleStatus is autoscale status of trainer jobs
	// TODO(ZhengQi): this will used in autoscale mode in future.
	ScaleRecords TrainerJobScaleRecords `json:"scale_records"`
	// ReplicaStatuses is detail status of resources
	// TODO(ZhengQi): should we only considered trainer job now?
	ReplicaStatuses []*TrainingResourceStatus `json:"replica_statuses"`
	// StartTime marks when the trainingjob is Running
	StartTime metav1.Time `json:"startTime"`
	// Released marks resource have been released
	Released bool `json:"released"`
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

//Annotations that offering additional metadata.
type Annotations struct {
	Usergroupid string `json:"usergroupid"`
	Userid      string `json:"userid"`
	Priority    string `json:"priority"`
	Scheduler   string `json:"scheduler"`
	Walltime    int    `json:"walltime"`
}

//Framework which operator support.
type Framework struct {
	Name FrameworkName `json:"name"`
	Type JobType       `json:"type"`
}

//FrameworkName that operator support.
type FrameworkName string

//Framework name const.
const (
	Paddle     FrameworkName = "paddle"
	TensorFlow               = "tensorflow"
)

//JobType that operator support.
type JobType string

//Job type const.
const (
	Local JobType = "local"
	Nccl2         = "nccl2"
	Multi         = "multi"
)
