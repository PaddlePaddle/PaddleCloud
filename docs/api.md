# Api

## CRD Structure

Paddle on Kubernetes Operator uses Kubernetes CRD for specifying Paddle training job. CRD is widely used after Kubernetes 1.8+ for creating native workload on Kubernetes.

Paddle CRD structure is as follows: 

```
TrainingJob
|_TrainingJobSpec

TrainingJobSpec
|_Trainer
  |_TrainerSpec
|_Pserver
  |_PserverSpec
|_Master
  |_MasterSpec

```


## API Definition

### TrainingJobMeta
parameter | type | explanation
 --- | --- | ---
job_name | str | the unique name for the training job

### TrainingJobSpec
parameter | type | explanation
 --- | --- | ---
image|str|the paddle docker image
trainer_package | str | trainer package file path which user have the access right
port|str|open port for training job
port_num|int|number of open port
port_num_for_sparse|int|number of sparse open port
schedulerName|str|the name of selected scheduler
podGroupName|str|the name of pod group which is used in co-scheduling
mountPath|str|the path to mount into container for each trainer and pserver

### TrainerSpec
entry_point | str | entry point for startup trainer process
workspace | str | workspace in kubernetes
passes | int | training pass number
min-instance | int | min instance number for auto scale 
max-instance | int | max instance number for auto scale 
trainer_cpu|int| CPU count for each Trainer process
trainer_mem|str| memory allocated for each Trainer process, a plain integer using one of these suffixes: E, P, T, G, M, K
trainer_gpu|int| GPU count for each Trainer process, if you only want CPU, do not set this parameter

### PserverSpec
min-instance | int | min instance number for auto scale 
max-instance | int | max instance number for auto scale 
pserver_cpu|int| CPU count for each Parameter Server process
pserver_mem|str| memory allocated for each Parameter Server process, a plain integer using one of these suffixes: E, P, T, G, M, K

### MasterSpec
master_cpu|int| CPU count for each Master process
master_mem|str| memory allocated for each Master process, a plain integer using one of these suffixes: E, P, T, G, M, K
