# PP-YOLOE 云端训练

**适用场景：基于Kubernetes的生产环境、多机分布式训练。**

上一部分的内容我们介绍了在开发和单机部署场景中，使用飞桨模型套件Docker大礼包来训练以及部署PP-YOLOE模型；本部分内容主要介绍如何在生产环境的Kubernetes集群中使用云上飞桨组件来训练PP-YOLOE模型。


## 飞桨K8S大礼包概览

<div align="center">
  <img src="./img/PaddleCloudArch.png" width="800"/>
</div>

- **简单易用的编程接口**。

  云上飞桨大礼包给用户提供了简单易用的编程接口，用户可以使用Python SDK构建模型工作流，也可以通过使用Kubernetes自定义资源单独使用各个云上飞桨功能组件。


- **规范的套件镜像版本管理。**

  云上飞桨大礼包基于Teckton Pipeline为飞桨模型套件提供镜像的持续集成构建功能，后续还会为用户提供模型套件镜像和云原生组件下载站，


- **提供各领域场景的工作流模板。**

  云上飞桨模型套件Operator为用户提供了领域常见的工作流模板。如何PaddleOCR/PaddleDetection等模型套件的场景案例，我们都有提供云上部署解决方案。


- **具有丰富的云上飞桨组件。**

  云上飞桨具有丰富的功能组件，包括 VisualDL可视化工具、样本数据缓存组件、模型训练组件、推理服务部署组件等，这些功能组件可以大幅加速用户开发和部署模型的效率。


- **针对飞桨框架的定制优化。**

  除了提升研发效率的功能组件，我们还针对飞桨框架进行了针对性优化，如样本缓存组件加速云上飞桨分布式训练作业、基于飞桨框架和调度器协同设计的GPU利用率优化。

|         安装         |        部署         |          MLOps      |       多套件编排      |
| :-----------------: | :-----------------: | :-----------------: | :-----------------: |
|一键安装<br>版本对齐与自动更新<br>|集群<br>多设备<br>多端<br>多云<br>全链条<br>|自动扩缩容<br>A/B测试<br>容错<br>弹性训练<br>|模型训练配置<br>从训练到部署<br>可视化界面|


### 数据与训练组件

- 模型训练组件旨在为 Kubernetes 上运行飞桨分布式训练任务提供简单易用的标准化接口，并为训练任务的管理提供定制化的完整支持。
- 样本缓存组件基于开源项目JuiceFS实现了样本缓存，旨在解决Kubernetes中计算与存储分离的结构带来的高网络IO开销问题，提升云上飞桨分布式训练作业的执行效率。

<div align="center">
  <img src="../../docs/images/sampleset-arch.jpeg" width="600"/>
</div>

## 使用云原生组件训练PP-YOLOE

**环境要求**：

- Kubernetes v1.21
- kubectl
- helm

如果您没有Kubernetes环境，可以使用MicroK8S来搭建一个本地Kubernetes环境，更多详情请参考我们开源的[PaddleCloud项目](https://github.com/PaddlePaddle/PaddleCloud)

### 安装云上飞桨组件

使用Helm一键安装所有组件和所有依赖

```bash
# 添加PaddleCloud Chart仓库
$ helm repo add paddlecloud https://paddleflow-public.hkg.bcebos.com/charts
$ helm repo update

# 安装云上飞桨组件
$ helm install pdc paddlecloud/paddlecloud --set tags.all-dep=true --namespace kubeflow --create-namespace

# 检查所有云上飞桨组件是否成功启动，命名空间下的所有Pod都为Runing状态则安装成功。
$ kubectl get pods -n kubeflow
```

更多安装参数请参考我们开源的[PaddleCloud项目](https://github.com/PaddlePaddle/PaddleCloud)

### coco数据集准备

使用数据缓存组件来准备数据集，编写SampleSet Yaml文件如下：

```yaml
# coco.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleSet
metadata:
  name: coco
  namespace: kubeflow
spec:
  partitions: 1
  source:
    uri: bos://paddleflow-public.hkg.bcebos.com/coco
    secretRef:
      name: none
  secretRef:
    name: data-center
```

```bash
# 创建coco数据集
$ kubectl apply -f coco.yaml
sampleset.batch.paddlepaddle.org/coco created

# 查看数据集的状态
$ kubectl get sampleset coco -n kubeflow
NAME   TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
coco   20 GiB       2.4 GiB       13 GiB        1/1       Ready   33h
```

### 训练PP-YOLOE模型

使用训练组件在Kubernetes集群上训练PP-YOLOE模型，编写PaddleJob Yaml文件如下：

```yaml
# ppyoloe.yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: ppyoloe
  namespace: kubeflow
spec:
  cleanPodPolicy: OnCompletion
  sampleSetRef:
    name: coco
    namespace: kubeflow
    mountPath: /root/.cache/paddle
  worker:
    replicas: 1
    template:
      spec:
        containers:
          - name: ppyoloe
            image: registry.baidubce.com/paddleflow-public/paddledetection:2.4
            command:
            - python
            args:
            - "tools/train.py"
            - "-c"
            - "configs/ppyoloe/ppyoloe_crn_l_300e_coco.yml"
            - "-o"
            - "use_gpu=false"
            - "epoch=10"
            - "worker_num=2"
            - "TrainReader.batch_size=2"
            - "log_iter=1"
            - "snapshot_epoch=1"
```

```bash
# 创建PaddleJob训练模型
$ kubectl apply -f ppyoloe.yaml
paddlejob.batch.paddlepaddle.org/ppyoloe created

# 查看PaddleJob状态
$ kubectl get pods -n kubeflow -l paddle-res-name=ppyoloe-worker-0
NAME               READY   STATUS    RESTARTS   AGE
ppyoloe-worker-0   1/1     Running   0          4s

# 查看训练日志
$ kubectl logs -f ppyoloe-worker-0 -n kubeflow
```

## 更多资源

欢迎关注云上飞桨项目PaddleCloud，我们为您提供了飞桨模型套件标准镜像以及全栈的云原生模型套件部署组件，如您有任何关于飞桨模型套件的部署问题，请联系我们。

- 如果你发现任何PaddleCloud存在的问题或者是建议, 欢迎通过[GitHub Issues](https://github.com/PaddlePaddle/PaddleCloud/issues)给我们提issues。
