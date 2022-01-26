# JupyterHub Notebook使用手册

**JupyterHub Notebook 组件链接：** https://github.com/jupyterhub/zero-to-jupyterhub-k8s

## 概述

云上飞桨产品使用 JupyterHub Notebook 作为用户开发模型的交互式编程入口。JupyterHub 是一个多用户的 Jupyter 门户，在设计之初就把多用户创建、资源分配、数据持久化等功能做成了插件模式。

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-c15ccbdce6e861e20dfd8d0d5308f7ce2b9293ca)

上图是 JupyterHub 系统架构图，其由四个子系统组成：

 - **服务管理中心** 基于 tornado 构建，它是 JupyterHub 的核心模块。
 - **HTTP 代理模块** 可以作为节点代理，用于接收来自客户端浏览器的请求。
 - **Spawners模块**  用于监控的多个用户的 Jupyter notebook 服务器。
 - **身份验证模块** 用于管理用户权限并授权用户如何访问系统界面。

zero to jupyterhub k8s 项目支持多租户隔离、动态资源分配、数据持久化、数据隔离、高可用、权限控制等功能，能够支持公司级别（上万）用户规模。

## 在 Kubernetes 集群安装 JupyterHub

在安装 JupyterHub 前您需要有 Kubernetes 集群环境，并在当前节点上安装好 [Helm，](https://helm.sh/docs/)Helm 可以通过以下命令来安装：

```bash
curl https://raw.githubusercontent.com/helm/helm/HEAD/scripts/get-helm-3 | bash
```

检查 Helm 状态与版本

```bash
helm version
```

### 1）初始化 Helm Chart 配置文件

Helm Chart 可以渲染需要安装的 Kubernetes 资源的模板。通过 使用 Helm Chart，用户可以覆盖 Chart 的默认值进行自定义安装。

在这一步中，我们将初始化一个图表配置文件，以便您修改 JupyterHub 安装时的初试值。从版本 1.0.0 开始，您无需进行任何配置即可开始使用，因此只需创建一个带有一些有用注释的 config.yaml 文件即可。创建  config.yaml 文件如下：

```yaml
# This file can update the JupyterHub Helm chart's default configuration values.
#
# For reference see the configuration reference and default values, but make
# sure to refer to the Helm chart version of interest to you!
#
# Introduction to YAML:     https://www.youtube.com/watch?v=cdLNKUoMc6c
# Chart config reference:   https://zero-to-jupyterhub.readthedocs.io/en/stable/resources/reference.html
# Chart default values:     https://github.com/jupyterhub/zero-to-jupyterhub-k8s/blob/HEAD/jupyterhub/values.yaml
# Available chart versions: https://jupyterhub.github.io/helm-chart/
#
```

### 2）将 JupyterHub Chart 添加到 Helm 仓库

```bash
helm repo add jupyterhub https://jupyterhub.github.io/helm-chart/
helm repo update
```

输出如下：

```bash
Hang tight while we grab the latest from your chart repositories...
...Skip local chart repository
...Successfully got an update from the "stable" chart repository
...Successfully got an update from the "jupyterhub" chart repository
Update Complete. ⎈ Happy Helming!⎈
```

### 3）安装 JupyterHub Chart

在包含上述 config.yaml 文件的目录中，执行如下命令：

```bash
helm upgrade --cleanup-on-fail \
  --install <helm-release-name> jupyterhub/jupyterhub \
  --namespace <k8s-namespace> \
  --create-namespace \
  --version=<chart-version> \
  --values config.yaml
```

其中：

- `<helm-release-name>` 指的是 [Helm 版本名称](https://helm.sh/docs/glossary/#release)，用于区分 Chart 安装的标识符，当您更改或删除此 Chart 安装的配置时，版本号是必填项。 如果您的 Kubernetes 集群将包含多个 JupyterHub，请确保区分它们，您可以使用 `helm list` 查看各个 JupyterHub Chart 的版本。
- `<k8s-namespace>` 指的是 [Kubernetes 命名空间](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)，在本例中是用于对 Kubernetes 资源进行分组的标识符 与 JupyterHub Chart 关联的所有 Kubernetes 资源。 你需要命名空间标识符来执行任何带有 `kubectl` 的命令。
- 此步骤可能需要一些时间，在此期间您的终端不会有任何输出。JupyterHub 正在后台安装。
- 如果安装过程中出现 `release named <helm-release-name> already exists` 错误，那么您应该通过运行 `helm delete <helm-release-name>` 删除该版本。 然后重复此步骤重新安装。 如果仍然存在，请执行 `kubectl delete namespace <k8s-namespace>` 并重试。
- 如果安装步骤出现错误，请在重新运行安装命令之前通过运行 `helm delete <helm-release-name>` 删除 Helm 版本。
- 如果拉取 Docker 镜像出现错误，比如 `Error: timed out waiting for the condition ，您可以在 `helm` 命令中添加一个 `--timeout=<number-of-minutes>m` 参数。
- `--version` 参数对应的是 Helm Chart 的*版本*，而不是 JupyterHub 的版本。 每个版本的 JupyterHub Helm Chart 都与特定版本的 JupyterHub 配对。 例如，Helm 图表的 `0.11.1` 运行 JupyterHub `1.3.0`。 有关每个版本的 JupyterHub Helm Chart 中安装了哪个 JupyterHub 版本的列表，请参阅 [Helm Chart 仓库 ](https://jupyterhub.github.io/helm-chart/)。

### 4）查看 Pod 状态

在第 2 步运行时，您可以在另一个终端中输入下面的命令来查看正在创建的 pod：

```bash
kubectl get pod --namespace jhub
```

### 5）等待所有 Pod 成功运行

等待 hub 和 proxy pod 进入 Running 状态。

```bash
NAME                    READY     STATUS    RESTARTS   AGE
hub-5d4ffd57cf-k68z8    1/1       Running   0          37s
proxy-7cb9bc4cc-9bdlp   1/1       Running   0          37s
```

### 6）获取 JupyterHub Notebook 访问 IP

运行以下命令，直到 proxy-public 服务的 EXTERNAL-IP 可用，如示例输出中所示。

```bash
$ kubectl get service --namespace <k8s-namespace>
NAME           TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)        AGE
hub            ClusterIP      10.51.243.14    <none>          8081/TCP       1m
proxy-api      ClusterIP      10.51.247.198   <none>          8001/TCP       1m
proxy-public   LoadBalancer   10.51.248.230   104.196.41.97   80:31916/TCP   1m
```

最后，请在浏览器中输入代理公共服务的外部 IP。 JupyterHub 使用默认的虚拟身份验证器运行，因此输入任何用户名和密码组合都可以让您登入 Notebook。用户界面如下：

![图片](http://bos.bj.bce-internal.sdns.baidu.com/agroup-bos-bj/bj-f8b9d6a5a2b4e342aad987f27799bcc6abc6ad60)

更多自定义安装文档请查看：https://zero-to-jupyterhub.readthedocs.io/en/latest/jupyterhub/installation.html

