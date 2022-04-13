# Controller Architecture

The controller of paddle-operator was built with kuberbuilder v3 which based on sigs.k8s.io/controller-runtime v0.7.0, 
most of the common staffs like informer, indexer, clientset are handled by the skeleton.
Therefore, the major implementation took place 
in the **Reconcile** function of PaddleJobReconciler type which located in **controllers/paddlejob_controller.go**.

The mission of developer is handling well the resources related to PaddleJob, 
reconciling the job spec with status in real time.

The PaddleJob required kubernetes resources are quite simple, 

* the workloads are presented by pods which are labeled as PS(Parameter Server) or Worker;
* if the intranet network mode is set to service, each pod will be bound to a service with specified ports;
* all the configuration information are stored in a configmap which will be mounted to every pods as env.

| component | kubernets res | replicas | dependency | 
| --- | --- | --- | --- |
| PS | pods | - | 0+ | None | 
| Worker | pods | 1+ | None |
| Network | service | = PS + Worker | PS, Worker, Service Mode |
| Env Config | configmap | 1 | | 

Briefly, the workflow are show as follows.

![Workflow](../images/pd-op-reconcile.svg)

It shows that pods are created one by one in the order of PS first, Worker follows,
services are created alongside the pods if intranet network set to service mode.

Since the ip or alternative information of pods are collected to the configmap, 
the configmap will be created after the pods allocated but the pods will not running until configmap ready.


