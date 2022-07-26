# PP-Human Docker部署

**适用场景：测试开发环境、单机、异构硬件。**

本节内容主要介绍下如何在本地环境测试和开发环境中使用PaddleDetection模型套件的官方标准Docker镜像来进行PP-Human案例的开发与部署


## 背景介绍

社区是城市的关键组成部分，社区治理是围绕社区场景下的人、地、物、情、事的管理与服务。

随着城市化的快速推进及人口流动的快速增加，传统社区治理在人员出入管控、安防巡逻、车辆停放管理等典型场景下都面临着人力不足、效率低下、响应不及时等诸多难题。而人工智能技术代替人力，实现人、车、事的精准治理，大幅降低人力、物质、时间等成本，以最低成本发挥最强大的管理效能，有效推动城市治理向更“数字化、自动化、智慧化”的方向演进。

百度飞桨针对智慧社区中的典型场景，提供了**实时行人分析工具PP-Human**，基于行人检测与跟踪，实现**26种人体属性分析以及摔倒等异常行为识别**。

**详细文档可参考：https://github.com/PaddlePaddle/PaddleDetection/tree/develop/deploy/pphuman**

### 场景一：社区人员信息管理

传统社区视频监控80%都依靠人工实现，随着摄像头在社区中的大规模普及，日超千兆的视频图像数据、人员信息的日渐繁杂已远超人工的负荷。

因此，上海天覆科技灵活应用飞桨行人分析PP-Human中的人体跟踪和属性识别算法实现社区视频监控的结构化留痕，实时识别进出小区的人员的**性别、年龄、衣着打扮等26种属性并记录其运动轨迹**，实现**AI算法完全代替人力**，高效、准确地完成**出入口管理、快速寻人、轨迹分析**等任务。

<div align="center">
  <img src="https://user-images.githubusercontent.com/48054808/159898428-5bda0831-7249-4889-babd-9165f26f664d.gif" width="700"/>
  <br>
  街道人员属性识别
</div>

## 场景二：摔倒识别
社区的安全防护是重中之重，如何高效保障社区居民人身安全成为一大难题，传统的方式以人工视频监控为主，人力巡逻为辅，往往面临异常情况响应不及时、人力消耗极大的问题。

飞桨行人分析PP-Human中提供的摔倒识别算法，采用了**关键点+时空图卷积网络**的技术，**对摔倒姿势无限制、背景环境无要求**，助力多家安防龙头企业实现了不同方向、不同姿态、不同光照情况下40毫秒的实时摔倒识别，避免因人力监管不到位造成的救援拖延，完成社区安防系统智能化的全面转型。

<div align="center">
  <img src="https://user-images.githubusercontent.com/48054808/159898395-061cdca2-a076-46ac-9251-9d2ca9fcb8a4.gif" width="700"/>
  <br>
  办公区域摔倒检测
</div>

#### 接下来将手把手带您使用PP-Human实现以上两个场景！

## 安装

使用的Docker环境可以快速上手体验，我们为您提供了CPU和GPU版本的镜像。

**使用CPU版本的Docker镜像**

```bash
docker run --name pphuman -v $PWD:/mnt -p 8888:8888 -it registry.baidubce.com/paddleflow-public/paddledetection:2.4 /bin/bash
```

**使用GPU版本的Docker镜像**

```bash
docker run --name pphuman --runtime=nvidia -v $PWD:/mnt -p 8888:8888 -it registry.baidubce.com/paddleflow-public/paddledetection:2.4-gpu-cuda10.2-cudnn7 /bin/bash
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

## 下载对应模型

场景一社区人员信息留存涉及模型：目标检测、属性识别

场景二人员摔倒涉及模型：目标检测、关键点检测、行为识别

其中，属性分析包含26种不同属性：
```
- 性别：男、女
- 年龄：小于18、18-60、大于60
- 朝向：朝前、朝后、侧面
- 配饰：眼镜、帽子、无
- 正面持物：是、否
- 包：双肩包、单肩包、手提包
- 上衣风格：带条纹、带logo、带格子、拼接风格
- 下装风格：带条纹、带图案
- 短袖上衣：是、否
- 长袖上衣：是、否
- 长外套：是、否
- 长裤：是、否
- 短裤：是、否
- 短裙&裙子：是、否
- 穿靴：是、否
```
行为识别目前仅支持摔倒，后续会补齐打架、抽烟、玩手机、睡觉等异常行为检测。

接下来，逐一下载上述模型。

```bash
#进入到 PaddleDetection 目录 /home/PaddleDetection
cd /home/PaddleDetection

#下载检测模型
wget https://bj.bcebos.com/v1/paddledet/models/pipeline/mot_ppyoloe_l_36e_pipeline.zip

#下载属性模型
wget https://bj.bcebos.com/v1/paddledet/models/pipeline/strongbaseline_r50_30e_pa100k.zip

#下载关键点模型
wget https://bj.bcebos.com/v1/paddledet/models/pipeline/dark_hrnet_w32_256x192.zip

#下载行为识别模型
wget https://bj.bcebos.com/v1/paddledet/models/pipeline/STGCN.zip

#解压至./output_inference文件夹
unzip -d output_inference mot_ppyoloe_l_36e_pipeline.zip
unzip -d output_inference strongbaseline_r50_30e_pa100k.zip
unzip -d output_inference dark_hrnet_w32_256x192.zip
unzip -d output_inference STGCN.zip
```

## 下载示例数据

本案例提供了一些示例数据，用于测试模型效果，您也可以使用自己的数据来体验下。

```bash
# 切换目录到挂盘路径/mnt，下载示例数据包并解压
cd /mnt && wget https://paddleflow-public.hkg.bcebos.com/ppdet/pphuman.zip && unzip pphuman.zip
```

## 配置文件说明

PP-Human相关配置位于`deploy/pphuman/config/infer_cfg.yml`中，存放模型路径，完成不同功能需要设置不同的任务类型。 (完整的路径是:　`/home/PaddleDetection/deploy/pphuman/config/infer_cfg.yml`)

功能及任务类型对应表单如下：

| 输入类型 | 功能 | 任务类型 | 配置项 |
|-------|-------|----------|-----|
| 图片 | 属性识别 | 目标检测 属性识别 | DET ATTR |
| 单镜头视频 | 属性识别 | 多目标跟踪 属性识别 | MOT ATTR |
| 单镜头视频 | 行为识别 | 多目标跟踪 关键点检测 行为识别 | MOT KPT ACTION |

本项目应用到的是：
- 社区人员信息留存：单镜头视频或图片输入的属性识别
- 摔倒检测：单镜头视频输入的摔倒识别

### **本次项目不需要修改配置文件信息，可通过模型预测时的参数选择对应任务。**

## 模型预测

选取视频样本进行属性识别与摔倒检测。

模型预测参数选择分为两部分：
- 功能选择：将对应参数设置为True
    - 属性识别：enable_attr
    - 行为识别：enable_action
- 模型路径修改：设置对应任务(DET, MOT, ATTR, KPT, ACTION)的模型路径
    - 例如 图片输入的属性识别：

``` 
--model_dir det=output_inference/mot_ppyoloe_l_36e_pipeline/ attr=output_inference/strongbaseline_r50_30e_pa100k/
``` 

### 场景一：社区人员信息留存

支持开发者根据具体情况选择**视频**或**单帧图片**输入进行属性识别。

注意事项：
- video_file or image_dir后是输入视频or图片的路径，开发者可上传自己的数据集进行尝试

```bash
# 视频行人属性识别
python deploy/pphuman/pipeline.py \
    --config deploy/pphuman/config/infer_cfg.yml \
    --model_dir mot=output_inference/mot_ppyoloe_l_36e_pipeline/ attr=output_inference/strongbaseline_r50_30e_pa100k/ \
    --video_file=/mnt/pphuman/attr.mp4 \
    --enable_attr=True \
    --device=cpu
```

```bash
# 图片行人属性识别
python deploy/pphuman/pipeline.py \
    --config deploy/pphuman/config/infer_cfg.yml \
    --model_dir det=output_inference/mot_ppyoloe_l_36e_pipeline/ attr=output_inference/strongbaseline_r50_30e_pa100k/ \
    --image_dir=/mnt/pphuman/img \
    --enable_attr=True \
    --device=cpu
```

### 场景二：摔倒检测

```bash
# 摔倒检测
python deploy/pphuman/pipeline.py \
    --config deploy/pphuman/config/infer_cfg.yml \
    --model_dir mot=output_inference/mot_ppyoloe_l_36e_pipeline/ kpt=output_inference/dark_hrnet_w32_256x192/ action=output_inference/STGCN \
    --video_file=/mnt/pphuman/fall.mp4 \
    --enable_action=True \
    --device=cpu
```

## 更多资源

飞桨目标检测开发套件PaddleDetection，内置190个主流目标检测、实例分割、跟踪、关键点检测算法，其中包括服务器端和移动端产业级SOTA模型、冠军方案和学术前沿算法，并提供配置化的网络模块组件、十余种数据增强策略和损失函数等高阶优化支持和多种部署方案，在打通数据处理、模型开发、训练、压缩、部署全流程的基础上，提供丰富的案例及教程，加速算法产业落地应用。

- 如果你发现任何PaddleDetection存在的问题或者是建议, 欢迎通过[GitHub Issues](https://github.com/PaddlePaddle/PaddleDetection/issues)给我们提issues。

- 欢迎加入PaddleDetection QQ、微信（添加并回复小助手“检测”）用户群

<div align="center">
<img src="https://user-images.githubusercontent.com/48054808/157800129-2f9a0b72-6bb8-4b10-8310-93ab1639253f.jpg"  width = "190" />  
<img src="https://user-images.githubusercontent.com/48054808/157800287-a4fced21-2821-4e55-8fd6-6aae167122c2.png"  width = "200" />  
</div>
