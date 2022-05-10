# 快速上手教程

该教程使用 Tekton pipeline 和 Tekton triggers，定时自动拉取 paddle 套件最新版本，制作镜像推送到 docker hub 镜像仓库。

## 安装 Tekton

本教程所有步骤均在 kubernates v1.21 版本进行测试，并默认您有一些 kubernates 的基础。如果您是其他版本 kubernates，请参考 [Tekton Pipeline](https://github.com/tektoncd/pipeline)  和 [Tekton triggers](https://github.com/tektoncd/triggers)，根据版本依赖关系选取正确的版本安装。

首先安装 Tekton pipelines

```bash
$ kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/previous/v0.34.1/release.yaml
```

接着安装 Tekton trigger

```bash
$ kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/previous/v0.19.1/release.yaml
$ kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/previous/v0.19.1/interceptors.yaml
```

（可选）如果您想使用 Tekton dashboard 更直观的查看各个任务，请安装 dashboard，使用方法参考 [Tekton Dashboard](https://tekton.dev/docs/dashboard/) 

```bash
$ kubectl apply --filename https://github.com/tektoncd/dashboard/releases/latest/download/tekton-dashboard-release.yaml
```

查看安装的 pods

```
$ kubectl get pods -n tekton-pipelines
NAME                                                 READY   STATUS    RESTARTS   AGE
tekton-triggers-webhook-79c866dc85-g42h8             1/1     Running   0          9d
tekton-triggers-controller-75b9b7b77d-pksdq          1/1     Running   0          9d
tekton-pipelines-webhook-6f44cfb768-rkf79            1/1     Running   0          9d
tekton-pipelines-controller-799fd96f87-wjdh4         1/1     Running   0          9d
tekton-dashboard-56fcdc6756-8dmpr                    1/1     Running   0          9d
tekton-triggers-core-interceptors-7769dc7cbc-n7nxd   1/1     Running   0          9d
```

如果您的网络不能访问 google 等网站，可能会无法拉取部分 docker 镜像导致安装失败，请搜索国内替代源。

## 安装自动构建镜像的 pipeline

### 配置 push docker 权限

本示例使用 docker hub 作为推送镜像的 registry，如有其他镜像站需求，请参考 [kaniko](https://github.com/GoogleContainerTools/kaniko) 完成相应的身份认证。

1. 在 docker hub 创建项目。

   如果没有 docker hub 账户的话，请先注册账户，然后点击 Create Repository 创建项目，例如账户名为 paddletest，创建的项目为 paddlecloud，则 push 镜像时，格式为 `docker push paddletest/paddlecloud:${tag}`

2. 登录 docker 并创建身份认证。

   ```bash
   $ docker login --usrname=paadletest		# paddletest 更改为你的账户
   ```

   登录成功后，会创建 `~/.docker/config.json` 认证文件

   接着在默认命名空间下创建名为 docker-push 的 secret

   ```bash
   # root 改为当前 linux 用户
   # 如果准备在其他命名空间创建 pipeline，这里请加上 -n $namespace_name
   $ kubectl create secret generic docker-push  --from-file=.dockerconfigjson=/root/.docker/config.json  --type=kubernetes.io/dockerconfigjson
   ```

### 安装 pipeline

``` bash
$ cd $PaddleCloud_path/tekton/example
$ kubectl create -f rbac.yaml					
$ kubectl create -f pipeline.yaml
```

## 配置并运行 pipeline

教程提供了两种使用 pipeilne 的方式，分别是单次直接运行构建镜像和定时运行构建镜像的 cronjob 方式。建议您先通过直接运行的方式进行配置测试，再使用 cronjob 方式进行部署。

### 单次运行方式

 `tekton/example` 下的 `pipelinerun.yaml` 文件如下，请根据表格中的详细说明，更改相应参数。

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  generateName: single-pipelinerun-
spec:
  serviceAccountName: tekton-triggers-example-sa
  pipelineRef:
    name: build-single-image
  workspaces:
  - name: shared-data
    volumeClaimTemplate:
      spec:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  params:
  - name: image-registry
    value: paddlecloud/paddleocr
  - name: toolkit-repo-name
    value: PaddleOCR
  - name: toolkit-revision
    value: release/2.4
  - name: toolkit-base-image-tag
    value: 2.2.2
  - name: docker-repo-url
    value: https://github.com/freeliuzc/PaddleCloud.git
  - name: docker-revision
    value: dev-tekton
  - name: dockerfile-path
    value: tekton/dockerfiles/Dockerfile
```

参数说明：

| 参数                   | 含义                                                         |
| ---------------------- | ------------------------------------------------------------ |
| image-registry         | 要推送到的 container image 仓库，例如账户名为 paddletest，仓库为 paddlecloud，这里填写 paadletest/paddlecloud |
| toolkit-repo-name      | paddle 的模型套件名称，目前支持 PaddleOCR、PaddleDetection、PaddleNLP、PaddleSeg、PaddleRec、PaddleClas、PaddleSpeech，其他套件可自行配置 |
| toolkit-revision       | 跟踪的套件分支，例如 release/2.X，develop，master            |
| toolkit-base-image-tag | 所要制作镜像依赖的 paddlepaddle 基础镜像 tag。模型套件和 paddlepaddle 版本依赖关系可在套件官网获取，已有的镜像可在 [docker hub](https://hub.docker.com/r/paddlepaddle/paddle/tags) 查找 |
| docker-repo-url        | 包含 Dockerfile 文件的 github 仓库                           |
| docker-revision        | 此仓库所使用的 branch                                        |
| dockerfile-path        | Dockerfile 文件在仓库的相对路径，例如放在仓库根目录下，填写 Dockerfile |

更改完成后，创建 pipelinerun 实例

```
$ kubectl create -f pipelinerun.yaml
pipelinerun.tekton.dev/single-pipelinerun-gvw2j created
# gvw2j为附加的随机码，每次运行都会不同
```

查看 pipelinerun

```
$ kubectl get pipelinerun
NAME                       SUCCEEDED   REASON      STARTTIME   COMPLETIONTIME
single-pipelinerun-gvw2j   True        Succeeded   94m         74m
```

使用 Tekton 工具查看 pipelinerun 日志

```
$ tkn pipelinerun logs -f single-pipelinerun-gvw2j -n default
```

或者您可以使用 [Tekton dashboard](https://tekton.dev/docs/dashboard/) 更直观的查看各个任务和日志。

删除 pipelinerun 及相应资源

```
$ kubectl delete pipelinerun single-pipelinerun-gvw2j
```

### 定时运行方式

该方式使用 Tekton 的 Triggers 作为事件接收器， 使用 kubernetes 的 cronjob 作为事件触发器，定时发送 curl 请求，完成流程。

首先安装 Trigger

```
$ cd $PaddleCloud_path/tekton/example
$ kubectl create -f trigger.yaml
```

接着更改 cronjob.yaml 文件中参数，参数含义参考上表

```
env:
- name: IMAGE_REGISTRY
  value: paddlecloud/paddleocr
- name: TOOLKIT_NAME
  value: PaddleOCR
- name: TOOLKIT_REVISION
  value: release/2.4
- name: TOOLKIT_BASE_IMAGE_TAG
  value: 2.2.2
- name: DOCKER_REPO_URL
  value: https://github.com/freeliuzc/PaddleCloud.git
- name: DOCKER_REVISION
  value: dev-tekton
- name: DOCKERFILE_PATH
  value: tekton/dockerfiles/Dockerfile
```

根据定时需求更改 cronjob.yaml 中 `schedule: "*/0 0 * * 0"` 字段内容，例如每周日的凌晨0:00运行

```
schedule: "*/0 0 * * 0"
```

安装 cronjob 

```
$ kubectl create -f cronjob.yaml
```

查看运行的 cronjob

```
$ kubectl get cronjobs
```

假如设置了每三分钟运行一次，三分钟后，可以通过命令查看 pipelinerun 的运行情况

```
$ kubectl get pipelinerun
$ kubectl get pods
```

删除 crobjob

```
$ kubectl delete cronjobs cleanup-cronjob
$ kubectl delete cronjobs curl-cronjob
```

