# 分布式训练组件（PaddleJob Operator）概览

## 背景

[飞桨（PaddlePaddle）](https://github.com/paddlepaddle) 是全面开源开放、技术领先、功能完备、广受欢迎的国产深度学习平台。
然后，当前在 Kubernetes 上使用飞桨框架进行分布式训练并非易事，因此我们基于 Operator 机制，通过提供 PaddleJob 自定义资源为用户简化云上分布式训练的工作。

> 更多关于使用飞桨框架进行分布式训练的文档可以参考 [FleetX](https://fleet-x.readthedocs.io/en/latest/index.html)

## 目标

* 通过为用户提供 PaddleJob 自定义资源来简化在云上分布式训练的工作，并能够支持参数服务器（PS）与集合通信（Collective）两种架构模式。
* 实现 PaddleJob Controller 来自动创建依赖的资源（如 Service 和 ConfigMap等），并且能够自动管理作业，使得作业达到预期状态。

## 架构概览

分布式训练组件（PaddleJob Operator）是基于 [kuberbuilder](https://book.kubebuilder.io/) 构建的，kuberbuilder 能够自动生成框架代码，因此我们只需要关注实现 PaddleJob 的控制逻辑即可。

PaddleJob Controller 需要管理的 Kubernetes 资源相对比较简单：

* 训练作业的 Pod 会被标记为 PS(Parameter Server) 或 Worker 两种角色。
* 如果网络模式设置为 service，每个 pod 都会绑定一个指定端口的 service。
* 所有的配置信息都存储在一个 configmap 中，该 configmap 将作为 env 挂载到每个 pod。

| component | kubernets 资源 | replicas | 依赖 |
| --- | --- | --- | --- |
| PS | pods | - | 0+ | None |
| Worker | pods | 1+ | None |
| Network | service | = PS + Worker | PS, Worker, Service Mode |
| Env Config | configmap | 1 | |

PaddleJob Controller 的工作流程如下图所示：

![Workflow](../images/pd-op-reconcile.svg)

PaddleJob Controller 会先创建 PS 角色的 Pod，然后在创建 Worker Pod，如果 Intranet 网络设置为 Service 模式，则同时创建 Service 服务。

由于 Pod 的 ip 或替代信息被收集到 configmap 中，configmap 将在 Pod 分配后创建，但 Pod 在 configmap 创建好之前不会运行。


### 示例

下面的例子是使用 PaddleJob 进行 Wide & Deep 模型的训练，采用的是参数服务器架构。

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-and-deep
spec:
  withGloo: 1
  intranet: PodIP
  cleanPodPolicy: OnCompletion
  worker:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
  ps:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
```

**注意：**

* 参数 withGloo 的可选配置有0不启用、1只启动worker端、2启动所有（worker和server），默认设置为1。
* cleanPodPolicy 可以选择配置为 Always/Never/OnFailure/OnCompletion，表示在任务终止（失败或成功）时是否删除 Pod。 调试时建议 Never，生产时建议 OnCompletion；
* intranet 可以选择配置为Service/PodIP，即Pod之间的通信方式。 用户无需配置，默认使用PodIP；
* ps 和 worker 的主体内容是 podTemplateSpec。 用户可以根据 Kubernetes 规范添加更多内容，例如 GPU 配置。

同时，我们还提供了使用 GPU 来并采用集合通信（Collective）架构模式来训练 ResNet50 模型的案例：

```
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: resnet
spec:
  cleanPodPolicy: Never
  worker:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-resnet:v1
            command:
            - python
            args:
            - "-m"
            - "paddle.distributed.launch"
            - "train_fleet.py"
            volumeMounts:
            - mountPath: /dev/shm
              name: dshm
            resources:
              limits:
                nvidia.com/gpu: 1
        volumes:
        - name: dshm
          emptyDir:
            medium: Memory
```

**注意：**

* 这里需要将主机内存挂载进 Worker Pod，以防止防止内存溢出错误。
* 本示例使用内置数据集。程序启动后，数据集会被下载在容器内，可能会等待很长时间。

更多使用文档，请参考 [Paddle Operator快速上手指南](./paddlejob-tutorails.md)