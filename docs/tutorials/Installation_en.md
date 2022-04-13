English | [简体中文](./Installation.md)
- [Installation](#installation)
  - [1. Prerequisites](#1-prerequisites)
  - [2. Install charts](#2-install-charts)
    - [2.1 Quick installation of all components](#21-quick-installation-of-all-components)
    - [2.2 Custom installation](#22-custom-installation)
    - [2.3 Install configuration for using volcano](#23-install-configuration-for-using-volcano)
  - [3. Paddlejob test](#3-paddlejob-test)
  - [4. Uninstall](#4-uninstall)
  - [5. Advanced usage](#5-advanced-usage)
# Installation

## 1. Prerequisites

* Kubernetes, version: v1.21
* kubectl
* helm

> Major components such as Paddlejob can run on kubernetes v1.16+ and we test all the contents at version v1.21. If you want to experience PaddleCloud in a easy way, it is recommended to use version v1.21 for the stable operation of other dependencies.

If you do not have a Kubernetes environment, you can refer to [microk8s documentation]() for installation. If you are using the macOS system, or encounter installation problems, you can refer to the documentation [macOS installation microk8s](./macOS_install_microk8s.md).

## 2. Install charts

> Please make sure that you have completed the installation and configuration of kubernates, and you can connect the cluster through kubectl and helm.

If you are new to kubernates or experience PaddleCloud in a brand new kubernates cluster, it is recommended to install all components and dependencies with one command in **Section 2.1**. If you will deploy PaddleCloud in an existing kubernates cluster or in a production environment, it is recommended to  use a custom installation method in **Section 2.2**.

### 2.1 Quick installation of all components

> In this tutorial, we use paddlecloud as namespace by default

Add and update helm's charts repositories.

```bash
$ helm repo add paddlecloud https://paddleflow-public.hkg.bcebos.com/charts
$ helm repo update
```

Install all components and all dependencies using helm.

```bash
# create namespace in k8s
$ kubectl create namespace paddlecloud
# install
$ helm install -n paddlecloud test paddlecloud/paddlecloud --set tags.all-dep=true 
```

Check installed components.

```bash
$ kubectl -n paddlecloud get deployments
NAME                         READY   UP-TO-DATE   AVAILABLE   AGE
test-paddlecloud-paddlejob   1/1     1            1           97s
test-paddlecloud-serving     1/1     1            1           97s
test-paddlecloud-sampleset   1/1     1            1           97s

$ kubectl -n paddlecloud get pods
NAMESPACE     NAME                                          READY   STATUS    RESTARTS      AGE
paddlecloud   test-paddlecloud-paddlejob-5c46b5b5dc-gnrlq   1/1     Running   0             30s
paddlecloud   test-paddlecloud-serving-6bc9f77bf6-4njl2     2/2     Running   0             30s
paddlecloud   test-paddlecloud-sampleset-654c876446-bd6x7   1/1     Running   0             30s

$ kubectl -n paddlecloud get replicasets
NAME                                    DESIRED   CURRENT   READY   AGE
test-paddlecloud-paddlejob-5c46b5b5dc   1         1         1       2m39s
test-paddlecloud-sampleset-654c876446   1         1         1       2m39s
test-paddlecloud-serving-6bc9f77bf6     1         1         1       2m39s
```

### 2.2 Custom installation

**Component custom installation**: All components will be installed by default, namely the model training component **paddlejob**, the data caching component **sampleset** and the model inference component **serving**. You can cancel the installation of some components by `--set {$components_name}.enable=false`.

**Dependent charts custom installation**: By default, no dependent charts will be installed. You can specify the parameter `tags.all-dep=true` to install all dependent charts, or use `{$chart_name}.enable=true` to install some dependent charts. The following table shows the provided dependent charts and corresponding versions:

| chart name           | alias     | version |
| -------------------- | --------- | ------- |
| hostpath-provisioner | hostpath  | 0.2.13  |
| juicefs-csi-driver   | juicefs   | 0.8.1   |
| redis                | /         | 16.5.4  |
| jupyterhub           | /         | 1.1.2   |
| kubeflow-pipelines   | pipelines | 0.1.0   |
| knative-serving      | knative   | 0.1.0   |

Helm provides **two ways to set parameters**, one is  using `--set` on the command line, and the other is using yaml files. If you will change many parameters, we recommend the yaml file method, otherwise the `--set` method.

**Example 1. Install paddlejob, sampleSet component and all dependent charts in namespace paddlecloud named test:**

**''--set" method：**

Use `tags.all-dep=true` to install all the dependent charts, and use `serving.enabled=false` to cancel the installation of serving.

```bash
$ kubectl create namespace paddlecloud
$ helm install test paddlecloud/paddlecloud --namespace paddlecloud --set tags.all-dep=true,serving.enabled=false
```

**"yaml file" method：**

Create `values.yaml`, set parameters in type of yaml.

```yaml
tags:
  all-dep: true
serving:
  enabled: false
```

Installation chart.

```bash
$ helm install test paddlecloud/paddlecloud -f valuse.yaml --namespace paddlecloud
```

**Example 2. Install all components in default namespace, and only install redis and kubeflow-pipelines dependent charts:**

**"--set" method:**

Use `pipelines.enabled=true,redis.enabled=true` to install kubeflow-pipelines and redies charts.

```bash
$ helm install test paddlecloud/paddlecloud --set pipelines.enabled=true,redis.enabled=true
```

**"yaml file" method：**

Create `values.yaml`, set parameters in type of yaml.

```yaml
pipelines:
  enabled: true
redis:
  enabled: true
```

Installation chart.

```bash
$ helm install test paddlecloud/paddlecloud -f values.yaml
```

More information about `helm install` can be obtained through the `helm install --help` command.

### 2.3 Install configuration for using volcano

> This step is not required, use this configuration if and only if you want to use volcano to schedule jobs.

Please install Volcano refer to [official tutorial](https://github.com/volcano-sh/volcano) first.

Then create values.yaml with the following contents.

```yaml
paddlejob:
  args:
    - --leader-elect
    - --namespace=paddlecloud  
    - --scheduling=volcano       
```

Parameter Description：

- **namespace**：The namespace here must be the same as the installation namespace.
- **scheduling**：Set volcano to schedule jobs.

Install chart with `values.yaml`.

```bash
$ helm install test paddlecloud/paddlecloud -n paddlecloud -f values.yaml
```

## 3. Paddlejob test

Here is a simple example to check installation. You can refer to [paddlejob tutorial](./Paddlejob) for more detailed instructions and more examples. This example adopts PS mode, uses cpu to train the model. We need to configure PS and worker.

Deploy paddlejob.

```shell
$ kubectl -n paddlecloud apply -f $path_to_PaddleCloud/samples/paddlejob/wide_and_deep.yaml
```

Check pods state.

```shell
kubectl -n paddlecloud get pods
```

Check PaddleJob state.

```shell
kubectl -n paddlecloud get pdj
```

Please click [paddlejob tutorial](./Paddlejob_en.md) get more about paddlejob.

## 4. Uninstall

List the charts installed by helm.

```shell
$ helm list -A
NAME  NAMESPACE  	  REVISION	 UPDATED                                STATUS  	CHART               APP VERSION
test  paddlecloud   1       	2022-04-11 18:08:38.119978 +0800 CST   	deployed	paddlecloud-0.1.0   0.4.0
```

Uninstall charts.

```bash
$ helm uninstall test -n paddlecloud
```

## 5. Advanced usage

More configuration can be found in Makefile, clone this repo and enjoy it.
If you have any questions or concerns about the usage, please do not hesitate to contact us.
