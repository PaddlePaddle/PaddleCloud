# Training Job Fault Tolerant

Originally from [EDL Design Doc](https://github.com/PaddlePaddle/Paddle/blob/develop/doc/v2/design/cluster_train/README.md#fault-tolerant)

The training job will pause if the coordinator server processes is dead, or any of the parameter server process is dead. They will be started by [Kubernetes](https://kubernetes.io/) and recover in few minutes. Please refer to [fault recovery](#fault-recovery).

The training job will continue to make progress if there is at least one training process running. The strategy depends on the type of optimization algorithm:

- sync-SGD

	TODO

- async-SGD

	Since async-SGD does not require synchronization between mini-batches, the system will by definition make process if at least one trainer is running.

## Fault Recovery

PaddlePaddle uses [etcd](https://github.com/coreos/etcd) to keep track of the states of processes. Because etcd is a distributed reliable key-value store, the restarted process can recover its states from etcd. The model parameters are periodically saved into distributed file system, so a restarted parameter server can recover its parameters from the saved file.


### Coordinator Server Process

When the coordinator is started by the Kubernetes, it executes the following steps at startup:

1. Grabs a unique *coordinator* lock in etcd, which prevents concurrent coordinator instantiations.
1. Recovers the task queues from etcd if they already exist, otherwise, the coordinator will create them.
1. Write its ip address to */coordinator/addr* so that trainers can discover it.
1. Listens to trainers' request of task, dispatch one upon request, and updates task queue using an etcd transaction to ensure lock is held during the update.

When the coordinator server process is dead for any reason, Kubernetes will restart it. It will be online again with all states recovered from etcd in few minutes.

### Trainer Process

When the trainer is started by the Kubernetes, it executes the following steps at startup:

1. Watches the available parameter server prefix keys `/ps/` on etcd and waits until the count of parameter servers reaches the desired count */ps_desired*.
1. Finds and watches */coordinator/addr* to get coordinator's address.
1. Requests for tasks from the coordinator to start training.

When a trainer fails, Kuberentes would try to restart it. The recovered trainer would fetch tasks from coordinator and go on training.

### Parameter Server Process

When the parameter server is started by Kubernetes, it executes the following steps at startup:

1. Read desired total number of parameter servers from etcd `/ps_desired`
1. Search through etcd keys `/ps/<index>` (`/ps/0`, `/ps/1`, ...) to find the first non-existant key whose index is smaller than the total number of parameter servers. Set the key using a transaction to avoid concurrent writes. The parameter server's index is inferred from the key name.


1. The parameter server can load parameters if there are already saved parameters in the save path (inferred from its index).
1. Now the parameter server is ready for the trainers' requests.

If the parameter server's etcd lease expires, the parameter server will kill itself.
