

目前基于飞桨框架和 Kubernetes 调度器协同设计的 GPU 利用率优化项目还在开发中，飞桨框架部分的PR: https://github.com/PaddleFlow/Paddle/pull/2/files

## 背景

- GPU显存与算力（SM）利用率低下，显存使用率超过一半的GPU占比不足20%，算力利用率超过80%的GPU占比不足十分之一。

- 使用 gang-schedule 的调度策略，模型训练作业申请的GPU资源越多，算力资源空置等待的时间越长。
- 模型训练过程中对GPU资源的需求是动态变化的，一般会申请足够的资源保证任务正常运行，然而这种固定资源分配的方式也会带来一定程度的资源浪费。

## 框架与调度器的协同设计

传统的GPU共享方案：资源隔离（劫持CUDA API调用）、时间片模型、MPS模式。

缺点：作业间相互影响，难以保证高优作业的SLA，开发难度大，落地生产难。

![image-20220110154922617](/Users/chenenquan/Library/Application Support/typora-user-images/image-20220110154922617.png)



**框架与调度器的协同设计方案**

- 将任务划分为 Resource-guarantee job 和 Opportunistic job，保证高优作业执行效率的同时，使用机会作业提升集群利用率。
- 协同设计深度学习框架与 Kubernetes 调度器，实现动态调整作业的显存与算力上限，并根据统计的作业与设备信息，对不同作业类型实施不同的调度策略。

### 飞桨框架—显存动态调整

1. 作业启动后，框架根据模型申请的显存资源，设置一个合适的显存上界。
2. 当有个 mini-batch 显存需求突增且设备显存不足时，则将待创建的 Tensor申明到内存中（Pinned Memory），保证作业能够正常运行。
3. 将同设备机会作业的显存上界下调，出让显存给高优先级作业，保证高优作业的执行效率。
4. 机会作业出让显存后，高优作业上调显存上界，下个 mini-batch 的 Tensor 又可以申明在显存中。
5. 框架持续对作业显存使用情况打点，配合 Local Coordinator 动态调整显存上界。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-055d31e5c57d86553720a3af9eb2a242f5888fe1)

**飞桨框架显存动态调整实现代码类图：**

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-1536f7d1d5bbead5f6c848bbcfdd97ec81569971)

### 飞桨框架—算力动态调整

模型训练作业特点：

- 训练过程中包含特征提取、数据增强等不使用GPU算力的流程，独占模式会浪费GPU算力（Exclusive mode）。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-ab3bbad60ec9f2d39e59ba4150a855027a2b87b5)

- 作业打包模式会带来 GPU kernel 排队延迟和 PCIe带宽抢占的问题，影响高优作业执行效率。

  ![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-a31dc14cddbc768889c9a6e780925b1a52c1e1d4)

算力动态调整方案：

接管算子（Op）执行流程，通过给 Op 加上延迟执行时间实现 Op 级别的分时复用。

1. 实现 GpuOpManager 接管 Op 执行流程，待执行 Op 依序进入执行队列，并简单分配延迟执行时间。
2. GpuOpManager 不断监控算子执行时间和 GPU 算力利用率并打点，配合 Local Coordinator 根据一定策略动态调整后续 Op 延迟执行时间。如高优作业 Op 的算力被抢占时，则延长机会作业 Op 的等待执行时间。

### Kubernets 调度器—整体框架

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

