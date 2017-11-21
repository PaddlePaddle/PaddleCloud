# 背景

Kubernetes(k8s)中有使用宿主机网络的需求，比如高性能训练的时候需要使用RDMA，或者有些用户没有容器网络来实现k8s的网络模型。在k8s中，有HostNetwork这个选项，决定pod是否使用宿主机网络。

另一方面，HostNetwork可以快速使kubernetes集群运行于RDMA网络或轻量级网络配置环境。而且我们在使用flannel和calico作为k8s底层网络时，遇到一些bug，导致paddle节点直接无法正常通信。这些问题很有可能是开源overlay网络方案自身存在的bug。采用HostNetwork的方案，可以在实现更复杂的overlay network之前作为较稳定的生产环境。

PaddlePaddle在使用k8s进行训练的时候，用户的任务可以不依赖容器网络进行训练。只要让pserver与trainer使用HostNetwork，训练任务也能正常运行。但是如果使用HostNetwork，就需要对宿主机的端口资源进行管理，如果端口分配混乱，就有可能造成pod无法启动的情况。同时用户在提交任务时，也不应该关注端口的分配，这种场景下，就需要有一个组件来实现宿主机端口分配，在PaddleCloud发起训练任务时，能为pserver和trainer分配出宿主机端口。

# 设计

## port manager

![](portmanager.jpg)

Port-manager对kubernetes apiserver进行watch与list。对于需要使用HostNetwork的任务进行端口的分配。然后更新资源的配置信息。

这里存在两个问题：

1. 并不是所有的`HostNetwork`任务都需要分配端口，port-manager不应该影响k8s集群正常的运行方式。

1. 这些真需要port-manager分配端口的任务，应该在分配完`host port`后才进行调度，执行任务流程。
    
    
对于问题1，Port-manager应该不作用于所有使用HostNetwork的job,rc,rs等资源（因为有些用户自己指定了端口）。所以，port-manager可以使用kubernetes的[annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/)功能来确定一个任务是否需要进行端口分配。对于需要使用port-manager来进行宿主机端口分配的replica set，job等，可在其metadata中增加使用port-manager的annotations。

例如：

```yaml
port-manager/hostnetwork: "true"
```


对于问题2，由于用户的训练请求是通过paddle cloud首先传递到了k8s，port-manager是对kubernetes apiserver进行watch与list后才知道有哪些任务需要分配端口。
这种情况下，就需要在paddle cloud向k8s发起请求后，创建出资源时让k8s暂时先不运行此任务，待完成端口分配后，将宿主机端口信息增加到相关资源的spec中才可以继续运行。
一种解决方式是可以通过把rs,job中相关的启动pod数量的字段置为0，而将真正的期望的数量写入annotations中：

例如rc：

```rc
annotations:
    port-manager/rc.replicas: "5"    
......
......
replicas: 0
```

job:

```
annotations:
    port-manager/parallelism: "5"    
......
......
parallelism: 0
```

在port-manager对这些任务分配完端口后，会update相关的宿主机端口，同时把真正需要的replicas、parallelism填入spec中。





