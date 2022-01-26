[TOC]

## 一、产品概览

### 1. 背景介绍

飞桨（PaddlePaddle）是全面开源开放、技术领先、功能完备的产业级深度学习平台，目前已广泛应用于工业、农业、服务业等。在众多飞桨用户中，一些客户还具有业务上云的需求，这就要求飞桨生态能提供完善且易用的云上组件，方便用户在云上落地飞桨深度学习工作流。虽然目前云上飞桨已有样本缓存、模型训练、模型服务等云原生组件，但缺少将各组件串起来共同完成飞桨深度学习工作流的工具。此外，由于大部分飞桨套件用户是各领域的模型开发人员，可能对 Kubernetes 相关的知识并不熟悉。因此，基于各飞桨生态套件构建该领域的常见深度学习工作流是十分有必要的，此外还需要为用户提供简单易用的编程接口，方便用户构建其自定义的工作流。基于以上原因，我们构建了云上飞桨产品旨在协助飞桨用户快速地完成深度学习工作流上云的需求。

### 2. 架构概览

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-b17c081128b2a58d8f84f66d975ba2fd6eb53cfa)

上图是云上飞桨产品的架构图。

在用户交互层，云上飞桨产品给用户提供了 Python SDK 和 Kubernetes 自定义资源两种编程接口。用户可以使用 Pipeline Python SDk 快速便捷的构建自己的飞桨深度学习工作流，也可以通过 Kubernetes CRD 单独使用某些功能组件。除了编程接口，云上飞桨产品在交互层还给用户提供了多租户JupyterHub Notebook服务，各种 UI 管理界面，如 Pipeline UI / 模型管理中心 UI / VisualDL可视化UI 等。

飞桨具有丰富的模型套件库，如 PaddleOCR、PaddleClas、PaddleDetection、PaddleNLP等，不过由于模型套件并不是面向云原生设计的，再者由于大部分模型套件用户对 Kubernetes 的知识并不熟悉，这就导致在云上基于飞桨模型套件落地深度学习工作流并非易事。所以我们将各个飞桨模型套件进行云原生化，各个模型套件的 Operator 共同组成了架构图中的模型套件层。目前 PaddleOCR 已经完成 Operator 化，其他飞桨模型套件的 Operator 开发工作将在后续开展。

模型套件的下一层是云上飞桨功能组件层，每个模型套件 Operator 都包含样本数据缓存、分布式训练、推理服务部署、VisualDL训练可视化等的功能。丰富的云上飞桨工作组件可以协助用户快速的开发部署模型，可大幅加速研发效率。

### 3. 产品优势

- **简单易用的编程接口**。云上飞桨产品给用户提供了简单易用的编程接口，用户可以使用 Python SDK 构建自定义云上飞桨工作流，也可以通过使用 Kubernetes 自定义资源单独使用各个云上飞桨功能组件。

- **规范的套件镜像版本管理。** 由于飞桨模型套件并不是面向云原生设计的，当前各套件的镜像管理相对混乱。此外，用户实际使用过程中可能需要对模型套件的代码进行修改，这就导致无法更新线上的套件镜像到最新版本。云上飞桨产品规范了模型套件镜像版本管理，后续也将基于 Teckton Pipeline 为飞桨模型套件提供镜像的持续集成与发布功能，方便用户在修改套件代码后，能够快速部署上线。

- **提供各领域场景的工作流模板。**云上飞桨模型套件 Operator 为用户提供了领域常见的工作流模板。如OCR场景，为用户提供了文字检查模型的工作流模板。用户只需修改少量参数即可基于模块快速部署上线飞桨工作流。

- **具有丰富的云上飞桨组件。** 云上飞桨具有丰富的功能组件，包括 VisualDL可视化工具、样本数据缓存组件、模型训练组件、推理服务部署组件等，这些功能组件可以大幅加速用户开发和部署模型的效率。

- **针对飞桨框架进行了定制优化。** 除了提升研发效率的功能组件，我们还针对飞桨框架进行了正对性优化，如样本缓存组件加速云上飞桨分布式训练作业、基于飞桨框架和调度器协同设计的GPU利用率优化。

  

## 二、各组件使用手册

### 1. JupyterHub Notebook使用手册

**JupyterHub Notebook 组件链接：** https://github.com/jupyterhub/zero-to-jupyterhub-k8s

云上飞桨产品使用 JupyterHub Notebook 作为用户开发模型的交互式编程入口。JupyterHub 是一个多用户的 Jupyter 门户，在设计之初就把多用户创建、资源分配、数据持久化等功能做成了插件模式。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-c15ccbdce6e861e20dfd8d0d5308f7ce2b9293ca)

上图是 JupyterHub 系统架构图，其由四个子系统组成：

 - **服务管理中心** 基于 tornado 构建，它是 JupyterHub 的核心模块。
 - **HTTP 代理模块** 可以作为节点代理，用于接收来自客户端浏览器的请求。
 - **Spawners模块**  用于监控的多个用户的 Jupyter notebook 服务器。
 - **身份验证模块** 用于管理用户权限并授权用户如何访问系统界面。

zero to jupyterhub k8s 项目支持多租户隔离、动态资源分配、数据持久化、数据隔离、高可用、权限控制等功能，能够支持公司级别（上万）用户规模。

**在 Kubernetes 集群安装 JupyterHub**

在安装 JupyterHub 前您需要有 Kubernetes 集群环境，并在当前节点上安装好 [Helm，](https://helm.sh/docs/)Helm 可以通过以下命令来安装：

```bash
curl https://raw.githubusercontent.com/helm/helm/HEAD/scripts/get-helm-3 | bash
```

检查 Helm 状态与版本

```bash
helm version
```

**1）初始化 Helm Chart 配置文件**

Helm Chart 可以渲染需要安装的 Kubernetes 资源的模板。通过 使用 Helm Chart，用户可以覆盖 Chart 的默认值进行自定义安装。

在这一步中，我们将初始化一个图表配置文件，以便您修改 JupyterHub 安装时的初试值。从版本 1.0.0 开始，您无需进行任何配置即可开始使用，因此只需创建一个带有一些有用注释的 config.yaml 文件即可。创建  config.yaml 文件如下：

```yaml
# This file can update the JupyterHub Helm chart's default configuration values.
#
# For reference see the configuration reference and default values, but make
# sure to refer to the Helm chart version of interest to you!
#
# Introduction to YAML:     https://www.youtube.com/watch?v=cdLNKUoMc6c
# Chart config reference:   https://zero-to-jupyterhub.readthedocs.io/en/stable/resources/reference.html
# Chart default values:     https://github.com/jupyterhub/zero-to-jupyterhub-k8s/blob/HEAD/jupyterhub/values.yaml
# Available chart versions: https://jupyterhub.github.io/helm-chart/
#
```

**2）将 JupyterHub Chart 添加到 Helm 仓库**

```bash
helm repo add jupyterhub https://jupyterhub.github.io/helm-chart/
helm repo update
```

输出如下：

```bash
Hang tight while we grab the latest from your chart repositories...
...Skip local chart repository
...Successfully got an update from the "stable" chart repository
...Successfully got an update from the "jupyterhub" chart repository
Update Complete. ⎈ Happy Helming!⎈
```

**3）安装 JupyterHub Chart**

在包含上述 config.yaml 文件的目录中，执行如下命令：

```bash
helm upgrade --cleanup-on-fail \
  --install <helm-release-name> jupyterhub/jupyterhub \
  --namespace <k8s-namespace> \
  --create-namespace \
  --version=<chart-version> \
  --values config.yaml
```

其中：

- `<helm-release-name>` 指的是 [Helm 版本名称](https://helm.sh/docs/glossary/#release)，用于区分 Chart 安装的标识符，当您更改或删除此 Chart 安装的配置时，版本号是必填项。 如果您的 Kubernetes 集群将包含多个 JupyterHub，请确保区分它们，您可以使用 `helm list` 查看各个 JupyterHub Chart 的版本。
- `<k8s-namespace>` 指的是 [Kubernetes 命名空间](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)，在本例中是用于对 Kubernetes 资源进行分组的标识符 与 JupyterHub Chart 关联的所有 Kubernetes 资源。 你需要命名空间标识符来执行任何带有 `kubectl` 的命令。
- 此步骤可能需要一些时间，在此期间您的终端不会有任何输出。JupyterHub 正在后台安装。
- 如果安装过程中出现 `release named <helm-release-name> already exists` 错误，那么您应该通过运行 `helm delete <helm-release-name>` 删除该版本。 然后重复此步骤重新安装。 如果仍然存在，请执行 `kubectl delete namespace <k8s-namespace>` 并重试。
- 如果安装步骤出现错误，请在重新运行安装命令之前通过运行 `helm delete <helm-release-name>` 删除 Helm 版本。
- 如果拉取 Docker 镜像出现错误，比如 `Error: timed out waiting for the condition ，您可以在 `helm` 命令中添加一个 `--timeout=<number-of-minutes>m` 参数。
- `--version` 参数对应的是 Helm Chart 的*版本*，而不是 JupyterHub 的版本。 每个版本的 JupyterHub Helm Chart 都与特定版本的 JupyterHub 配对。 例如，Helm 图表的 `0.11.1` 运行 JupyterHub `1.3.0`。 有关每个版本的 JupyterHub Helm Chart 中安装了哪个 JupyterHub 版本的列表，请参阅 [Helm Chart 仓库 ](https://jupyterhub.github.io/helm-chart/)。

**4）查看 Pod 状态**

在第 2 步运行时，您可以在另一个终端中输入下面的命令来查看正在创建的 pod：

```bash
kubectl get pod --namespace jhub
```

**5）等待所有 Pod 成功运行**

等待 hub 和 proxy pod 进入 Running 状态。

```bash
NAME                    READY     STATUS    RESTARTS   AGE
hub-5d4ffd57cf-k68z8    1/1       Running   0          37s
proxy-7cb9bc4cc-9bdlp   1/1       Running   0          37s
```

**6）获取 JupyterHub Notebook 访问 IP**

运行以下命令，直到 proxy-public 服务的 EXTERNAL-IP 可用，如示例输出中所示。

```bash
$ kubectl get service --namespace <k8s-namespace>
NAME           TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)        AGE
hub            ClusterIP      10.51.243.14    <none>          8081/TCP       1m
proxy-api      ClusterIP      10.51.247.198   <none>          8001/TCP       1m
proxy-public   LoadBalancer   10.51.248.230   104.196.41.97   80:31916/TCP   1m
```

最后，请在浏览器中输入代理公共服务的外部 IP。 JupyterHub 使用默认的虚拟身份验证器运行，因此输入任何用户名和密码组合都可以让您登入 Notebook。用户界面如下：

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-f8b9d6a5a2b4e342aad987f27799bcc6abc6ad60)

更多自定义安装文档请查看：https://zero-to-jupyterhub.readthedocs.io/en/latest/jupyterhub/installation.html



### 2. 样本数据缓存组件使用手册

**样本数据缓存组件代码：** https://github.com/PaddleFlow/paddle-operator/tree/sampleset

在 Kubernetes 的架构体系中，计算与存储是分离的，这给数据密集型的深度学习作业带来较高的网络IO开销。为了解决该问题，我们基于 JuiceFS 在开源项目 Paddle Operator 中实现了样本缓存组件，大幅提升了云上飞桨分布式训练作业的执行效率。

**背景介绍**

由于云计算平台具有高可扩展性、高可靠性、廉价性等特点，越来越多的机器学习任务运行在Kubernetes集群上。因此我们开源了Paddle Operator项目，通过提供PaddleJob自定义资源，让云上用户可以很方便地在Kubernetes集群使用飞桨（PaddlePaddle）深度学习框架运行模型训练作业。

然而，在深度学习整个pipeline中，样本数据的准备工作也是非常重要的一环。目前云上深度学习模型训练的常规方案主要采用手动或脚本的方式准备数据，这种方案比较繁琐且会带来诸多问题。比如将HDFS里的数据复制到计算集群本地，然而数据会不断更新，需要定期的同步数据，这个过程的管理成本较高；或者将数据导入到远程对象存储，通过制作PV和PVC来访问样本数据，从而模型训练作业就需要访问远程存储来获取样本数据，这就带来较高的网络IO开销。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-6d25edebdced1033430a7990c0bdd731dbf0e6c1)



为了方便云上用户管理样本数据，加速云上飞桨框架分布式训练作业，我们与JuiceFS社区合作，联合推出了面向飞桨框架的样本缓存与管理方案，该方案期望达到如下目标：

 - **基于JuiceFS加速远程数据访问。** JuiceFS是一款面向云环境设计的高性能共享文件系统，其在数据组织管理和访问性能上进行了大量针对性的优化。基于JuiceFS实现样本数据缓存引擎，能够提供高效的文件访问性能。
 - **充分利用本地存储，缓存加速模型训练。** 能够充分利用计算集群本地存储，比如内存和磁盘，来缓存热点样本数据集，并配合缓存亲和性调度，在用户无感知的情况下，智能地将作业调度到有缓存的节点上。这样就不用反复访问远程存储，从而加速模型训练速度，一定程度上也能提升GPU资源的利用率。
 - **数据集及其管理操作的自定义资源抽象。** 将样本数据集及其管理操作抽象成Kubernetes的自定义资源，屏蔽数据操作的底层细节，减轻用户心智负担。用户可以很方便地通过操作自定义资源对象来管理数据，包括数据同步、数据预热、清理缓存、淘汰历史数据等，同时也支持定时任务。
 - **统一数据接口，支持多种存储后端。** 样本缓存组件要能够支持多种存储后端，并且能提供统一的POSIX协议接口，用户无需在模型开发和训练阶段使用不同的数据访问接口，降低模型开发成本。同时样本缓存组件也要能够支持从多个不同的存储源导入数据，适配用户现有的数据存储状态。

**面临的挑战**

然而，在Kubernetes的架构体系中，计算与存储是分离的，这种架构给上诉目标的实现带来了些挑战，主要体现在如下几点：

- **Kubernetes 调度器是缓存无感知的，**也就是说kube-scheduler并没有针对本地缓存数据的调度策略，因此模型训练作业未必能调度到有缓存的节点，从而导致缓存无法重用。如何实现缓存亲和性调度，协同编排训练作业与缓存数据，是我们面临的首要问题。
- **在数据并行的分布式训练任务中，单机往往存放不下所有的样本数据，**因此样本数据是要能够以分区的形式分散缓存在各计算节点上。然而，我们知道负责管理自定义资源的控制器（Controller Manager）不一定运行在缓存节点上，如何通过自定义控制器来管理分布式的缓存数据，也是实现该方案时要考虑的难点问题。
- 除了前述两点，如何结合飞桨框架合理地对样本数据进行分发和预热，提高本地缓存命中率，减少作业访问远程数据的次数，从而提高作业执行效率，这还需要进一步的探索。

针对上述问题，我们在开源项目Paddle Operator中提供了样本缓存组件，较好地解决了这些挑战，下文将详细阐述我们的解决方案。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-f7af13a5235fe75ede11cbd48e947e335c2e0f51)

上图是Paddle Operator的整体架构，其构建在Kubernetes上，包含如下三个主要部分：

**1.自定义API资源（Custom Resource）**

Paddle Operator定义了三个CRD，用户可编写和修改对应的YAML文件来管理训练作业和样本数据集。

 - **PaddleJob**是飞桨分布式训练作业的抽象，它将Parameter
   Server（参数服务器）和Collective（集合通信）两种分布式深度学习架构模式统一到一个CRD中，用户通过创建PaddleJob可以很方便地在Kubernetes集群运行分布式训练作业。

 - **SampleSet**是样本数据集的抽象，数据可以来自远程对象存储、HDFS或Ceph等分布式文件系统，并且可以指定缓存数据的分区数、使用的缓存引擎、多级缓存目录等配置。

 - **SampleJob**定义了些样本数据集的管理作业，包括数据同步、数据预热、清除缓存、淘汰历史旧数据等操作，支持用户设置各个数据操作命令的参数，
   同时还支持以定时任务的方式运行数据管理作业。

**2. 自定义控制器（Controller Manager）**

控制器在 Kubernetes 的 Operator 框架中是用来监听 API 对象的变化（比如创建、修改、删除等），然后以此来决定实际要执行的具体工作。

 - **PaddleJob Controller**负责管理PaddleJob的生命周期，比如创建参数服务器和训练节点的Pod，并维护工作节点的副本数等。

 - **SampleSet Controller**负责管理SampleSet的生命周期，其中包括创建
   PV/PVC等资源对象、创建缓存运行时服务、给缓存节点打标签等工作。

 - **SampleJob Controller**负责管理SampleJob的生命周期，通过请求缓存运行时服务的接口，触发缓存引擎异步执行数据管理操作，并获取执行结果。

**3.缓存引擎（Cache Engine）**

缓存引擎由缓存运行时服务（Cache Runtime Server）和JuiceFS存储插件（JuiceFS CSI Driver）两部分组成，提供了样本数据存储、缓存、管理的功能。

 - **Cache Runtime Server** 负责样本数据的管理工作，接收来自SampleSet Controller和SampleJob Controller的数据操作请求，并调用JuiceFS客户端完成相关操作执行。
 - **JuiceFS CSI Driver** 是JuiceFS社区提供的CSI插件，负责样本数据的存储与缓存工作，将样本数据缓存到集群本地并将数据挂载进PaddleJob的训练节点。

**难点突破与优化**

在上述的整体架构中，Cache Runtime Server是非常重要的一个组件，它由Kubernetes原生的API资源StatefulSet实现，在每个缓存节点上都会运行该服务，其承担了缓存数据分区管理等工作，也是解决难点问题的突破口。下图是用户创建SampleSet后，Paddle Operator对PaddleJob完成缓存亲和性调度的大概流程。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-6fc880b7d7c11d8b211f0ca5ea8bcecf57d124d8)

当用户创建SampleSet后，SampleSet Controller就会根据 SampleSet中的配置创建出PV和PVC。在PV和PVC完成绑定后，SampleSet Controller则会将PVC添加到Runtime Pod模板中，并创建出指定分区数的Cache Runtime Server。在Cache Runtime Server成功调度到相应节点后，SampleSet Controller则会对该节点做标记，且标记中还带有缓存的分区数。这样等下次用户提交PaddleJob时，Paddle Controller会自动地给Paddle Worker Pod添加nodeAffinity和PVC字段，这样调度器（Scheduler）就能将Paddle Worker调度到指定的缓存分区节点上，这即实现了对模型训练作业的缓存亲和性调度。

值得一提的是，在该调度方案中，Paddle框架的训练节点能够做到与缓存分区一一对应的，这能够最大程度上地利用本地缓存的优势。当然，该方案同时也支持对PaddleJob的扩缩容，当PaddleJob的副本数大于SampleSet的分区数时（这也是可以调整的），PaddleJob Controller并不会对多出来Paddle Worker做nodeAffinity限制，这些Paddle Worker还可以通过挂载的Volume访问远程存储来获取样本数据。

解决了缓存亲和性调度的问题，我们还面临着SampleSet/SampleJob Controller如何管理分布式样本缓存数据集，如何合理地做分布式预热的挑战。使用JuiceFS CSI存储插件可以解决样本数据存储与缓存的问题，但由于Kubernetes CSI插件提供的接口有限（只提供了与 Volume 挂载相关的接口），所以需要有额外的数据管理服务驻守在缓存节点，即：Cache Runtime Server。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-c9e28e6464e2c42f5e0163b5251d1f7de8b672a2)

上图是 Cache Runtime Server 内部工作流程示意图，其中封装了JuiceFS客户端的可执行文件。Runtime Server提供了三种类型的接口，它们的作用分别是：1）上传数据操作命令的参数；2）获取数据操作命令的结果；3）获取样本数据集及其缓存的状态。

Server在接收到Controller上传的命令参数后，会将参数写到指定路径，然后会触发Worker进程异步地执行相关操作命令，并将命令执行结果写到结果路径，然后Controller可以通过调用相关接口获取数据管理作业的执行结果。

数据同步、清理缓存、淘汰旧数据这三个操作比较容易实现，而数据预热操作的实现相对会复杂些。因为，当样本数据量比较大时，单机储存无法缓存所有数据，这时就要考虑对数据进行分区预热和缓存。为了最大化利用本地缓存和存储资源，我们期望对数据预热的策略要与飞桨框架读取样本数据的接口保持一致。

因此，对于WarmupJob我们目前实现了两种预热策略:**Sequence**和**Random**，分别对应飞桨框架SequenceSampler和DistributedBatchSampler两个数据采样API。JuiceFS的Warmup命令支持通过--file参数指定需要预热的文件路径，故将0号Runtime Server作为Master，负责给各个分区节点分发待预热的数据，即可实现根据用户指定的策略对样本数据进行分布式预热的功能。

至此，难点问题基本都得以解决，该方案将缓存引擎与飞桨框架紧密结合，充分利用了本地缓存来加速云上飞桨的分布式训练作业。

**性能测试**

为了验证Paddle Operator样本缓存加速方案的实际效果，我们选取了常规的ResNet50模型以及ImageNet数据集来进行性能测试，并使用了PaddleClas项目中提供的模型实现代码。具体的实验配置如下：

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-c4a3dd375368d1051ae3735409af488ce516c77f)

基于以上配置，我们做了两组实验。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-ac4c625aa0a843ddad8722b495155fb835d38bdb)第一组实验对比了样本数据缓存前后的性能差异，从而验证了使用样本缓存组件来加速云上飞桨训练作业的必要性。

如上图，在样本数据缓存到训练节点前，1机1卡和1机2卡的训练速度分别为211.26和224.23 images/s；在样本数据得以缓存到本地后，1机1卡和1机2卡的训练速度分别为383.01和746.76 images/s。可以看出，在计算与存储分离的Kubernetes集群里，由于带宽有限（本实验的带宽为50Mbps），训练作业的主要性能瓶颈在于远程样本数据IO上。带宽的瓶颈并不能通过调大训练作业的并行度来解决，并行度越高，算力浪费越为严重。

因此，使用样本缓存组件提前将数据预热到训练集群本地，可以大幅加速云上飞桨训练作业的执行效率。在本组实验中，1机1卡训练效率提升了81.3%，1机2卡的训练速度提升了233%。

此外，使用样本缓存组件预热数据后，1机1卡383.01 images/s的训练速度与直接在宿主机上的测试结果一致，也就是说，缓存引擎本身基本上没有带来性能损耗。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-fd12235449456cc4ea16cd0e05daa3950bd082d5)

第二组实验对比了使用1机1卡时样本数据预热前后JuiceFS与BOS FS的性能差异，相比于BOS FS，JuiceFS在访问远程小文件的场景下具有更优的性能表现。

如上图，在样本数据预热到本地前，需要通过CSI插件访问远程对象存储（如BOS）中的样本数据，使用BOS FS CSI插件与JuiceFS CSI插件的训练速度分别是69.71和211.26 images/s；在样本数据得以缓存在本地后，使用BOS FS CSI插件与JuiceFS CSI插件的训练速度分别是381.98和382.43 images/s，性能基本没有差异。

由此可以看出，在访问远程小文件的场景下，JuiceFS相比BOS FS有近3倍的性能提升。两者的性能差异可能与JuiceFS的文件存储格式和BOS FS的实现有关，这有待进一步验证。除了性能上的考量，JuiceFS提供的文件系统挂载服务更加稳定，功能完善且易用，这也是优先选择JuiceFS作为底层缓存引擎的重要原因。



**安装样本缓存组件**

下文将讲述如何安装部署 Paddle Operator 样本缓存组件，并通过实际的示例演示了缓存组件的基础功能。

**前提条件**

- Kubernetes >= 1.8
- kubectl

**创建 PaddleJob / SampleSet / SampleJob 自定义资源**

```bash
kubectl apply -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/main/deploy/v1/crd.yaml
```

创建成功后，可以通过以下命令来查看创建的自定义资源

```bash
$ kubectl get crd | grep batch.paddlepaddle.org
NAME                                    CREATED AT
paddlejobs.batch.paddlepaddle.org       2021-08-23T08:45:17Z
samplejobs.batch.paddlepaddle.org       2021-08-23T08:45:18Z
samplesets.batch.paddlepaddle.org       2021-08-23T08:45:18Z
```

**部署 Operator**

```bash
kubectl apply -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/main/deploy/v1/operator.yaml
```

**部署样本缓存组件的 Controller**

以下命令中的 YAML 文件包括了自定义资源 SampleSet 和 SampleJob 的 Controller

```bash
kubectl apply -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/main/deploy/extensions/controllers.yaml
```

通过以下命令可以查看部署好的 Controller

```
$ kubectl -n paddle-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
paddle-controller-manager-776b84bfb4-5hd4s   1/1     Running   0          60s
paddle-samplejob-manager-69b4944fb5-jqqrx    1/1     Running   0          60s
paddle-sampleset-manager-5cd689db4d-j56rg    1/1     Running   0          60s
```

**安装 CSI 存储插件**

目前 Paddle Operator 样本缓存组件仅支持 [JuiceFS](https://github.com/juicedata/juicefs/blob/main/README_CN.md) 作为底层的样本缓存引擎，样本访问加速和缓存相关功能主要由缓存引擎来驱动。

部署 JuiceFS CSI Driver

```bash
kubectl apply -f https://raw.githubusercontent.com/juicedata/juicefs-csi-driver/master/deploy/k8s.yaml
```

部署好 CSI 驱动后，您可以通过以下命令查看状态，Kubernetes 集群中的每一个 worker 节点应该都会有一个 **juicefs-csi-node** Pod。

```bash
$ kubectl -n kube-system get pods -l app.kubernetes.io/name=juicefs-csi-driver
NAME                                          READY   STATUS    RESTARTS   AGE
juicefs-csi-controller-0                      3/3     Running   0          13d
juicefs-csi-node-87f29                        3/3     Running   0          13d
juicefs-csi-node-8h2z5                        3/3     Running   0          13d
```

**注意**：如果 Kubernetes 无法发现 CSI 驱动程序，并出现类似这样的错误：**driver name csi.juicefs.com not found in the list of registered CSI drivers**，这是由于 CSI 驱动没有注册到 kubelet 的指定路径，您可以通过下面的步骤进行修复。

在集群中的 worker 节点执行以下命令来获取 kubelet 的根目录

```
ps -ef | grep kubelet | grep root-dir
```

在上述命令打印的内容中，找到 `--root-dir` 参数后面的值，这既是 kubelet 的根目录。然后将以下命令中的 `{{KUBELET_DIR}}` 替换为 kubelet 的根目录并执行该命令。

```
curl -sSL https://raw.githubusercontent.com/juicedata/juicefs-csi-driver/master/deploy/k8s.yaml | sed 's@/var/lib/kubelet@{{KUBELET_DIR}}@g' | kubectl apply -f -
```

更多详情信息可参考 [JuiceFS CSI Driver](https://github.com/juicedata/juicefs-csi-driver)



**缓存组件使用示例**

由于 JuiceFS 缓存引擎依靠 Redis 来存储文件的元数据，并支持多种对象存储作为数据存储后端，为了方便起见，本示例使用 Redis 同时作为元数据引擎和数据存储后端。

**1）准备 Redis 数据库**

您可以很容易的在云计算平台购买到各种配置的云 Redis 数据库，本示例使用 Docker 在 Kubernetes 集群的 worker 节点上运行一个 Redis 数据库实例。

```bash
docker run -d --name redis \
	-v redis-data:/data \
	-p 6379:6379 \
	--restart unless-stopped \
	redis redis-server --appendonly yes
```

**注意**：以上命令将本地目录 `redis-data` 挂载到 Docker 容器的 `/data` 数据卷中，您可以按需求挂载不同的文件目录。

**2）创建 Secret**

准备好 JuiceFS 用于格式化文件系统需要的字段，需要的字段如下：

| 字段       | 含义                  | 说明                                                         |
| ---------- | --------------------- | ------------------------------------------------------------ |
| name       | JuiceFS 文件系统名称  | 可以是任意的字符串，用来给 JuiceFS 文件系统命名              |
| storage    | 对象存储类型          | 例如：bos [JuiceFS 支持的对象存储和设置指南](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/how_to_setup_object_storage.md) |
| bucket     | 对象存储的 URL        | 存储数据的桶路径，例如： bos://imagenet.bj.bcebos.com/imagenet |
| metaurl    | 元数据存储的 URL      | 如 redis://username:password@host:6379/0，[JuiceFS 支持的元数据存储引擎](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/databases_for_metadata.md) |
| access-key | 对象存储的 Access Key | 可选参数；如果对象存储使用的是 username 和 password，例如使用 Redis 作为存储后端，这里填用户名。 |
| secret-key | 对象存储的 Secret key | 可选参数；如果对象存储使用的是 username 和 password，例如使用 Redis 作为存储后端，这里填密码。 |

更多详情信息可以参考：[JuiceFS 快速上手](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/quick_start_guide.md) ；[JuiceFS命令参考](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/command_reference.md)

本示例中各字段配置如下，因为没有给 Redis 配置密码，字段 `access-key` 和 `secret-key` 可以不设置。

| 字段    | 说明                         |
| ------- | ---------------------------- |
| name    | imagenet                     |
| storage | redis                        |
| bucket  | redis://192.168.7.227:6379/1 |
| metaurl | redis://192.168.7.227:6379/0 |

**注意**：请将 IP `192.168.7.227` 替换成您在第一步中运行 Redis 容器的宿主机的 IP 地址。

然后创建 secret.yaml 文件

```yaml
apiVersion: v1
data:
  name: aW1hZ2VuZXQ=
  storage: cmVkaXM=
  bucket: cmVkaXM6Ly8xOTIuMTY4LjcuMjI3OjYzNzkvMQ==
  metaurl: cmVkaXM6Ly8xOTIuMTY4LjcuMjI3OjYzNzkvMA==
kind: Secret
metadata:
  name: imagenet
  namespace: paddle-system
type: Opaque
```

其中各字段的值经过 Base64 编码，您可以通过下面的示例命令来获取个字段 Base64 编码后的值。

```bash
$ echo "redis://192.168.7.227:6379/0" | base64
cmVkaXM6Ly8xOTIuMTY4LjcuMjI3OjYzNzkvMAo=
```

使用 kubectl 命令创建 secret

```bash
$ kubectl create -f secret.yaml
secret/imagenet created
```

**3）创建 SampleSet**

编写 imagenet.yaml 如下，数据源来自 bos (百度对象存储)公开的 bucket。

```bash
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleSet
metadata:
  name: imagenet
  namespace: paddle-system
spec:
  # 分区数，一个Kubernetes节点表示一个分区
  partitions: 1
  source:
    uri: bos://paddleflow-public.hkg.bcebos.com/imagenet/demo
  secretRef:
    name: imagenet
```

创建 SampleSet，等待数据完成同步，并查看 SampleSet 的状态。数据同步操作可能比较耗时，请耐心等待。

```bash
$ kubectl create -f imagenet.yaml
sampleset.batch.paddlepaddle.org/imagenet created
$ kubectl get sampleset -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   58 MiB       200 B         12 GiB        1/1       Ready   76s
$ kubectl get pods -n paddle-system
NAME                                         READY   STATUS    RESTARTS   AGE
imagenet-runtime-0                           1/1     Running   0          30s
paddle-controller-manager-776b84bfb4-qs67f   1/1     Running   0          11h
paddle-samplejob-manager-685449d6d7-cmqfb    1/1     Running   0          11h
paddle-sampleset-manager-69bc7fb85d-4rjcg    1/1     Running   0          11h
```

等创建的 SampleSet 其 PHASE 状态为 Ready 时，表示该数据集可以使用了。

**4）体验 SampleJob（可选）**

缓存组件提供了 SampleJob 用来做样本数据集的管理，目前支持了4种 Job 类型，分别是 sync/warmup/clear/rmr。

**注意**：由于 SampleSet 的数据缓存信息不是实时更新的，SampleJob 执行完成后缓存信息会在 30s 内完成更新。

删除缓存引擎中的数据，rmr job 里的 rmrOptions 是必填的，paths 里的参数是数据的相对路径。

```bash
$ cat rmr-job.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleJob
metadata:
  name: imagenet-rmr
  namespace: paddle-system
spec:
  type: rmr
  sampleSetRef:
    name: imagenet
  rmrOptions:
    paths:
      - n01514859

$ kubectl create -f rmr-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-rmr created
$ kubectl get samplejob imagenet-rmr -n paddle-system
NAME           PHASE
imagenet-rmr   Succeeded
$ kubectl get sampleset imagenet -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   4.0 KiB      200 B         12 GiB        1/1       Ready   29m
```

将远程数据源中的样本数据同步到缓存引擎中，数据大概4GiB左右，12.8万张图片，请确保有足够的内存空间。

```bash
$ cat sync-job.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleJob
metadata:
  name: imagenet-sync
  namespace: paddle-system
spec:
  type: sync
  sampleSetRef:
    name: imagenet
  # sync job 需要填写 secret 信息
  secretRef:
    name: imagenet
  syncOptions:
    source: bos://paddleflow-public.hkg.bcebos.com/imagenet

$ kubectl create -f sync-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-sync created
$ kubectl get samplejob imagenet-sync -n paddle-system
NAME            PHASE
imagenet-sync   Succeeded
$ kubectl get sampleset imagenet -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   4.2 GiB      4.2 GiB       7.3 GiB       1/1       Ready   83m
```

清理集群中的缓存数据

```bash
$ cat clear-job.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleJob
metadata:
  name: imagenet-clear
  namespace: paddle-system
spec:
  type: clear
  sampleSetRef:
    name: imagenet
    
$ kubectl create -f clear-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-clear created
$ kubectl get samplejob imagenet-clear -n paddle-system
NAME             PHASE
imagenet-clear   Succeeded
$ kubectl get sampleset imagenet -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   4.2 GiB      80 B          12 GiB        1/1       Ready   85m
```

将缓存引擎中的数据预热到 Kubernetes 集群中

```bash
$ cat warmup-job.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleJob
metadata:
  name: imagenet-warmup
  namespace: paddle-system
spec:
  type: warmup
  sampleSetRef:
    name: imagenet

$ kubectl create -f warmup-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-warmup created
$ kubectl get samplejob imagenet-warmup -n paddle-system
NAME              PHASE
imagenet-warmup   Succeeded
$ kubectl get sampleset imagenet -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   4.2 GiB      4.2 GiB       7.3 GiB       1/1       Ready   90m
```

**5）创建 PaddleJob**

以下示例使用 nginx 镜像来简单示范下如何在 PaddleJob 中声明使用 SampleSet 样本数据集。

编写 ps-demo.yaml 文件如下：

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: ps-demo
  namespace: paddle-system
spec:
  sampleSetRef:
    # 申明要使用的 SampleSet
    name: imagenet
    # SampleSet 样本数据集在 Worker Pods 的挂载路径
    mountPath: /mnt/imagenet
  worker:
    replicas: 1
    template:
      spec:
        containers:
          - name: sample
            image: nginx
  ps:
    replicas: 1
    template:
      spec:
        containers:
          - name: sample
            image: nginx
```

查看 PaddleJob 的状态

```bash
$ kubectl get paddlejob -n paddle-system
NAME      STATUS    MODE   AGE
ps-demo   Running   PS     112s
```

查看挂载在 PaddleJob worker pod 的样本数据

```bash
$ kubectl exec -it ps-demo-worker-0 -n paddle-system -- /bin/bash
$ ls /mnt/imagenet
demo  train  train_list.txt
```



### 3. 模型训练组件使用手册

**模型训练组件代码：** https://github.com/PaddleFlow/paddle-operator

模型训练组件（Paddle-Operator）旨在为 Kubernetes 上运行飞桨任务提供标准化接口、训练任务管理的定制化完整支持。PaddleJob 是 Paddle-Operator 在 Kubernetes 云平台上的 CRD，并且 Paddle-Operator 定义了与其对应的 Controller。借助 PaddleJob，飞桨深度学习作业可以快速分布在 Kubernetes 集群中运行。数据科学家和机器学习工程师可以通过 PaddleJob CRD 创建飞桨深度学习作业，监控和查看深度学习的模型训练进度和状态，管理飞桨深度学习作业的生命周期。

**Paddle-Operator 主要解决在 Kubernetes 上运行飞桨分布式深度学习任务的以下几个痛点：**

- 缺少针对飞桨任务的定制

- 缺少统一（将一个分布式任务作为一个整体）的资源调度
- 缺少统一任务启动
- 手动指定分布式任务的地址和端口
- 缺少统一的资源回收

**Paddle-Operator适用范围：**

- 可以无缝集成到云原生Kubernetes平台（百度云CCE、阿里云ACK、华为云CCE等公有云原生Kubernetes平台，或者私有Kubernetes平台）中单独使用，使Paddle任务在Kubernetes平台上快速便捷的运行起来。
- 可以无缝集成到基于Kubeflow生态中，支持基于Kubernetes+Kubeflow搭建的机器学习平台中集成PaddlePaddle，提升PaddlePaddle的在主流平台社区的兼容性。
- 可以配合Kubernetes系统中的第三方批量调度组件（Kube-batch、Volcano、Coscheduling等）一起使用，减少业务和平台方集成PaddlePaddle的成本。

**目前 PaddleJob 支持定义如下功能：**

 - 支持 PS/Collective/Heter 三种分布式训练模式。
 - 支持使用 Volcano 进行模型训练作业的调度。
 - 支持训练作业的弹性扩缩容并具有容错保障。

**架构概览**

下图是模型训练组件的架构流程图，描述了从 PaddleJob CRD注册，接收 PaddleJob 任务，到创建 PaddleJob 和调度 PaddleJob 的流程。

![架构流程图](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-83d85ee7f9b603f0be6ef384dd2d627a4322b1ab)

以下是 PaddleJob 进行 Reconcile 的机制：
![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-334d770898f31403b2823a9e9b82d372dc291799)

下面我们通过训练 Wide & Deep 和 ResNet50 模型来简要说明如何使用 PaddleJob 进行分布式训练。

**安装paddle-operator** 
安装 paddle-operator 需要有已经安装的 kubernetes (v1.8+) 集群和 kubectl (v1.8+) 工具。本节所需配置文件和示例可以在 [paddle-operator项目](https://github.com/PaddleFlow/paddle-operator) 找到， 可以通过 git clone 或者复制文件内容保存。

```
deploy
|-- examples
|   |-- resnet.yaml
|   |-- wide_and_deep.yaml
|   |-- wide_and_deep_podip.yaml
|   |-- wide_and_deep_service.yaml
|   `-- wide_and_deep_volcano.yaml
|-- v1
|   |-- crd.yaml
|   `-- operator.yaml
`-- v1beta1
    |-- crd.yaml
    `-- operator.yaml
```

**部署 CRD**

注意：kubernetes 1.15 及以下使用 v1beta1 目录，1.16 及以上使用目录 v1.

执行以下命令，

```bash
$ kubectl create -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/dev/deploy/v1/crd.yaml
```

或者

```bash
$ kubectl create -f deploy/v1/crd.yaml
```

注意：v1beta1 请根据报错信息添加 –validate=false 选项

通过以下命令查看是否成功，

```bash
$ kubectl get crd
NAME                                    CREATED AT
paddlejobs.batch.paddlepaddle.org       2021-02-08T07:43:24Z
```

**部署 controller 及相关组件**
注意：默认部署的 namespace 为 paddle-system，如果希望在自定义的 namespace 中运行或者提交任务， 需要先在 operator.yaml 文件中对应更改 namespace 配置，其中

 - namespace: paddle-system 表示该资源部署的 namespace，可理解为系统 controller namespace；
 - Deployment 资源中 containers.args 中 –namespace=paddle-system 表示 controller 监控资源所在 namespace，即任务提交 namespace。

执行以下部署命令，

```bash
$ kubectl create -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/dev/deploy/v1/operator.yaml
```

或者

```bash
$ kubectl create -f deploy/v1/operator.yaml
```

通过以下命令查看部署结果和运行状态，

```bash
$ kubectl -n paddle-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
paddle-controller-manager-698dd7b855-n65jr   1/1     Running   0          1m
```

通过查看 controller 日志以确保运行正常，

```bash
$ kubectl -n paddle-system logs paddle-controller-manager-698dd7b855-n65jr
```

提交 demo 任务查看效果，

```bash
$ kubectl -n paddle-system create -f deploy/examples/wide_and_deep.yaml
```

查看 paddlejob 任务状态, pdj 为 paddlejob 的缩写，

```bash
$ kubectl -n paddle-system get pdj
NAME                     STATUS      MODE   PS    WORKER   AGE
wide-ande-deep-service   Completed   PS     2/2   0/2      4m4s
```

以上信息可以看出：训练任务已经正确完成，该任务为 ps 模式，配置需求 2 个 pserver, 2 个在运行，需求 2 个 woker，0 个在运行（已完成退出）。 可通过 cleanPodPolicy 配置任务完成/失败后的 pod 删除策略，详见任务配置。

查看 pod 状态，

```bash
$ kubectl -n paddle-system get pods
```

**卸载**

通过以下命令卸载部署的组件，

```bash
$ kubectl delete -f deploy/v1/crd.yaml -f deploy/v1/operator.yaml
```

注意：重新安装时，建议先卸载再安装

**paddlejob 任务提交**

在上述安装过程中，我们使用了 wide-and-deep 的例子作为提交任务演示，本节详细描述任务配置和提交流程供用户参考提交自己的任务， 镜像的制作过程可在 *docker 镜像* 章节找到。

**示例 wide and deep**

本示例采用 PS 模式，使用 cpu 进行训练，所以需要配置 ps 和 worker。


准备配置文件，

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  cleanPodPolicy: Never
  withGloo: 1
  worker:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
  ps:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
```

说明：

- 提交命名需要唯一，如果存在冲突请先删除原 paddlejob 确保已经删除再提交;
- ps 模式时需要同时配置 ps 和 worker，collective 模式时只需要配置 worker 即可；
- withGloo 可选配置为 0 不启用， 1 只启动 worker 端， 2 启动全部(worker端和Server端)， 建议设置 1；
- cleanPodPolicy 可选配置为 Always/Never/OnFailure/OnCompletion，表示任务终止（失败或成功）时，是否删除 pod，调试时建议 Never，生产时建议 OnCompletion；
- intranet 可选配置为 Service/PodIP，表示 pod 间的通信方式，用户可以不配置, 默认使用 PodIP；
- ps 和 worker 的内容为 podTemplateSpec，用户可根据需要遵从 kubernetes 规范添加更多内容, 如 GPU 的配置.

提交任务: 使用 kubectl 提交 yaml 配置文件以创建任务，

```bash
$ kubectl -n paddle-system create -f demo-wide-and-deep.yaml
```

**示例 resnet**

本示例采用 Collective 模式，使用 gpu 进行训练，所以只需要配置 worker，且需要配置 gpu。

准备配置文件，

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: resnet
spec:
  cleanPodPolicy: Never
  worker:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-resnet:v1
            command:
            - python
            args:
            - "-m"
            - "paddle.distributed.launch"
            - "train_fleet.py"
            volumeMounts:
            - mountPath: /dev/shm
              name: dshm
            resources:
              limits:
                nvidia.com/gpu: 1
        volumes:
        - name: dshm
          emptyDir:
            medium: Memory
```

注意：

- 这里需要添加 shared memory 挂载以防止缓存出错；
- 本示例采用内置 flower 数据集，程序启动后会进行下载，根据网络环境可能等待较长时间。

提交任务: 使用 kubectl 提交 yaml 配置文件以创建任务，

```bash
$ kubectl -n paddle-system create -f resnet.yaml
```

**更多配置**

**Volcano 支持**

paddle-operator 支持使用 volcano 进行复杂任务调度，使用前请先 [安装](https://github.com/volcano-sh/volcano) 。

本节使用 volcano 实现 paddlejob 运行的 gan-scheduling。

使用此功能需要进行如下配置：

- 创建 paddlejob 同名 podgroup，具体配置信息参考 volcano 规范；
- 在 paddlejob 任务配置中添加声明：schedulerName: volcano , 注意：需要且只需要在 worker 中配置。

配置示例，

```yaml
---
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  cleanPodPolicy: Never
  withGloo: 1
  worker:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        schedulerName: volcano
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
  ps:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1

---
apiVersion: scheduling.volcano.sh/v1beta1
kind: PodGroup
metadata:
  name: wide-ande-deep
spec:
  minMember: 4
```

在以上配置中，我们通过创建最小调度单元为 4 的 podgroup，并将 paddlejob 任务标记使用 volcano 调度，实现了任务的 gan-scheduling。

可以通过以下命运提交上述任务查看结果，

```bash
$ kubectl -n paddle-system create -f deploy/examples/wide_and_deep.yaml
```

**GPU 和节点选择**

更多配置示例，

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  intranet: Service
  cleanPodPolicy: OnCompletion
  worker:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
            resources:
              limits:
                nvidia.com/gpu: 1
        nodeSelector:
          accelerator: nvidia-tesla-p100
  ps:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
            resources:
              limits:
                nvidia.com/gpu: 1
        nodeSelector:
          accelerator: nvidia-tesla-p100
```

**数据存储**

在 kubernentes 中使用挂载存储建议使用 pv/pvc 配置，详见 [persistent-volumes](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) 。

这里使用 nfs 云盘作为存储作为示例，配置文件如下，

```
$ cat pv-pvc.yaml
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Recycle
  storageClassName: slow
  mountOptions:
    - hard
    - nfsvers=4.1
  nfs:
    path: /nas
    server: 10.12.201.xx

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-pvc
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 10Gi
  storageClassName: slow
  volumeName: nfs-pv
```

使用以下命令在 namespace paddle-system 中 创建 pvc 名为 nfs-pvc 的存储声明，实际引用为 10.12.201.xx 上的 nfs 存储。

```
$ kubectl -n paddle-system apply -f pv-pvc.yaml
```

注意 pvc 需要绑定 namespace 且只能在该 namespace 下使用。

提交 paddlejob 任务时，配置 volumes 引用以使用对应存储，

```
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: paddlejob-demo-1
spec:
  cleanPolicy: OnCompletion
  worker:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/paddle-ubuntu:2.0.0-18.04
            command: ["bash","-c"]
            args: ["cd /nas/wide_and_deep; python3 train.py"]
            volumeMounts:
            - mountPath: /nas
              name: data
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: nfs-pvc
  ps:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/paddle-ubuntu:2.0.0-18.04
            command: ["bash","-c"]
            args: ["cd /nas/wide_and_deep; python3 train.py"]
            volumeMounts:
            - mountPath: /nas
              name: data
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: nfs-pvc
```

该示例中，镜像仅提供运行环境，训练代码和数据均通过存储挂载的方式添加。



### 4. 推理服务组件使用手册

**模型服务部署项目链接：** https://github.com/PaddleFlow/ElasticServing

ElasticServing 通过提供自定义资源 PaddleService，支持用户在 Kubernetes 集群上使用 TensorFlow、onnx、PaddlePaddle 等主流框架部署模型服务。 ElasticServing 构建在 Knative Serving 之上，其提供了自动扩缩容、容错、健康检查等功能，并且支持在异构硬件上部署服务，如 Nvidia GPU 或 昆仑芯片。 ElasticServing 采用的是 serverless 架构，当没有预估请求时，服务规模可以缩容到零，以节约集群资源，同时它还支持并蓝绿发版等功能。

**架构概览**

下图是 ElasticServing 的架构流程图，其使用了 Knative 提供的自动扩缩容机制，并使用 Istio 作为推理服务的统一网关。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-cdc4d860be727666b9bd73c08e04a955bd68883b)

**安装 ElasticServing 组件**

本示例使用的模型服务镜像基于 [Paddle Serving CPU 版](https://github.com/PaddlePaddle/Serving/blob/v0.6.0/README_CN.md) 构建而成.

前提条件：

- Kubernetes >= 1.18
- 安装 Knative Serving 依赖的网络插件 请查考 [安装指南](https://knative.dev/v0.21-docs/install/any-kubernetes-cluster/#installing-the-serving-component) 或者执行脚本： `hack/install_knative.sh`(knative serving v0.21 with istio) / `hack/install_knative_kourier.sh`(knative serving v0.22 with kourier).

**1）通过以下命令安装 CRD 和 Operator**

```bash
# 下载 ElasticServing
git clone https://github.com/PaddleFlow/ElasticServing.git
cd ElasticServing

# 安装 CRD
kubectl apply -f assets/crd.yaml

# 安装自定义 Controller
kubectl apply -f assets/elasticserving_operator.yaml
```

**2）运行示例**

```bash
# 部署 paddle service
kubectl apply -f assets/sample_service.yaml
```

**3）检查服务状态**

```bash
# 查看命名空间 paddleservice-system 下的 Service
kubectl get svc -n paddleservice-system

# 查看命名空间 paddleservice-system 下的 knative service
kubectl get ksvc -n paddleservice-system

# 查看命名空间 paddleservice-system 下的 pod
kubectl get pods -n paddleservice-system

# 查看 Paddle Service Pod 的日志信息
kubectl logs <pod-name> -n paddleservice-system -c paddleserving
```

本示例使用 Istio 插件作为 Knative Serving 的网络方案，您也可以使用其他的网络插件比如：Kourier 和 Ambassador。

```bash
# Find the public IP address of the gateway (make a note of the EXTERNAL-IP field in the output)
kubectl get service istio-ingressgateway --namespace=istio-system
# If the EXTERNAL-IP is pending, get the ip with the following command
kubectl get po -l istio=ingressgateway -n istio-system -o jsonpath='{.items[0].status.hostIP}'
# If you are using minikube, the public IP address of the gateway will be listed once you execute the following command (There will exist four URLs and maybe choose the second one)
minikube service --url istio-ingressgateway -n istio-system

# Get the port of the gateway
kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}'

# Find the URL of the application. The expected result may be http://paddleservice-sample.paddleservice-system.example.com
kubectl get ksvc paddle-sample-service -n paddleservice-system
```



**示例1 部署 ResNet50 图片分类服务**

下面我们以 ResNet50 模型为例，部署图片分类服务，并简要说明如何使用 ElasticServing 进行模型服务部署。

1）编写 sample_service.yaml 如下:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  labels:
    istio-injection: enabled
  name: paddleservice-system
---
apiVersion: elasticserving.paddlepaddle.org/v1
kind: PaddleService
metadata:
  name: paddleservice-sample
  namespace: paddleservice-system
spec:
  canary:
    arg: cd Serving/python/examples/imagenet && python3 resnet50_web_service_canary.py
      ResNet50_vd_model cpu 9292
    containerImage: jinmionhaobaidu/resnetcanary
    port: 9292
    tag: latest
  canaryTrafficPercent: 50
  default:
    arg: cd Serving/python/examples/imagenet && python3 resnet50_web_service.py ResNet50_vd_model
      cpu 9292
    containerImage: jinmionhaobaidu/resnet
    port: 9292
    tag: latest
  runtimeVersion: paddleserving
  service:
    minScale: 0
    window: 10s
```

**注意：**上述 Yaml 文件 Spec 部分只有 default 是必填的字段，其他字段可以是为空。如果您自己的 paddleservice 不需要字段 canary 和 canaryTrafficPercent，可以不填。

执行如下命令来创建 PaddleService

```bash
kubectl apply -f /dir/to/this/yaml/sample_service.yaml
```

2）检查服务是否可用
执行如下命令查看服务是否可用。

```bash
curl -H "host:paddleservice-sample.paddleservice-system.example.com" -H "Content-Type:application/json" -X POST -d '{"feed":[{"image": "https://paddle-serving.bj.bcebos.com/imagenet-example/daisy.jpg"}], "fetch": ["score"]}' http://<IP-address>:<Port>/image/prediction
```

3）预期输出结果

```bash
# 期望的输出结果如下
default:
{"result":{"label":["daisy"],"prob":[0.9341399073600769]}}

canary:
{"result":{"isCanary":["true"],"label":["daisy"],"prob":[0.9341399073600769]}}
```



**示例2 中文分词模型服务**

本示例采用 lac 中文分词模型来做服务部署，更多模型和代码的详情信息可以查看 [Paddle Serving](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/lac/README_CN.md).

**1）构建服务镜像（可选）**

本示例模型服务镜像基于 `registry.baidubce.com/paddlepaddle/serving:0.6.0-devel` 构建而成，并上传到公开可访问的镜像仓库 `registry.baidubce.com/paddleflow-public/lac-serving:latest` 。 如您需要 GPU 或其他版本的基础镜像，可以查看文档 [Docker 镜像](https://github.com/PaddlePaddle/Serving/blob/v0.6.0/doc/DOCKER_IMAGES_CN.md), 并按照如下步骤构建镜像。

1. 下载 Paddle Serving 代码

```bash
$ wget https://github.com/PaddlePaddle/Serving/archive/refs/tags/v0.6.0.tar.gz
$ tar xzvf Serving-0.6.0.tar.gz
$ mv Serving-0.6.0 Serving
$ cd Serving
```

2. 编写如下 Dockerfile

```dockerfile
FROM registry.baidubce.com/paddlepaddle/serving:0.6.0-devel

WORKDIR /home

COPY . /home/Serving

WORKDIR /home/Serving

# install depandences
RUN pip install -r python/requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple && \
    pip install paddle-serving-server==0.6.0 -i https://pypi.tuna.tsinghua.edu.cn/simple && \
    pip install paddle-serving-client==0.6.0 -i https://pypi.tuna.tsinghua.edu.cn/simple

WORKDIR /home/Serving/python/examples/lac

RUN python3 -m paddle_serving_app.package --get_model lac && \
    tar xzf lac.tar.gz && rm -rf lac.tar.gz

ENTRYPOINT ["python3", "-m", "paddle_serving_server.serve", "--model", "lac_model/", "--port", "9292"]
```

3. 构建镜像

```
docker build . -t registry.baidubce.com/paddleflow-public/lac-serving:latest
```

**2）创建 PaddleService**

1. 编写 YAML 文件

```yaml
# lac.yaml
apiVersion: elasticserving.paddlepaddle.org/v1
kind: PaddleService
metadata:
  name: paddleservice-sample
  namespace: paddleservice-system
spec:
  default:
    arg: python3 -m paddle_serving_server.serve --model lac_model/ --port 9292
    containerImage: registry.baidubce.com/paddleflow-public/lac-serving
    port: 9292
    tag: latest
  runtimeVersion: paddleserving
  service:
    minScale: 1
```

2. 创建 PaddleService

```bash
$ kubectl apply -f lac.yaml
paddleservice.elasticserving.paddlepaddle.org/paddleservice-lac created
```

**3）查看服务状态**

1. 您可以使用以下命令查看服务状态

```bash
# Check service in namespace paddleservice-system
kubectl get svc -n paddleservice-system | grep paddleservice-lac

# Check knative service in namespace paddleservice-system
kubectl get ksvc paddleservice-lac -n paddleservice-system

# Check pods in namespace paddleservice-system
kubectl get pods -n paddleservice-system
```

2. 运行以下命令获取 ClusterIP

```bash
$ kubectl get svc paddleservice-lac-default-private -n paddleservice-system
```

**3）测试 BERT 模型服务**

模型服务支持 HTTP / BRPC / GRPC 三种客户端访问，客户端代码和环境配置详情请查看文档 [中文分词模型服务](https://github.com/PaddlePaddle/Serving/blob/develop/python/examples/lac/README_CN.md) 。

通过以下命令可以简单测试下服务是否正常

```bash
# 注意将 IP-address 和 Port 替换成上述 paddleservice-criteoctr-default-private service 的 cluster-ip 和端口。
curl -H "Host: paddleservice-sample.paddleservice-system.example.com" -H "Content-Type:application/json" -X POST -d '{"feed":[{"words": "我爱北京天安门"}], "fetch":["word_seg"]}' http://<IP-address>:<Port>/lac/prediction
```

预期结果

```json
{"result":[{"word_seg":"\u6211|\u7231|\u5317\u4eac|\u5929\u5b89\u95e8"}]}
```



### 5. Pipeline相关组件使用手册

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



### 6. VisualDL可视化组件使用手册

**VisualDL项目链接：** https://github.com/PaddlePaddle/VisualDL

**VisualDL Pipeline Components：**https://github.com/PaddleFlow/paddlex-backend/blob/main/components/visualdl.yaml

VisualDL是飞桨可视化分析工具，以丰富的图表呈现训练参数变化趋势、模型结构、数据样本、高维数据分布等。可帮助用户更清晰直观地理解深度学习模型训练过程及模型结构，进而实现高效的模型优化。

VisualDL提供丰富的可视化功能，支持标量、图结构、数据样本（图像、语音、文本）、超参数可视化、直方图、PR曲线、ROC曲线及高维数据降维呈现等诸多功能，同时VisualDL提供可视化结果保存服务，通过VDL.service生成链接，保存并分享可视化结果。具体功能使用方式，请参见 [**VisualDL使用指南**](https://github.com/PaddlePaddle/VisualDL/blob/develop/docs/components/README_CN.md)。如欲体验最新特性，欢迎试用我们的[**在线演示系统**](https://www.paddlepaddle.org.cn/paddle/visualdl/demo)。

**核心功能**

- API设计简洁易懂，使用简单。模型结构一键实现可视化。
- 功能覆盖标量、数据样本、图结构、直方图、PR曲线及数据降维可视化。
- 全面支持Paddle、ONNX、Caffe等市面主流模型结构可视化，广泛支持各类用户进行可视化分析。
- 与飞桨服务平台及工具组件全面打通，为您在飞桨生态系统中提供最佳使用体验。

**使用方式**

云上飞桨基于 Kubeflow Pipeline 将 VirsualDL 以 Components 的形式方便用户部署可视化组件，如果您使用的是 Training Components 则只需要再构建 Training Op 时将参数 use_visualdl 设置为 True 即可。开启训练日志可视化服务，需要在模型训练代码中进行日志打点，下面将详细讲述如何在模型代码中使用 VisualDL Python SDK 来打印特定格式的指标日志。

**1）安装 VisualDL Python SDK**

使用 pip 安装

```bash
python -m pip install visualdl -i https://mirror.baidu.com/pypi/simple
```

**2）记录日志**

VisualDL将训练过程中的数据、参数等信息储存至日志文件中后，启动面板即可查看可视化结果。

```python
class LogWriter(logdir=None,
                max_queue=10,
                flush_secs=120,
                filename_suffix='',
                display_name='',
                file_name='',
                **kwargs)
```

接口参数如下：

| 参数            | 格式   | 含义                                                         |
| --------------- | ------ | ------------------------------------------------------------ |
| logdir          | string | 日志文件所在的路径，VisualDL将在此路径下建立日志文件并进行记录，如果不填则默认为`runs/${CURRENT_TIME}` |
| max_queue       | int    | 日志记录消息队列的最大容量，达到此容量则立即写入到日志文件   |
| flush_secs      | int    | 日志记录消息队列的最大缓存时间，达到此时间则立即写入到日志文件（日志消息队列到达最大缓存时间或最大容量，都会立即写入日志文件） |
| filename_suffix | string | 为默认的日志文件名添加后缀                                   |
| display_name    | string | 指定面板启动后显示的路径，如不指定此项则显示日志所在的实际路径，当日志所在路径过长或想隐藏日志所在路径时可指定此参数 |
| file_name       | string | 指定写入的日志文件名，如果指定的文件名已经存在，则将日志续写在此文件中，因此可通过此参数实现日志续写的功能，文件名必须包括`vdlrecords` |

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-ee85ee8b66d8cc706aa991537221e52b72a00a1b)

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-ee8dd2c694449690976b1a7e1c89f31e9aecea63)

**示例**

设置日志文件并记录标量数据：

```python
from visualdl import LogWriter

# 在`./log/scalar_test/train`路径下建立日志文件
with LogWriter(logdir="./log/scalar_test/train") as writer:
    # 使用scalar组件记录一个标量数据,将要记录的所有数据都记录在该writer中
    writer.add_scalar(tag="acc", step=1, value=0.5678)
    writer.add_scalar(tag="acc", step=2, value=0.6878)
    writer.add_scalar(tag="acc", step=3, value=0.9878)
# 如果不想使用上下文管理器`with`，可拆解为以下几步完成：
"""
writer = LogWriter(logdir="./log/scalar_test/train")

writer.add_scalar(tag="acc", step=1, value=0.5678)
writer.add_scalar(tag="acc", step=2, value=0.6878)
writer.add_scalar(tag="acc", step=3, value=0.9878)

writer.close()
"""
```

注：调用LogWriter(logdir="./log/scalar_test/train")将会在./log/scalar_test/train目录下生成一个日志文件， 运行一次程序所产生的训练数据应该只记录到一个日志文件中，因此应该只调用一次LogWriter，用返回的LogWriter对象来记录所有数据， 而不是每记录一个数据就创建一次LogWriter。

如下是错误示范：

```python
from visualdl import LogWriter
with LogWriter(logdir="./log/scalar_test/train") as writer:  # 将会创建日志文件vdlrecords.xxxx1.log
    writer.add_scalar(tag="acc", step=1, value=0.5678)  # 数据写入./log/scalar_test/train/vdlrecords.xxxx1.log
with LogWriter(logdir="./log/scalar_test/train") as writer:  # 将会创建日志文件vdlrecords.xxxx2.log
    writer.add_scalar(tag="acc", step=2, value=0.6878)  # 数据将会写入./log/scalar_test/train/vdlrecords.xxxx2.log
```

**3）启动面板（本地体验可选）**

本地体验可以按照下面步骤进行操作，如果是使用 Pipeline Components 的话只需要再构建 Op 时指定函数参数即可。

在上述示例中，日志已记录三组标量数据，现可启动VisualDL面板查看日志的可视化结果。

使用命令行启动VisualDL面板，命令格式如下：

```bash
visualdl --logdir <dir_1, dir_2, ... , dir_n> --model <model_file> --host <host> --port <port> --cache-timeout <cache_timeout> --language <language> --public-path <public_path> --api-only
```

参数详情：

| 参数            | 意义                                                         |
| --------------- | ------------------------------------------------------------ |
| --logdir        | 设定日志所在目录，可以指定多个目录，VisualDL将遍历并且迭代寻找指定目录的子目录，将所有实验结果进行可视化 |
| --model         | 设定模型文件路径(非文件夹路径)，VisualDL将在此路径指定的模型文件进行可视化，目前可支持PaddlePaddle、ONNX、Keras、Core ML、Caffe等多种模型结构，详情可查看[graph支持模型种类](https://github.com/PaddlePaddle/VisualDL/blob/develop/docs/components/README_CN.md#功能操作说明-4) |
| --host          | 设定IP，默认为`127.0.0.1`，若想使得本机以外的机器访问启动的VisualDL面板，需指定此项为`0.0.0.0`或自己的公网IP地址 |
| --port          | 设定端口，默认为`8040`                                       |
| --cache-timeout | 后端缓存时间，在缓存时间内前端多次请求同一url，返回的数据从缓存中获取，默认为20秒 |
| --language      | VisualDL面板语言，可指定为'en'或'zh'，默认为浏览器使用语言   |
| --public-path   | VisualDL面板URL路径，默认是'/app'，即访问地址为'http://<host>:<port>/app' |
| --api-only      | 是否只提供API，如果设置此参数，则VisualDL不提供页面展示，只提供API服务，此时API地址为'http://<host>:<port>/<public_path>/api'；若没有设置public_path参数，则默认为'http://<host>:<port>/api' |

针对上一步生成的日志，启动命令为：

```
visualdl --logdir ./log
```

**可视化图表说明请查考链接：** https://github.com/PaddlePaddle/VisualDL



**在 Pipeline 中使用 VisualDL**

如您需在 Kubernetes 集群上部署 VisualDL 服务，可以使用原生的 Deployment 来进行部署。如需集成到飞桨工作流中，云上飞桨基于 kubeflow Pipeline 组件为您封装好了 VisualDL Components 并在 Training Components 也有集成，下面将详细介绍如何使用 VisualDL Component 或在 Training Component 中开启可视化功能。

VisualDL Component Yaml 文件链接：https://github.com/PaddleFlow/paddlex-backend/blob/main/components/visualdl.yaml

Training Component Yaml 文件链接：https://github.com/PaddleFlow/paddlex-backend/blob/main/operator/yaml/training.yaml

**示例1  用 VisualDL Component 部署可视化服务**

在构建 Paddle Workflow 的代码中添加 VisualDL Op，并在模型训练任务完成之后运行。需要注意的是当前 VisualDL 不支持分布式训练任务日志的可视化，所以 PaddleJob 的 work 并行度需要设置为 1。下面的代码展示了如何在工作流代码中构建 VisualDL Op:

```python
import kfp
import kfp.dsl as dsl
from kfp import components

def create_volume_op():
    return dsl.VolumeOp(
        name="PPOCR Detection PVC",
        resource_name="ppocr-detection-pvc",
        storage_class="task-center",
        size="10Gi",
        modes=dsl.VOLUME_MODE_RWM
    ).set_display_name("create pvc and pv for PaddleJob"
    ).add_pod_annotation(name="pipelines.kubeflow.org/max_cache_staleness", value="P0D")

def create_visualdl_op(volume_op):
    create_visualdl = components.load_component_from_file("../../components/visualdl.yaml")
    visualdl_op = create_visualdl(
        name="ppocr-visualdl",
        namespace="paddleflow",
        pvc_name=volume_op.volume.persistent_volume_claim.claim_name,
        mount_path="/mnt/task-center/",
        logdir="/mnt/task-center/models/vdl/",
        model=f"/mnt/task-center/ppocr/server/__model__",
    )
    visualdl_op.set_display_name("deploy VisualDL")
    return visualdl_op

def create_training_op(volume_op):
    worker_replicas = 1
    ...

@dsl.pipeline(
  name="ppocr-detection-demo",
  description="An example for using ppocr train .",
)
def ppocr_detection_demo():
    volume_op = create_volume_op()
    training_op = create_training_op(volume_op)
    training_op.after(volume_op)
    visualdl_op = create_visualdl_op(volume_op)
    visualdl_op.after(training_op)
    ....

....
```

**示例2  在 Training Component 中开启 VisualDL 可视化服务**

除了使用 VisualDL Component，在 Pipeline 组件的 Training Component 中也内置的常规使用 VisualDL 的方式，如特殊需求，只需要在 Training Component 中开始 VisualDL 功能即可，即设置 use_visualdl=True。下面的代码演示了如何在构建 Training Op 时开启 VisualDL 的功能：

```python
...

def create_training_op(volume_op):
    """
    使用飞桨生态套件进行模型训练的组件，支持PS和Collective两种架构模式
    :param volume_op: 共享存储盘
    :return: TrainingOp
    """
    training_op = components.load_component_from_file("./yaml/training.yaml")
    return training_op(
        name="ppocr-det",
        dataset="icdar2015",  # 数据集
        project="PaddleOCR",  # Paddle生态项目名
        worker_replicas=1,    # Collective模式Worker并行度
        gpu_per_node=1,       # 指定每个worker所需的GPU个数
        use_visualdl=True,    # 是否启动模型训练日志可视化服务
        train_label="train_icdar2015_label.txt",   # 训练集的label
        test_label="test_icdar2015_label.txt",     # 测试集的label
        config_path="configs/det/det_mv3_db.yml",  # 模型训练配置文件
        pvc_name=volume_op.volume.persistent_volume_claim.claim_name,  # 共享存储盘
        # 模型训练镜像
        image="registry.baidubce.com/paddleflow-public/paddleocr:2.1.3-gpu-cuda10.2-cudnn7",
        # 修改默认模型配置
        config_changes="Global.epoch_num=10,Global.log_smooth_window=2,Global.save_epoch_step=5",
        # 预训练模型URI
        pretrain_model="https://paddle-imagenet-models-name.bj.bcebos.com/dygraph/MobileNetV3_large_x0_5_pretrained.pdparams",
    ).set_display_name("model training")
  
  
...
```

VisualDL 服务部署好后，使用下面的命令即可将服务的端口暴露给宿主机。

```bash
$ kubectl port-forward svc/ppocr-det-visualdl -n kubeflow 8040:8040
```

注意：ppocr-det-visualdl 是 Service 的名称，kubeflow 是 Service ppocr-det-visualdl 所在的命名空间。

然后，访问 https://<local-ip>:8040 即可访问 VisualDL 的 Web UI。

![image-20220110144410952](/Users/chenenquan/Library/Application Support/typora-user-images/image-20220110144410952.png)

### 7. 小模型场景GPU利用率优化

目前基于飞桨框架和 Kubernetes 调度器协同设计的 GPU 利用率优化项目还在开发中，飞桨框架部分的PR: https://github.com/PaddleFlow/Paddle/pull/2/files

**背景**

- GPU显存与算力（SM）利用率低下，显存使用率超过一半的GPU占比不足20%，算力利用率超过80%的GPU占比不足十分之一。
- 使用 gang-schedule 的调度策略，模型训练作业申请的GPU资源越多，算力资源空置等待的时间越长。
- 模型训练过程中对GPU资源的需求是动态变化的，一般会申请足够的资源保证任务正常运行，然而这种固定资源分配的方式也会带来一定程度的资源浪费。

**框架与调度器的协同设计**

传统的GPU共享方案：资源隔离（劫持CUDA API调用）、时间片模型、MPS模式。

缺点：作业间相互影响，难以保证高优作业的SLA，开发难度大，落地生产难。

![image-20220110154922617](/Users/chenenquan/Library/Application Support/typora-user-images/image-20220110154922617.png)



框架与调度器的协同设计方案

- 将任务划分为 Resource-guarantee job 和 Opportunistic job，保证高优作业执行效率的同时，使用机会作业提升集群利用率。
- 协同设计深度学习框架与 Kubernetes 调度器，实现动态调整作业的显存与算力上限，并根据统计的作业与设备信息，对不同作业类型实施不同的调度策略。

**飞桨框架—显存动态调整**

1. 作业启动后，框架根据模型申请的显存资源，设置一个合适的显存上界。
2. 当有个 mini-batch 显存需求突增且设备显存不足时，则将待创建的 Tensor申明到内存中（Pinned Memory），保证作业能够正常运行。
3. 将同设备机会作业的显存上界下调，出让显存给高优先级作业，保证高优作业的执行效率。
4. 机会作业出让显存后，高优作业上调显存上界，下个 mini-batch 的 Tensor 又可以申明在显存中。
5. 框架持续对作业显存使用情况打点，配合 Local Coordinator 动态调整显存上界。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-055d31e5c57d86553720a3af9eb2a242f5888fe1)

飞桨框架显存动态调整实现代码类图：

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-1536f7d1d5bbead5f6c848bbcfdd97ec81569971)

**飞桨框架—算力动态调整**

模型训练作业特点：

- 训练过程中包含特征提取、数据增强等不使用GPU算力的流程，独占模式会浪费GPU算力（Exclusive mode）。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-ab3bbad60ec9f2d39e59ba4150a855027a2b87b5)

- 作业打包模式会带来 GPU kernel 排队延迟和 PCIe带宽抢占的问题，影响高优作业执行效率。

  ![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-a31dc14cddbc768889c9a6e780925b1a52c1e1d4)

算力动态调整方案：

接管算子（Op）执行流程，通过给 Op 加上延迟执行时间实现 Op 级别的分时复用。

1. 实现 GpuOpManager 接管 Op 执行流程，待执行 Op 依序进入执行队列，并简单分配延迟执行时间。
2. GpuOpManager 不断监控算子执行时间和 GPU 算力利用率并打点，配合 Local Coordinator 根据一定策略动态调整后续 Op 延迟执行时间。如高优作业 Op 的算力被抢占时，则延长机会作业 Op 的等待执行时间。

**Kubernets 调度器—整体框架**

两种类型作业

- resource-guarantee jobs：声明有 resource quota，保证高优作业的SLA。
- opportunistic jobs：使用碎片化的GPU资源( best-effort )，提升集群资源利用率。

调度器部分由两个模块进行配合

Local Coordinator

- 收集本地设备信息和作业信息，包括GPU的显存算力利用率、作业 mini-batch 执行时间等信息，上传给 Global Scheduler。
- 负责本地作业端到端的执行管理，根据框架统计的作业信息，动态调整作业的显存与算力上限，并写入文件。

Global Scheduler

- 维护多个租户队列，并对高优作业与机会作业使用不同的调度策略， 保证资源分配的公平性。
- 根据 Local Coordinator 上传的信息和集群拓扑状态（如 NVLink 等），负责全局的作业调度。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-f8f2dd2521f9449c199e41058135533db32ede3e)



## 三、案例教程

### 1. 文字检查模型全链路案例

**安装 Pipeline及相关组件**

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



**创建 Notebook 用于 Pipeline 开发**

安装好云上飞桨产品后，您可以通过以下命令来获取安装在集群中的 JupyterHub 服务的 IP 地址。

```bash
$ kubectl get service -n jhub
NAME           TYPE           CLUSTER-IP       EXTERNAL-IP                  PORT(S)        AGE
hub            ClusterIP      172.16.135.111   <none>                       8081/TCP       34d
proxy-api      ClusterIP      172.16.237.93    <none>                       8001/TCP       34d
proxy-public   LoadBalancer   172.16.234.64    120.48.xx.xx,192.168.48.18   80:30185/TCP   34d
```

从上述命令的输出中找到 proxy-public 对应的 EXTERNAL-IP 地址，然后访问 http://120.48.xx.xx:80 链接，打开 JupyterHub UI 界面，然后创建用户名和密码即可使用。

进入 notebook 页面后，在 work 目录下创建 yaml 文件夹，并将 [Components YAML](https://github.com/PaddleFlow/paddlex-backend/tree/main/operator/yaml) 等文件上传到该目录中，如下图：
![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-c5c731354f0f376bbfb52dea6da94fecf13e9b18)

然后，在work目录下创建新的 Notebook 用于编写工作流代码。构建工作流需要使用 Python SDK，所以需要安装相关依赖包：

```python
In[1]  !pip install kfp -i https://mirror.baidu.com/pypi/simple
```

**编写 PP-OCR 模型训练与部署工作流代码**

PP-OCR 模型训练与部署示例代码：https://github.com/PaddleFlow/paddlex-backend/blob/main/operator/pipeline-demo.py

以下通过 OCR 领域的文本检测场景为例，展示了 PP-OCR 模型云上训练到服务部署工作流的落地实践。

**1）准备 PP-OCR Pipeline 的共享盘**

由于从模型训练到模型服务部署各阶段间，会产出临时文件，比如模型训练日志、模型快照、不同格式的模型文件等，所以需要创建一个共享盘，这样工作流中各阶段任务就能够共享这些临时文件，完成整个流程。以下为创建 Pipeline 共享盘的示例代码：

```Python
import kfp
import kfp.dsl as dsl
from kfp import components

def create_volume_op():
    """
    创建 PaddleOCR Pipeline 所需的共享存储盘
    :return: VolumeOp
    """
    return dsl.VolumeOp(
        name="PPOCR Detection PVC",
        resource_name="ppocr-detection-pvc",
        storage_class="task-center",
        size="10Gi",
        modes=dsl.VOLUME_MODE_RWM
    ).set_display_name("create volume for pipeline")
```

**2）准备样本数据集并缓存到集群本地**

在进行模型训练任务之前，需要将样本数据从远程的存储服务中拉取到本地，并以分布式的形式缓存到训练集群的节点上，这样能够大幅加上模型训练作业的效率。这里用到的样本数据集是 icdar2015。

```python
def create_dataset_op():
    """
    将样本数据集拉取到训练集群本地并缓存
    :return: DatasetOp
    """
    dataset_op = components.load_component_from_file("./yaml/dataset.yaml")
    return dataset_op(
        name="icdar2015",
        partitions=1,                 # 缓存分区数
        source_secret="data-source",  # 数据源的秘钥
        source_uri="bos://paddleflow-public.hkg.bcebos.com/icdar2015/"  # 样本数据URI
    ).set_display_name("prepare sample data")
```

**3）开始进行PP-OCR模型的训练任务**

Training API 中提供开启 VisualDL 模型训练日志可视化的接口，同时还提供了模型转化和预训练模型接口，您可以通过指定相关参数来进行配置。

```python
def create_training_op(volume_op):
    """
    使用飞桨生态套件进行模型训练的组件，支持PS和Collective两种架构模式
    :param volume_op: 共享存储盘
    :return: TrainingOp
    """
    training_op = components.load_component_from_file("./yaml/training.yaml")
    return training_op(
        name="ppocr-det",
        dataset="icdar2015",  # 数据集
        project="PaddleOCR",  # Paddle生态项目名
        worker_replicas=1,    # Collective模式Worker并行度
        gpu_per_node=1,       # 指定每个worker所需的GPU个数
        use_visualdl=True,    # 是否启动模型训练日志可视化服务
        train_label="train_icdar2015_label.txt",   # 训练集的label
        test_label="test_icdar2015_label.txt",     # 测试集的label
        config_path="configs/det/det_mv3_db.yml",  # 模型训练配置文件
        pvc_name=volume_op.volume.persistent_volume_claim.claim_name,  # 共享存储盘
        # 模型训练镜像
        image="registry.baidubce.com/paddleflow-public/paddleocr:2.1.3-gpu-cuda10.2-cudnn7",
        # 修改默认模型配置
        config_changes="Global.epoch_num=10,Global.log_smooth_window=2,Global.save_epoch_step=5",
        # 预训练模型URI
        pretrain_model="https://paddle-imagenet-models-name.bj.bcebos.com/dygraph/MobileNetV3_large_x0_5_pretrained.pdparams",
    ).set_display_name("model training")
```

**4）将训练好的模型上传到模型存储服务并进行版本管理**
云上飞桨产品使用 Minio 组件来提供模型存储服务，您可以通过 ModelHub 相关 API 来进行版本管理。

```python
def create_modelhub_op(volume_op):
    """
    模型转换、存储、版本管理组件
    :param volume_op:
    :return:
    """
    modelhub_op = components.load_component_from_file("./yaml/modelhub.yaml")
    return modelhub_op(
        name="ppocr-det",
        model_name="ppocr-det",  # 模型名称
        model_version="latest",  # 模型版本号
        pvc_name=volume_op.volume.persistent_volume_claim.claim_name,  # 共享存储盘
    ).set_display_name("upload model file")
```

**5）部署模型在线推理服务**
Serving 组件支持蓝绿发版、自动扩缩容等功能，您可以通过相关参数进行配置。

```python
def create_serving_op():
    """
    部署模型服务
    :return: ServingOp
    """
    serving_op = components.load_component_from_file("./yaml/serving.yaml")
    return serving_op(
        name="ppocr-det",
        model_name="ppocr-det",  # 模型名称
        model_version="latest",  # 模型版本
        port=9292,               # Serving使用的端口
        # PaddleServing镜像
        image="registry.baidubce.com/paddleflow-public/serving:v0.6.2",
    ).set_display_name("model serving")
```

**6）编译 PP-OCR Pipeline 并提交任务**
通过 Python SDK 构建 PP-OCR 工作流，并将工作流提交到集群上运行。

```python
@dsl.pipeline(
    name="ppocr-detection-demo",
    description="An example for using ppocr train.",
)
def ppocr_detection_demo():
    # 创建 ppocr pipeline 各步骤所需的存储盘
    volume_op = create_volume_op()

    # 拉取远程数据（BOS/HDFS）到训练集群本地，并缓存
    dataset_op = create_dataset_op()
    dataset_op.execution_options.caching_strategy.max_cache_staleness = "P0D"

    # 采用Collective模型分布式训练ppocr模型，并提供模型训练可视化服务
    training_op = create_training_op(volume_op)
    training_op.execution_options.caching_strategy.max_cache_staleness = "P0D"
    training_op.after(dataset_op)

    # 将模型转换为 PaddleServing 可用的模型格式，并上传到模型中心
    modelhub_op = create_modelhub_op(volume_op)
    modelhub_op.execution_options.caching_strategy.max_cache_staleness = "P0D"
    modelhub_op.after(training_op)

    # 从模型中心下载模型，并启动 PaddleServing 服务
    serving_op = create_serving_op()
    serving_op.execution_options.caching_strategy.max_cache_staleness = "P0D"
    serving_op.after(modelhub_op)
 
if __name__ == "__main__":
    import kfp.compiler as compiler

    pipeline_file = "ppocr_detection_demo.yaml"
    compiler.Compiler().compile(ppocr_detection_demo, pipeline_file)
    client = kfp.Client(host="http://www.my-pipeline-ui.com:80")
    run = client.create_run_from_pipeline_package(
        pipeline_file,
        arguments={},
        run_name="paddle ocr detection demo",
        service_account="pipeline-runner"
    )
```

**7）查看 PP-OCR 工作流执行情况**
将任务提交后，可以在 Pipeline UI 中查看 PP-OCR 模型工作流的执行情况。
![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-79c84aa29e425f8eaeb17e6eb7f7cd1b619a16c5)

**8）查看模型训练可视化指标**

使用 Training Components 创建飞桨模型训练作业时，可以通过 use_visualdl 参数来指定是否开启可视化服务，通过以下命令可以将 VisualDL的 Service 映射到宿主机上,

```bash
$ kubectl port-forward svc/ml-pipeline-ui -n kubeflow 8765:80
```

然后访问 http://localhost:8765 就可以进入 VisualDL 的UI界面了。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-4e44ef0581edccd40ac9db679d32b33ddd188e58)

**9）模型存储中心与版本管理**

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-6f785ac0b0f591b1ffc02ca892ee27b0ee6d965b)



### 2. 基于Paddle Serving部署边缘推理服务

Paddle Serving作为飞桨（PaddlePaddle）开源的服务化部署服务化方案，提供了C++ Serving和Python Pipeline两套框架，旨在帮助深度学习开发者和企业提供高性能、灵活易用的工业级在线推理服务，助力人工智能落地应用。在最新的Paddle Serving v0.7.0中，提供了丰富的模型示例，总计有42个，具体模型信息可查看Model_Zoo：https://github.com/PaddlePaddle/Serving/blob/v0.7.0/doc/Model_Zoo_CN.md。

百度智能边缘（Baidu Intelligent Edge，BIE）由云端管理平台和BAETYL 开源边缘计算框架两部分组成，实现将云计算能力拓展至用户现场，可以提供临时离线、低延时的计算服务，包括消息规则、函数计算、AI 推断。智能边缘配合百度智能云，形成“云管理，端计算”的端云一体解决方案。

通过Paddle Serving赋能BIE，可以实现产业级的边缘AI服务发布解决方案，达到如下的云边端能力：

- **管理边缘节点：**纳管多种类型的边缘节点，包括服务器、边缘计算盒子。如果边缘侧是一个多机集群，也支持通过BIE统一管理。
- **状态检查：**支持监控边缘节点运行状态、资源使用（CPU、内存、GPU、磁盘、网络流量等）。
- **下发Serving：**支持云端将Paddle Serving下发至边缘侧，作为边缘侧服务化推理 + Serving版本升级。
- **下发模型：**支持云端动态下发PaddlePaddle模型至边缘侧，模型版本升级。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-1d0dd46103242cc55dc2247b74aafb16590b93e2)

以下教程详细描述使用Paddle Serving和BIE实现云边端服务发布的能力。主要包括实验准备、模型准备、Paddle Serving镜像准备、模型应用创建、模型应用部署、测试验证、测试效果展示。

**1）实验设备**

一台x86架构的ubuntu 18.04虚拟机，不依赖GPU

**2）模型文件准备**

1. 在宿主机上下载Paddle Serving代码

   ```bash
   git clone https://github.com/PaddlePaddle/Serving.git
   ```

2. 下载模型，参考文档:

https://github.com/PaddlePaddle/Serving/tree/v0.7.0/examples/Pipeline/PaddleDetection/yolov3

```bash
# 进入到yolov3实例模型目录
cd Serving/examples/Pipeline/PaddleDetection/yolov3/
# 下载模型
wget --no-check-certificate https://paddle-serving.bj.bcebos.com/pddet_demo/2.0/yolov3_darknet53_270e_coco.tar
# 解压模型
tar xf yolov3_darknet53_270e_coco.tar
# 解压以后删除模型压缩包
rm -r yolov3_darknet53_270e_coco.tar
```

3. 制作模型压缩包

```bash
cd Serving/examples/Pipeline/PaddleDetection/yolov3/
压缩当前目录下的文件
zip -r paddle_serving_yolov3_darknet53_270e_coco.zip ./*
# 查看md5
md5sum paddle_serving_yolov3_darknet53_270e_coco.zip 
7a2ca27f2f444c6ac169d19922ff89ab  paddle_serving_yolov3_darknet53_270e_coco.zip
```

4. 将模型上传到bos

**3）Paddle Serving镜像准备**

1. 下载Paddle Serving开发镜像

```bash
docker pull registry.baidubce.com/paddlepaddle/serving:0.7.0-devel
```

2. 运行Paddle Serving开发镜像

```bash
docker run --rm -dit --name pipeline_serving_demo registry.baidubce.com/paddlepaddle/serving:0.7.0-devel bash
```

3. Paddle Serving开发镜像当中安装依赖程序

```bash
# 进入容器
docker exec -it pipeline_serving_demo bash
# 下载代码
git clone https://github.com/PaddlePaddle/Serving.git
# 进入Paddle Serving代码目录
cd Serving
# 安装依赖
pip3 install -r python/requirements.txt -i https://pypi.tuna.tsinghua.edu.cn/simple

# CPU环境安装内容如下

# 安装Paddle Serving
pip3 install paddle-serving-client==0.7.0 -i https://pypi.tuna.tsinghua.edu.cn/simple
pip3 install paddle-serving-server==0.7.0 -i https://pypi.tuna.tsinghua.edu.cn/simple 
pip3 install paddle-serving-app==0.7.0 -i https://pypi.tuna.tsinghua.edu.cn/simple 

# 安装Paddle相关Python库
pip3 install paddlepaddle==2.2.0
```

加上 *-i https://pypi.tuna.tsinghua.edu.cn/simple* 表示使用国内源，提升下载速度，非必须，可以不加。

4. 提交镜像，固化上面的安装内容

```bash
docker commit pipeline_serving_demo paddle_serving:0.7.0-cpu-py36
```

这里将我制作的镜像推送到了百度公有云CCR，可以直接下载使用

```bash
docker pull registry.baidubce.com/pp/paddle-serving:0.7.0-cpu-py36
```

**4）模型应用创建**

4.1 创建模型文件配置项

① 创建配置项*paddle-yolov3-model*

② 点击引入文件

- 类型：HTTP
- URL：https://bie-document.gz.bcebos.com/paddlepaddle/paddle_serving_yolov3_darknet53_270e_coco.zip
- 文件名称：paddle_serving_yolov3_darknet53_270e_coco.zip
- 是否解压：是

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-e33ea2a7739701fcdf3db57b361b4a60e57b5ea9)

4.2 创建启动脚本配置项

① 创建配置项*paddle-yolov3-run-script*

② 添加配置数据如下

- 变量名：run.sh
- 变量值：如下述代码

```bash
#! /usr/bin/env bash
cd /home/work/yolov3
python3 web_service.py
```

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-0ec1fe4a1d649fbb017fbe69ee5fc251bf9cb856)

4.3 创建paddle-serving应用并挂载

① 创建应用*paddle-serving*

② 配置服务

- 基础信息：
- 名称：paddle-serving

- 镜像：paddle-serving:0.7.0-cpu-py36

- 卷配置：
- /home/work/script：运行脚本位置，与*启动参数*一致

- /home/work/yolov3：模型位置，与*运行脚本*一致

- 启动参数
- */bin/bash*

- */home/work/script/run.sh*，与前面的*卷配置*一致

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-70aacdcd5e01a88bdbf281c21cfb969ed9090291)

**5）模型应用部署**

5.1 进入到 paddle-serving

5.2 定位到目标节点，点击**单节点匹配**，选择目标节点 **paddle-serving-test**。等待几分钟，部署状态将变为**已部署**。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-7b571af0b4dd84a11c244f2e8d82093615e5908d)

5.3 进入边缘节点，可以查看服务在边缘测的运行状态，如下图所示：

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-997c90e4c647f1af422e59348dc7dd2ec714147a)

**6）测试验证**

6.1 使用 paddle-serving-client 验证

① ssh 登录边缘节点

② 查看边缘节点 BIE 应用状态

```bash
kubectl get pod -n baetyl-edge
NAME                             READY   STATUS    RESTARTS   AGE
paddle-serving-dd6d8986c-d89k7   1/1     Running   0          3m7s
```

③ 进入边缘容器 

```bash
kubectl exec -it paddle-serving-dd6d8986c-d89k7 -n baetyl-edge /bin/bash
# 进去以后，工作目录为/home
λ paddle-serving-dd6d8986c-d89k7 /home 
# 查看/home/work目录，检查云端模型是否下发成功
λ paddle-serving-dd6d8986c-d89k7 /home/work cd /home/work/
λ paddle-serving-dd6d8986c-d89k7 /home/work ls
script/  yolov3/
```

④ 执行测试命令

```bash
# 进入yolov3目录
λ paddle-serving-dd6d8986c-d89k7 /home/work/yolov3 cd /home/work/yolov3/
# 查看内容
λ paddle-serving-dd6d8986c-d89k7 /home/work/yolov3 ls -l
total 221M
--rw-r-- 1 root root 136K Dec 17 09:21 000000570688.jpg
-rw-rw-r-- 1 root root  509 Dec 17 09:21 benchmark_config.yaml
-rw-rw-r-- 1 root root 4.2K Dec 17 09:21 benchmark.py
-rw-rw-r-- 1 root root 2.1K Dec 17 09:21 benchmark.sh
-rw-rw-r-- 1 root root 1.5K Dec 17 09:21 config.yml
-rw-rw-r-- 1 root root  621 Dec 17 09:21 label_list.txt
-rwxr-xr-x 1 root root 220M Dec 17 09:21 paddle_serving_yolov3_darknet53_270e_coco.zip
-rw-rw-r-- 1 root root 1.2K Dec 17 09:21 pipeline_http_client.py
drwxr-xr-x 2 root root 4.0K Dec 17 09:26 PipelineServingLogs/
-rw-r--r-- 1 root root   89 Dec 17 09:26 ProcessInfo.json
-rw-rw-r-- 1 root root  368 Dec 17 09:21 README_CN.md
-rw-rw-r-- 1 root root  374 Dec 17 09:21 README.md
drwxr-xr-x 2 root root 4.0K Dec 17 09:21 serving_client/
drwxr-xr-x 2 root root 4.0K Dec 17 09:21 serving_server/
-rw-rw-r-- 1 root root 2.8K Dec 17 09:21 web_service.py
λ paddle-serving-dd6d8986c-d89k7 /home/work/yolov3 python3 
# 执行客户端测试脚本
pipeline_http_client.py
```

返回结果如下：

```bash
{
    'err_no': 0,
    'err_msg': '',
    'key': ['bbox_result'],
    'value': ["[{'category_id': 0, 'bbox': [215.16099548339844, 438.1199951171875, 43.29920959472656, 186.94189453125], 'score': 0.9860591292381287}, ..."],
    'tensors': []
}
```

6.2 使用 postman 验证

① 我们知道上述 yolov3 模型服务的容器内端口是18082，当前我们需要在单独的一台测试机器上使用 postman 去调用 yolov3 服务接口，那么就需要将容器内的18082端口映射到宿主机上，我们在云端 BIE 控制台配置 paddle-serving 这个服务，添加端口映射，如下图所示：

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-a6e8d3eb9c97e607440d2ac4118a2362d9b7113c)

② 下载测试图片 dog.jpeg

下载地址：https://bce.bdstatic.com/doc/bce-doc/BIE/dog_831f56a.jpeg

③ 执行一下命令，将这张图片的 base64 编码输出到 dog.base64 文件当中

```bash
base64 -i dog.jpeg -o dog.base64
```

④ 组装 postman 输出参数，如下所示：

```bash
{
    "key":[
        "image"
    ],
    "value":[
         "..."
    ]
}
```

⑤ postman 调用 url 为 ：

http://[ip]:18082/yolov3/prediction，如下图所示：

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-da4724c5c4acc97490ecc6057851735a23d61363)

⑥ 查看 postman 返回结果如下，如下图所示：

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-4b8c1627b1372dcf3e9df07a20538bb8ffa227aa)

```bash
{
    "err_no":0,
    "err_msg":"",
    "key":[
        "bbox_result"
    ],
    "value":[
        "[{'category_id': 16, 'bbox': [138.06399536132812, 54.169952392578125, 464.1486511230469, 552.6064147949219], 'score': 0.9826956987380981}, {'category_id': 57, 'bbox': [142.67298889160156, 29.47320556640625, 401.58738708496094, 564.1134033203125], 'score': 0.02150958590209484}]"
    ],
    "tensors":[

    ]
}
```

⑦ 我们看到 category_id 为16，查看该模型的 label_list.txt，我们看到16刚好对应 dog

模型链接地址：https://github.com/PaddlePaddle/Serving/blob/v0.7.0/examples/Pipeline/PaddleDetection/yolov3/label_list.txt

说明：label_list 当中的 id，从0开始

测试效果

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-19e0cfdedf58d8e6ee8b6759190262466b4edb3d)

**基于 Paddle Serving+BIE 的迁移扩展能力：**

通过 BIE 部署 Paddle Serving 的主要思路

- 文中所用到的 paddle-serving 应用和运行脚本是通用的，若想运行不同模型，只需替换下发模型文件即可。
- GPU 镜像构建逻辑与 CPU 镜像一致，可参考官网文档，本文统一使用 CPU 镜像。

Paddle Serving 不断拓展异构硬件和边缘端部署能力，将与百度智能云天工智能边缘框架 BIE 深度合作。智能边缘框架 BIE 凭借其核心技术优势已在各领域落地部署中提供解决方案，未来，BIE 将携手更多开发者共创智能边缘发展新机遇，推动边缘计算平台稳步向前，助力众多行业实现智慧化转型。

Paddle Serving 即将发布 v0.8.0版本将提供更多硬件上 AI 服务化部署，如华为昇腾310、昇腾910、海光 DCU、以及英伟达 Jetson。对服务化部署感兴趣的小伙伴欢迎来 Paddle Serving 的 github 了解：https://github.com/PaddlePaddle/Serving
