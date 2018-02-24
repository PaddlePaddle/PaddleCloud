# PaddlePaddle Cloud

[![Build Status](https://travis-ci.org/PaddlePaddle/cloud.svg?branch=develop)](https://travis-ci.org/PaddlePaddle/cloud)

PaddlePaddle Cloud is a combination of PaddlePaddle and Kubernetes. It
supports fault-recoverable and fault-tolerant large-scaled distributed
deep learning. We can deploy it on public cloud and on-premise
clusters.

PaddlePaddle Cloud includes the following components:

- paddlectl: A command-line tool that talks to paddlecloud and
  paddle-fs.
- paddlecloud: An HTTP server that exposes Kubernetes as a Web
  service.
- paddle-fs: An HTTP server that exposes the CephFS distributed
  filesystem as a Web service.
- EDL (elastic deep learning): A Kubernetes controller that supports
  elastic scheduling of deep learning jobs and other jobs.
- Fault-tolerant distributed deep learning: This part is in
  the [Paddle](https://github.com/PaddlePaddle/paddle) repo.

## Tutorials

- [快速开始](./doc/tutorial_cn.md)
- [中文手册](./doc/usage_cn.md)


## How To

- [Build PaddlePaddle Cloud](./doc/howto/build.md)
- [Deploy PaddlePaddle Cloud](./doc/howto/deploy.md)
- [Elastic Deep Learning using EDL](./doc/howto/edl.md)
- [PaddlePaddle Cloud on Minikube](./doc/howto/minikube.md)

## Directory structure

```
.
├── demo: distributed version of https://github.com/PaddlePaddle/book programs
├── doc: documents
├── docker: scripts to build Docker image to run PaddlePaddle distributed
├── go
│   ├── cmd
│   │   ├── edl: entry of EDL controller binary
│   │   ├── paddlecloud: the command line client of PaddlePaddle Cloud (will be deprecated)
│   │   ├── paddlectl: the command line client of PaddlePaddle Cloud
│   │   └── pfsserver: entry of PaddleFS binary
│   ├── edl: EDL implementation
│   ├── filemanager: PaddleFS implementation
│   ├── paddlecloud: command line client implement (will be deprecated)
│   ├── paddlectl: command line client implement
│   ├── scripts: scripts for Go code generation
├── k8s: YAML files to create different components of PaddlePaddle Cloud
│   ├── edl: TPR definition and EDL controller for TraningJob resource
│   │   ├── autoscale_job: A sample TrainingJob that can scale
│   │   └── autoscale_load: A sample cluster job demonstrating a common workload
│   ├── minikube: YAML files to deploy on local mini-kube environment
│   └── raw_job: A demo job demonstrates how to run PaddlePaddle jobs in cluster
└── python: PaddlePaddle Cloud REST API server
```
