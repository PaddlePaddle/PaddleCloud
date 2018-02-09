# PaddlePaddle Cloud

[![Build Status](https://travis-ci.org/PaddlePaddle/cloud.svg?branch=develop)](https://travis-ci.org/PaddlePaddle/cloud)

PaddlePaddle Cloud is a Distributed Deep-Learning Cloud Platform for both cloud
providers and enterprises. PaddlePaddle Cloud provides cluster Deep-Learning
features including:

- Manage cluster Deep-Learning jobs use a command-line client very easily.
- Elastic Deep-Learning ([EDL](./doc/edl/README.md)) enables maximizing cluster resource utilities
  and reduce average job pending time.
- Manages thousands of GPUs in the cluster.
- Fault tolerant distributed training with zero downtime.

PaddlePaddle Cloud use [Kubernetes](https://kubernetes.io) as it's backend job
dispatching and cluster resource management center. And use [PaddlePaddle](https://github.com/PaddlePaddle/Paddle.git)
as the deep-learning framework. 

## Components

- Server-side components:
  - Cloud Server
    A REST API server accepts request from command-line client and submit
    `TrainingJob` resource to Kubernetes cluster. Cloud Server is written using
    `Django` framework under `python/` directory.
  - PaddleFS
    On cloud file management server used to upload user training job python programs,
    downloading trained models to user's desktop etc. The code is under `go/cmd/pfsserver`.
  - EDL Controller
    A Kubernetes [Controller](https://kubernetes.io/docs/concepts/api-extension/custom-resources/#custom-controllers)
    to enable automatically scale up/down jobs to maximize cluster performance.
  - PaddlePaddle Cloud Job runtime Docker image.
- Client-side component:
  - Command-Line client
    Client to submit, list and kill cluster jobs. The code is under
    `go/cmd/paddlecloud`. Code under `go/cmd/paddlecloud` will be deprecated and
    move on to `go/cmd/paddlectl`.

## User Manuals

[快速开始](./doc/tutorial_cn.md)

[中文手册](./doc/usage_cn.md)

English tutorials(coming soon...)

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

Contributors can create pull requests to describe one design under `doc/desgin`
or write code according to current design docs.