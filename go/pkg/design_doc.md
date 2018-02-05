# Design Doc for PaddlePaddle TrainingJob k8s CRD
# Goals
We are going to define PaddlePaddle cluster job using Kubernetes [CRD](https://kubernetes.io/docs/concepts/api-extension/custom-resources/). Each job is described using a yaml representation called  
TrainingJob. With CRD, we can manage the PaddlePaddle cluster job as easy as k8s primary resources. The goals including:

- Implement a controller and register a TrainingJob CRD
- Create and delete TrainingJob
- TrainingJob life cycle management

# Requirements

K8s 1.8+  and [Golang 1.9+](https://golang.org/dl/) are required for this project.

# Design
## concepts
### TrainingJob
The TrainingJob CRD defines a PaddlePaddle job resource for k8s, including cluster training parameters and a set of 
PaddlePaddle roles such as master, pserver and trainer.

- master is a k8s `ReplicaSet`
- pserver is a k8s `ReplicaSet`
- trainer is a k8s `job`

Here is an example yaml for a PaddlePaddle cluster job.
 
```yaml
apiVersion: paddlepaddle.org/v1alpha1
kind: TrainingJob
metadata:
  name: paddlejob
  namespace: testspace
spec:
  #paddle image with training script
  image: "paddlepaddle/paddlecloud-job"
  # base port of pserver
  port: 7164
  # ports num default 1
  ports_num: 1
  ports_num_for_sparse: 1
  fault_tolerant: false
  mountPath: "/home/work/namespace/"
  master:
    resources:
      limits:
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
  pserver:
    min-instance: 2
    max-instance: 2
    resources:
      limits:
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
  trainer:
    entrypoint: "python train.py"
    workspace: "/home/job-1/"
    passes: 10
    # max should equal min while fault_tolerant is disable
    min-instance: 2
    max-instance: 6
    resources:
      limits:
        cpu: "200m"
        memory: "200Mi"
      requests:
        cpu: "200m"
        memory: "200Mi"

```                                                                                         

The example will create a PaddlePaddle cluster job with a master, 2 pserver and 2 trainer.

### TrainingJober
The TrainingJober manages a specific TrainingJob, including job spec, current status and events of this job. Here is
 the struct of TrainingJober.
 
 ```go
 type trainingJobEventType string
 
 const (
 	trainingJobEventDelete trainingJobEventType = "Delete"
 	trainingJobEventModify trainingJobEventType = "Modify"
 )
 
 type trainingJobEvent struct {
 	// pet is the TrainingJobEventType of TrainingJob
 	pet trainingJobEventType
 	// The job transfer the information fo job
 	job *v1alpha1.TrainingJob
 }
 
 // TrainingJober is to manager a specific TrainingJob
 type TrainingJober struct {
 	// job is the job the TrainingJober manager.
 	job *v1alpha1.TrainingJob
 
 	// kubeCli is standard kubernetes client.
 	kubeCli kubernetes.Interface
 
 	// trainingJobClient is the client of TrainingJob.
 	trainingJobClient trainingJobClient.Interface
 
 	// status is the status in memory, update when TrainingJob status changed and update the CRD resource status.
 	status v1alpha1.TrainingJobStatus
 
 	// eventCh is the channel received by Controller, include Modify and Delete.
 	// When trainingJobEvent is Delete it will delete all resources
 	// The maximum is 1000.
 	eventCh chan *trainingJobEvent
 }
 
 ```
 When user submit a TrainingJob, controller start a TrainingJober to manage the TrainingJob. 
 - It will parser the config to TrainingJob spec including PSERVER, MASTER and TRAINER.
 - It will create resource orderly. 
 - It will sync the status of the job at regular time. While the status changed, it will update the job's status to k8s. 
 - It will release all the resource while the job succeeded or failed.
 
### Controller

The controller manages distributed PaddlePaddle jobs by creating a series of TrainingJobers. Here is the struct of 
Controller.
```go
type Controller struct {
	// KubeCli is a standard kubernetes clientset
	KubeCli kubernetes.Interface
	// ApiCli is the extension kubernetes clientset
	ApiCli apiextensionsclient.Interface
	// PaddleCli is a clientset for our own API group
	PaddleCli paddleclientset.Interface

	trainingjobLister paddlelisters.TrainingJobLister
	trainingjobSynced cache.InformerSynced

	jobtracker map[string]*trainingjober.TrainingJober

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

```

- It start to confirm that there are only one TrainingJob controller in cluster.
- It will register TrainingJob CRD to cluster if there is no TrainingJob resource.
- It will create a TrainingJober while PaddlePaddle job was submitted and delete the TrainingJober while job was 
deleted.

### Overall design

The whole life cycle of TrainingJob is managed through the two layer control of Controller and TrainingJober. As 
shown in the following figure:

![image](https://github.com/qizheng09/figure/blob/master/paddlecloud/overall_design.png?raw=true)

## State machine

When a job was submitted to cluster. Controller will start a jobber to manager the lifecycle. Here is the state 
machine of a TrainingJob. 

![image](https://github.com/qizheng09/figure/blob/master/paddlecloud/state_machine.png?raw=true)


