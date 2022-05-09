# PP-OCRv3 识别训推一体 Docker 部署实战

PaddleOCR 是百度开源的超轻量级OCR模型库，提供了数十种文本检测、识别模型，旨在打造一套丰富、领先、实用的文字检测、识别模型/工具库，助力使用者训练出更好的模型，并应用落地。

本教程旨在帮助使用者基于 [云上飞桨（PaddleCloud）]((https://github.com/PaddlePaddle/PaddleCloud)) 来快速体验和部署 PP-OCRv3 识别模型，并掌握其使用方式，包括：

1. PP-OCR3 识别快速使用
2. 文件识别模型的训练和预测方式

## 1. PaddleCloud 简介

[云上飞桨（PaddleCloud）](https://github.com/PaddlePaddle/PaddleCloud) 是面向飞桨 （PaddlePaddle）框架及其模型套件的部署工具箱，为用户提供了模型套件 Docker 化部署和 Kubernetes 集群部署两种方式，满足不同场景与环境的部署需求。
在本案例中，我们使用 PaddleCloud 提供的 PaddleOCR 镜像大礼包来进行 Docker 化部署。

### 1.1 准备环境

如果您所使用的机器上还没有安装 Docker，您可以参考 [Docker 官方文档](https://docs.docker.com/get-docker/) 来进行安装。 如果您需要使用支持 GPU 版本的镜像，则还需安装好 NVIDIA 相关驱动和 nvidia-docker，详情请参考[官方文档](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html#docker) 。

**使用CPU版本的Docker镜像**

```bash
docker run --name ppocr -v $PWD:/mnt -p 8888:8888 -it paddlecloud/paddleocr:dygraph-cpu-364817 /bin/bash
```

**使用GPU版本的Docker镜像**

```bash
docker run --name ppocr --runtime=nvidia -v $PWD:/mnt -p 8888:8888 -it paddlecloud/paddleocr:dygraph-gpu-cuda10.2-cudnn7-364817 /bin/bash
```

进入容器内，则可进行 PP-OCRv3 模型的训练和部署工作。

### 1.2 PP-OCRv3 检测模型介绍

PP-OCRv3 采用CML的蒸馏策略，训练配置文件为`/home/PaddleOCR/configs/det/ch_PP-OCRv3/ch_PP-OCRv3_det_cml.yml`，CML蒸馏训练策略包含三个模型，分别是蒸馏教师模型以及两个蒸馏学生模型。

网络结构配置如下：

```yaml
Architecture:
  name: DistillationModel
  algorithm: Distillation
  model_type: det
  Models:                   
    Student:                 # CML蒸馏的Student模型配置
      model_type: det
      algorithm: DB
      Transform: null
      Backbone:
        name: MobileNetV3    # Student模型backbone使用mobilev3
        scale: 0.5
        model_name: large
        disable_se: true
      Neck:
        name: RSEFPN         # Student模型neck部分使用PaddleOCR中的RSEFPN
        out_channels: 96
        shortcut: True
      Head:
        name: DBHead
        k: 50
    Student2:                # Student2模型的配置同Student
      model_type: det
      algorithm: DB
      Transform: null
      Backbone:
        name: MobileNetV3
        scale: 0.5
        model_name: large
        disable_se: true
      Neck:
        name: RSEFPN
        out_channels: 96
        shortcut: True
      Head:
        name: DBHead
        k: 50
    Teacher:                 # Teacher模型配置
      freeze_params: true
      return_all_feats: false
      model_type: det
      algorithm: DB
      Backbone:
        name: ResNet         # Teacher使用resnet50作为backbone 
        in_channels: 3
        layers: 50         
      Neck:
        name: LKPAN          # Teacher模型使用PaddleOCR中的LKPAN作为neck网络
        out_channels: 256
      Head:
        name: DBHead
        kernel_list: [7,2,2]  
        k: 50
```

注：PP-OCRv3 模型分别在网络结构做了以下优化
- Student模型使用RSEPAN提升模型召回和精度；
- Teacher模型使用LKPAN提升模型精度和召回；

详细策略介绍参考链接：https://github.com/PaddlePaddle/PaddleOCR/blob/dygraph/doc/doc_ch/PP-OCRv3_introduction.md


### 1.3 准备训练数据

本教程以HierText数据集为例，介绍PP-OCRv3检测模型的蒸馏训练方式。 HierText是第一个具有自然场景和文档中文本分层注释的数据集。该数据集包含从 Open Images 数据集中选择的 11639 张图像，提供高质量的单词 (~1.2M)、行和段落级别的注释。HierText数据集下载地址：https://github.com/google-research-datasets/hiertext

值得注意的是该数据集的标注格式与ppocrlabel格式不一样，我们需要对其数据标签格式进行相应的转换。

您可以从AI Studio中直接下载标签格式转换后的HierText数据集：https://aistudio.baidu.com/aistudio/datasetdetail/143700

您可以通过运行如下指令，完成数据集的下载和解压操作：

```bash
# 下载数据集
$ cd /mnt && wget https://paddleflow-public.hkg.bcebos.com/ppocr/hiertext1.tar

# 解压数据集
$ cd /mnt && tar xf hiertext1.tar && mv hiertext1 hiertext
```

运行上述命令后，在 `/mnt` 目录下包含以下文件：

```
/mnt/hiertext
  └─ train/     HierText训练集数据
  └─ validation/     HierText验证集数据
  └─ label_hiertext_train.txt  HierText训练集的行标注
  └─ label_hiertext_val.txt    HierText验证集的行标注
```

其中，paddleocr支持的标注文件格式为：

```
# 图像文件的路径                            json.dumps编码的图像标注信息"
hiertext/train/1b1b8bd73eb47995.jpg       [{"points": [[758, 283], [971, 267], [972, 279], [758, 294]], "transcription": "We are not programming in 1969 anymore"}, ...]
```

其中图像标注信息中包含两种参数：

- $points$表示文本框的四个点的绝对坐标(x, y)，从左上角的点开始顺时针排列。
- $transcription$表示当前文本框的文字内容，在文本检测任务中无需使用这个信息。

如果您想在其他数据集上训练PaddleOCR，可以按照上述形式构建标注文件。

之后您需要修改训练配置文件ch_PP-OCRv3_det_cml.yml中的训练数据为HierText数据。

- 修改训练数据配置：

```yaml
Train:
  dataset:
    name: SimpleDataSet
    data_dir: ./train_data/icdar2015/text_localization/
    label_file_list:
      - ./train_data/icdar2015/text_localization/train_icdar2015_label.txt
```

修改为：

```yaml
Train:
  dataset:
    name: SimpleDataSet
    data_dir: /mnt/
    label_file_list:
      - /mnt/hiertext/label_hiertext_train.txt
```

- 修改验证数据配置：

```yaml
Eval:
  dataset:
    name: SimpleDataSet
    data_dir: ./train_data/icdar2015/text_localization/
    label_file_list:
      - ./train_data/icdar2015/text_localization/test_icdar2015_label.txt
```

修改为：

```yaml
Eval:
  dataset:
    name: SimpleDataSet
    data_dir: /mnt/
    label_file_list:
      - /mnt/hiertext/label_hiertext_val.txt
```

### 1.4 启动训练

下载 PP-OCRv3 的蒸馏预训练模型并进行训练的方式如下

```bash
# 下载预训练模型到~/PaddleOCR/pre_train文件夹下
$ mkdir /home/PaddleOCR/pre_train

$ cd /home/PaddleOCR/pre_train

$ wget https://paddleocr.bj.bcebos.com/PP-OCRv3/chinese/ch_PP-OCRv3_det_distill_train.tar

$ tar xf ch_PP-OCRv3_det_distill_train.tar
```

启动训练，训练模型默认保存在 output 目录下，加载 PP-OCRv3 检测预训练模型。

```bash
$ cd /home/PaddleOCR
 
# 这里以 GPU 训练为例，使用 CPU 进行训练的话，需要指定参数 Global.use_gpu=false
$ python3 tools/train.py -c configs/det/ch_PP-OCRv3/ch_PP-OCRv3_det_cml.yml -o Global.save_model_dir=./output/ Global.pretrained_model=./pre_train/ch_PP-OCRv3_det_distill_train/best_accuracy
```

如果要使用多GPU分布式训练，请使用如下命令：

```bash
# 启动训练，训练模型默认保存在output目录下，--gpus '0,1,2,3'表示使用0，1，2，3号GPU训练
python3 -m paddle.distributed.launch --log_dir=./debug/ --gpus '0,1,2,3' tools/train.py -c configs/det/ch_PP-OCRv3/ch_PP-OCRv3_det_cml.yml -o Global.save_model_dir=./output/ Global.pretrained_model=./pre_train/ch_PP-OCRv3_det_distill_train/best_accuracy
```

### 1.5 模型评估

训练过程中保存的模型在output目录下，包含以下文件：

```
best_accuracy.states    
best_accuracy.pdparams  # 默认保存最优精度的模型参数
best_accuracy.pdopt     # 默认保存最优精度的优化器相关参数
latest.states    
latest.pdparams  # 默认保存的最新模型参数
latest.pdopt     # 默认保存的最新模型的优化器相关参数
```

其中，best_accuracy是保存的最优模型，可以直接使用该模型评估

```bash
# 进行模型评估
cd /home/PaddleOCR/

python3 tools/eval.py -c configs/det/ch_PP-OCRv3/ch_PP-OCRv3_det_cml.yml -o Global.checkpoints=./output/best_accuracy
```

### 1.6 基于训练模型的预测

使用上述步骤训练好的模型，测试文本检测效果。我们在 ./doc/imgs_en/文件夹下准备了一些测试图像，您也可以上传自己的图像测试我们的OCR检测模型。

```bash
# 进行检测
cd /home/PaddleOCR/

python3 tools/infer_det.py -c configs/det/ch_PP-OCRv3/ch_PP-OCRv3_det_cml.yml -o Global.checkpoints=./output/best_accuracy Global.infer_img=./doc/imgs_en/img_12.jpg
```

预测可视化的图像默认保存在./checkpoints/det_db/目录下，运行下述代码进行可视化

```python
import matplotlib.pyplot as plt
from PIL import Image
## 显示原图，读取名称为12.jpg的测试图像
img_path= "./checkpoints/det_db/det_results_Student/img_12.jpg"
img = Image.open(img_path)
plt.figure("test_img", figsize=(10,10))
plt.imshow(img)
plt.show()
```

<div align="center">
  <img src="./img/figure-1.png" title="architecture" width="80%" height="80%" alt="">
</div>


### 1.7 基于推理引擎预测

模型训练好后，可以将模型固化为文件，以便于部署

运行如下指令，可将训练好的模型导出为预测部署模型

```bash
# 导出为预测部署模型
cd /home/PaddleOCR/

python3 tools/export_model.py -c configs/det/ch_PP-OCRv3/ch_PP-OCRv3_det_cml.yml -o Global.checkpoints=./output/best_accuracy Global.save_inference_dir=./inference/ 
```

运行完后，导出的预测部署模型位于inference目录下，组织结构为：

```
inference
├── Student      # 保存的精度最高的Student模型
│   ├── inference.pdiparams
│   ├── inference.pdiparams.info
│   └── inference.pdmodel
├── Student2    # CML训练方法中的第二个student模型，精度低于Student
│   ├── inference.pdiparams
│   ├── inference.pdiparams.info
│   └── inference.pdmodel
└── Teacher     # 蒸馏教师模型
    ├── inference.pdiparams
    ├── inference.pdiparams.info
    └── inference.pdmodel
```

Student下的模型为导出的精度最高的模型。下面以Student的inference模型为例，介绍inference模型的使用方法。

注：关于inference模型的更多使用示例，参考：https://github.com/PaddlePaddle/PaddleOCR/blob/dygraph/doc/doc_ch/inference_ppocr.md

```bash
# 使用inference模型进行文字检测
cd /home/PaddleOCR/

python3 tools/infer/predict_det.py --image_dir=./doc/imgs_en/img_10.jpg --det_model_dir=./inference/Student/
```

```python
## 显示轻量级模型识别结果
## 可视化det_res_img_10.jpg的文本检测效果
import matplotlib.pyplot as plt
from PIL import Image
img_path= "./inference_results/det_res_img_10.jpg"
img = Image.open(img_path)
plt.figure("results_img", figsize=(20,20))
plt.imshow(img)
plt.show()
```

<div align="center">
  <img src="./img/figure-2.png" title="architecture" width="80%" height="80%" alt="">
</div>
