# ElasticTraining

ElasticTraining currently consists of two prjects EDL and Training Job K8S Operator, and leveraging on Volcano project.

* EDL: https://github.com/elasticdeeplearning/edl

* Training Job K8S Operator: https://github.com/elasticdeeplearning/trainingjob-operator && https://github.com/baidu/paddle-on-k8s-operator

* Volcano: https://github.com/volcano-sh/volcano

EDL with K8S Operator mainly simplifies distributed training programming. Through the ability of checkpoint, EDL can tolerate worker errors during the training process, so that training process can have flexible amount of workers. Through the serverless mode, the entire training task can be started from a relatively small number of workers. When the cluster resources are sufficient, expand the number of workers in the entire training task, shorten the time for job startup, and see the results of the first iteration as soon as possible. At the same time, the overall utilization of the cluster is improved through online/offline service joint deployment, and R&D efficiency is improved.

At the scheduler level, gang scheduling in Volcano is used to send the task as a whole, but the number of workers can be increased or decreased at any time. In this case, the training can still converge completely. EDL has been verified on Wide & Deep model and xDeepFM model.

The ability of online/offline service joint deployment is reflected in the production clusters running various online services, and it is usually necessary to set aside surplus resources to cope with the sudden increase in user requests. We hope to use these "margins" for AI training to improve cluster utilization. By running the EDL training job with a lower priority, the cluster will automatically expand the online service when user requests increase; at this time, the EDL job automatically releases resources to cooperate with the online service expansion. When the traffic peak has passed, the cluster automatically shrinks the online service. At this time, EDL automatically uses the released resources.

## Use Case Project

PaddleCTR use ElasticTraining to provide the end to end CTR capabilities to the industries.

* PaddleCTR: https://github.com/PaddlePaddle/ElasticCTR

In terms of the definition of the network structure for prediction model, the PaddleCTR model library provides many cutting-edge CTR prediction models. Users can easily call these models to construct their own prediction models.

The number of samples used in prediction scenarios is generally large, ranging from several million to tens of millions. Single-machine training is very slow to meet the training efficiency of the model, and it is often necessary to accelerate model training on distributed clusters. Because the number of prediction scenes is large, dividing the resource training model for each scene separately will undoubtedly greatly increase the work of the cluster administrator. However, less resource division will affect the training speed, and too much division may cause resource waste. Therefore, the usual practice is that the model training of these niche estimation scenarios share a resource pool. However, sharing a resource pool is difficult to balance user experience and cluster resource utilization. Model training tasks for niche prediction scenarios are often more and less frequent. When there are few jobs, the resource pool is idle and waste resources; when there are many jobs, the tasks submitted later need to be queued.

EDL's elastic training can solve this problem well. Usually the resources on a cluster are shared by multiple tenants, and these tenants may run various computing tasks, such as online service tasks, data computing tasks, and so on. In order to ensure the SLO of different tenants, the cluster manager will allocate resource quotas to each tenant. Each tenant has a high priority and uses its own resource quota to perform computing tasks. If the resources in the configuration are free, other tenants can use the free resources in the tenant quota with low priority. If the original tenant's computing tasks increase during use, other tenants need to return the used resources. Since the peaks and valleys of usage of different tenants in the cluster are generally staggered, there are often idle resources in the cluster. Model training tenants can use EDL to second the idle resources of other tenants to train the model in a low-priority manner. Even if the worker of the EDL job is preempted by the original tenant during the training process, the training job will not terminate and fail. EDL will look for idle resources of other tenants in the cluster to start new workers and add the new workers to the training job.