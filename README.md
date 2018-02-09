# PaddlePaddle Cloud

[![Build Status](https://travis-ci.org/PaddlePaddle/cloud.svg?branch=develop)](https://travis-ci.org/PaddlePaddle/cloud)

PaddlePaddle Cloud is a Distributed Deep-Learning Cloud Platform for both cloud
providers and enterprises. PaddlePaddle Cloud provide cluster Deep-Learning
features including:

- Manage cluster Deep-Learning jobs use a command-line client very easily.
- Elastic Deep-Learning ([EDL](./doc/edl/README.md)) enables maximizing cluster resource utilities
  and reduce average job pending time.
- Manages thousands of GPUs in cluster.
- Fault tolerant distributed training with zero down time.

PaddlePaddle Cloud use [Kubernetes](https://kubernetes.io) as it's backend job
dispatching and cluster resource management center. And use [PaddlePaddle](https://github.com/PaddlePaddle/Paddle.git)
as the deep-learning frame work. 

## Components

- Server side components includes:
  - Cloud Server
    A REST API server accepts incomming request from lommand-line client, and submit
    `TrainingJob` resource to Kubernetes cluster. Cloud Server is written using
    `Django` framework under `python` directory.
  - PaddleFS
    On cloud file management server used to upload user training job python programs,
    downloading trained models to user's desktop. The code is under `go/cmd/pfsserver`.
  - EDL Controller
    A Kubernetes [Controller](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#custom-controllers)
    to enable automatically scale up/down jobs to maximize cluster performance.
  - PaddlePaddle Cloud Job runtime Docker image.
- Client side component:
  - Command-Line client
    Client for submit, list and kill cluster training jobs. Then code is under
    `go/cmd/paddlecloud`, And, code under `go/cmd/paddlecloud` will deprecate and
    move on to `go/cmd/paddlectl`.

## User Manuals

[快速开始](./doc/tutorial_cn.md)

[中文手册](./doc/usage_cn.md)

English tutorials(comming soon...)

## Build

[Build all components and Docker images](./doc/build/build.md)

## Deploy

[Deploy Cloud Server](./doc/deploy/deploy.md)

[Deploy EDL](./doc/deploy/deploy_edl.md)

[Run on Minikube (for developers)](./doc/deploy/run_on_minikube.md)

## Contribute

We appreciate your contributions! You can contribute on any of the components
according to the following code structure:

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
│   ├── edl: EDL implementaion
│   ├── filemanager: PaddleFS implementaion
│   ├── paddlecloud: command line client implement (will be deprecated)
│   ├── paddlectl: command line client implement
│   ├── scripts: scripts for Go code generation
├── k8s: YAML files to create different componets of PaddlePaddle Cloud
│   ├── edl: TPR defination and EDL controller for TraningJob resource
│   │   ├── autoscale_job: A sample TrainingJob that can scale
│   │   └── autoscale_load: A sample cluster job demonstrating a common workload
│   ├── minikube: YAML files to deploy on local mini-kube environment
│   └── raw_job: A demo job demostrates how to run PaddlePaddle jobs in cluster
└── python: PaddlePaddle Cloud REAT API server
```

Contributors can create pull requests to describe one design under `doc/desgin`
or write code according to current design docs.