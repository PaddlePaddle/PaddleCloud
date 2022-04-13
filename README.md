English | [简体中文](./README-zh_CN.md)

# PaddleCloud

## Overview

PaddleCloud aims to provide a set of easy-to-use cloud components based on PaddlePaddle and related kits to meet customers' business cloud requirements. In order to get through the whole process from training to deployment, the model training component paddlejob, the model inference component serving, and the sample caching component sampleset for acceleration have been developed. The components provide users with almost zero-based experience tutorials and easy-to-use programming interfaces. You can get a clearer understanding of PaddleCloud through Architecture Overview.

## Components introduction

- The model training component **paddlejob** is designed to provide a simple and easy-to-use standardized interface for running Paddle distributed training tasks on Kubernetes, and to provide customized and complete support for the management of training tasks. [More About paddlejob](./docs/design/paddlejob).
- The sample cache component **sampleset** implements sample cache in paddlejob based on the open source project JuiceFS. It aims to solve the problem of high network IO overhead caused by the separation structure of computing and storage in Kubernetes, so as to improve the efficiency of distributed training jobs on the cloud. [More about sampleset](./docs/design/sampleset).
- The model inference component **serving** is developed based on Knative Serving and provides functions such as automatic scaling, fault tolerance, and health check. It supports deploying model services on Kubernetes clusters using mainstream frameworks such as PaddlePaddle, TensorFlow, and onnx. [More about serving](./docs/design/serving).

## Quick Start

### Prerequisites

* Kubernetes,  version: v1.21
* kubectl
* helm

> Major components such as Paddlejob can run on kubernetes v1.16+ and we test all the contents at version v1.21. If you want to experience PaddleCloud in a easy way, it is recommended to use version v1.21 for the stable operation of other dependencies.

If you do not have a Kubernetes environment, you can refer to [microk8s official documentation](https://microk8s.io/docs/getting-started) for installation. If you use macOS system, or encounter installation problems, you can refer to the document [macOS install microk8s](./docs/macOS_install_microk8s.md).

### Installation

We assume that you have installed the kubernates cluster environment and you can access the cluster through command such as **helm** and **kubectl**. Otherwise, please refer to the more detailed [installation tutorial](./docs/tutorials/Installation_en.md) for help. If you deploy components in the production environment or have custom installation requirements, please also refer to [Installation Tutorial](./docs/tutorials/Installation_en.md).

Add and update helm's charts repositories,

```bash
$ helm repo add paddlecloud https://paddleflow-public.hkg.bcebos.com/charts
$ helm repo update
```

Install all components and all dependencies using helm.

```bash
# create namespace in k8s
$ kubectl create namespace paddlecloud
# install
$ helm install test paddlecloud/paddlecloud --set tags.all-dep=true --namespace paddlecloud
```

You can find the specific meaning of all the parameters in [Installation Tutorial](./docs/tutorials/Installation_en.md).

### Run demo paddlejob

You can get more detailed usage examples in [here](./docs/tutorials/Paddlejob_en.md).

Deploy your first paddlejob demo with

```shell
$ kubectl -n paddlecloud apply -f $path_to_project/samples/paddlejob/wide_and_deep.yaml
```

Check pods status
```shell
kubectl -n paddlecloud get pods
```

Check paddle job status
```shell
kubectl -n paddlecloud get pdj
```

## Tutorials

**Quick Start**

- [Installation](./docs/tutorials/Installation_en.md)
-  [paddlejob tutorials](./docs/tutorials/Paddlejob_en.md)
- sampleset tutorials
- serving tutorials

#### Design Introduction

- PaddleCloud 
- paddlejob 
  - [design-total](./docs/design/paddlejob/design.md)
  - [design-arch](./docs/design/paddlejob/design-arch.md)
  - [design-controller](./docs/design/paddlejob/design_controller.md)
  - [design-fault-tolerant](./docs/design/paddlejob/design_fault_tolerant.md)
- sampleset
- serving

## License

PaddleCloud is released under the [Apache 2.0 license]().
