# ElasticTraining

ElasticTraining currently consists of two prjects EDL and Training Job K8S Operator, and leveraging on Volcano project.

EDL: https://github.com/elasticdeeplearning/edl
Training Job K8S Operator: https://github.com/elasticdeeplearning/trainingjob-operator && https://github.com/baidu/paddle-on-k8s-operator
Volcano: https://github.com/volcano-sh/volcano

EDL with K8S Operator mainly simplifies distributed training programming. Through the ability of checkpoint, EDL can tolerate worker errors during the training process, so that training process can have flexible amount of workers. Through the serverless mode, the entire training task can be started from a relatively small number of workers. When the cluster resources are sufficient, expand the number of workers in the entire training task, shorten the time for job startup, and see the results of the first iteration as soon as possible. At the same time, the overall utilization of the cluster is improved through online/offline service joint deployment, and R&D efficiency is improved.

At the scheduler level, gang scheduling in Volcano is used to send the task as a whole, but the number of workers can be increased or decreased at any time. In this case, the training can still converge completely. EDL has been verified on Wide & Deep model and xDeepFM model.

The ability of online/offline service joint deployment is reflected in the production clusters running various online services, and it is usually necessary to set aside surplus resources to cope with the sudden increase in user requests. We hope to use these "margins" for AI training to improve cluster utilization. By running the EDL training job with a lower priority, the cluster will automatically expand the online service when user requests increase; at this time, the EDL job automatically releases resources to cooperate with the online service expansion. When the traffic peak has passed, the cluster automatically shrinks the online service. At this time, EDL automatically uses the released resources.
