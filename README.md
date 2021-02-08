# Paddle Operator

## Overview

Paddle Operator makes it easy to run [padddle](https://www.paddlepaddle.org.cn/)
distributed training on kubernetes by providing PaddleJob custom resource etc.

## Prerequisites

* Kubernetes >= 1.8
* kubectl

## Quick start

With kubernetes ready, you can install paddle operator with configuration in *deploy* folder (you can get them by clone this repo or just copy them).


Create PaddleJob crd,
```shell
kubectl create -f deploy/v1/crd.yaml
```

Then deploy controller,

```shell
kubectl create -f deploy/v1/operator.yaml
```

> paddle controller configured to run in namespace *paddle-system* and only job in this namespace will be handled by default,
if you prefer other namespace, change related setting in operator.yaml before creation.
Note that the namespace running operator/controller may different from the one your job submit to.

If you got controller ready, for example,
```
$kubectl -n paddle-system get pods
NAME                                         READY   STATUS    RESTARTS   AGE
paddle-controller-manager-698dd7b855-n65jr   1/1     Running   0          1m
```

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

## Advanced usage

More configuration can be found in Makefile, clone this repo and enjoy it. 
If you have any questions or concerns about the usage, please do not hesitate to contact us.


