# Paddle Operator 样本缓存组件概览

由于云计算平台的易扩展性等特点，越来越多的机器学习任务跑在 Kubernetes 集群里。然而在 Kubernetes 的架构体系中，计算和存储是分离的，模型训练作业需要访问远程存储来获取训练样本，这就给数据密集型的机器学习作业带来较高的网络 IO 开销。受 [Fluid](https://github.com/fluid-cloudnative/fluid) 项目的启发，我们在 Paddle Operator 项目中添加了样本缓存组件，旨在通过将样本数据缓存在 Kubernetes 集群本地来加速云上飞桨分布式训练作业。

## 功能要点

1. 将远程样本数据集和数据操作抽象成 SampleSet 和 SampleJob 两个自定义资源对象，用户可以简单地通过操作 CRD 来管理样本数据，而不用关心底层存储接口和缓存引擎等细节。
2. 基于 [JuiceFS](https://github.com/juicedata/juicefs) 加速远程样本数据访问速度，特别是在海量小文件场景，经过和 [BOSFS](http://baidu.netnic.com.cn/doc/BOS/BOSCLI/8.5CBOS.20FS.html) 的对比测试发现，JuiceFS 给模型训练作业带来了近3倍的性能提升。
3. Paddle Operator 样本缓存组件会将样本数据预热并缓存在集群本地，并对 PaddleJob 实现了缓存亲和性调度，能够大幅缩短 PaddleJob 执行时间，一定程度也能提高 GPU 资源利用率。

## 整体架构

<div align="center">
  <img src="http://paddleflow-public.hkg.bcebos.com/Static/ext-arch.png" title="architecture" width="60%" height="60%" alt="">
</div>

上图是 Paddle Operator 的整体架构，其构建在 Kubernetes 上，包含如下三个主要部分：

一是自定义 API 资源（Custom Resource），Paddle Operator 定义了三个 CRD ，用户可编写和修改对应的 YAML 来管理训练作业和样本数据集。

- **PaddleJob**：是 Paddle 分布式训练作业的抽象，它将 Parameter Server 和 Collective 两种分布式深度学习架构模式统一到一个 CRD 中，用户通过该创建 PaddleJob 可以很方便地在 Kubernetes 集群运行分布式训练作业。
- **SampleSet**：是样本数据集的抽象，数据可以来自远程对象存储、HDFS 或 Ceph 等分布式文件系统，并且可以指定缓存数据的分区数、使用的缓存引擎、 多级缓存目录、缓存节点等配置。
- **SampleJob**：定义了些样本数据集的管理作业，包括数据同步、数据预热、清除缓存、淘汰历史旧数据等操作，支持用户设置各个数据操作命令的参数， 同时还指定以定时任务的方式运行数据管理作业。

二是自定义控制器（Controller Manager），控制器在 Kubernetes 的 Operator 框架中是用来监听 API 对象的变化（比如创建、修改、删除等），然后以此来决定实际要执行的具体工作。

- **PaddleJob Controller**：负责管理 PaddleJob 的生命周期，比如创建参数服务器和训练节点的 Pod，并维护工作节点的副本数等。
- **SampleSet Controller**：负责管理 SampleSet 的生命周期，其中包括创建 PV/PVC 等资源对象、创建缓存运行时服务、给缓存节点打标签等工作。
- **SampleJob Controller**：负责管理 SampleJob 的生命周期，通过请求缓存运行时服务的接口，触发缓存引擎异步执行数据管理操作，并获取执行结果。

三是缓存引擎（Cache Engine），缓存引擎由缓存运行时服务（Cache Runtime Server）和 JuiceFS 存储插件（JuiceFS CSI Driver）两部分组成，提供了样本数据存储、缓存、管理的功能。

- **Cache Runtime Server**：负责样本数据的管理工作，接收来自 SampleSet Controller 和 SampleJob Controller 的数据操作请求，调用 JuiceFS 的命令完成相关操作执行。
- **JuiceFS CSI Driver**：负责样本数据的存储与缓存工作，将样本数据缓存到集群本地并将数据挂载进 PaddleJob 的训练节点。

## 快速开始
查看文档 [Paddle Operator 样本缓存组件快速上手](./ext-get-start.md) 来体验下吧。

## 性能测试
关于性能测试相关的文档请参考：[性能测试](./ext-benchmark.md)

## 更多资料
想了解更多关于自定义资源的详细，请查看文档 [API docs](../en/api_doc.md)。
