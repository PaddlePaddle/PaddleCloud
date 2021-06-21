# hostport-manager

This repository implements a hostport-manager for watching  resources as
defined with a ThirdPartyResources.

## Purpose

This is an hostport manager to alloc/free host port for pods.

## Running


```
http:
hostport-manager --kubernetes=http://xxx/?inClusterConfig=false   --logtostderr=true --v=4  --hostport-range=10000-20000

https:
 ./hostport-manager -kubeconfig=/root/.kube/config -logtostderr=true

TODO
```

## Use

create job:

```bash
kubectl create -f example_job.yaml
# add annotation: hostport-manager/portnum: "2"
```



example_job.yaml
```
apiVersion: paddlepaddle.org/v1
kind: TrainingJob
metadata:
  name: example
  annotations:
    hostport-manager/portnum: "2"
spec:
  image: "paddlepaddle/paddlecloud-job"
  port: 7164
  ports_num: 1
  ports_num_for_sparse: 1
  fault_tolerant: true
  trainer:
    entrypoint: "python train.py"
    workspace: "/home/job-1/"
    passes: 10
    min-instance: 2
    max-instance: 6
    resources:
      limits:
        #alpha.kubernetes.io/nvidia-gpu: 1
        cpu: "200m"
        memory: "200Mi"
      requests:
        cpu: "200m"
        memory: "200Mi"
  pserver:
    min-instance: 2
    max-instance: 2
    resources:
      limits:
        cpu: "800m"
        memory: "1Gi"
      requests:
        cpu: "500m"
        memory: "600Mi"
```


see hostport:
```
kubectl get trainingjobs -o yaml
```

```yaml
apiVersion: v1
items:
- apiVersion: paddlepaddle.org/v1
  kind: TrainingJob
  metadata:
    annotations:
      hostport-manager/hostport: 19935,23149
      hostport-manager/portnum: "2"
    creationTimestamp: 2017-12-04T03:16:56Z
    name: example
    resourceVersion: "135924934"
    selfLink: /apis/paddlepaddle.org/v1/namespaces/yanxu05-baidu-com/trainingjobs/example
    uid: 95110e0e-d8a1-11e7-aa74-6c92bf4727a8
....
```


