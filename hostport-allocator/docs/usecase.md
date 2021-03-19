# hostport-manager

This repository implements a hostport-manager for watching Foo resources as
defined with a CustomResourceDefinition/ThirdPartyResources  (CRD).

## Purpose

This is an hostport manager  to alloc/free host port for pods.

## Running



```
http:
hostport-manager --kubernetes=http://xxx/?inClusterConfig=false   --logtostderr=true --v=4  --hostport-range=10000-20000

https:
 ./hostport-manager -kubeconfig=/root/.kube/config -logtostderr=true

TODO
```


## Use Cases

CustomResourceDefinitions can be used to implement custom resource types for your Kubernetes cluster.
These act like most other Resources in Kubernetes, and may be `kubectl apply`'d, etc.

Some example use cases:

* Provisioning/Management of external datastores/databases (eg. CloudSQL/RDS instances)
* Higher level abstractions around Kubernetes primitives (eg. a single Resource to define an etcd cluster, backed by a Service and a ReplicationController)

## Defining types

Each instance of your custom resource has an attached Spec, which should be defined via a `struct{}` to provide data format validation.
In practice, this Spec is arbitrary key-value data that specifies the configuration/behavior of your Resource.

For example, if you were implementing a custom resource for a paddle job, you might provide a TrainingJob like the following:

``` go
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
```