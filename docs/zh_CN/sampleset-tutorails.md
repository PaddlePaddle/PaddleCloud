# 样本缓存组件（SampleSet Operator）快速上手

本文档主要讲述了如何快速上手样本缓存组件（SampleSet Operator），并通过实际的示例演示了缓存组件的基础功能。

## 前提条件

* Kubernetes >= 1.8
* kubectl
* Helm

## 安装缓存组件及其依赖

样本缓存组件依赖于 [JuiceFS CSI Driver](https://github.com/juicedata/juicefs-csi-driver)，且 JuiceFS 需要使用对象存储（如 Minio）来存储数据、使用 Redis 来存储元数据。
为了方便用户快速上手体验，本项目提供的 paddlecloud helm chart 中包含了 JuiceFS CSI Driver、Redis、Minio 等依赖，您可以通过指定 `--set tags.all-dep=true` 来安装所有依赖。

添加并更新 helm 的 charts repositories

```bash
$ helm repo add paddlecloud https://paddleflow-public.hkg.bcebos.com/charts
$ helm repo update
```

使用 helm 一键安装

```bash
# install
$ helm install test paddlecloud/paddlecloud --set tags.all-dep=true -n paddlecloud --create-namespace
```

注意：如果 JuiceFS CSI Driver 没有正常启动，即 Pod 不是 Running 状态，并出现类似这样的错误：driver name csi.juicefs.com not found in the list of registered CSI drivers，
这是由于 CSI 驱动没有注册到 kubelet 的指定路径，你可以通过指定 `juicefs.kubeletDir` 来正确安装 JuiceFS CSI Driver。

在集群中的 worker 节点执行以下命令来获取 kubelet 的根目录

```
ps -ef | grep kubelet | grep root-dir
```

在上述命令打印的内容中，找到 --root-dir 参数后面的值，这既是 kubelet 的根目录。

然后在安装时将 `juicefs.kubeletDir` 指定为上步找到的 kubelet 的根目录。

```bash
# install
$ helm install test paddlecloud/paddlecloud --set tags.all-dep=true --set juicefs.kubeletDir=<kubelet root dir> -n paddlecloud --create-namespace
```

更详细的安装文档请参考[安装PaddleCloud](./installation.md)。

## 缓存组件使用示例

为了方便用户快速上手体验样本缓存组件，我们将本示例的数据集存放在百度对象存储(BOS)公开课可访问的 bucket 中，无需秘钥即可访问数据。
同时，在上述的 PaddleCloud 安装步骤中会默认创建 `none` 和 `data-center` 两个 Secret，分别用来访问 BOS 数据源和 Minio。
如果您需要使用自己的数据集，并且数据源需要通过 `access-key` 和 `secret-key` 来访问，您可以参考下面的步骤来创建 Secret。

### 1. 创建 Secret (可选)

> 本步骤是可选的，如果您直接使用本文档提供的数据，可以直接从创建 SampleSet 开始操作。

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
  namespace: paddlecloud
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

### 2. 创建 SampleSet

编写 imagenet.yaml 如下，数据源来自 bos (百度对象存储)公开的 bucket。
```yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleSet
metadata:
  name: imagenet
  namespace: paddlecloud
spec:
  # 分区数，一个Kubernetes节点表示一个分区
  partitions: 1
  source:
    uri: bos://paddleflow-public.hkg.bcebos.com/imagenet/demo
    secretRef:  # 用于访问数据源的 Secret
      name: none
  secretRef:  # 用于访问对象存储 Minio 的 Secret
    name: data-center
```

创建 SampleSet，等待数据完成同步，并查看 SampleSet 的状态。数据同步操作可能比较耗时，请耐心等待。
```bash
$ kubectl create -f imagenet.yaml
sampleset.batch.paddlepaddle.org/imagenet created

$ kubectl get sampleset -n paddlecloud
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   58 MiB       200 B         12 GiB        1/1       Ready   76s

$ kubectl get pods -n paddlecloud
NAME                                         READY   STATUS    RESTARTS   AGE
imagenet-runtime-0                           1/1     Running   0          30s
paddle-controller-manager-776b84bfb4-qs67f   1/1     Running   0          11h
paddle-samplejob-manager-685449d6d7-cmqfb    1/1     Running   0          11h
paddle-sampleset-manager-69bc7fb85d-4rjcg    1/1     Running   0          11h
```

等创建的 SampleSet 其 PHASE 状态为 Ready 时，表示该数据集可以使用了。

### 3. 体验 SampleJob（可选）

缓存组件提供了 SampleJob 用来做样本数据集的管理，目前支持了4种 Job 类型，分别是 sync/warmup/clear/rmr。

**注意**：由于 SampleSet 的数据缓存信息不是实时更新的，SampleJob 执行完成后缓存信息会在 30s 内完成更新。

删除缓存引擎中的数据，rmr job 里的 rmrOptions 是必填的，paths 里的参数是数据的相对路径。

```bash
$ cat rmr-job.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleJob
metadata:
  name: imagenet-rmr
  namespace: paddlecloud
spec:
  type: rmr
  sampleSetRef:
    name: imagenet
  rmrOptions:
    paths:
      - n01514859

$ kubectl create -f rmr-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-rmr created

$ kubectl get samplejob imagenet-rmr -n paddlecloud
NAME           PHASE
imagenet-rmr   Succeeded

$ kubectl get sampleset imagenet -n paddlecloud
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
  namespace: paddlecloud
spec:
  type: sync
  sampleSetRef:
    name: imagenet
  # sync job 需要填写 secret 信息
  secretRef:
    name: none
  syncOptions:
    source: bos://paddleflow-public.hkg.bcebos.com/imagenet

$ kubectl create -f sync-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-sync created

$ kubectl get samplejob imagenet-sync -n paddlecloud
NAME            PHASE
imagenet-sync   Succeeded

$ kubectl get sampleset imagenet -n paddlecloud
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
  namespace: paddlecloud
spec:
  type: clear
  sampleSetRef:
    name: imagenet
    
$ kubectl create -f clear-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-clear created

$ kubectl get samplejob imagenet-clear -n paddlecloud
NAME             PHASE
imagenet-clear   Succeeded

$ kubectl get sampleset imagenet -n paddlecloud
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
  namespace: paddlecloud
spec:
  type: warmup
  sampleSetRef:
    name: imagenet

$ kubectl create -f warmup-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-warmup created

$ kubectl get samplejob imagenet-warmup -n paddlecloud
NAME              PHASE
imagenet-warmup   Succeeded

$ kubectl get sampleset imagenet -n paddlecloud
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   4.2 GiB      4.2 GiB       7.3 GiB       1/1       Ready   90m
```

### 4. 创建 PaddleJob 

以下示例使用 nginx 镜像来简单示范下如何在 PaddleJob 中声明使用 SampleSet 样本数据集。 如果您的集群中有 GPU 硬件资源，并且想要测试缓存组件给模型训练带来的提升效果，请参考文档：[性能测试](./ext-benchmark.md)

编写 ps-demo.yaml 文件如下：

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: ps-demo
  namespace: paddlecloud
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
$ kubectl get paddlejob -n paddlecloud
NAME      STATUS    MODE   AGE
ps-demo   Running   PS     112s
```

查看挂载在 PaddleJob worker pod 的样本数据

```bash
# 进入 PaddleJob worker pod
$ kubectl exec -it ps-demo-worker-0 -n paddlecloud -- /bin/bash

$ ls /mnt/imagenet
demo  train  train_list.txt
```