## 简介

该 task 基于 [Tekton 示例](https://hub.tekton.dev/tekton/task/kaniko) 进行更改，构建 docker 镜像并 push 到 docker hub 仓库，如果要推送到其他仓库，请参考 [kaniko](https://github.com/GoogleContainerTools/kaniko)。使用前，请确保您已经安装 [Tekton pipeline](https://github.com/tektoncd/pipeline)。

## 使用示例

首先在本地创建一个空文件夹，此示例中为 `/root/docker`；

使用 `docker login --username=$you_username` 登录 docker hub 后，会生成`~/.docker/config.json`认证文件，类似下面的格式:

```
{
    "auths": {
        "https://index.docker.io/v1/": {
            "auth": "xxxxxxxxxxxx"
        }
    }
}
```

将 config.json 和你的 Dockerfile 复制到 `/root/docker`；

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

根据需求更改 taskrun.yaml 配置

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

运行任务

```
$ kubectl create -f taskrun.yaml
```

使用 Tekton 工具查看任务日志

```
$ tkn taskrun logs -f -n default kaniko-run
```

删除任务

```
$ kubectl delete taskrun kaniko-run
```

