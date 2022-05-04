[English](../en/Installation_en.md) | 简体中文
- [安装教程](#安装教程)
  - [1. 环境需求](#1-环境需求)
  - [2. 安装](#2-安装)
    - [2.1 一键安装所有组件](#21-一键安装所有组件)
    - [2.2 自定义安装](#22-自定义安装)
    - [2.3 使用 volcano 的安装配置](#23-使用-volcano-的安装配置)
  - [3. paddlejob 任务测试](#3-paddlejob-任务测试)
  - [4. 卸载](#4-卸载)
  - [5. 高级用法](#5-高级用法)
# 安装教程

## 1. 环境需求

* Kubernetes, version = v1.21
* kubectl
* helm

> Paddlejob 等主要组件可以在 kubernetes v1.16+ 版本运行。如果您想要体验 PaddleCloud，为了其他依赖的稳定运行，建议选取已测试过的 v1.21 版本。

如果您没有Kubernetes环境，可以参考 [microk8s官方文档](https://microk8s.io/docs/getting-started) 进行安装；如果您使用的是 macOS 系统， 或遇到了安装问题，可以参考文档 [macOS 安装 microk8s](./macOS_install_microk8s.md)。

## 2. 安装

> 请确保您已完成kubernates的安装配置，并可以通过kubectl、helm对集群进行操作。

如果您是 kubernates 新手或在一个全新的 kubernates 环境中体验组件，建议按照章节 2.1 一键安装所有组件和依赖；如果您在在已有的 kubernates 环境中部署，建议阅读章节 2.2，使用自定义安装方法

### 2.1 一键安装所有组件

添加并更新 helm 的 charts repositories

```bash
$ helm repo add paddlecloud https://paddleflow-public.hkg.bcebos.com/charts
$ helm repo update
```

使用 helm 一键安装

> 此篇教程内，namespace 默认使用 paddlecloud，如需更改，请自行替换

```bash
# create namespace in k8s
$ kubectl create namespace paddlecloud
# install
$ helm install -n paddlecloud test paddlecloud/paddlecloud --set tags.all-dep=true 
```

安装完成后，查看安装的各个内容

```bash
$ kubectl -n paddlecloud get deployments
NAME                         READY   UP-TO-DATE   AVAILABLE   AGE
test-paddlecloud-paddlejob   1/1     1            1           97s
test-paddlecloud-serving     1/1     1            1           97s
test-paddlecloud-sampleset   1/1     1            1           97s

$ kubectl -n paddlecloud get pods
NAMESPACE     NAME                                          READY   STATUS    RESTARTS      AGE
paddlecloud   test-paddlecloud-paddlejob-5c46b5b5dc-gnrlq   1/1     Running   0             30s
paddlecloud   test-paddlecloud-serving-6bc9f77bf6-4njl2     2/2     Running   0             30s
paddlecloud   test-paddlecloud-sampleset-654c876446-bd6x7   1/1     Running   0             30s

$ kubectl -n paddlecloud get replicasets
NAME                                    DESIRED   CURRENT   READY   AGE
test-paddlecloud-paddlejob-5c46b5b5dc   1         1         1       2m39s
test-paddlecloud-sampleset-654c876446   1         1         1       2m39s
test-paddlecloud-serving-6bc9f77bf6     1         1         1       2m39s
```

### 2.2 自定义安装

**组件自定义安装**：默认安装全部组件，分别是模型训练组件 paddlejob、数据缓存组件 sampleset 和模型推理组件 serving，可通过 `--set {$components_name}.enable=false` 取消部分组件安装。

**依赖 charts 自定义安装**：默认不安装任何依赖 charts，指定参数 `tags.all-dep=true` 安装所有依赖，或使用`{$chart_name}.enable=true` 安装部分依赖。下表展示了提供的依赖和相应的版本：

| chart name           | alias     | version |
| -------------------- | --------- | ------- |
| hostpath-provisioner | hostpath  | 0.2.13  |
| juicefs-csi-driver   | juicefs   | 0.8.1   |
| redis                | /         | 16.5.4  |
| jupyterhub           | /         | 1.1.2   |
| kubeflow-pipelines   | pipelines | 0.1.0   |
| knative-serving      | knative   | 0.1.0   |

helm 提供了**两种指定参数的方法**，一种是在命令行**使用 --set 指定参数**，另一种是**使用 yaml 文件指定参数**。一般来说，如果您更改了较多参数，或使用次数多，建议使用 yaml 文件方式，否则使用 --set 方式。

**示例1，在命名空间 paddlecloud 安装 paddlejob、sampleSet 组件和所有依赖，名字叫 test：**

**"--set" 安装方式：**

直接将要设置的参数用 --set 进行指定

```bash
$ kubectl create namespace paddlecloud
$ helm install test paddlecloud/paddlecloud --namespace paddlecloud --set tags.all-dep=true,serving.enabled=false
```

**yaml 文件安装方式：**

创建 values.yaml 文件，将自定义值以 yaml 的形式写在文件中

```yaml
tags:
  all-dep: true
serving:
  enabled: false
```

安装 chart

```bash
$ helm install test paddlecloud/paddlecloud -f valuse.yaml --namespace paddlecloud
```

**示例2，在默认命名空间安装所有组件，只安装 redis 和 kubeflow-pipelines 依赖：**

**"--set"安装方式：**

```bash
$ helm install test paddlecloud/paddlecloud --set pipelines.enabled=true,redis.enabled=true
```

**yaml文件安装方式：**

创建 values.yaml

```yaml
pipelines:
  enabled: true
redis:
  enabled: true
```

安装

```bash
$ helm install test paddlecloud/paddlecloud -f values.yaml
```

更多关于 `helm install` 的使用可通过 `helm install --help` 命令获取

### 2.3 使用 volcano 的安装配置

> 此步骤非必须步骤，当且仅当您想要使用 volcano 调度作业时，使用此配置

请您先按照[官网教程](https://github.com/volcano-sh/volcano)安装 Volcano 。

创建 values.yaml 文件，加入以下内容

```yaml
paddlejob:
  args:
    - --leader-elect
    - --namespace=paddlecloud  
    - --scheduling=volcano       
```

参数说明：

- namespace：此处的namespace必须和安装时设置的namespace一致，如没有显式设置，namespace默认为default
- scheduling：使用volcano进行调度

安装方式和上面一致：

```bash
$ helm install test paddlecloud/paddlecloud -n paddlecloud -f values.yaml
```

## 3. paddlejob 任务测试

这里使用一个简单案例测试组件，查看[paddlejob 使用教程](Paddlejob.md)获取更详细的使用说明和更多的实验样例。此示例采用 PS 模式，使用 cpu 进行训练，需要配置 PS 和 worker 。

1. 提交任务：

```shell
$ kubectl -n paddlecloud apply -f wide_and_deep.yaml
```

2. 查看 pods 状态

```shell
$ kubectl -n paddlecloud get pods
```

3. 查看 PaddleJob 状态

```shell
$ kubectl -n paddlecloud get pdj
```

请点击 [paddlejob 使用教程](Paddlejob.md) 查看更多使用案例和更为详细的讲解。

## 4. 卸载

首先查看当前 helm 安装的 chart

```shell
$ helm list -A
NAME  NAMESPACE  	  REVISION	 UPDATED                                STATUS  	CHART               APP VERSION
test  paddlecloud   1       	2022-04-11 18:08:38.119978 +0800 CST   	deployed	paddlecloud-0.1.0   0.4.0
```

卸载安装的 chart

```bash
$ helm uninstall test -n paddlecloud
```

## 5. 高级用法

您可以在 Makefile 文件中查看更多的设置，也可以 clone [该项目](https://github.com/PaddlePaddle/PaddleCloud)来进行修改。 如果您有任务问题或建议，欢迎联系我们。
