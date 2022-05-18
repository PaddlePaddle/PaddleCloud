# Paddle-toolkit-collection

本镜像仓库主要用于存储包含 [paddlepaddle](https://github.com/PaddlePaddle) 多个模型套件的标准镜像，方便模型套件用户进行 Docker 化部署或在云上部署。
Paddle-toolkit-collection 套件的标准镜像由 [PaddleCloud](https://github.com/PaddlePaddle/PaddleCloud) 项目基于 [Tekton Pipeline](https://github.com/tektoncd/pipeline) 自动构建， 
除了直接使用套件的标准镜像，如果您需要对模型套件进行二次开发并希望能够持续构建定制的镜像，
可以参考 [PaddleCloud Tekton文档](https://github.com/PaddlePaddle/PaddleCloud/blob/main/tekton/README.md)目录下的文档构建您自己的套件镜像CI流水线。

更多关于部署的内容可以参考云上飞桨项目 [PaddleCloud](https://github.com/PaddlePaddle/PaddleCloud) 。

目前镜像内维护的套件及相应版本如下表所示，其中**镜像仓库**为该模型套件单独的 docker 镜像。

| Paddle 套件                                                  | 镜像仓库                                                     | 维护版本    |
| ------------------------------------------------------------ | ------------------------------------------------------------ | ----------- |
| [PaddleOCR](https://github.com/PaddlePaddle/PaddleOCR)       | [PaddleOCR 镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddleocr) | release/2.5 |
| [PaddleDetection](https://github.com/PaddlePaddle/PaddleDetection) | [PaddleDetection 镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddledetection) | release/2.4 |
| [PaddleNLP](https://github.com/PaddlePaddle/PaddleNLP)       | [PaddleNLP 镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddlenlp) | develop     |
| [PaddleSeg](https://github.com/PaddlePaddle/PaddleSeg)       | [PaddleSeg 镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddleseg) | release/2.5 |
| [PaddleClas](https://github.com/PaddlePaddle/PaddleClas)     | [PaddleClas 镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddleclas) | release/2.3 |
| [PaddleSpeech](https://github.com/PaddlePaddle/PaddleSpeech) | [PaddleSpeech 镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddlespeech) | develop     |
| [PaddleRec](https://github.com/PaddlePaddle/PaddleRec)       | [PaddleRec 镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddlerec) | master      |

## PaddleCloud 优势

- **模型套件Docker镜像大礼包。**

  PaddleCloud 为用户提供了飞桨模型套件 Docker 镜像大礼包，这些镜像中包含运行模型套件案例的所有依赖并能持续更新，支持异构硬件环境和常见CUDA版本、开箱即用。

- **具有丰富的云上飞桨组件。**

  云上飞桨具有丰富的云原生功能组件，包括样本数据缓存组件、分部署训练组件、推理推理服务组件等，使用这些组件用户可以方便快捷的在 Kubernetes 集群上镜像模型的训练和部署工作。

- **功能强大的自运维能力。**

  云上飞桨组件基于 Kubernetes 的 Operator 机制提供了功能强大的自运维能力，如训练组件支持多种架构模式并具有分布式容错与弹性训练的能力，推理服务组件支持自动扩缩容与蓝绿发版等。

- **针对飞桨框架的定制优化。**

  除了部署便捷与自运维的优势，PaddleCloud 还针对飞桨框架进行了正对性优化，如通过缓存样本数据来加速云上飞桨分布式训练作业、基于飞桨框架和调度器的协同设计来优化集群GPU利用率等。


## 安装 Docker 

如果您所使用的机器上还没有安装 Docker，您可以参考 [Docker 官方文档](https://docs.docker.com/get-docker/) 来进行安装。
如果您需要使用支持 GPU 版本的镜像，则还需安装好 NVIDIA 相关驱动和 nvidia-docker，详情请参考[官方文档](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html#docker) 。

## 快速上手

使用的Docker环境可以快速上手体验，我们为您提供了CPU和GPU版本的镜像。 
如果您是Docker新手，建议您花费几分钟的时间学习下[docker基本用法](https://github.com/PaddlePaddle/PaddleCloud/blob/main/docs/zh_CN/docker-tutorial.md)。

**使用CPU版本的Docker镜像**

```bash
docker run --name dev -v $PWD:/mnt -p 8888:8888 -it paddlecloud/paddle-toolkit-collection:2.3.0-cpu /bin/bash
```

**使用GPU版本的Docker镜像**

```bash
docker run --name dev --runtime=nvidia -v $PWD:/mnt -p 8888:8888 -it paddlecloud/paddle-toolkit-collection:2.3.0-gpu-cuda10.2-cudnn7 /bin/bash
```

进入容器内，则可执行 PaddleDetection 套件中提供的案例。

**使用 Jupyterlab**

最新版镜像集成了 jupyterlab，进入容器后，可通过以下命令开启服务。
```bash
$ jupyter lab --ip=0.0.0.0 --port=8888 --allow-root --notebook-dir=/home
```
## 镜像列表

镜像 tag 的 2.x.x 代表 Paddle 版本，其中包含的套件为构建日期的最新 commit 版本


| 镜像路径                                                     | 构建时间       |
| ------------------------------------------------------------ | -------------- |
| paddlecloud/paddle-toolkit-collection:2.3.0-gpu-cpu          | 2022年05月18日 |
| paddlecloud/paddle-toolkit-collection:2.3.0-gpu-cuda10.2-cudnn7 | 2022年05月18日 |
| paddlecloud/paddle-toolkit-collection:2.3.0-gpu-cuda11.2-cudnn8 | 2022年05月18日 |
