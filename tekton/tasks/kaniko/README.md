## 简介

该 task 基于 [Tekton 示例](https://hub.tekton.dev/tekton/task/kaniko) 进行更改，完成构建 docker 镜像并 push 到相应 registry 的任务。

## 使用示例

首先在本地创建一个空文件夹，此示例中为 [/root/docker](/root/docker)

分别将 docker 认证文件（一般在`~/.docker/config.json`）和 Dockerfile 移入此文件夹内

创建 storage-class

```
$ cd $PaddleCloud_path/tekton/tasks/kaniko
$ kubectl create -f local-storage.yaml
```

接着将 pv-demo 文件里`.spec.local.path` 需要更换为刚刚创建的文件夹路径，此示例为`/root/docker`

```
$ cat pv-demo.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-kaniko
  labels:
    pv: kaniko
spec:
  capacity:
    storage: 8Gi
  volumeMode: Filesystem
  accessModes:
  - ReadWriteOnce
  persistentVolumeReclaimPolicy: Delete
  storageClassName: local-storage
  local:
    path: /root/docker				# It should containe config.json and Dockerfile
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - mynode
```

分别创建 pv 和 pvc

```
$ kubectl create -f pv-demo.yaml
$ kubectl create -f pvc-demo.yaml
```

最后，根据需求更改 taskrun.yaml 内容

```yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: kaniko-run
spec:
  taskRef:
    name: kaniko
  params:
  - name: IMAGE
    value: lzc842650834/jarvis_tekton:0.3			# image name
  - name: EXTRA_ARGS
    value:
    - "--build-arg=IMAGE_TAG=2.2.2"						# It can pass ARG value
  workspaces:
  - name: source
    persistentVolumeClaim:
      claimName: pvc-kaniko
  - name: dockerconfig
    persistentVolumeClaim:
      claimName: pvc-kaniko
```