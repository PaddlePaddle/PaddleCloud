[English](../en/Paddlejob_en.md) | 简体中文
- [功能概述](#功能概述)
- [示例1. CPU 训练 wide & deep 示例](#示例1-cpu-训练-wide--deep-示例)
  - [制作 docker 镜像](#制作-docker-镜像)
  - [提交 paddlejob 任务](#提交-paddlejob-任务)
  - [查看任务状态](#查看任务状态)
- [示例2. GPU 训练 resnet 示例](#示例2-gpu-训练-resnet-示例)
- [示例3. GPU 和 nodeselector 训练 wide & deep 示例](#示例3-gpu-和-nodeselector-训练-wide--deep-示例)
- [示例4. 使用 Volcano 调度任务](#示例4-使用-volcano-调度任务)
  - [volcano 引起的崩溃问题](#volcano-引起的崩溃问题)
- [示例5. 挂载外部数据](#示例5-挂载外部数据)
## 功能概述

使用 paddjob 自定义训练任务，一般需要三个步骤

1. 首先docker镜像
2. 编写 paddlejob 的 yaml 配置文件
3. 提交任务，查看任务运行状态。

目前 paddlejob 支持以下功能：

 - 支持 PS/Collective/Heter 三种分布式训练模式。
 - 支持使用 Volcano 进行模型训练作业的调度。
 - 支持训练作业的弹性扩缩容并具有容错保障。

教程提供了多个使用示例，包括使用 CPU、GPU 训练，使用 volcano 调度任务，使用外部数据等，其中示例1详细的介绍了镜像制作及相关参数说明，运行其他示例可参考示例1。

教程所用所有代码均可在`$path_to_PaddleCloud/samples/paddlejob`找到。

## 示例1. CPU 训练 wide & deep 示例

本示例采用 PS 模式，使用 cpu 对 wide & deep 网络进行训练，需要配置 ps 和 worker。

### 制作 docker 镜像

创建文件 Dockerfile，入口函数为 wide_and_deep 文件夹中的 ***train.py***

```dockerfile
$ ls
Dockerfile wide_and_deep

$ cat Dockerfile
FROM ubuntu:18.04

RUN apt update && \
    apt install -y python3 python3-dev python3-pip

RUN python3 -m pip install paddlepaddle==2.0.0 -i https://mirror.baidu.com/pypi/simple

## add user files

ADD wide_and_deep /wide_and_deep

WORKDIR /wide_and_deep

ENTRYPOINT ["python3", "train.py"]
```

构建 docker 镜像

```bash
docker build -t registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1 .
```

您可以更改最后几行以添加更多代码来运行作业，此外，还可以更改 *paddlepaddle* 版本或安装更多依赖

将镜像 push 到镜像库中以供 kubernetes使用

> 教程中使用的是百度的镜像库，您应该把它换为其他有权限推送的库

```bash
docker push registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
```

### 提交 paddlejob 任务

镜像准备完成后，创建 yaml 配置文件 wide_and_deep.yaml

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

- Metadata字段里的name需要唯一，如果存在冲突请先删除原 paddlejob 确保已经删除再提交;
- cleanPodPolicy 制定任务结束时（失败或完成）的策略，可选配置为 Always/Never/OnFailure/OnCompletion，调试时建议 Never，生产时建议 OnCompletion；
  - Always: 任何情况下都会删除pod
  - Never: 任何情况都不会删除pod
  - OnFailure: 只有任务失败会删除pod
  - Oncompletion: 只有任务完成会删除pod
- ps 模式时需要同时配置 ps 和 worker，collective 模式时只需要配置 worker 即可；
- withGloo 有三个可选配置，建议选1
  - 0：不启用
  - 1：启动 worker 端 
  - 2：启动worker和server端

- intranet 可选配置为 Service/PodIP，表示 pod 间的通信方式，用户可以不配置, 默认使用 PodIP；
- ps 和 worker 的内容为 podTemplateSpec，用户可根据需要遵从 kubernetes 规范添加更多内容, 如 GPU 的配置.

将任务提交到 kubernetes

```bash
kubectl -n paddlecloud apply -f wide_and_deep.yaml
```

### 查看任务状态

查看 pods 状态

```shell
$ kubectl -n paddlecloud get pods
```

查看 PaddleJob 状态（pdj为 paddlejob 缩写）

```shell
kubectl -n paddlecloud get pdj
```

## 示例2. GPU 训练 resnet 示例

本示例采用 Collective 模式，使用 gpu 对resnet网络进行训练，需要配置 worker和 gpu。

首先创建 Dockerfile，入口函数为 resnet 文件夹里的 train_fleet.py

```dockerfile
$ ls
Dockerfile   resnet

$ cat Dockerfile
FROM registry.baidubce.com/paddle-operator/paddle-base-gpu:cuda10.2-cudnn7-devel-ubuntu18.04

ADD resnet /resnet

WORKDIR /resnet

RUN pip install scipy opencv-python==4.2.0.32 -i https://mirror.baidu.com/pypi/simple

CMD ["python","-m","paddle.distributed.launch","train_fleet.py"]
```

> docker 镜像的 cuda 版本应该与集群之一兼容，更多细节可以通过运行 *nvidia-smi* 找到。

构建 docker 镜像

```
docker build -t registry.baidubce.com/paddle-operator/demo-resnet:v1 .
```

push 镜像

```
docker push registry.baidubce.com/paddle-operator/demo-resnet:v1
```

创建文件 resnet.yaml

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

提交及查看任务参考示例1。

## 示例3. GPU 和 nodeselector 训练 wide & deep 示例

本示例使用 nodeselector，选取 GPU p100 对 wide & deep 网络进行训练

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

## 示例4. 使用 Volcano 调度任务

本示例使用 volcano 实现 paddlejob 运行的 gan-scheduling。

1. 使用前请确保已[安装](https://github.com/volcano-sh/volcano) Volcano 。
2. 确保您已按照快速开始 [2.3 使用 volcano 的安装配置]() 安装paddlejob组件
3. 创建文件 wide_and_deep.yaml

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

使用此功能需要进行如下配置：

- 使用volcano提供的 api 创建 paddlejob 同名 PodGroup，具体配置信息参考 volcano 规范
- 在 paddlejob 任务配置中添加声明：schedulerName: volcano , 注意：需要且只需要在 worker 中配置。

在以上配置中，我们通过创建最小调度单元为 4 的 podgroup，并将 paddlejob 任务标记使用 volcano 调度，实现了任务的 gan-scheduling。

提交及查看任务参考示例1。

### volcano 引起的崩溃问题

如果您在卸载 volcano 后，发现所有的 deployments 都无法启动 pods，请执行以下命令

```bash
$ kubectl delete validatingwebhookconfigurations volcano-admission-service-jobs-validate volcano-admission-service-pods-validate volcano-admission-service-queues-validate
$ kubectl delete mutatingwebhookconfigurations volcano-admission-service-jobs-mutate volcano-admission-service-podgroups-mutate volcano-admission-service-pods-mutate volcano-admission-service-queues-mutate
```

volcano 当前版本在安装后，无法一键卸载相应的webhookconfig，导致无法核验成功，错误引起原因[参考](https://github.com/volcano-sh/volcano/issues/2102)。

## 示例5. 挂载外部数据

在上面的示例中，训练数据包含在 docker 映像中，这在实践中不是一个好的或可能的选择，kubernetes 提供了标准的 [PV(*persistent volume*) 和 PVC(*persistent volume claim*)](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) 挂载数据。

这里使用 nfs 云盘作为存储作为示例

1. 创建配置文件 pv-pvc.yaml

```yaml
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

2. 创建 pv 及 相应 pvc

pv/pvc 的 namespace 应与 paddlejob 一致，我们这里依然选用 paddlejob。该示例中，数据实际存储在10.12.201.xx的云空间中

```bash
$ kubectl -n paddlecloud apply -f pv-pvc.yaml
```

3. 为 paddlejob 配置 volumes 以使用相应存储。 

创建文件 data_demo.yaml

```yaml
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

4. 提交任务

```bash
$ kubectl -n paddlecloud apply -f data_demo.yaml
```

