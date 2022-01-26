# Pipeline相关组件使用手册

**Pipeline相关组件项目链接：**https://github.com/PaddleFlow/paddlex-backend

云上飞桨产品基于 Kubeflow Pipeline 和它的 Python SDK 为用户封装了高层 API，以 Components YAML 文件的形式为用户提供快速构建深度学习工作流的编程组件。下面我们将简要介绍下 Kubeflow Pipeline 以方便用户快速入手。

Kubeflow Pipelines 是一个基于 Docker 容器构建和部署可移植、可扩展的机器学习 (ML) 工作流的平台。

Kubeflow Pipelines 平台包括如下组件：

- 用于管理和跟踪实验、作业和运行的用户界面 (UI)。
- 用于调度 ML 工作流的引擎。
- 用于定义和操作 pipelines 和 components 的 SDK。
- 使用 SDK 与系统交互的 Notebook 。

Kubeflow Pipelines 的目标如下：

- 端到端编排：启用和简化机器学习管道的编排。
- 简化模型实验：让您轻松尝试多种想法和技术并管理您的各种试验/实验。
- 易于重用：使您能够重用组件和管道以快速创建端到端解决方案，而无需每次都重新构建。

下图是 Kubeflow Pipeline 的架构图，从高层次的角度来看，Pipeline 的执行流程如下：

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-529d237bef77d431aa8977abeba30a49ebcf0d01)

- **Python SDK：**您可以使用 Kubeflow Pipelines 领域特定语言 (DSL) 创建组件或指定 Pipeline。

- **DSL Compiler：**DSL 编译器将 Pipeline 的 Python 代码转换为静态 YAML 配置文件。

- **Pipeline Service：**您可以通过调用 pipeline service 的接口来提交并执行 Pipeline。

- **Kubernetes Resource：** Pipeline Service 调用 Kubernetes API Service 来创建必要的 Kubernetes 资源 (CRD) 以执行 Pipeline。

- **Orchestration Controllers：** 是一组编排控制器，用于启动并运行 Pipeline 所需的容器，如 Argo Workflow 控制器。

- **Artifact storage：** Pod 存储有两种数据：

  - **Metadata**： Kubeflow Pipelines 将元数据存储在 MySQL 数据库中。元数据包含实验、作业、pipeline 等产生的单一指标，这些指标会被聚合并用于排序和过滤。

  - **Artifacts：** Kubeflow Pipelines 将 Artifacts 存储在 Minio 等对象存储服务中。Artifacts 数据包含 Pipeline package、views 和大规模指标数据（如时间序列），并使用大规模指标数据来调试管道运行或调查单个运行的性能。

    MySQL 数据库和 Minio 服务均由 Kubernetes PersistentVolume 子系统提供支持。

- **Persistence agent and ML metadata：**Persistence agent 会监听由 Pipeline Service 创建的 Kubernetes 资源，并将这些资源的状态持久化在 ML 元数据服务中。 Pipeline Persistence Agent 会记录运行的容器及其输入和输出。 输入/输出由容器参数或数据 Artifacts  URI 组成。

- **Pipeline web server：**  Web 服务器从各种服务收集数据并以视图的方式展示，如当前运行的 Pipeline 列表、Pipeline 执行历史、Artifacts 列表等。



**安装 Pipeline 相关组件**

**前提条件**

 - [Kubernetes](https://kubernetes.io/docs/setup/production-environment/tools/) >= 1.8 
 - [kubectl](https://kubernetes.io/docs/tasks/tools/)
 - [Helm](https://helm.sh/docs/intro/install/) >= 3.0 
 - [Redis](https://redis.io/) （可以买云服务，或者本地安装单机版）

**1. Standalone模式安装（体验试用）**

如果您并不要多租户隔离功能，可以选着更加轻量级的 standalone 模式安装。

```bash
$ git clone https://github.com/PaddleFlow/paddlex-backend
$ cd hack && sh install-standalone.sh
```

以上命令会安装 Istio/knative/kubeflow pipeline/paddle-operator/JupyterHub 等组件，您可以通过以下命令查看各组件部署的情况。

```bash
$ kubectl get pods -n kubeflow
$ kubectl get pods -n istio
$ kubectl get pods -n knative
```

**多租户模式安装 （生产环境）**

如您需要完备的多租户隔离的功能，可以采用一键脚本安装的方式，在运行脚本前，需要先安装依赖的 [kustomize](https://kustomize.io/) 工具。

```bash
# 下载项目代码到本地
$ git clone https://github.com/PaddleFlow/paddlex-backend
$ cd hack && sh install.sh
```

以上命令会在 Kubernetes 集群中安装所有的依赖组件，需要等待一段时间。然后您可以通过以下命令来确认所有依赖的组件是否安装成功。

```bash
kubectl get pods -n cert-manager
kubectl get pods -n istio-system
kubectl get pods -n auth
kubectl get pods -n knative-eventing
kubectl get pods -n knative-serving
kubectl get pods -n kubeflow
kubectl get pods -n kubeflow-user-example-com
```

通过 Port-Forward 暴露 UI 界面的端口，从而访问 UI 界面。

```bash
kubectl port-forward svc/istio-ingressgateway -n istio-system 8080:80
```

打开浏览器访问 `http://localhost:8080`，默认用户名是 `user@example.com`，默认密码是 `12341234`

更多安装详情请参考链接：https://www.kubeflow.org/docs/components/pipelines/installation/overview/



**为云上飞桨定制的 Components**

云上飞桨产品以 Components YAML 文件的形式为用户提供快速构建深度学习工作流的编程组件。相关 Components YAML 文件存放在：https://github.com/PaddleFlow/paddlex-backend/tree/main/operator/yaml

下面将介绍 Components YAML 文件各个字段的含义。

**Dataset Component API 相关的字段：**

| 字段          | 类型    | 说明                                                |
| ------------- | ------- | --------------------------------------------------- |
| name          | String  | 数据集名称，如：imagenet                            |
| namespace     | String  | 命名空间，默认为kubeflow                            |
| action        | String  | 对自定义资源的操作，apply/create/patch/delete之一   |
| partitions    | Integer | 样本数据集缓存分区数，一个分区表示一个节点，默认为1 |
| source_uri    | String  | 样本数据的存储地址，如 bos://paddleflow/imagenet/   |
| source_secret | String  | 样本数据存储后端的访问秘钥                          |

**Training Component API 相关的字段：**

| 字段            | 类型    | 说明                                                  |
| --------------- | ------- | ----------------------------------------------------- |
| name            | String  | 模型名称，如: pp-ocr                                  |
| namespace       | String  | 命名空间，默认为kubeflow                              |
| action          | String  | 对自定义资源的操作，apply/create/patch/delete之一     |
| project         | String  | 飞桨生态套件项目名，如 PaddleOCR/PaddleClas等         |
| image           | String  | 包含飞桨生态套件代码的模型训练镜像                    |
| config_path     | String  | 模型配置文件的在套件项目中的相对路径                  |
| dataset         | String  | 模型训练任务用到的样本数据集                          |
| pvc_name        | String  | 工作流共享盘的PVC名称                                 |
| worker_replicas | Integer | 分布式训练 Worker 节点的并行度                        |
| config_changes  | String  | 模型配置文件修改内容                                  |
| pretrain_model  | String  | 预训练模型存储路径                                    |
| train_label     | String  | 训练集的样本标签文件名                                |
| test_label      | String  | 测试集的样本标签文件名                                |
| ps_replicas     | Integer | 分布式训练参数服务器并行度                            |
| gpu_per_node    | Integer | 每个 Worker 节点所需的GPU个数                         |
| use_visualdl    | Boolean | 是否开启模型训练日志可视化服务，默认为False           |
| save_inference  | Boolean | 是否需要保存推理可以格式的模型文件，默认为True        |
| need_convert    | Boolean | 是否需要将模型格式转化为Serving可用的格式，默认为True |

**ModelHub Component API 相关的字段：**

| 字段          | 类型   | 说明                                                    |
| ------------- | ------ | ------------------------------------------------------- |
| name          | String | 模型名称，如: pp-ocr                                    |
| namespace     | String | 命名空间，默认为kubeflow                                |
| action        | String | 对自定义资源的操作，apply/create/patch/delete之一       |
| endpoint      | String | 模型中心的URI，默认为http://minio-service.kubeflow:9000 |
| model_name    | String | 模型名称，与 Training API 的模型名称要保持一致          |
| model_version | String | 模型版本，默认为 latest                                 |
| pvc_name      | String | 工作流共享盘的PVC名称                                   |

**Serving  Component API 相关的字段：**

| 字段          | 类型    | 说明                                                    |
| ------------- | ------- | ------------------------------------------------------- |
| name          | String  | 模型名称，如: pp-ocr                                    |
| namespace     | String  | 命名空间，默认为kubeflow                                |
| action        | String  | 对自定义资源的操作，apply/create/patch/delete之一       |
| image         | String  | 包含 Paddle Serving 依赖的模型服务镜像                  |
| model_name    | String  | 模型名称，与 Training API 的模型名称要保持一致          |
| model_version | String  | 模型版本，与 ModelHub API 中的模型版本要保持一致        |
| endpoint      | String  | 模型中心的URI，默认为http://minio-service.kubeflow:9000 |
| port          | Integer | 模型服务的端口号，默认为8040                            |

更多 Pipeline 的使用案例可以参考案例教学 《文字检查模型全链路案例》