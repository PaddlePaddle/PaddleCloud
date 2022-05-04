<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Paddle Operator 样本缓存组件快速上手](#paddle-operator-%E6%A0%B7%E6%9C%AC%E7%BC%93%E5%AD%98%E7%BB%84%E4%BB%B6%E5%BF%AB%E9%80%9F%E4%B8%8A%E6%89%8B)
  - [前提条件](#%E5%89%8D%E6%8F%90%E6%9D%A1%E4%BB%B6)
  - [安装缓存组件](#%E5%AE%89%E8%A3%85%E7%BC%93%E5%AD%98%E7%BB%84%E4%BB%B6)
    - [1. 安装自定义资源](#1-%E5%AE%89%E8%A3%85%E8%87%AA%E5%AE%9A%E4%B9%89%E8%B5%84%E6%BA%90)
    - [2. 部署 Operator](#2-%E9%83%A8%E7%BD%B2-operator)
    - [3. 部署样本缓存组件的 Controller](#3-%E9%83%A8%E7%BD%B2%E6%A0%B7%E6%9C%AC%E7%BC%93%E5%AD%98%E7%BB%84%E4%BB%B6%E7%9A%84-controller)
    - [4. 安装 CSI 存储插件](#4-%E5%AE%89%E8%A3%85-csi-%E5%AD%98%E5%82%A8%E6%8F%92%E4%BB%B6)
  - [缓存组件使用示例](#%E7%BC%93%E5%AD%98%E7%BB%84%E4%BB%B6%E4%BD%BF%E7%94%A8%E7%A4%BA%E4%BE%8B)
    - [1. 准备 Redis 数据库](#1-%E5%87%86%E5%A4%87-redis-%E6%95%B0%E6%8D%AE%E5%BA%93)
    - [2. 创建 Secret](#2-%E5%88%9B%E5%BB%BA-secret)
    - [3. 创建 SampleSet](#3-%E5%88%9B%E5%BB%BA-sampleset)
    - [4. 体验 SampleJob（可选）](#4-%E4%BD%93%E9%AA%8C-samplejob%E5%8F%AF%E9%80%89)
    - [5. 创建 PaddleJob](#5-%E5%88%9B%E5%BB%BA-paddlejob)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Paddle Operator 样本缓存组件快速上手

本文档主要讲述了如何安装部署 Paddle Operator 样本缓存组件，并通过实际的示例演示了缓存组件的基础功能。

## 前提条件

* Kubernetes >= 1.8
* kubectl

## 安装缓存组件

### 1. 安装自定义资源

与 [README](../../README-zh_CN.md) 中描述的步骤一样，如果您使用的是0.4以前的版本，可以再次运行该命令来更新 CRD。
创建 PaddleJob / SampleSet / SampleJob 自定义资源,
```shell
kubectl apply -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/main/deploy/v1/crd.yaml
```

创建成功后，可以通过以下命令来查看创建的自定义资源
```shell
$ kubectl get crd | grep batch.paddlepaddle.org
NAME                                    CREATED AT
paddlejobs.batch.paddlepaddle.org       2021-08-23T08:45:17Z
samplejobs.batch.paddlepaddle.org       2021-08-23T08:45:18Z
samplesets.batch.paddlepaddle.org       2021-08-23T08:45:18Z
```

### 2. 部署 Operator

与 [README](../../README-zh_CN.md) 中描述的步骤一样，如果您使用的是0.4以前的版本，可以再次运行该命令来更新。
```shell
kubectl apply -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/main/deploy/v1/operator.yaml
```

### 3. 部署样本缓存组件的 Controller

以下命令中的 YAML 文件包括了自定义资源 SampleSet 和 SampleJob 的 Controller
```shell
kubectl apply -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/main/deploy/extensions/controllers.yaml
```

通过以下命令可以查看部署好的 Controller
```shell
$ kubectl -n paddle-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
paddle-controller-manager-776b84bfb4-5hd4s   1/1     Running   0          60s
paddle-samplejob-manager-69b4944fb5-jqqrx    1/1     Running   0          60s
paddle-sampleset-manager-5cd689db4d-j56rg    1/1     Running   0          60s
```

### 4. 安装 CSI 存储插件

目前 Paddle Operator 样本缓存组件仅支持 [JuiceFS](https://github.com/juicedata/juicefs/blob/main/README_CN.md) 作为底层的样本缓存引擎，样本访问加速和缓存相关功能主要由缓存引擎来驱动。

部署 JuiceFS CSI Driver
```shell
kubectl apply -f https://raw.githubusercontent.com/juicedata/juicefs-csi-driver/master/deploy/k8s.yaml
```

部署好 CSI 驱动后，您可以通过以下命令查看状态，Kubernetes 集群中的每一个 worker 节点应该都会有一个 **juicefs-csi-node** Pod。
```shell
$ kubectl -n kube-system get pods -l app.kubernetes.io/name=juicefs-csi-driver
NAME                                          READY   STATUS    RESTARTS   AGE
juicefs-csi-controller-0                      3/3     Running   0          13d
juicefs-csi-node-87f29                        3/3     Running   0          13d
juicefs-csi-node-8h2z5                        3/3     Running   0          13d
```

**注意**：如果 Kubernetes 无法发现 CSI 驱动程序，并出现类似这样的错误：**driver name csi.juicefs.com not found in the list of registered CSI drivers**，这是由于 CSI 驱动没有注册到 kubelet 的指定路径，您可以通过下面的步骤进行修复。

在集群中的 worker 节点执行以下命令来获取 kubelet 的根目录
```shell
ps -ef | grep kubelet | grep root-dir
```

在上述命令打印的内容中，找到 `--root-dir` 参数后面的值，这既是 kubelet 的根目录。然后将以下命令中的 `{{KUBELET_DIR}}` 替换为 kubelet 的根目录并执行该命令。
```shell
curl -sSL https://raw.githubusercontent.com/juicedata/juicefs-csi-driver/master/deploy/k8s.yaml | sed 's@/var/lib/kubelet@{{KUBELET_DIR}}@g' | kubectl apply -f -
```

更多详情信息可参考 [JuiceFS CSI Driver](https://github.com/juicedata/juicefs-csi-driver)

## 缓存组件使用示例

由于 JuiceFS 缓存引擎依靠 Redis 来存储文件的元数据，并支持多种对象存储作为数据存储后端，为了方便起见，本示例使用 Redis 同时作为元数据引擎和数据存储后端。

### 1. 准备 Redis 数据库

您可以很容易的在云计算平台购买到各种配置的云 Redis 数据库，本示例使用 Docker 在 Kubernetes 集群的 worker 节点上运行一个 Redis 数据库实例。
```shell
docker run -d --name redis \
	-v redis-data:/data \
	-p 6379:6379 \
	--restart unless-stopped \
	redis redis-server --appendonly yes
```

**注意**：以上命令将本地目录 `redis-data` 挂载到 Docker 容器的 `/data` 数据卷中，您可以按需求挂载不同的文件目录。

### 2. 创建 Secret

准备好 JuiceFS 用于格式化文件系统需要的字段，需要的字段如下：

| 字段  | 含义 | 说明  |
|:-----|:-----|:-----|
| name    | JuiceFS 文件系统名称 | 可以是任意的字符串，用来给 JuiceFS 文件系统命名 |
| storage | 对象存储类型         | 例如：bos [JuiceFS 支持的对象存储和设置指南](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/how_to_setup_object_storage.md) |
| bucket  | 对象存储的 URL       | 存储数据的桶路径，例如： bos://imagenet.bj.bcebos.com/imagenet |
| metaurl | 元数据存储的 URL  | 如 redis://username:password@host:6379/0，[JuiceFS 支持的元数据存储引擎](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/databases_for_metadata.md) |
| access-key | 对象存储的 Access Key | 可选参数；如果对象存储使用的是 username 和 password，例如使用 Redis 作为存储后端，这里填用户名。 |
| secret-key | 对象存储的 Secret key | 可选参数；如果对象存储使用的是 username 和 password，例如使用 Redis 作为存储后端，这里填密码。  |

更多详情信息可以参考：[JuiceFS 快速上手](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/quick_start_guide.md) ；[JuiceFS命令参考](https://github.com/juicedata/juicefs/blob/main/docs/zh_cn/command_reference.md)

本示例中各字段配置如下，因为没有给 Redis 配置密码，字段 `access-key` 和 `secret-key` 可以不设置。

| 字段  | 说明  |
|:-----|:-----|
| name    | imagenet | 
| storage | redis |
| bucket  | redis://192.168.7.227:6379/1 |
| metaurl | redis://192.168.7.227:6379/0  |

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

### 3. 创建 SampleSet

编写 imagenet.yaml 如下，数据源来自 bos (百度对象存储)公开的 bucket。
```yaml
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

### 4. 体验 SampleJob（可选）

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

### 5. 创建 PaddleJob 

以下示例使用 nginx 镜像来简单示范下如何在 PaddleJob 中声明使用 SampleSet 样本数据集。 如果您的集群中有 GPU 硬件资源，并且想要测试缓存组件给模型训练带来的提升效果，请参考文档：[性能测试](./ext-benchmark.md)

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