# 缓存组件性能测试

为了验证 Paddle Operator 样本缓存加速方案的实际效果，我们选取了常规的 ResNet50 模型以及 ImageNet 数据集来进行性能测试，并使用了 [PaddleClas](https://github.com/PaddlePaddle/PaddleClas) 项目中提供的模型实现代码。
测试方案从两个方面来验证缓存组件（JuiceFS）的性能，分别是：远程文件加速访问的场景和利用本地缓存加速模型端到端训练效率的场景，并主要对比了缓存组件（JuiceFS）与 [BOS FS](http://baidu.netnic.com.cn/doc/BOS/BOSCLI/8.5CBOS.20FS.html) (百度云对象存储基于 FUSE 实现的挂载工具）的性能差异。

## 测试方案

- 模型：ResNet 50 v1.5
- 数据集：ImageNet（使用的是子集）
- 训练集：128,660 张图片 + 标签
- 分布式架构：Collective
- 模型实现代码: [PaddleClas](https://github.com/PaddlePaddle/PaddleClas)

## 准备样本数据

数据的准备工作可以参考 [Paddle Operator 样本缓存组件快速上手](./ext-get-start.md) 中创建和同步 ImageNet SampleSet 的步骤。
ImageNet 样本数据的子集存放在 BOS 的公开 bucket 中：`bos://paddleflow-public.hkg.bcebos.com/imagenet` ，数据存放目录如下：

```
.
├── demo
├── n01514859
├── train            # 训练样本图片集
└── train_list.txt   # 训练样本标签
```

## 准备测试镜像

我们基于 PaddlClas 项目构建好了测试镜像，并存放在可公开访问的镜像仓库中：`registry.baidubce.com/paddleflow-public/demo-resnet:v1` 。
您也可以参考 [Perf 项目](https://github.com/PaddlePaddle/Perf/tree/master/ResNet50V1.5) 或 [PaddleClas 项目](https://github.com/PaddlePaddle/PaddleClas) 自行构建 ResNet50 模型的训练镜像。

## 提交测试任务

1. 编写如下 resnet.yaml 文件

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: resnet50
  namespace: paddle-system
spec:
  cleanPodPolicy: Never
  sampleSetRef:
    name: imagenet
    mountPath: /data
  worker:
    # 指定多机配置
    replicas: 1
    template:
      spec:
        containers:
          - name: resent
            image: registry.baidubce.com/paddleflow-public/demo-resnet:v1
            command:
            - python
            args:
            - "-m"
            - "paddle.distributed.launch"
            - "--log_dir"    # 指定日志输出路径
            - "/data/log/"   # 这里将日志输出到挂盘的 /data 目录下，您可以更具实际需求更改日志输出路径
            - "./tools/train.py"
            - "-c"
            - "ResNet50.yaml"
            # - "-o"
            # - "DataLoader.sampler.batch_size=96"
            # 将宿主机的内存盘挂载进 Worker 容器内，防止读取样本时出现 OOM 错误。
            volumeMounts:
            - mountPath: /dev/shm
              name: dshm
            resources:
              limits:
                # 指定多卡配置
                nvidia.com/gpu: 1
        volumes:
        - name: dshm
          emptyDir:
            medium: Memory
```

其中 `ResNet50.yaml` 模型配置文件是参考 [文档](https://github.com/PaddlePaddle/PaddleClas/blob/release/2.2/ppcls/configs/ImageNet/ResNet/ResNet50.yaml) 修改的，您可以根据实际的硬件条件调整 `batch_size` 等参数，如添加参数 `-o DataLoader.sampler.batch_size=96`。

2. 提交 PaddleJob 完成模型训练

```bash
kubectl apply -f resnet.yaml
```
