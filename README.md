# PaddleCloud

[English](./README_en.md) | 简体中文

云上飞桨（PaddleCloud）项目是面向[飞桨](https://github.com/PaddlePaddle/Paddle) （PaddlePaddle）框架及其模型套件的部署工具箱，为用户提供了模型套件 Docker 化部署和 Kubernetes 集群部署两种方式，满足不同场景与环境的部署需求。

## 云上飞桨优势

- **模型套件Docker镜像大礼包。**

  PaddleCloud 为用户提供了飞桨模型套件 Docker 镜像大礼包，这些镜像中包含运行模型套件案例的所有依赖并能持续更新，支持异构硬件环境和常见CUDA版本、开箱即用。

- **具有丰富的云上飞桨组件。**

  云上飞桨具有丰富的云原生功能组件，包括样本数据缓存组件、分布式训练组件、推理服务组件等，使用这些组件用户可以方便快捷地在 Kubernetes 集群上进行训练和部署工作。

- **功能强大的自运维能力。**

  云上飞桨组件基于 Kubernetes 的 Operator 机制提供了功能强大的自运维能力，如训练组件支持多种架构模式并具有分布式容错与弹性训练的能力，推理服务组件支持自动扩缩容与蓝绿发版等。

- **针对飞桨框架的定制优化。**

  除了部署便捷与自运维的优势，PaddleCloud 还针对飞桨框架进行了正对性优化，如通过缓存样本数据来加速云上飞桨分布式训练作业、基于飞桨框架和调度器的协同设计来优化集群GPU利用率等。


## 模型套件Docker化部署

PaddleCloud 基于 [Tekton](https://github.com/tektoncd/pipeline) 为飞桨模型套件（PaddleOCR等）提供了镜像持续构建的能力，使用这些镜像用户可以快速在本地环境体验和部署套件中的案例。
目前 [PaddleOCR](https://github.com/PaddlePaddle/PaddleOCR) 、[PaddleDetection](https://github.com/PaddlePaddle/PaddleDetection) 、[PaddleNLP](https://github.com/PaddlePaddle/PaddleNLP)
等飞桨模型套件已经接入CI流水线，后续还将接入更多的模型套件。除了直接使用套件的标准镜像，如果您需要对模型套件进行二次开发并希望能够持续构建定制的镜像，可以参考 [tekton](./tekton/README.md ) 
目录下的文档构建您的套件镜像CI流水线。

> **适用场景**：本地测试开发环境、单机部署环境。

### 文档教程

- 飞桨模型套件镜像列表
  - [PaddleOCR镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddleocr)
  - [PaddleDetection镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddledetection)
  - [PaddleNLP镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddlenlp)
  - [PaddleClas镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddleclas)
  - [PaddleSeg镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddleseg)
  - [PaddleSpeech镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddlespeech)
  - [PaddleRec镜像仓库](https://hub.docker.com/repository/docker/paddlecloud/paddlerec)
- 模型套件Docker化部署案例
  - [PP-Human行人检查](./samples/pphuman/pphuman-docker.md)
  - [训练PP-YOLOE模型](./samples/pphuman/ppyoloe-docker.md)
  - [PP-OCRv3训推一体部署实战](./samples/PaddleOCR/PP-OCRv3.md)

## Kubernetes集群部署

PaddleCloud 基于 Kubernetes 的 Operator 机制为您提供了多个功能强大的云原生组件，如样本数据缓存组件、分布式训练组件、 以及模型推理服务组件，使用这些组件您可以快速地在云上进行分布式训练和模型服务化部署。
此外，PaddleCloud 还基于 [Kubeflow Pipeline](https://github.com/kubeflow/pipelines) 为您提供了云上模型全链路的功能。您可以使用 Python SDK 来快速构建自定义的飞桨深度学习工作流，并能够利用上各飞桨云原生组件的能力，使得数据准备、模型训练、超参调优、模型服务化部署等每个步骤的工作都变得简单。

> **适用场景**：基于 Kubernetes 的多机部署环境。

### 文档教程

- [安装文档](./docs/zh_CN/installation.md)
- [架构概览](./docs/zh_CN/paddlecloud-overview.md)
- 数据集缓存组件
  - [组件概览](./docs/zh_CN/sampleset-overview.md)
  - [快速上手](./docs/zh_CN/sampleset-tutorails.md)
  - [性能测试](./docs/zh_CN/sampleset-benchmark.md)
- 分布式训练组件
  - [组件概览](./docs/zh_CN/paddlejob-overview.md)
  - [快速上手](./docs/zh_CN/paddlejob-tutorails.md)
- 模型推理服务组件
  - [快速上手](./docs/zh_CN/serving-tutorials.md)
- 云上飞桨组件使用案例
  - [训练PP-YOLOE目标检测模型](./samples/pphuman/ppyoloe-k8s.md)
  - [训练PP-OCRv3文本识别模型](./samples/PaddleOCR/PP-OCRv3.md)
- 云上模型全链路案例
  - [PP-OCR文字检测模型全链路案例](./samples/pipelines/README.md)


## 许可证书

本项目的发布受[Apache 2.0 license](./LICENSE)许可认证。

