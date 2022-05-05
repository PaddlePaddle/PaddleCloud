# 介绍

该项目使用 Tekton pipelines 和 Triggers，实现了自动拉取 paddle 模型套件最新版本（例如 PaddleOCR、PaddleDetection）并制作镜像推送到 docker hub 的流程，并使用 kubernates cronjob 支持该流程的长期运行。

# 使用指南

## 安装 Tekton

本文所有步骤均在 kubernates v1.21 版本进行测试，并默认您有一些 kubernates 的基础。如果您是其他版本 kubernates，请参考 [Tekton Pipeline](https://github.com/tektoncd/pipeline)  和 [Tekton triggers](https://github.com/tektoncd/triggers)，根据 kubernates 版本选取对应的版本安装。

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

如果您没有 vpn，可能会遇到无法 pull docker image 的问题，请使用国内的替代资源。

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

我们提供了两种使用 pipeilne 的方式，分别是单次执行的 pipelinerun 方式和定时执行的 cronjob 方式。建议您先通过 pipelinerun 方式配置并测试 pipeline 的运行情况，再使用 cronjob 方式进行部署。

### pipelinerun 运行方式

pipelinerun 示例文件如下，请根据参数的详细说明，更改为您自己的使用配置

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  Name: auto-build-pipelinerun
spec:
  serviceAccountName: tekton-triggers-example-sa
  pipelineRef:
    name: auto-build-pipeline
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
  - name: toolkit-repo-name
    value: PaddleOCR
  - name: toolkit-revision      
    value: release/2.4
  - name: toolkit-base-image-tag
    value: 2.2.2
  - name: docker-repo-url
    value: https://github.com/freeliuzc/PaddleCloud.git
  - name: docker-revision      
    value: main
  - name: dockerfile-path
    value: images/tekton/dockerfiles/paddleocr/Dockerfile.cpu
  - name: image-registry
    value: paddlecloud/paddleocr
# - name: kaniko-image
#   value: gcr.io/kaniko-project/executor:latest  
```

参数说明：

| 参数                   | 含义                                                         |
| ---------------------- | ------------------------------------------------------------ |
| toolkit-repo-name      | paddle 的模型套件名称，目前支持 PaddleOCR、PaddleDetection   |
| toolkit-revision       | 套件的 release 分支，例如 release/2.3，release/2.4           |
| toolkit-base-image-tag | 依赖的 paddle 基础镜像，模型套件和基础镜像依赖关系可在套件官网获取，已有的镜像 tag 可在 [docker hub](https://hub.docker.com/r/paddlepaddle/paddle/tags) 查找 |
| docker-repo-url        | 包含 dockerfile 的 github 仓库                               |
| docker-revision        | 此仓库所使用的 branch / commit 版本                          |
| dockerfile-path        | Dockerfile 文件在仓库的相对路径，例如放在仓库根目录下，填写 Dockerfile |
| image-registry         | 要推送到的 container image 仓库，例如账户名为 paddletest，仓库为 paddlecloud，这里填写 paadletest/registry |

 创建 pipelinerun

```
$ kubectl create -f pipelinerun.yaml
```

 查看 pipelinerun 日志

```
$ tkn pipelinerun logs -f auto-build-pipeline -n default
```

或者您可以使用 [Tekton dashboard](https://tekton.dev/docs/dashboard/) 更直观的查看各个任务和日志。

删除 pipelinerun 及相应资源

```
$ kubectl delete pipelinerun auto-build-pipeline
```

### Cronjob 运行方式

1. 首先和 pipelinerun.yaml 一样，更爱 cron-template 中的 pipeline 配置参数，各个参数的具体含义可参考上表

2. （可选）使用 cron 语法更改`schedule: "*/0 0 * * 0"`字段内容，以满足您的定时需求

```
schedule: "*/0 0 * * 0"
```

3. 安装 cronjob

```
$ cd $PaddleCloud/tekton/example
$ kubectl create -f cronjob.yaml
```

4. 查看运行的 cronjob

```
$ kubectl get cronjobs
```

5. 删除 crobjob

```
$ kubectl delete cronjobs cleanup-cronjob
$ kubectl delete cronjobs curl-cronjob
```

