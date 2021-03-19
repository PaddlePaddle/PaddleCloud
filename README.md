# Paddle Operator

## Overview

Paddle Operator makes it easy to run [padddle](https://www.paddlepaddle.org.cn/)
distributed training job on kubernetes by providing PaddleJob custom resource etc.

## Prerequisites

* Kubernetes >= 1.8
* kubectl

## Installation

With kubernetes ready, you can install paddle operator with configuration in *deploy* folder 
(use *deploy/v1* for kubernetes v1.16+ or *deploy/v1beta1* for kubernetes 1.15-).

Create PaddleJob crd,
```shell
kubectl create -f deploy/v1/crd.yaml
```

A succeed creation leads to result as follows,
```
$ kubectl get crd
NAME                                    CREATED AT
paddlejobs.batch.paddlepaddle.org       2021-02-08T07:43:24Z
```

Then deploy controller,

```shell
kubectl create -f deploy/v1/operator.yaml
```

the ready state of controller would be as follow,
```
$kubectl -n paddle-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
paddle-controller-manager-698dd7b855-n65jr   1/1     Running   0          1m
```

> paddle controller configured above to run in namespace *paddle-system* and only job in this namespace will be handled by default,
if you prefer other namespace, change related setting in operator.yaml before creation.
Note that the namespace running operator/controller may different from the one your job submit to.

## Quick start

Deploy your first paddlejob demo with
```shell
kubectl -n paddle-system create -f deploy/examples/wide_and_deep.yaml
```

Check pods status by
```shell
kubectl -n paddle-system get pods
```

especially,
```shell
kubectl -n paddle-system get pdj
```
may give you abstract summary of your job.

Fin, you can play with your own job.

## Uninstall
Simply
```shell
kubectl delete -f deploy/v1/crd.yaml -f deploy/v1/operator.yaml
```
## Advanced usage

More configuration can be found in Makefile, clone this repo and enjoy it. 
If you have any questions or concerns about the usage, please do not hesitate to contact us.

## More Information

Please refer to the
[中文文档](https://fleet-x.readthedocs.io/en/latest/paddle_fleet_rst/paddle_on_k8s.html) 
for more information about paddle configuration.



