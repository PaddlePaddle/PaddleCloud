# Tekton

## 简介

该模块使用 Tekton pipelines 和 Triggers，每周定时拉取 paddlepaddle 模型套件最新版本（例如 PaddleOCR、PaddleDetection 等）制作 docker 镜像， 并推送到 docker hub 镜像仓库。您可以更改参数或 pipeline，实现自定义需求。

## 特性

- 目前支持7个 paddlepaddle 模型套件的自动构建流程，并可以轻松的新增其他模型套件。
- 可通过设置 cronjob 参数，在任意时间完成镜像自动构建任务。
- 沙盒式结构，可根据需求轻松的制定出自定义任务流程。

## 快速上手

教程：[制作 paddle 模型组件镜像并 push 到 docker hub 镜像仓库](./example/README.md)

[Kaniko 使用指南](./tasks/kaniko/README.md)

## 项目概览

### 项目结构

<img src="../docs/images/tekton-arch.png" alt="tekton-arch" style="zoom:25%;" />

如上图所示，该项目主要流程分为两大块，一个是完成业务需求的 pipeline 流程（流程图下部），另一个是为了自动维护套件最新版本镜像的 cronjob && Tekton triggers 流程（流程图上部）。

#### Task

目前已有三个 Task，分别是基于 Tekton 工具库 git clone、kaniko 和对 Paddle 套件预处理的 prepare-build-env， 其中 git clone 和 kaniko 较为固定，若有自定义需求，可更改 prepare-build-env 中的处理逻辑。

#### Pipeline && Pipelinerun

Pipeline 由一个或多个 Task 组合而成，项目提供了针对单套件的单版本镜像制作（single-image）和针对单套件多版本镜像制作（multi-image） 两个Pipeline，用户可根据需求，更改 pipeline 中 prepare-build-env 的步骤。Pipelinrun 对 Pipeline 中的参数进行配置，是 Pipeline 的实例化。此外，用户既可以使用 Pipelinerun 脚本单次运行 Pipeline，也可以通过 cronjob/webhook 定期执行

#### Cronjob && Trigger

项目定义了两个 cronjob，一个是使用 curl 发送 POST 请求，将 Pipeline 用到的参数作为  POST 的 data 发送给 Tekton eventListener，再通过Tekton Trigger 机制将参数应用于 Pipelinerun。这样设计的好处是，Pipeline 和 Tigger 等组件只需要安装一次，是参数无关的，用户只需要将参数（包括定时和镜像配置）写入 cronjob，创建 cronjob 即可完成部署或更改部署。

另一个是自动清理 Pipelinerun 资源的 cronjob，鉴于 Tekton Pipeline 和 Trigger 均无自动清理策略，这里使用 cronjob 定时检查 Pipeline 的实例，即 Pipelinerun 的运行数量，并按照时间戳和设定的存在阈值清理掉部分 Pipelinerun。

#### Dockerfile 说明

为了尽量增强 Pipeline 的通用性，目前只使用了一个 Dockerfile 模板制作所有组件的 Docker 镜像。一方面，在 prepare-build-env 里对部分信息进行解析，通过 ARG 传递到 Dockerfile；另一方面，Dockerfile 是冗余的，可通过在 prepare-build-env 里添加逻辑判断，使用 sed -i 等命令少量更改 Dockerfile 内容。例如 PaddleRec 和 PaddleSpeech 无需安装 requirement.txt 中的包，而 Dockerfile 有`RUN pip3.7 install -r requirements.txt` 的内容，可通过如下逻辑进行处理

```
NO_REQUIREMENT=(PaddleRec PaddleSpeech)

if [[ ${NO_REQUIREMENT[*]} =~ ${TOOLKIT_NAME} ]]
then
    sed -i "/requirements.txt/d" Dockerfile
fi
```

### 文件结构

```
.
├── cronjobs													# k8s 的 cronjob 和 Tekton 的 Trigger
│   ├── build-image-triggers.yaml			# Tekton trigger，接收 curl 请求并触发 pipelinerun
│   ├── cleanup.yaml									# 自动清理 pipelinerun 资源的 cronjob
│   ├── curl													# 自动触发 curl 请求的 cronjobs，维护多个 paddle 套件
├── dockerfiles												# 制作镜像使用的 Dockerfile
├── example														# 快速上手教程
├── pipelines													# 包含了两个 pipeline 和对应的 pipelinerun
├── rbac															# 用户注册					
└── tasks															# Tekton 的 task 任务
    ├── git-clone											# 针对 github 的 git clone task
    ├── kaniko												# 基于 kaniko 的镜像制作推送流程
    └── prepare-build-env							# 针对 paddlepaddle 组件的处理 task
```

## 飞桨模型组件镜像仓库

我们使用该模块维护了多个 Paddle 组件镜像仓库，包括基于 GPU 和 CPU 的镜像，如果您有其他需求，请联系我们。

| Paddle 套件     | 镜像仓库                                                     | 维护版本    |
| --------------- | ------------------------------------------------------------ | ----------- |
| PaddleOCR       | [PaddleOCR 镜像仓库](https://hub.docker.com/r/paddlecloud/paddleocr) | release/2.4 |
| PaddleDetection | [PaddleDetection 镜像仓库](https://hub.docker.com/r/paddlecloud/paddledetection) | release/2.4 |
| PaddleNLP       | [PaddleNLP 镜像仓库](https://hub.docker.com/r/paddlecloud/paddlenlp) | develop     |
| PaddleSeg       | [PaddleSeg 镜像仓库](https://hub.docker.com/r/paddlecloud/paddleseg) | release/2.5 |
| PaddleClas      | [PaddleClas 镜像仓库](https://hub.docker.com/r/paddlecloud/paddleclas) | release/2.3 |
| PaddleSpeech    | [PaddleSpeech 镜像仓库](https://hub.docker.com/r/paddlecloud/paddlespeech) | develop     |
| PaddleRec       | [PaddleRec 镜像仓库](https://hub.docker.com/r/paddlecloud/paddlerec) | master      |



