# PP-OCR文字检测模型全链路案例

## 安装相关组件

### 环境条件

 - [Kubernetes](https://kubernetes.io/docs/setup/production-environment/tools/) >= 1.8 
 - [kubectl](https://kubernetes.io/docs/tasks/tools/)
 - [Helm](https://helm.sh/docs/intro/install/) >= 3.0 
 - [Redis](https://redis.io/) （可以买云服务，或者本地安装单机版）

### 1. Standalone模式安装（体验试用）

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

创建数据存储所需的 Secret，

### 2. 多租户模式安装 （生产环境）

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

### 创建 StorageClass 和 Secret

由于使用的 JuiceFS CSI 插件需要使用 Redis 作为元数据存储，您可以选着使用云服务商提供的 Redis 服务或者是自己安装。然后，将 Redis 服务的连接URI进行Base64编码，URI的格式为 `redis://[username]:[passward][@]IP:PORT/DATABASE`。如 `redis://default:123456@127.0.0.1:6379/0`，使用命令 `echo "redis://default:123456@127.0.0.1:6379/0" | base64 ` 获取 Base64 编码。然后将 Base64 编码后的字符串填入到 hack/prerequisites 目录下文件 data-center.yaml 和 task-center 的字段 metaurl 上。最后执行如下命令，创建 StorageClass 和 Secret。

```bash
$ kubectl apply -f ./prerequisites/data-center.yaml
$ kubectl apply -f ./prerequisites/task-center.yaml
```

### 创建 Notebook 用于 Pipeline 开发（可选）

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