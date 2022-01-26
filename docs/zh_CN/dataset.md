# 样本数据缓存组件使用手册

**样本数据缓存组件代码：** https://github.com/PaddleFlow/paddle-operator/tree/sampleset

在 Kubernetes 的架构体系中，计算与存储是分离的，这给数据密集型的深度学习作业带来较高的网络IO开销。为了解决该问题，我们基于 JuiceFS 在开源项目 Paddle Operator 中实现了样本缓存组件，大幅提升了云上飞桨分布式训练作业的执行效率。

## 背景介绍

由于云计算平台具有高可扩展性、高可靠性、廉价性等特点，越来越多的机器学习任务运行在Kubernetes集群上。因此我们开源了Paddle Operator项目，通过提供PaddleJob自定义资源，让云上用户可以很方便地在Kubernetes集群使用飞桨（PaddlePaddle）深度学习框架运行模型训练作业。

然而，在深度学习整个pipeline中，样本数据的准备工作也是非常重要的一环。目前云上深度学习模型训练的常规方案主要采用手动或脚本的方式准备数据，这种方案比较繁琐且会带来诸多问题。比如将HDFS里的数据复制到计算集群本地，然而数据会不断更新，需要定期的同步数据，这个过程的管理成本较高；或者将数据导入到远程对象存储，通过制作PV和PVC来访问样本数据，从而模型训练作业就需要访问远程存储来获取样本数据，这就带来较高的网络IO开销。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-6d25edebdced1033430a7990c0bdd731dbf0e6c1)



为了方便云上用户管理样本数据，加速云上飞桨框架分布式训练作业，我们与JuiceFS社区合作，联合推出了面向飞桨框架的样本缓存与管理方案，该方案期望达到如下目标：

 - **基于JuiceFS加速远程数据访问。** JuiceFS是一款面向云环境设计的高性能共享文件系统，其在数据组织管理和访问性能上进行了大量针对性的优化。基于JuiceFS实现样本数据缓存引擎，能够提供高效的文件访问性能。
 - **充分利用本地存储，缓存加速模型训练。** 能够充分利用计算集群本地存储，比如内存和磁盘，来缓存热点样本数据集，并配合缓存亲和性调度，在用户无感知的情况下，智能地将作业调度到有缓存的节点上。这样就不用反复访问远程存储，从而加速模型训练速度，一定程度上也能提升GPU资源的利用率。
 - **数据集及其管理操作的自定义资源抽象。** 将样本数据集及其管理操作抽象成Kubernetes的自定义资源，屏蔽数据操作的底层细节，减轻用户心智负担。用户可以很方便地通过操作自定义资源对象来管理数据，包括数据同步、数据预热、清理缓存、淘汰历史数据等，同时也支持定时任务。
 - **统一数据接口，支持多种存储后端。** 样本缓存组件要能够支持多种存储后端，并且能提供统一的POSIX协议接口，用户无需在模型开发和训练阶段使用不同的数据访问接口，降低模型开发成本。同时样本缓存组件也要能够支持从多个不同的存储源导入数据，适配用户现有的数据存储状态。

## 面临的挑战

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

## 难点突破与优化

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

## 性能测试

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

## 安装样本缓存组件

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

## 缓存组件使用示例

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