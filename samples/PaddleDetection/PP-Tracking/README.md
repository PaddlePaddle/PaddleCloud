# PP-Tracking多目标跟踪部署实战

**PP-Tracking**是基于飞桨深度学习框架的业界首个开源实时跟踪系统。针对实际业务的难点痛点，PP-Tracking内置行人车辆跟踪、跨镜头跟踪、多类别跟踪、小目标跟踪及流量计数等能力与产业应用，同时提供可视化开发界面。模型集成多目标跟踪，目标检测，ReID轻量级算法，进一步提升PP-Tracking在服务器端部署性能。同时支持python，C++部署，适配Linux，Nvidia Jetson多平台环境。

![](https://ai-studio-static-online.cdn.bcebos.com/c8ffcde8c1974a69863806e3aeaf81c6680f3a8de50d4d56be9a14f155979f9e)

![](https://ai-studio-static-online.cdn.bcebos.com/ad76397a0c954b86b024a4330343e073e8ad0c1261da4a29be2a0f3bdbb196c6)
![](https://ai-studio-static-online.cdn.bcebos.com/38c90f74410041a3964e4f52b98f3ee881b443b6c7264f6ebd063b3c03ff8e4c)


在如下示例中，将介绍如何使用示例代码基于您在BML中已创建的数据集来完成单镜头跟踪模型的训练，评估和推理。以及多镜头的部署。

## Docker化部署

PaddleCloud基于 [Tekton](https://github.com/tektoncd/pipeline) 为Detection模型套件提供了镜像持续构建的能力，并支持CPU、GPU以及常见CUDA版本的镜像。
您可以查看 [PaddleDetection镜像仓库](https://hub.docker.com/r/paddlecloud/paddledetection) 来获取所有的镜像列表。
同时我们也将PP-Tracking多目标跟踪案例放置到了AI Studio平台上，您可以点击 [PP-Tracking之手把手玩转多目标跟踪](https://aistudio.baidu.com/aistudio/projectdetail/3916206?channelType=0&channel=0) 在平台上快速体验。

> **适用场景**：本地测试开发环境、单机部署环境。

## 1. 安装Docker

如果您所使用的机器上还没有安装 Docker，您可以参考 [Docker 官方文档](https://docs.docker.com/get-docker/) 来进行安装。
如果您需要使用支持 GPU 版本的镜像，则还需安装好NVIDIA相关驱动和 [nvidia-docker](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html#docker) 。

**注意**：如果您使用的是Windows系统，需要开启 [WSL2（Linux子系统功能）功能](https://docs.microsoft.com/en-us/windows/wsl/install)。

### 2. 启动容器

**使用CPU版本的Docker镜像**

```bash
# 这是加上参数 --shm-size=32g 是为了防止容器里内存不足
docker run --name ppdet -v $PWD:/mnt -p 8888:8888 -it --shm-size=32g paddlecloud/paddledetection:2.4-cpu-e9a542 /bin/bash
```

**使用GPU版本的Docker镜像**

```bash
docker run --name ppdet --runtime=nvidia -v $PWD:/mnt -p 8888:8888 -it --shm-size=32g paddlecloud/paddleocr:2.4-gpu-cuda10.2-cudnn7-e9a542 /bin/bash
```

进入容器内，则可进行PP-Tracking案例的实战体验。

### 3. 数据集准备

修改mot的配置文件，文件目录如下
整理之前：

```azure
MOT16
  └——————train
  └——————test
```

整理之后：

```
MOT16
|——————images
|        └——————train
|        └——————test
└——————labels_with_ids
└——————train
```

如果网速不好可以自行下载和解压数据集

```bash
# 下载数据集
cd ./data && wget https://bj.bcebos.com/v1/paddledet/data/mot/demo/MOT16.zip

mv ./data/MOT16.zip ./PaddleDetection/dataset/mot

# 解压数据集
cd ./PaddleDetection/dataset/mot && unzip MOT16.zip
```

生成labels_with_ids

```bash
cd ./PaddleDetection/dataset/mot/MOT16 && mkdir -p images

cd ./PaddleDetection/dataset/mot/MOT16 && mv ./train ./images && mv ./test ./images

cd ./PaddleDetection/dataset/mot && python gen_labels_MOT.py
```

生成mot16.train文件并且复制到 image_lists下面

```python
import glob
import os.path as osp
image_list = []
for seq in sorted(glob.glob('PaddleDetection/dataset/mot/MOT16/images/train/*')):
    for image in glob.glob(osp.join(seq, "img1")+'/*.jpg'):
        image = image.replace('PaddleDetection/dataset/mot/','')
        image_list.append(image)
with open('mot16.train','w') as image_list_file:
    image_list_file.write(str.join('\n',image_list))
```

```bash
mkdir -p ./PaddleDetection/dataset/mot/image_lists && cp -r mot16.train ./PaddleDetection/dataset/mot/image_lists
```

### 4. 修改配置文件里面的数据集

在/home/PaddleDetection/configs/mot/fairmot/fairmot_dla34_30e_1088x608.yml文件最后添加

```yaml
... ...
# for MOT training
# for MOT training
TrainDataset:
  !MOTDataSet
    dataset_dir: dataset/mot
    image_lists: ['mot16.train']
    data_fields: ['image', 'gt_bbox', 'gt_class', 'gt_ide']

# for MOT evaluation
# If you want to change the MOT evaluation dataset, please modify 'data_root'
EvalMOTDataset:
  !MOTImageFolder
    dataset_dir: dataset/mot
    data_root: MOT16/images/train
    keep_ori_im: False # set True if save visualization images or video, or used in DeepSORT

# for MOT video inference
TestMOTDataset:
  !MOTImageFolder
    dataset_dir: dataset/mot
    keep_ori_im: True # set True if save visualization images or video
```

### 5. 开始训练

使用MOT16-02序列作为训练数据，训练30epoch，V100环境下大约需要30分钟

```bash
# 这里使用GPU设备来训练
cd /home/PaddleDetection/ && python -m paddle.distributed.launch --log_dir=./fairmot_dla34_30e_1088x608/ --gpus 0 tools/train.py -c configs/mot/fairmot/fairmot_dla34_30e_1088x608.yml
```

### 6. 模型评估

为了方便我们可以下载训练好的模型进行eval https://paddledet.bj.bcebos.com/models/mot/fairmot_dla34_30e_1088x608.pdparams

```bash
# 将 model.pdparams 下载并放置到 `output` 文件中
mkdir -p PaddleDetection/output && cd PaddleDetection/output/ && wget https://bj.bcebos.com/v1/paddledet/models/mot/fairmot_dla34_30e_1088x608.pdparams

# 进行模型评估
cd /home/PaddleDetection && CUDA_VISIBLE_DEVICES=0 python tools/eval_mot.py -c configs/mot/fairmot/fairmot_dla34_30e_1088x608.yml -o weights=output/fairmot_dla34_30e_1088x608.pdparams
```

### 7. 模型推理

使用下载好的模型进行推理，为了方便我们只推理了dataset/mot/MOT16/images/test/MOT16-01/img1下面的数据

跟踪输出视频保存在output/mot_outputs/img1_vis.mp4

txt文件结果保存在output/mot_results/img1.txt,输出格式表示为frame_id, id, bbox_left, bbox_top, bbox_width, bbox_height, score, x, y, z

```bash
cd /home/PaddleDetection/ && CUDA_VISIBLE_DEVICES=0 python tools/infer_mot.py -c configs/mot/fairmot/fairmot_dla34_30e_1088x608.yml -o weights=output/fairmot_dla34_30e_1088x608.pdparams --image_dir=dataset/mot/MOT16/images/test/MOT16-01/img1  --save_videos
```

### 8. 导出模型

```bash
cd /home/PaddleDetection && CUDA_VISIBLE_DEVICES=0 python tools/export_model.py -c configs/mot/fairmot/fairmot_dla34_30e_1088x608.yml -o weights=output/fairmot_dla34_30e_1088x608.pdparams
```

### 9. 使用导出的模型进行推理

PP-Tracking中在部署阶段提供了多种跟踪相关功能，例如流量计数，出入口统计，绘制跟踪轨迹等，具体使用方法可以[参考文档](https://github.com/PaddlePaddle/PaddleDetection/tree/release/2.3/deploy/pptracking/python#%E5%8F%82%E6%95%B0%E8%AF%B4%E6%98%8E)

```bash
cd /home/PaddleDetection && wget https://bj.bcebos.com/v1/paddledet/data/mot/demo/person.mp4 && wget https://bj.bcebos.com/v1/paddledet/data/mot/demo/entrance_count_demo.mp4

# 输出视频保存在output/person.mp4中
cd /home/PaddleDetection && python deploy/pptracking/python/mot_jde_infer.py --model_dir=output_inference/fairmot_dla34_30e_1088x608 --video_file=person.mp4 --device=GPU

cd /home/PaddleDetection && python deploy/pptracking/python/mot_jde_infer.py --model_dir=output_inference/fairmot_dla34_30e_1088x608 --do_entrance_counting --draw_center_traj --video_file=entrance_count_demo.mp4 --device=GPU
```

## MTMCT 跨镜跟踪体验

跨镜头多目标跟踪是对同一场景下的不同摄像头拍摄的视频进行多目标跟踪，是监控视频领域一个非常重要的研究课题。

相较于单镜头跟踪，跨镜跟踪将不同镜头获取到的跟踪轨迹进行融合，得到跨镜跟踪的输出轨迹。PP-Tracking选用DeepSORT方案实现跨镜跟踪，为了达到实时性选用了PaddleDetection自研的PP-YOLOv2和PP-PicoDet作为检测器，选用PaddleClas自研的轻量级网络PP-LCNet作为ReID模型。更多内容可参考文档

本项目展示城市主干道场景下的车辆跨镜跟踪预测流程，数据来自[AIC21开源数据集](https://www.aicitychallenge.org/)

### 1. 下载预测部署模型

首先我们下载目标检测和ReID预测模型，[下载地址](https://github.com/PaddlePaddle/PaddleDetection/tree/develop/configs/mot/mtmct#deepsort%E5%9C%A8-aic21-mtmctcityflow-%E8%BD%A6%E8%BE%86%E8%B7%A8%E5%A2%83%E8%B7%9F%E8%B8%AA%E6%95%B0%E6%8D%AE%E9%9B%86test%E9%9B%86%E4%B8%8A%E7%9A%84%E7%BB%93%E6%9E%9C) ，然后统一放在~/PaddleDetection/output_inference下

```bash
wget https://paddledet.bj.bcebos.com/models/mot/deepsort/ppyolov2_r50vd_dcn_365e_aic21mtmct_vehicle.tar

wget https://paddledet.bj.bcebos.com/models/mot/deepsort/deepsort_pplcnet_vehicle.tar

cd /home/PaddleDetection/ && mkdir -p output_inference

mv ppyolov2_r50vd_dcn_365e_aic21mtmct_vehicle.tar ~/PaddleDetection/output_inference

mv deepsort_pplcnet_vehicle.tar ~/PaddleDetection/output_inference

cd ~/PaddleDetection/output_inference && tar -xvf ppyolov2_r50vd_dcn_365e_aic21mtmct_vehicle.tar && tar -xvf deepsort_pplcnet_vehicle.tar
```

### 2. 跨镜跟踪预测

在完成模型下载后，需要修改PaddleDetection/deploy/pptracking/python路径下的mtmct_cfg.yml，这份配置文件中包含了跨镜跟踪中轨迹融合的相关参数。首先需要确定cameras_bias中对应的名称与输入视频名称对应；其次，我们本次项目使用轨迹融合中的通用方法，将zone和camera相关的方法设置为False。修改后配置如下：

```yaml
# config for MTMCT
MTMCT: True
cameras_bias:
  c003: 0
  c004: 0
# 1.zone releated parameters
use_zone: False #True
zone_path: dataset/mot/aic21mtmct_vehicle/S06/zone
# 2.tricks parameters, can be used for other mtmct dataset
use_ff: True
use_rerank: True
# 3.camera releated parameters
use_camera: False #True
use_st_filter: False
# 4.zone releated parameters
use_roi: False #True
roi_dir: dataset/mot/aic21mtmct_vehicle/S06
```

配置完成后即可运行如下命令，输入视频为c003.mp4和c004.mp4两个不同视角的摄像头拍摄结果，跨镜跟踪输出视频保存在

```bash
wget https://bj.bcebos.com/v1/paddledet/data/mot/demo/mtmct-demo.tar && mv mtmct-demo.tar ~/PaddleDetection && cd ~/PaddleDetection && tar xvf mtmct-demo.tar 

cd ~/PaddleDetection && python deploy/pptracking/python/mot_sde_infer.py --model_dir=output_inference/ppyolov2_r50vd_dcn_365e_aic21mtmct_vehicle/ --reid_model_dir=output_inference/deepsort_pplcnet_vehicle/ --mtmct_dir=./mtmct-demo --device=GPU --mtmct_cfg=deploy/pptracking/python/mtmct_cfg.yml --scaled=True --save_mot_txts --save_images
```
