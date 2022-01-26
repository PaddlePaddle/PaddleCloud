**模型训练组件代码：** https://github.com/PaddleFlow/paddle-operator

模型训练组件（Paddle-Operator）旨在为 Kubernetes 上运行飞桨任务提供标准化接口、训练任务管理的定制化完整支持。PaddleJob 是 Paddle-Operator 在 Kubernetes 云平台上的 CRD，并且 Paddle-Operator 定义了与其对应的 Controller。借助 PaddleJob，飞桨深度学习作业可以快速分布在 Kubernetes 集群中运行。数据科学家和机器学习工程师可以通过 PaddleJob CRD 创建飞桨深度学习作业，监控和查看深度学习的模型训练进度和状态，管理飞桨深度学习作业的生命周期。

## 功能说明

**Paddle-Operator 主要解决在 Kubernetes 上运行飞桨分布式深度学习任务的以下几个痛点：**

- 缺少针对飞桨任务的定制

- 缺少统一（将一个分布式任务作为一个整体）的资源调度
- 缺少统一任务启动
- 手动指定分布式任务的地址和端口
- 缺少统一的资源回收

**Paddle-Operator适用范围：**

- 可以无缝集成到云原生Kubernetes平台（百度云CCE、阿里云ACK、华为云CCE等公有云原生Kubernetes平台，或者私有Kubernetes平台）中单独使用，使Paddle任务在Kubernetes平台上快速便捷的运行起来。
- 可以无缝集成到基于Kubeflow生态中，支持基于Kubernetes+Kubeflow搭建的机器学习平台中集成PaddlePaddle，提升PaddlePaddle的在主流平台社区的兼容性。
- 可以配合Kubernetes系统中的第三方批量调度组件（Kube-batch、Volcano、Coscheduling等）一起使用，减少业务和平台方集成PaddlePaddle的成本。

**目前 PaddleJob 支持定义如下功能：**

 - 支持 PS/Collective/Heter 三种分布式训练模式。
 - 支持使用 Volcano 进行模型训练作业的调度。
 - 支持训练作业的弹性扩缩容并具有容错保障。

## 架构概览

下图是模型训练组件的架构流程图，描述了从 PaddleJob CRD注册，接收 PaddleJob 任务，到创建 PaddleJob 和调度 PaddleJob 的流程。

![架构流程图](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-83d85ee7f9b603f0be6ef384dd2d627a4322b1ab)

以下是 PaddleJob 进行 Reconcile 的机制：
![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-334d770898f31403b2823a9e9b82d372dc291799)

下面我们通过训练 Wide & Deep 和 ResNet50 模型来简要说明如何使用 PaddleJob 进行分布式训练。

## 安装paddle-operator
安装 paddle-operator 需要有已经安装的 kubernetes (v1.8+) 集群和 kubectl (v1.8+) 工具。本节所需配置文件和示例可以在 [paddle-operator项目](https://github.com/PaddleFlow/paddle-operator) 找到， 可以通过 git clone 或者复制文件内容保存。

```
deploy
|-- examples
|   |-- resnet.yaml
|   |-- wide_and_deep.yaml
|   |-- wide_and_deep_podip.yaml
|   |-- wide_and_deep_service.yaml
|   `-- wide_and_deep_volcano.yaml
|-- v1
|   |-- crd.yaml
|   `-- operator.yaml
`-- v1beta1
    |-- crd.yaml
    `-- operator.yaml
```

**1部署 CRD**

注意：kubernetes 1.15 及以下使用 v1beta1 目录，1.16 及以上使用目录 v1.

执行以下命令，

```bash
$ kubectl create -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/dev/deploy/v1/crd.yaml
```

或者

```bash
$ kubectl create -f deploy/v1/crd.yaml
```

注意：v1beta1 请根据报错信息添加 –validate=false 选项

通过以下命令查看是否成功，

```bash
$ kubectl get crd
NAME                                    CREATED AT
paddlejobs.batch.paddlepaddle.org       2021-02-08T07:43:24Z
```

**2部署 controller 及相关组件**
注意：默认部署的 namespace 为 paddle-system，如果希望在自定义的 namespace 中运行或者提交任务， 需要先在 operator.yaml 文件中对应更改 namespace 配置，其中

 - namespace: paddle-system 表示该资源部署的 namespace，可理解为系统 controller namespace；
 - Deployment 资源中 containers.args 中 –namespace=paddle-system 表示 controller 监控资源所在 namespace，即任务提交 namespace。

执行以下部署命令，

```bash
$ kubectl create -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/dev/deploy/v1/operator.yaml
```

或者

```bash
$ kubectl create -f deploy/v1/operator.yaml
```

通过以下命令查看部署结果和运行状态，

```bash
$ kubectl -n paddle-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
paddle-controller-manager-698dd7b855-n65jr   1/1     Running   0          1m
```

通过查看 controller 日志以确保运行正常，

```bash
$ kubectl -n paddle-system logs paddle-controller-manager-698dd7b855-n65jr
```

提交 demo 任务查看效果，

```bash
$ kubectl -n paddle-system create -f deploy/examples/wide_and_deep.yaml
```

查看 paddlejob 任务状态, pdj 为 paddlejob 的缩写，

```bash
$ kubectl -n paddle-system get pdj
NAME                     STATUS      MODE   PS    WORKER   AGE
wide-ande-deep-service   Completed   PS     2/2   0/2      4m4s
```

以上信息可以看出：训练任务已经正确完成，该任务为 ps 模式，配置需求 2 个 pserver, 2 个在运行，需求 2 个 woker，0 个在运行（已完成退出）。 可通过 cleanPodPolicy 配置任务完成/失败后的 pod 删除策略，详见任务配置。

查看 pod 状态，

```bash
$ kubectl -n paddle-system get pods
```

**3）卸载）**

通过以下命令卸载部署的组件，

```bash
$ kubectl delete -f deploy/v1/crd.yaml -f deploy/v1/operator.yaml
```

注意：重新安装时，建议先卸载再安装

## 使用示例

在上述安装过程中，我们使用了 wide-and-deep 的例子作为提交任务演示，本节详细描述任务配置和提交流程供用户参考提交自己的任务， 镜像的制作过程可在 *docker 镜像* 章节找到。

### 示例1 wide and deep

本示例采用 PS 模式，使用 cpu 进行训练，所以需要配置 ps 和 worker。


准备配置文件，

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  cleanPodPolicy: Never
  withGloo: 1
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

说明：

- 提交命名需要唯一，如果存在冲突请先删除原 paddlejob 确保已经删除再提交;
- ps 模式时需要同时配置 ps 和 worker，collective 模式时只需要配置 worker 即可；
- withGloo 可选配置为 0 不启用， 1 只启动 worker 端， 2 启动全部(worker端和Server端)， 建议设置 1；
- cleanPodPolicy 可选配置为 Always/Never/OnFailure/OnCompletion，表示任务终止（失败或成功）时，是否删除 pod，调试时建议 Never，生产时建议 OnCompletion；
- intranet 可选配置为 Service/PodIP，表示 pod 间的通信方式，用户可以不配置, 默认使用 PodIP；
- ps 和 worker 的内容为 podTemplateSpec，用户可根据需要遵从 kubernetes 规范添加更多内容, 如 GPU 的配置.

提交任务: 使用 kubectl 提交 yaml 配置文件以创建任务，

```bash
$ kubectl -n paddle-system create -f demo-wide-and-deep.yaml
```

### 示例2 resnet

本示例采用 Collective 模式，使用 gpu 进行训练，所以只需要配置 worker，且需要配置 gpu。

准备配置文件，

```yaml
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

注意：

- 这里需要添加 shared memory 挂载以防止缓存出错；
- 本示例采用内置 flower 数据集，程序启动后会进行下载，根据网络环境可能等待较长时间。

提交任务: 使用 kubectl 提交 yaml 配置文件以创建任务，

```bash
$ kubectl -n paddle-system create -f resnet.yaml
```

## 更多配置

### Volcano 支持

paddle-operator 支持使用 volcano 进行复杂任务调度，使用前请先 [安装](https://github.com/volcano-sh/volcano) 。

本节使用 volcano 实现 paddlejob 运行的 gan-scheduling。

使用此功能需要进行如下配置：

- 创建 paddlejob 同名 podgroup，具体配置信息参考 volcano 规范；
- 在 paddlejob 任务配置中添加声明：schedulerName: volcano , 注意：需要且只需要在 worker 中配置。

配置示例，

```yaml
---
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  cleanPodPolicy: Never
  withGloo: 1
  worker:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        schedulerName: volcano
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
  ps:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1

---
apiVersion: scheduling.volcano.sh/v1beta1
kind: PodGroup
metadata:
  name: wide-ande-deep
spec:
  minMember: 4
```

在以上配置中，我们通过创建最小调度单元为 4 的 podgroup，并将 paddlejob 任务标记使用 volcano 调度，实现了任务的 gan-scheduling。

可以通过以下命运提交上述任务查看结果，

```bash
$ kubectl -n paddle-system create -f deploy/examples/wide_and_deep.yaml
```

### GPU 和节点选择

更多配置示例，

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  intranet: Service
  cleanPodPolicy: OnCompletion
  worker:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
            resources:
              limits:
                nvidia.com/gpu: 1
        nodeSelector:
          accelerator: nvidia-tesla-p100
  ps:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
            resources:
              limits:
                nvidia.com/gpu: 1
        nodeSelector:
          accelerator: nvidia-tesla-p100
```

### 数据存储

在 kubernentes 中使用挂载存储建议使用 pv/pvc 配置，详见 [persistent-volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) 。

这里使用 nfs 云盘作为存储作为示例，配置文件如下，

```
$ cat pv-pvc.yaml
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Recycle
  storageClassName: slow
  mountOptions:
    - hard
    - nfsvers=4.1
  nfs:
    path: /nas
    server: 10.12.201.xx

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-pvc
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 10Gi
  storageClassName: slow
  volumeName: nfs-pv
```

使用以下命令在 namespace paddle-system 中 创建 pvc 名为 nfs-pvc 的存储声明，实际引用为 10.12.201.xx 上的 nfs 存储。

```
$ kubectl -n paddle-system apply -f pv-pvc.yaml
```

注意 pvc 需要绑定 namespace 且只能在该 namespace 下使用。

提交 paddlejob 任务时，配置 volumes 引用以使用对应存储，

```
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: paddlejob-demo-1
spec:
  cleanPolicy: OnCompletion
  worker:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/paddle-ubuntu:2.0.0-18.04
            command: ["bash","-c"]
            args: ["cd /nas/wide_and_deep; python3 train.py"]
            volumeMounts:
            - mountPath: /nas
              name: data
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: nfs-pvc
  ps:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/paddle-ubuntu:2.0.0-18.04
            command: ["bash","-c"]
            args: ["cd /nas/wide_and_deep; python3 train.py"]
            volumeMounts:
            - mountPath: /nas
              name: data
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: nfs-pvc
```

该示例中，镜像仅提供运行环境，训练代码和数据均通过存储挂载的方式添加。