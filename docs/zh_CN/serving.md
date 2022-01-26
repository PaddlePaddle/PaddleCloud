# 推理服务组件使用手册

**模型服务部署项目链接：** https://github.com/PaddleFlow/ElasticServing

ElasticServing 通过提供自定义资源 PaddleService，支持用户在 Kubernetes 集群上使用 TensorFlow、onnx、PaddlePaddle 等主流框架部署模型服务。 ElasticServing 构建在 Knative Serving 之上，其提供了自动扩缩容、容错、健康检查等功能，并且支持在异构硬件上部署服务，如 Nvidia GPU 或 昆仑芯片。 ElasticServing 采用的是 serverless 架构，当没有预估请求时，服务规模可以缩容到零，以节约集群资源，同时它还支持并蓝绿发版等功能。

## 架构概览

下图是 ElasticServing 的架构流程图，其使用了 Knative 提供的自动扩缩容机制，并使用 Istio 作为推理服务的统一网关。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-cdc4d860be727666b9bd73c08e04a955bd68883b)

## 安装 ElasticServing 组件

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

## 示例

### 示例1 部署 ResNet50 图片分类服务

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

### 示例2 中文分词模型服务

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