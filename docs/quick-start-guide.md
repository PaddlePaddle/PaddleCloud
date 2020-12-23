# Quick Start Guide

## Installation and Configuration

The easiest way to install Paddle Operator is to use Helm chart.

```
# mkdir -p $GOPATH/src/github.com/baidu/
# cd $GOPATH/src/github.com/baidu/
# git clone http://github.com/paddleflow/elastictraining

```
Then modify the configurations in `deployment/elastictraining/values.yaml`, change the kubeconfig/location to point to the upper directory of kubeconfig. 

```
# helm install $GOPATH/src/github.com/paddleflow/elastictraining/deployment/elastictraining --namespace kube-system
# helm list

```

## Running examples

You can have a quick start by using `example/examplejob.yaml`.

```
# kubectl apply -f example/examplejob.yaml
# kubectl get trainingjob examplejob -o=yaml
# kubectl describe trainingjob examplejob

```

You can also try to use new gang-scheduling mode to schedule your trainingjob, first you can follow [this link](https://github.com/kubernetes-sigs/kube-batch/blob/master/doc/usage/tutorial.md) to start up the kube-batch in helm, and then try `example/examplejob_with_kube_batch.yaml`.
