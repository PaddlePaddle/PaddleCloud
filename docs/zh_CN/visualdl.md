# VisualDL可视化组件使用手册

**VisualDL项目链接：** https://github.com/PaddlePaddle/VisualDL

**VisualDL Pipeline Components：**https://github.com/PaddleFlow/paddlex-backend/blob/main/components/visualdl.yaml

VisualDL是飞桨可视化分析工具，以丰富的图表呈现训练参数变化趋势、模型结构、数据样本、高维数据分布等。可帮助用户更清晰直观地理解深度学习模型训练过程及模型结构，进而实现高效的模型优化。

VisualDL提供丰富的可视化功能，支持标量、图结构、数据样本（图像、语音、文本）、超参数可视化、直方图、PR曲线、ROC曲线及高维数据降维呈现等诸多功能，同时VisualDL提供可视化结果保存服务，通过VDL.service生成链接，保存并分享可视化结果。具体功能使用方式，请参见 [**VisualDL使用指南**](https://github.com/PaddlePaddle/VisualDL/blob/develop/docs/components/README_CN.md)。如欲体验最新特性，欢迎试用我们的[**在线演示系统**](https://www.paddlepaddle.org.cn/paddle/visualdl/demo)。

## 核心功能

- API设计简洁易懂，使用简单。模型结构一键实现可视化。
- 功能覆盖标量、数据样本、图结构、直方图、PR曲线及数据降维可视化。
- 全面支持Paddle、ONNX、Caffe等市面主流模型结构可视化，广泛支持各类用户进行可视化分析。
- 与飞桨服务平台及工具组件全面打通，为您在飞桨生态系统中提供最佳使用体验。

## 使用方式

云上飞桨基于 Kubeflow Pipeline 将 VirsualDL 以 Components 的形式方便用户部署可视化组件，如果您使用的是 Training Components 则只需要再构建 Training Op 时将参数 use_visualdl 设置为 True 即可。开启训练日志可视化服务，需要在模型训练代码中进行日志打点，下面将详细讲述如何在模型代码中使用 VisualDL Python SDK 来打印特定格式的指标日志。

### 1. 安装 VisualDL Python SDK

使用 pip 安装

```bash
python -m pip install visualdl -i https://mirror.baidu.com/pypi/simple
```

### 2. 记录日志

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

#### 示例

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

### 3. 启动面板（本地体验可选）

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

## 在 Pipeline 中使用 VisualDL

如您需在 Kubernetes 集群上部署 VisualDL 服务，可以使用原生的 Deployment 来进行部署。如需集成到飞桨工作流中，云上飞桨基于 kubeflow Pipeline 组件为您封装好了 VisualDL Components 并在 Training Components 也有集成，下面将详细介绍如何使用 VisualDL Component 或在 Training Component 中开启可视化功能。

VisualDL Component Yaml 文件链接：https://github.com/PaddleFlow/paddlex-backend/blob/main/components/visualdl.yaml

Training Component Yaml 文件链接：https://github.com/PaddleFlow/paddlex-backend/blob/main/operator/yaml/training.yaml

### 示例1  用 VisualDL Component 部署可视化服务

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

### 示例2  在 Training Component 中开启 VisualDL 可视化服务

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

