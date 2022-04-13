# Training Job Architecture

Originally from [EDL Design Doc](https://github.com/PaddlePaddle/Paddle/blob/develop/doc/v2/design/cluster_train/README.md#training-job)

A training job will be created once user asks Paddle cloud to train a model. The training job is made up of different processes that collaboratively consume data and produce a trained model. There are three kinds of processes:

1. the *coordinator server process*, which dispatches tasks to
1. one or more *trainer processes*, which run distributed training and synchronize gradients/models via
1. one or more *parameter server processes*, where each holds a shard of the global model, and receive the uploaded gradients from every *trainer process*, so they can run the optimize functions to update their parameters.


By coordinating these processes, PaddlePaddle supports use both Synchronize Stochastic Gradient Descent (sync SGD) and Asynchronous Stochastic Gradient Descent (async SGD) to train user-defined neural network topologies.

When training with sync SGD, parameter servers wait for all trainers to finish gradients update and then send the updated parameters to trainers, training can not proceed until the trainer received the updated parameters. This creates a synchronization point between trainers. When training with async SGD, each trainer upload gradient and download new parameters individually, without the synchronization with other trainers. Using asyc SGD will be faster in terms of time per pass, but have more noise in gradient since trainers are likely to have a stale model.

### Coordinator Server Process

The coordinator server process will:

- Partition a dataset into [tasks](#task) and dispatch tasks to trainers.
- Keep track of training progress on the dataset with [task queue](#task-queue). A training job will iterate on the dataset for a full pass until it goes into next pass.


#### Task

A task is a data shard to be trained. The total number of tasks will be much bigger than the total number of trainers. The number of data instances inside a task will be much bigger than the mini-batch size.

#### Task Queue

The coordinator server has three task queues to track training progress. 

- The todo queue holds tasks to be dispatched. When a job starts, the coordinator server fills in the todo queue with all tasks.
- The pending queue holds tasks that are currently training by trainers.
- the done queue holds tasks that are already trained.


1. When a new pass of training starts, all tasks will be placed in the todo queue.
1. Upon trainer requests for new task, the coordinator server will dispatch a task from todo queue to it, put the task in the pending queue and wait for completion.
1. The trainer will work on its task and tell the coordinator server once the task is completed and ask for new task. The coordinator server will dispatch a new task to that trainer.
1. If a task fails for any reason in trainer, or takes longer than a specific period of time,  the coordinator server will move the task back to the todo queue. The timeout count for that task will increase by one. If the timeout count is above a threshold, the task is likely to cause a trainer to crash, then it will be discarded.
1. The coordinator server will move completed task to the done queue. When the todo queue is empty, the coordinator server will start a new pass by moving all tasks in the done queue to todo queue and reset the timeout counter of all tasks to zero.

### Trainer Process

The trainer process will:

- Request tasks from the coordinator.
- Work on the tasks
- Upload gradient to parameter servers, and update local model by downloading new parameters from parameter servers.

### Parameter Server Process

Parameter server processes hold the parameters collaboratively. The parameters are partitioned on different parameter servers.

The parameter server will:

- Receive gradient from the trainers, update its parameters, and give the trainers the latest parameters.
- Periodically save its parameters to distributed file system by overriding the previous save.

### Optimization Algorithms

The communication pattern between the trainers and the parameter servers depends on the category of optimization algorithm:

- Synchronous Stochastic Gradient Descent (sync-SGD)

	Parameter server will wait for all trainer finish n-th mini-batch calculation and send their gradients before broadcasting new parameters to every trainer. Every trainer will wait for the new parameters before starting n+1-th mini-batch.

- Asynchronous Stochastic Gradient Descent (async-SGD)

	There will no synchronization between different trainers, and parameter server updates its parameter as soon as it receives new gradient:

	- Each trainer uploads its accumulated gradient every n mini-batches.
	- Every m mini-batches, the trainer downloads new parameters from parameter server.
	- n and m do not have to be equal.
