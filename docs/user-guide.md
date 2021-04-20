# User guide

This doc refer to the user with paddle-operator installed and exploring a advanced use of paddlepaddle in training scenario.

## preparation

We will fully explain two examples *wide and deep* [source](https://github.com/PaddlePaddle/FleetX/tree/develop/examples/wide_and_deep)
and *resnet* [source](https://github.com/PaddlePaddle/FleetX/tree/develop/examples/resnet),
the resources can be downloaded following the links.

## Example : wide and deep

This job is running in PS mode with CPU only, 
the entrypoint of program is *train.py* in folder *wide_and_deep*.

With the provided *Dockerfile* showing as follows,

```
$ ls
Dockerfile   wide_and_deep

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

```
docker build -t registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1 .
```

User may change the last lines to add more code to run the job, 
furthermore, one can also change the *paddlepaddle* version or install more dependency.

The docker image should be pushed to the register before it can be used in the kubernetes cluster,
```
docker push registry.baidubce.com/paddle-operator/demo-wide-and-deep:v1
```

> User should replace the registry and add authority before pushing it.

With docker image ready, we can define the yaml config file for *kubectl*,

```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: wide-ande-deep
spec:
  withGloo: 1
  intranet: PodIP
  cleanPodPolicy: OnCompletion
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

* The optional configuration of withGloo is 0 not enabled, 1 only starts the worker side, 2 starts all (worker and server), it is recommended to set 1;
* The cleanPodPolicy can be optionally configured as Always/Never/OnFailure/OnCompletion, which indicates whether to delete the pod when the task is terminated (failed or successful). It is recommended to Never during debugging and OnCompletion during production;
* The intranet can be optionally configured as Service/PodIP, which means the communication method between pods. The user does not need to configure it, and PodIP is used by default;
* The content of ps and worker is podTemplateSpec. Users can add more content according to the Kubernetes specification, such as GPU configuration.

Finally, the job can be created by
```
kubectl -n paddle-system create -f demo-wide-and-deep.yaml
```

## Example : resnet

This job is running in Collective mode with GPU, 
the entrypoint of program is *train_fleet.py* in folder *resnet*.

With the provided *Dockerfile* showing as follows,
```
$ ls
Dockerfile   resnet

$ cat Dockerfile

FROM registry.baidubce.com/paddle-operator/paddle-base-gpu:cuda10.2-cudnn7-devel-ubuntu18.04

ADD resnet /resnet

WORKDIR /resnet

RUN pip install scipy opencv-python==4.2.0.32 -i https://mirror.baidu.com/pypi/simple

CMD ["python","-m","paddle.distributed.launch","train_fleet.py"]

```

> The cuda version of the docker image should be compatible with the one of cluster, more detail can be found by running *nvidia-smi*.

The docker image is build by
```
docker build -t registry.baidubce.com/paddle-operator/demo-resnet:v1 .
```

and pushed with authority

```
docker push registry.baidubce.com/paddle-operator/demo-resnet:v1
```

The yaml config file for *kubectl* is something like

```
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
* Here you need to add shared memory to mount to prevent cache errors;
* This example uses the built-in data set. After the program is started, it will be downloaded. Depending on the network environment, it may wait a long time.


Finally, the job can be created by

```
kubectl -n paddle-system create -f resnet.yaml
```

## More configuration

### Volcano support

[Volcano](https://github.com/volcano-sh/volcano) should be installed alongside the kubernetes cluster with namespaces will configured.

The following example shows the ability of *gan-scheduling* of paddlejob with benefit to volcano.

Here is the yaml config file,
```
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
            image: registry.baidubce.com/kuizhiqing/demo-wide-and-deep:v1
  ps:
    replicas: 2
    template:
      spec:
        restartPolicy: "Never"
        containers:
          - name: paddle
            image: registry.baidubce.com/kuizhiqing/demo-wide-and-deep:v1

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

### GPU and Node selector

This example shows the ability of paddle-operator work with gpu and nodeselector which is very useful in practice.

```
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
            image: registry.baidubce.com/kuizhiqing/demo-wide-and-deep:v1
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
            image: registry.baidubce.com/kuizhiqing/demo-wide-and-deep:v1
            resources:
              limits:
                nvidia.com/gpu: 1
        nodeSelector:
          accelerator: nvidia-tesla-p100
```

### Data Storage

In the examples above, training data are contained in the docker image which is not a good or possible choice in practice,
however, kubernetes provide the standard PV(*persistent volume*) and PVC(*persistent volume claim*) abstraction.

The following example show the usage of pv/pvc with paddlejob, one may use alternative backend pv in specific circumstance.
```
$ cat pv-pvc.yaml
---
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

> pvc is namespace related.

```
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
            image: registry.baidubce.com/kuizhiqing/paddle-ubuntu:2.0.0-18.04
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
            image: registry.baidubce.com/kuizhiqing/paddle-ubuntu:2.0.0-18.04
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
