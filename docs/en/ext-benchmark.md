# Benchmark of Cache Component

In order to verify the actual effect of the cache component of Paddle Operator, we used the ResNet50 model and ImageNet data set for performance testing, and the ResNet50 model is implemented [PaddleClas](https://github.com/PaddlePaddle/PaddleClas) project.
The benchmark plan verifies the performance of the cache component (JuiceFS) from two aspects: small file acceleration and cache acceleration, and mainly compares the cache component (JuiceFS) and [BOS FS](http://baidu.netnic.com.cn/doc/BOS/BOSCLI/8.5CBOS.20FS.html) (Baidu Cloud Object Storage).

## Benchmark Plan

- Model: ResNet 50 v1.5
- DataSet: ImageNet Subset 128,660 pictures and labels
- Distributed Architecture：Collective
- Model Implementation Code: [PaddleClas](https://github.com/PaddlePaddle/PaddleClas)

## Prepare DataSet

For data preparation, please refer to the document of [quick start for cache component](./ext-get-start.md).
The ImageNet subset training data is stored in the public bucket of BOS: `bos://paddleflow-public.hkg.bcebos.com/imagenet` as follows:

```
.
├── demo
├── n01514859
├── train            # train set
└── train_list.txt   # labels of train set
```

## Prepare Test Container

We built a test container based on the PaddlClas project and stored it in a public image registry: `registry.baidubce.com/paddleflow-public/demo-resnet:v1`.
If you want build your own container, please refer to [Perf project](https://github.com/PaddlePaddle/Perf/tree/master/ResNet50V1.5) or [PaddleClas project](https://github.com/PaddlePaddle/PaddleClas).

## Submit Benchmark Job

1. prepare resnet.yaml file

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
    # Specified the number of host machines
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
            - "--log_dir"    # log dir
            - "/data/log/"   # you can change the log output path according to your needs
            - "./tools/train.py"
            - "-c"
            - "ResNet50.yaml"
            # - "-o"
            # - "DataLoader.sampler.batch_size=96"
            # Mount the host's memory disk into the Worker container to prevent OOM errors when reading samples.
            volumeMounts:
            - mountPath: /dev/shm
              name: dshm
            resources:
              limits:
                # Specified the number of GPU card in host
                nvidia.com/gpu: 1
        volumes:
        - name: dshm
          emptyDir:
            medium: Memory
```

The model configuration file is modified from [ResNet50.yaml](https://github.com/PaddlePaddle/PaddleClas/blob/release/2.2/ppcls/configs/ImageNet/ResNet/ResNet50.yaml), you can Adjust the parameters such as `batch_size` according to the actual hardware conditions, such as adding the parameter `-o DataLoader.sampler.batch_size=96`.

2. submit PaddleJob

```bash
kubectl apply -f resnet.yaml
```
