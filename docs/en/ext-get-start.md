<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Quick Start for Cache Component](#quick-start-for-cache-component)
  - [Requirements](#requirements)
  - [Install Cache Component](#install-cache-component)
    - [1. Install CRD](#1-install-crd)
    - [2. Deploy Operator](#2-deploy-operator)
    - [3. Deploy Controller](#3-deploy-controller)
    - [4. Install JuiceFS CSI Driver](#4-install-juicefs-csi-driver)
  - [Example](#example)
    - [1. Deploy Redis with Docker](#1-deploy-redis-with-docker)
    - [2. Create Secret](#2-create-secret)
    - [3. Create SampleSet](#3-create-sampleset)
    - [4. Experience SampleJob (Optional)](#4-experience-samplejob-optional)
    - [5. Create PaddleJob](#5-create-paddlejob)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Quick Start for Cache Component

This document mainly describes how to install and deploy the cache component of Paddle Operator, and demonstrates the basic functions of the cache component through actual examples.

## Requirements

* Kubernetes >= 1.8
* kubectl

## Install Cache Component

### 1. Install CRD

As described in the [README](../../README-zh_CN.md), if you are using a version earlier than 0.4, you can run this command again to update the CRD.
Create PaddleJob / SampleSet / SampleJob custom resources,
```shell
kubectl apply -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/main/deploy/v1/crd.yaml
```

Then you can use the command below to view the created custom resource
```shell
$ kubectl get crd | grep batch.paddlepaddle.org
NAME                                    CREATED AT
paddlejobs.batch.paddlepaddle.org       2021-08-23T08:45:17Z
samplejobs.batch.paddlepaddle.org       2021-08-23T08:45:18Z
samplesets.batch.paddlepaddle.org       2021-08-23T08:45:18Z
```

### 2. Deploy Operator

As described in the [README](../../README-zh_CN.md), if you are using a version earlier than 0.4, you can run this command again to update Operator.
```shell
kubectl apply -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/main/deploy/v1/operator.yaml
```

### 3. Deploy Controller

The YAML file in the following command includes the Controller of SampleSet and SampleJob
```shell
kubectl apply -f https://raw.githubusercontent.com/PaddleFlow/paddle-operator/main/deploy/extensions/controllers.yaml
```

You can view the deployed Controller through the following command
```shell
$ kubectl -n paddle-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
paddle-controller-manager-776b84bfb4-5hd4s   1/1     Running   0          60s
paddle-samplejob-manager-69b4944fb5-jqqrx    1/1     Running   0          60s
paddle-sampleset-manager-5cd689db4d-j56rg    1/1     Running   0          60s
```

### 4. Install JuiceFS CSI Driver

Now, the Paddle Operator cache component only supports [JuiceFS](https://github.com/juicedata/juicefs/blob/main/README_CN.md) as cache engine.

deploy JuiceFS CSI Driver
```shell
kubectl apply -f https://raw.githubusercontent.com/juicedata/juicefs-csi-driver/master/deploy/k8s.yaml
```

After deploying the CSI driver, you can check the status with the following command. Each worker node in the Kubernetes cluster should have a **juicefs-csi-node** Pod.
```shell
$ kubectl -n kube-system get pods -l app.kubernetes.io/name=juicefs-csi-driver
NAME                                          READY   STATUS    RESTARTS   AGE
juicefs-csi-controller-0                      3/3     Running   0          13d
juicefs-csi-node-87f29                        3/3     Running   0          13d
juicefs-csi-node-8h2z5                        3/3     Running   0          13d
```

**Note**: If Kubernetes cannot find the CSI driver, and an error like this: **driver name csi.juicefs.com not found in the list of registered CSI drivers**, this is because the CSI driver is not registered to the specified path of kubelet, You can fix it by the steps below.

Run the following command on any non-master node in your Kubernetes cluster:
```shell
ps -ef | grep kubelet | grep root-dir
```

If the result isn't empty, modify the CSI driver deployment k8s.yaml file with the new path and redeploy the CSI driver again. please replace {{KUBELET_DIR}} in the above command with the actual root directory path of kubelet.
```shell
curl -sSL https://raw.githubusercontent.com/juicedata/juicefs-csi-driver/master/deploy/k8s.yaml | sed 's@/var/lib/kubelet@{{KUBELET_DIR}}@g' | kubectl apply -f -
```

More Information please refer to [JuiceFS CSI Driver](https://github.com/juicedata/juicefs-csi-driver)

## Example

Since the JuiceFS cache engine relies on Redis to store metadata and supports multiple object storage as the data storage backend. In this example, we samply uses Redis as both the metadata engine and the data storage backend.

### 1. Deploy Redis with Docker

You can easily buy cloud Redis services on the cloud computing platform. This example uses Docker to run a Redis instance on the worker nodes of the Kubernetes cluster.
```shell
docker run -d --name redis \
	-v redis-data:/data \
	-p 6379:6379 \
	--restart unless-stopped \
	redis redis-server --appendonly yes
```

**Note**: The above command will mount the local directory `redis-data` to the `/data` the Docker container. You can mount different file directories as required.

### 2. Create Secret

Prepare the fields required by JuiceFS to format the file system. The required fields are as follows:

| Field  | Description |
|:-----|:-----|
| name    | The name of JuiceFS file system, it can be any string |
| storage | The name of object storage，e.g.: bos. More information please refer to [How to Setup Object Storage](https://github.com/juicedata/juicefs/blob/main/docs/en/how_to_setup_object_storage.md) |
| bucket  | The uri for data in object storage，e.g.: bos://imagenet.bj.bcebos.com/imagenet |
| metaurl | The uri of metadata，e.g.: redis://username:password@host:6379/0. More information please refer to [Databases for_Metadata](https://github.com/juicedata/juicefs/blob/main/docs/en/databases_for_metadata.md) |
| access-key | The access key of object storage (optional).If the object storage uses username and password, such as using Redis as the storage backend, fill in the username here. |
| secret-key | The secret key of object storage (optional). If the object storage uses username and password, such as using Redis as the storage backend, fill in the password here.  |

More information about JuiceFS please refer to：[Quick Start Guide](https://github.com/juicedata/juicefs/blob/main/docs/en/quick_start_guide.md) and [Command Reference](https://github.com/juicedata/juicefs/blob/main/docs/en/command_reference.md).

The fields in this example are configured as follows, because Redis is not configured with a password, the fields `access-key` and `secret-key` need not be set.

| Field  | Value  |
|:-----|:-----|
| name    | imagenet | 
| storage | redis |
| bucket  | redis://192.168.7.227:6379/1 |
| metaurl | redis://192.168.7.227:6379/0  |

**Note**: Please replace the IP `192.168.7.227` with the IP address of the host where you run the Redis container in the first step.

Then create secret.yaml file as below
```yaml
apiVersion: v1
data:
  name: aW1hZ2VuZXQ=
  storage: cmVkaXM=
  bucket: cmVkaXM6Ly8xOTIuMTY4LjcuMjI3OjYzNzkvMQ==
  metaurl: cmVkaXM6Ly8xOTIuMTY4LjcuMjI3OjYzNzkvMA==
kind: Secret
metadata:
  name: imagenet
  namespace: paddle-system
type: Opaque
```

The value of each field is Base64-encoded. You can obtain the Base64-encoded value of each field through the following command.
```bash
$ echo "redis://192.168.7.227:6379/0" | base64
cmVkaXM6Ly8xOTIuMTY4LjcuMjI3OjYzNzkvMAo=
```

Create secret by kubectl
```bash
$ kubectl create -f secret.yaml
secret/imagenet created
```

### 3. Create SampleSet

Write imagenet.yaml as follows, and the data source comes from a bucket publicized by bos (Baidu Object Storage).
```yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleSet
metadata:
  name: imagenet
  namespace: paddle-system
spec:
  # Number of partitions, a Kubernetes node represents a partition
  partitions: 1
  source:
    uri: bos://paddleflow-public.hkg.bcebos.com/imagenet/demo
  secretRef:
    name: imagenet
```

Create a SampleSet, wait for the data to complete synchronization, and check the status of the SampleSet. This operation may be time-consuming, please be patient.
```bash
$ kubectl create -f imagenet.yaml
sampleset.batch.paddlepaddle.org/imagenet created
$ kubectl get sampleset -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   58 MiB       200 B         12 GiB        1/1       Ready   76s
$ kubectl get pods -n paddle-system
NAME                                         READY   STATUS    RESTARTS   AGE
imagenet-runtime-0                           1/1     Running   0          30s
paddle-controller-manager-776b84bfb4-qs67f   1/1     Running   0          11h
paddle-samplejob-manager-685449d6d7-cmqfb    1/1     Running   0          11h
paddle-sampleset-manager-69bc7fb85d-4rjcg    1/1     Running   0          11h
```

When the PHASE status of the created SampleSet is Ready, it means that the data set is ready for use.

### 4. Experience SampleJob (Optional)

The cache component provides SampleJob to manage sample data sets. Currently, it supports 4 job types, namely sync/warmup/clear/rmr.

**Note**: Since the cache status of SampleSet is not updated in real time, the cache status will be updated within 30s after the SampleJob is created.

To delete the data in the cache engine, rmrOptions in rmr job is required, and the parameter in paths is the relative path of the data.
```bash
$ cat rmr-job.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleJob
metadata:
  name: imagenet-rmr
  namespace: paddle-system
spec:
  type: rmr
  sampleSetRef:
    name: imagenet
  rmrOptions:
    paths:
      - n01514859

$ kubectl create -f rmr-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-rmr created
$ kubectl get samplejob imagenet-rmr -n paddle-system
NAME           PHASE
imagenet-rmr   Succeeded
$ kubectl get sampleset imagenet -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   4.0 KiB      200 B         12 GiB        1/1       Ready   29m
```

Synchronize the sample data in the remote storage to the cache engine. The data is about 4G and have 128,000 pictures. Please ensure there is enough memory space.
```bash
$ cat sync-job.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleJob
metadata:
  name: imagenet-sync
  namespace: paddle-system
spec:
  type: sync
  sampleSetRef:
    name: imagenet
  # secretRef is required by SyncJob
  secretRef:
    name: imagenet
  syncOptions:
    source: bos://paddleflow-public.hkg.bcebos.com/imagenet

$ kubectl create -f sync-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-sync created
$ kubectl get samplejob imagenet-sync -n paddle-system
NAME            PHASE
imagenet-sync   Succeeded
$ kubectl get sampleset imagenet -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   4.2 GiB      4.2 GiB       7.3 GiB       1/1       Ready   83m
```

Clean up cached data in the cluster
```bash
$ cat clear-job.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleJob
metadata:
  name: imagenet-clear
  namespace: paddle-system
spec:
  type: clear
  sampleSetRef:
    name: imagenet
    
$ kubectl create -f clear-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-clear created
$ kubectl get samplejob imagenet-clear -n paddle-system
NAME             PHASE
imagenet-clear   Succeeded
$ kubectl get sampleset imagenet -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   4.2 GiB      80 B          12 GiB        1/1       Ready   85m
```

Prefetch the data in the cache engine to the Kubernetes cluster
```bash
$ cat warmup-job.yaml
apiVersion: batch.paddlepaddle.org/v1alpha1
kind: SampleJob
metadata:
  name: imagenet-warmup
  namespace: paddle-system
spec:
  type: warmup
  sampleSetRef:
    name: imagenet

$ kubectl create -f warmup-job.yaml
samplejob.batch.paddlepaddle.org/imagenet-warmup created
$ kubectl get samplejob imagenet-warmup -n paddle-system
NAME              PHASE
imagenet-warmup   Succeeded
$ kubectl get sampleset imagenet -n paddle-system
NAME       TOTAL SIZE   CACHED SIZE   AVAIL SPACE   RUNTIME   PHASE   AGE
imagenet   4.2 GiB      4.2 GiB       7.3 GiB       1/1       Ready   90m
```

### 5. Create PaddleJob

The following example uses nginx to simply demonstrate how to mount SampleSet in PaddleJob. You can refer to [Benchmark](./ext-benchmark.md) for more information about performance testing.

Create ps-demo.yaml
```yaml
apiVersion: batch.paddlepaddle.org/v1
kind: PaddleJob
metadata:
  name: ps-demo
  namespace: paddle-system
spec:
  sampleSetRef:
    # Declare the SampleSet to be used
    name: imagenet
    # The mounting path of the SampleSet in the Worker Pods
    mountPath: /mnt/imagenet
  worker:
    replicas: 1
    template:
      spec:
        containers:
          - name: sample
            image: nginx
  ps:
    replicas: 1
    template:
      spec:
        containers:
          - name: sample
            image: nginx
```

Check the status of PaddleJob
```bash
$ kubectl get paddlejob -n paddle-system
NAME      STATUS    MODE   AGE
ps-demo   Running   PS     112s
```

Check the sample data mounted on PaddleJob worker pod
```bash
$ kubectl exec -it ps-demo-worker-0 -n paddle-system -- /bin/bash
$ ls /mnt/imagenet
demo  train  train_list.txt
```