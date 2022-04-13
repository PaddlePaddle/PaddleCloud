[toc]

## Introduction

You need three steps to ues paddlejob:

1. Build docker image.
2. Write configuration file for paddlejob.
3. Deploy paddlejob and monitor the job status.

Paddlejob supports the following functions：

 - Support PS, Collective and Heterthree training modes.
 - Support scheduling training jobs using Volcano.
 - Supports elastic expansion training with fault tolerance.支持训练作业

The tutorial provides multiple usage examples, including using CPU or GPU training, using volcano to schedule tasks, using external data, etc. Example 1 introduces  how to build docker image and related parameter descriptions in detail. For other examples, please refer to Example 1.

All the examples can be found in `$path_to_PaddleCloud/samples/paddlejob`.

## Example 1. Training wide & deep with CPU

This job is running in PS mode with CPU only.

### Build docker image

The entrypoint of program is **train.py** in folder **wide_and_deep**.

With the provided *Dockerfile* showing as follows,

```dockerfile
$ ls
Dockerfile wide_and_deep

$ cat Dockerfile
FROM ubuntu:18.04

RUN apt update && \
    apt install -y python3 python3-dev python3-pip

RUN python3 -m pip install paddlepaddle==2.0.0 -i https://mirror.baidu.com/pypi/simple

## add user files

ADD wide_and_deep /wide_and_deep

WORKDIR /wide_and_deep

ENTRYPOINT ["python3", "train.py"]
```

The docker image can be build by

```bash
docker build -t registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1 .
```

User may change the last lines to add more code to run the job, 
furthermore, one can also change the **paddlepaddle** version or install more dependency.

The docker image should be pushed to the register before it can be used in the kubernetes cluster.

> We use Baidu register here, user should replace the registry and add authority before pushing it.

```bash
docker push registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
```

### Deploy paddlejob

With docker image ready, we can define the yaml config file for kubectl.

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  cleanPodPolicy: Never
  withGloo: 1
  worker:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
  ps:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
```

- The **name in metadata** must be unique. You should delete the previous one if there is a conflict;
- The **cleanPodPolicy** can be optionally configured as Always/Never/OnFailure/OnCompletion, which indicates whether to delete the pod when the task is terminated (failed or successful). It is recommended to Never during debugging and OnCompletion during production;
- The optional configuration of **withGloo** is 0 not enabled, 1 only starts the worker side, 2 starts all (worker and server), it is recommended to set 1;
- The **intranet** can be optionally configured as Service/PodIP, which means the communication method between pods. The user does not need to configure it, and PodIP is used by default;
- The **content of ps and worker** is **podTemplateSpec**. Users can add more content according to the Kubernetes specification, such as GPU configuration.
- In ps mode, you need to configure ps and worker at the same time. In collective mode, you only need to configure worker；

Finally, the job can be created by

```bash
$ kubectl -n paddlecloud apply -f wide_and_deep.yaml
```

### Check job running state

Check pods state.

```shell
$ kubectl -n paddlecloud get pods
```

Check PaddleJob state(pdj is the abbreviation of paddlejob).

```shell
kubectl -n paddlecloud get pdj
```

## Example 2. Training resnet with GPU

This job is running in Collective mode with GPU, 
the entrypoint of program is **train_fleet.py** in folder **resnet**.

With the provided **Dockerfile** showing as follows.

```dockerfile
$ ls
Dockerfile   resnet

$ cat Dockerfile
FROM registry.baidubce.com/paddle-operator/paddle-base-gpu:cuda10.2-cudnn7-devel-ubuntu18.04

ADD resnet /resnet

WORKDIR /resnet

RUN pip install scipy opencv-python==4.2.0.32 -i https://mirror.baidu.com/pypi/simple

CMD ["python","-m","paddle.distributed.launch","train_fleet.py"]
```

> The cuda version of the docker image should be compatible with the one of cluster, more detail can be found by running *nvidia-smi*.

The docker image is build by

```bash
docker build -t registry.baidubce.com/paddle-operator/demo-resnet:v1 .
```

and pushed with authority

```bash
docker push registry.baidubce.com/paddle-operator/demo-resnet:v1
```

The yaml config file for kubectl is something like

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: resnet
spec:
  cleanPodPolicy: Never
  worker:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-resnet:v1
            command:
            - python
            args:
            - "-m"
            - "paddle.distributed.launch"
            - "train_fleet.py"
            volumeMounts:
            - mountPath: /dev/shm
              name: dshm
            resources:
              limits:
                nvidia.com/gpu: 1
        volumes:
        - name: dshm
          emptyDir:
            medium: Memory
```

- Here you need to add shared memory to mount to prevent cache errors;
- This example uses the built-in data set. After the program is started, it will be downloaded. Depending on the network environment, it may wait a long time.

Finally, the job can be created by

```bash
kubectl -n paddlecloud apply -f resnet.yaml
```

## Example 3. Training wide & deep with GPU and nodeselector

This example shows the ability of paddlejob work with gpu and nodeselector which is very useful in practice.

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  intranet: Service
  cleanPodPolicy: OnCompletion
  worker:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
            resources:
              limits:
                nvidia.com/gpu: 1
        nodeSelector:
          accelerator: nvidia-tesla-p100
  ps:
    replicas: 2
    template:
      spec:
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
            resources:
              limits:
                nvidia.com/gpu: 1
        nodeSelector:
          accelerator: nvidia-tesla-p100
```

## Example 4. Volcano support

The following example shows the ability of *gan-scheduling* of paddlejob with benefit to volcano.

1. Please make sure you have installed [Volcano](https://github.com/volcano-sh/volcano)。
2. Please make sure you have installed paddlejob as section [2.3 Install configuration for using volcano]().
3. Create file wide_and_deep_volcano.yaml.

```yaml
---
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  cleanPodPolicy: Never
  withGloo: 1
  worker:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        schedulerName: volcano
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
  ps:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1

---
apiVersion: scheduling.volcano.sh/v1beta1
kind: PodGroup
metadata:
  name: wide-ande-deep
spec:
  minMember: 4
```

Pay attention that paddlejob and podgroup should share the same name, 
here only the ability of gan0scheduling are benefit, more configuration are available for podgroup which is beyond the scope.

In this example, the total pods number are 4, so the paddlejob will scheduled until resources are fully ready.

### Crashed caused by volcano

We find a bug cauesd by volcano. If you uninstall volcano and find that all deployments fail to start pods, execute the following command.

```bash
$ kubectl delete validatingwebhookconfigurations volcano-admission-service-jobs-validate volcano-admission-service-pods-validate volcano-admission-service-queues-validate
$ kubectl delete mutatingwebhookconfigurations volcano-admission-service-jobs-mutate volcano-admission-service-podgroups-mutate volcano-admission-service-pods-mutate volcano-admission-service-queues-mutate
```

If you are interested about it, please refer to [github issue](https://github.com/volcano-sh/volcano/issues/2102).

## Example 5. Data Storage

In the examples above, training data are contained in the docker image which is not a good or possible choice in practice. However, kubernetes provide the standard [PV(*persistent volume*) and PVC(*persistent volume claim*)](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) abstraction.

The following example show the usage of pv/pvc with paddlejob and we choose nfc here.

1. Create file pv-pvc.yaml for deploying pv/pvc.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs-pv
spec:
  capacity:
    storage: 10Gi
  volumeMode: Filesystem
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Recycle
  storageClassName: slow
  mountOptions:
    - hard
    - nfsvers=4.1
  nfs:
    path: /nas
    server: 10.12.201.xx

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs-pvc
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 10Gi
  storageClassName: slow
  volumeName: nfs-pv
```

2. Deploy pv/pvc.

The namespace of pv/pvc should be the same with paddlejob, we still use paddlejob here. In this example, the data is actually stored in the cloud space on 10.12.201.xx.

```bash
$ kubectl -n paddlecloud apply -f pv-pvc.yaml
```

3. Create data_demo.yaml.

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: paddlejob-demo-1
spec:
  cleanPolicy: OnCompletion
  worker:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/paddle-ubuntu:2.0.0-18.04
            command: ["bash","-c"]
            args: ["cd /nas/wide_and_deep; python3 train.py"]
            volumeMounts:
            - mountPath: /nas
              name: data
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: nfs-pvc
  ps:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/paddle-operator/paddle-ubuntu:2.0.0-18.04
            command: ["bash","-c"]
            args: ["cd /nas/wide_and_deep; python3 train.py"]
            volumeMounts:
            - mountPath: /nas
              name: data
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: nfs-pvc
```

该示例中，镜像仅提供运行环境，训练代码和数据均通过存储挂载的方式添加。

4. The same step as previous examples.

```bash
$ kubectl -n paddlecloud apply -f data_demo.yaml
```

