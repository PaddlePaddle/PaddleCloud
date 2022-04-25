# PP-YOLOE Docker训练

**适用场景：测试开发环境、单机部署。**

本节内容主要介绍下如何在本地环境测试和开发环境中使用PaddleDetection模型套件的官方标准Docker镜像来训练PP-YOLOE模型。

## 背景知识
PP-YOLOE是基于PP-YOLOv2的卓越的单阶段Anchor-free模型，超越了多种流行的yolo模型。PP-YOLOE有一系列的模型，即s/m/l/x，可以通过width multiplier和depth multiplier配置。PP-YOLOE避免使用诸如deformable convolution或者matrix nms之类的特殊算子，以使其能轻松地部署在多种多样的硬件上。更多细节可以参考我们的[report](https://arxiv.org/abs/2203.16250)。

<div align="center">
  <img src="https://raw.githubusercontent.com/PaddlePaddle/PaddleDetection/release/2.4/docs/images/ppyoloe_map_fps.png" width=500 />
</div>

PP-YOLOE-l在COCO test-dev2017达到了51.4的mAP, 同时其速度在Tesla V100上达到了78.1 FPS。PP-YOLOE-s/m/x同样具有卓越的精度速度性价比, 其精度速度可以在[模型库](#模型库)中找到。

PP-YOLOE由以下方法组成
- 可扩展的backbone和neck
- [Task Alignment Learning](https://arxiv.org/abs/2108.07755)
- Efficient Task-aligned head with [DFL](https://arxiv.org/abs/2006.04388)和[VFL](https://arxiv.org/abs/2008.13367)
- [SiLU激活函数](https://arxiv.org/abs/1710.05941)

## 模型库
|          模型           | GPU个数 | 每GPU图片个数 |  骨干网络  | 输入尺寸 | Box AP<sup>val</sup> | Box AP<sup>test</sup> | Params(M) | FLOPs(G) | V100 FP32(FPS) | V100 TensorRT FP16(FPS) | 模型下载 | 配置文件  |
|:------------------------:|:-------:|:--------:|:----------:| :-------:| :------------------: | :-------------------: |:---------:|:--------:|:---------------:| :---------------------: | :------: | :------: |
| PP-YOLOE-s                  |     8      |    32    | cspresnet-s |     640     |       42.7        |        43.1         |   7.93    |  17.36   |       208.3 |          333.3          | [model](https://paddledet.bj.bcebos.com/models/ppyoloe_crn_s_300e_coco.pdparams) | [config](https://github.com/PaddlePaddle/PaddleDetection/tree/release/2.4/configs/ppyoloe/ppyoloe_crn_s_300e_coco.yml)                   |
| PP-YOLOE-m                  |     8      |    28    | cspresnet-m |     640     |       48.6        |        48.9         |   23.43   |  49.91   |   123.4   |  208.3   | [model](https://paddledet.bj.bcebos.com/models/ppyoloe_crn_m_300e_coco.pdparams) | [config](https://github.com/PaddlePaddle/PaddleDetection/tree/release/2.4/configs/ppyoloe/ppyoloe_crn_m_300e_coco.yml)                   |
| PP-YOLOE-l                  |     8      |    20    | cspresnet-l |     640     |       50.9        |        51.4         |   52.20   |  110.07  |   78.1    |  149.2   | [model](https://paddledet.bj.bcebos.com/models/ppyoloe_crn_l_300e_coco.pdparams) | [config](https://github.com/PaddlePaddle/PaddleDetection/tree/release/2.4/configs/ppyoloe/ppyoloe_crn_l_300e_coco.yml)                   |
| PP-YOLOE-x                  |     8      |    16    | cspresnet-x |     640     |       51.9        |        52.2         |   98.42   |  206.59  |   45.0    |   95.2   | [model](https://paddledet.bj.bcebos.com/models/ppyoloe_crn_x_300e_coco.pdparams) | [config](https://github.com/PaddlePaddle/PaddleDetection/tree/release/2.4/configs/ppyoloe/ppyoloe_crn_x_300e_coco.yml)                   |

**注意:**

- PP-YOLOE模型使用COCO数据集中train2017作为训练集，使用val2017和test-dev2017作为测试集，Box AP<sup>test</sup>为`mAP(IoU=0.5:0.95)`评估结果。
- PP-YOLOE模型训练过程中使用8 GPUs进行混合精度训练，如果训练GPU数和batch size不使用上述配置，须参考[FAQ](https://github.com/PaddlePaddle/PaddleDetection/blob/develop/docs/tutorials/FAQ)调整学习率和迭代次数。
- PP-YOLOE模型推理速度测试采用单卡V100，batch size=1进行测试，使用**CUDA 10.2**, **CUDNN 7.6.5**，TensorRT推理速度测试使用**TensorRT 6.0.1.8**。
- 参考[速度测试](#速度测试)以复现PP-YOLOE推理速度测试结果。
- 如果你设置了`--run_benchnark=True`, 你首先需要安装以下依赖`pip install pynvml psutil GPUtil`。

## 使用教程

### 安装

执行以下指令使用混合精度训练PP-YOLOE

**使用CPU版本的Docker镜像**

```bash
docker run --name pphuman -v $PWD:/root/.cache/paddle -p 8888:8888 -it registry.baidubce.com/paddleflow-public/paddledetection:2.4 /bin/bash
```

**使用GPU版本的Docker镜像**

```bash
docker run --name pphuman --runtime=nvidia -v $PWD:/root/.cache/paddle -p 8888:8888 -it registry.baidubce.com/paddleflow-public/paddledetection:2.4-gpu-cuda10.2-cudnn7 /bin/bash
```

**升级Jupyter Notebook**

```
pip3 install --upgrade ipykernel ipython
```

**启动Jupyter Notebook服务**

```
jupyter notebook --ip 0.0.0.0 --allow-root
```

**初次使用Jupyter Notebook密码**

不需要在登陆框中输入任何东西．

在服务器控制复制token输入到token框中，然后在下面的文本框中自己设定一个密码

启动Jupyter Notebook服务后，您就可以通过浏览器访问Docker容器内的Notebook服务啦。

### 训练

执行以下指令使用混合精度训练PP-YOLOE

```bash
python tools/train.py -c configs/ppyoloe/ppyoloe_crn_l_300e_coco.yml -o use_gpu=false
```

### 评估

在coco test-dev2017上评估，请先从[COCO数据集下载](https://cocodataset.org/#download)下载COCO test-dev2017数据集，然后解压到COCO数据集文件夹并像`configs/ppyolo/ppyolo_test.yml`一样配置`EvalDataset`。

执行以下命令在单个GPU上评估COCO val2017数据集

```bash
python tools/eval.py -c configs/ppyoloe/ppyoloe_crn_l_300e_coco.yml
```

### 部署

PP-YOLOE可以使用以下方式进行部署：
- Paddle Inference [Python](../../deploy/python) & [C++](../../deploy/cpp)
- [Paddle-TensorRT](../../deploy/TENSOR_RT.md)
- [PaddleServing](https://github.com/PaddlePaddle/Serving)

接下来，我们将介绍PP-YOLOE如何使用Paddle Inference在TensorRT FP16模式下部署

首先，参考[Paddle Inference文档](https://www.paddlepaddle.org.cn/inference/master/user_guides/download_lib.html#python)，下载并安装与你的CUDA, CUDNN和TensorRT相应的wheel包。

然后，运行以下命令导出模型

```bash
python tools/export_model.py -c configs/ppyoloe/ppyoloe_crn_l_300e_coco.yml -o trt=True
```

**注意：**
- TensorRT会根据网络的定义，执行针对当前硬件平台的优化，生成推理引擎并序列化为文件。该推理引擎只适用于当前软硬件平台。如果你的软硬件平台没有发生变化，你可以设置[enable_tensorrt_engine](https://github.com/PaddlePaddle/PaddleDetection/blob/release/2.4/deploy/python/infer.py#L660)的参数`use_static=True`，这样生成的序列化文件将会保存在`output_inference`文件夹下，下次执行TensorRT时将加载保存的序列化文件。
- PaddleDetection release/2.4及其之后的版本将支持NMS调用TensorRT，需要依赖PaddlePaddle release/2.3及其之后的版本

## 附录

PP-YOLOE消融实验

| 序号 |        模型                  | Box AP<sup>val</sup> | 参数量(M) | FLOPs(G) | V100 FP32 FPS |
| :--: | :---------------------------: | :-------------------: | :-------: | :------: | :-----------: |
|  A   | PP-YOLOv2          |         49.1         |   54.58   |  115.77   |     68.9     |
|  B   | A + Anchor-free    |         48.8         |   54.27   |  114.78   |     69.8     |
|  C   | B + CSPRepResNet   |         49.5         |   47.42   |  101.87   |     85.5     |
|  D   | C + TAL            |         50.4         |   48.32   |  104.75   |     84.0     |
|  E   | D + ET-Head        |         50.9         |   52.20   |  110.07   |     78.1     |
