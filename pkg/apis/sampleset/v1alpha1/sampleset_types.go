// Copyright 2021 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// SampleSetPhase indicates whether the loading is behaving
type SampleSetPhase string

// MediumType store medium type
type MediumType string

// DriverName specified the name of csi driver
type DriverName string

type Source struct {
	// URI should be in the following format: [NAME://]BUCKET[.ENDPOINT][/PREFIX]
	// Cannot be updated after SampleSet sync data to cache engine
	// More info: https://github.com/juicedata/juicesync
	// +kubebuilder:validation:MinLength=10
	// +kubebuilder:validation:Required
	URI string `json:"uri,omitempty"`
	// If the remote storage requires additional authorization information, set this secret reference
	// +optional
	SecretRef *corev1.SecretReference `json:"secretRef,omitempty"`
}

// MountOptions the mount options for csi drivers
type MountOptions struct {
	// JuiceFSMountOptions juicefs mount command options
	// +optional
	JuiceFSMountOptions *JuiceFSMountOptions `json:"juiceFSMountOptions,omitempty"`
}

// CSI describes csi driver name and mount options to support cache data
type CSI struct {
	// Name of cache runtime driver, now only support juicefs.
	// +kubebuilder:validation:Enum=juicefs
	// +kubebuilder:default=juicefs
	// +kubebuilder:validation:Required
	Driver DriverName `json:"driver,omitempty"`
	// Namespace of the runtime object
	// +optional
	MountOptions `json:",inline,omitempty"`
}

// CacheLevel describes configurations a tier needs
type CacheLevel struct {
	// Medium Type of the tier. One of the three types: `MEM`, `SSD`, `HDD`
	// +kubebuilder:validation:Enum=MEM;SSD;HDD
	// +kubebuilder:validation:Required
	MediumType MediumType `json:"mediumType,omitempty"`
	// directory paths of local cache, use colon to separate multiple paths
	// For example: "/dev/shm/cache1:/dev/ssd/cache2:/mnt/cache3".
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Path string `json:"path,omitempty"`
	// CacheSize size of cached objects in MiB
	// If multiple paths used for this, the cache size is total amount of cache objects in all paths
	// +optional
	CacheSize int `json:"cacheSize,omitempty"`
}

// Cache used to describe how cache data store
type Cache struct {
	// configurations for multiple storage tier
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Required
	Levels []CacheLevel `json:"levels,omitempty"`
}

// CacheStatus status of cache data
type CacheStatus struct {
	// TotalSize the total size of SampleSet data
	TotalSize string `json:"totalSize,omitempty"`
	// TotalFiles the total file number of SampleSet data
	TotalFiles string `json:"totalFiles,omitempty"`
	// CachedSize the total size of cached data in all nodes
	CachedSize string `json:"cachedSize,omitempty"`
	// DiskSize disk space on file system containing cache data
	DiskSize string `json:"diskSize,omitempty"`
	// DiskUsed disk space already been used, display by command df
	DiskUsed string `json:"diskUsed,omitempty"`
	// DiskAvail disk space available on file system, display by command df
	DiskAvail string `json:"diskAvail,omitempty"`
	// DiskUsageRate disk space usage rate display by command df
	DiskUsageRate string `json:"diskUsageRate,omitempty"`
	// ErrorMassage error massages collected when executing related command
	ErrorMassage string `json:"errorMassage,omitempty"`
}

// RuntimeStatus status of runtime StatefulSet
type RuntimeStatus struct {
	// RuntimeReady is use to display SampleSet Runtime pods status, format like {ReadyReplicas}/{SpecReplicas}.
	RuntimeReady string `json:"runtimeReady,omitempty"`
	// SpecReplicas is the number of Pods should be created by Runtime StatefulSet.
	SpecReplicas int32 `json:"specReplicas,omitempty"`
	// ReadyReplicas is the number of Pods created by the Runtime StatefulSet that have a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

// JobsName record the name of jobs triggered by SampleSet controller, it should store and load atomically.
type JobsName struct {
	// the name of sync data job, used by controller to request runtime server for get job information.
	SyncJobName types.UID `json:"syncJobName,omitempty"`
	// record the name of the last done successfully sync job name
	WarmupJobName types.UID `json:"warmupJobName,omitempty"`
}

// SampleSetSpec defines the desired state of SampleSet
type SampleSetSpec struct {
	// Partitions is the number of SampleSet partitions, partition means cache node.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Required
	Partitions int32 `json:"partitions,omitempty"`
	// Source describes the information of data source uri and secret name.
	// cannot update after data sync finish
	// +optional
	Source *Source `json:"source,omitempty"`
	// SecretRef is reference to the authentication secret for source storage and cache engine.
	// cannot update after SampleSet phase is Bound
	// +kubebuilder:validation:Required
	SecretRef *corev1.SecretReference `json:"secretRef,omitempty"`
	// If the data is already in cache engine backend storage, can set NoSync as true to skip Syncing phase.
	// cannot update after data sync finish
	// +optional
	NoSync bool `json:"noSync,omitempty"`
	// CSI defines the csi driver and mount options for supporting dataset.
	// Cannot update after SampleSet phase is Bound
	// +optional
	CSI *CSI `json:"csi,omitempty"`
	// Cache options used by cache runtime engine
	// Cannot update after SampleSet phase is Bound
	// +optional
	Cache Cache `json:"cache,omitempty"`
	// NodeAffinity defines constraints that limit what nodes this SampleSet can be cached to.
	// This field influences the scheduling of pods that use the cached dataset.
	// Cannot update after SampleSet phase is Bound
	// +optional
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`
	// If specified, the pod's tolerations.
	// Cannot update after SampleSet phase is Bound
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// SampleSetStatus defines the observed state of SampleSet
type SampleSetStatus struct {
	// Dataset Phase. One of the four phases: `None`, `Bound`, `NotBound` and `Failed`
	Phase SampleSetPhase `json:"phase,omitempty"`
	// CacheStatus the status of cache data in cluster
	CacheStatus *CacheStatus `json:"cacheStatus,omitempty"`
	// RuntimeStatus the status of runtime StatefulSet pods
	RuntimeStatus *RuntimeStatus `json:"runtimeStatus,omitempty"`
	// recorde the name of jobs, all names is generated by uuid
	JobsName JobsName `json:"jobsName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="TOTAL SIZE",type="string",JSONPath=`.status.cacheStatus.totalSize`
//+kubebuilder:printcolumn:name="CACHED SIZE",type="string",JSONPath=`.status.cacheStatus.cachedSize`
//+kubebuilder:printcolumn:name="AVAIL SPACE",type="string",JSONPath=`.status.cacheStatus.diskAvail`
//+kubebuilder:printcolumn:name="Runtime",type="string",JSONPath=`.status.runtimeStatus.runtimeReady`
//+kubebuilder:printcolumn:name="PHASE",type="string",JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=`.metadata.creationTimestamp`
//+genclient

// SampleSet is the Schema for the SampleSets API
type SampleSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SampleSetSpec   `json:"spec,omitempty"`
	Status SampleSetStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SampleSetList contains a list of SampleSet
type SampleSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SampleSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SampleSet{}, &SampleSetList{})
}
