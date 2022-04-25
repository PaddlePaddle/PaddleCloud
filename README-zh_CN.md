[English](./README.md) | 简体中文

# PaddleCloud

## 概述

PaddleCloud 旨在基于飞桨 PaddlePaddle 及相关套件，提供一套简单易用的云上组件，满足客户的业务上云需求。为打通从训练到部署的全流程，目前开发了模型训练组件 paddlejob、模型推理组件 serving 和用于加速的样本缓存组件 sampleset，并为用户提供了几乎零基础的体验教程和简单易用的编程接口，方便用户构建自定义的工作流。您可以通过架构概览部分更清晰的了解 PaddleCloud。

## 组件介绍                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             

- 模型训练组件（paddlejob）旨在为 Kubernetes 上运行飞桨分布式训练任务提供简单易用的标准化接口，并为训练任务的管理提供定制化的完整支持。[更多关于paddlejob的内容](./docs/design/paddlejob)。
- 样本缓存组件（sampleset）基于开源项目 JuiceFS，在 paddlejob 中实现了样本缓存，旨在解决 Kubernetes 中计算与存储分离的结构带来的高网络 IO 开销问题，提升云上飞桨分布式训练作业的执行效率。[更多关于sampleset 的内容](./docs/design/sampleset)。
- 模型推理组件（serving）基于 Knative Serving 构建，提供了自动扩缩容、容错、健康检查等功能，支持在 Kubernetes 集群上使用 PaddlePaddle、TensorFlow、onnx 等主流框架部署模型服务。[更多关于serving的内容](./docs/design/serving)。

## 快速开始

### 环境需求

* Kubernetes, version: v1.21
* kubectl
* helm

> Paddlejob 等主要组件可以在 kubernetes v1.16+ 版本运行。如果您想要体验 PaddleCloud，为了保证其他依赖的稳定运行，建议选取已测试过的 v1.21 版本。

如果您没有Kubernetes环境，可以参考 [microk8s官方文档](https://microk8s.io/docs/getting-started) 进行安装；如果您使用的是 macOS 系统，或遇到了安装问题，可以参考文档 [macOS 安装 microk8s](./docs/macOS_install_microk8s.md)。

### 安装

如果您在生产环境中部署组件，或有自定义安装需求，请访问[安装教程](./docs/tutorials/Installation.md)获取更详细的信息。

添加并更新 helm 的 charts repositories

```bash
$ helm repo add paddlecloud https://paddleflow-public.hkg.bcebos.com/charts
$ helm repo update
```

使用 helm 一键安装所有组件和所有依赖

```bash
# create namespace in k8s
$ kubectl create namespace paddlecloud
# install
$ helm install test paddlecloud/paddlecloud --set tags.all-dep=true --namespace paddlecloud
```

你可以在[安装教程](./docs/tutorials/Installation.md)内查看更详细的安装教程

### 运行 paddlejob 示例

你可以在 [paddlejob 使用教程](./docs/tutorials/Paddlejob.md)获取更详细的说明及更丰富的示例

提交 paddlejob

```shell
$ kubectl -n paddlecloud apply -f $path_to_paddlecloud/samples/paddlejob/wide_and_deep.yaml
```

 查看 pods 运行状态

```shell
kubectl -n paddlecloud get pods
```

查看 paddlejob 运行状态

```shell
kubectl -n paddlecloud get pdj
```

## 文档教程

#### 使用指南

- [飞桨模型套件Docker镜像大礼包](./images/)
- [安装教程](./docs/tutorials/Installation.md)
- [paddlejob 使用教程](./docs/tutorials/Paddlejob.md)
- sampleset 使用教程
- serving 使用教程

#### 设计说明

- PaddleCloud 项目架构
- paddlejob 
  - [总体设计](./docs/design/paddlejob/design.md)
  - [架构设计](./docs/design/paddlejob/design-arch.md)
  - [控制器设计](./docs/design/paddlejob/design_controller.md)
  - [训练作业容错](./docs/design/paddlejob/design_fault_tolerant.md)
- sampleset 设计
- serving 设计

## 许可证书

本项目的发布受[Apache 2.0 license]()许可认证。

