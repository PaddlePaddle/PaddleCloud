# Training Job Instance Co-Scheduling


## Use kube-batch

you need to install kube-batch as your second scheduler if you need co-scheduling. The tutorial for installation is [here](https://github.com/kubernetes-sigs/kube-batch/blob/master/doc/usage/tutorial.md).

The following is an example of how to run paddlepaddle job with kube-batch.


```yaml
apiVersion: paddlepaddle.org/v1
kind: TrainingJob
metadata:
  name: example-job-with-kube-batch
  namespace: default
spec:
  #paddle image with training script
  image: "paddlepaddle/paddlecloud-job"
  # base port of pserver
  port: 7164
  # ports num default 1
  ports_num: 1
  ports_num_for_sparse: 1
  fault_tolerant: true
  schedulerName: kube-batchd
  podGroupName: podgroup-1
  mountPath: "/home/work/namespace/"
  master:
    resources:
      limits:
        cpu: "100m"
        memory: "100Mi"
      requests:
        cpu: "100m"
        memory: "100Mi"
  pserver:
    min-instance: 2
    max-instance: 2
    resources:
      limits:
        cpu: "100m"
        memory: "100Mi"
      requests:
        cpu: "100m"
        memory: "100Mi"
  trainer:
    #entrypoint: "python train.py"
    entrypoint: "sleep 100"
    workspace: "/home/job-1/"
    passes: 10
    # max should equal min while fault_tolerant is disable
    min-instance: 2
    max-instance: 2
    resources:
      limits:
        #alpha.kubernetes.io/nvidia-gpu: 1
        cpu: "100m"
        memory: "100Mi"
      requests:
        cpu: "100m"
        memory: "100Mi"
---
apiVersion: scheduling.incubator.k8s.io/v1alpha1
kind: PodGroup
metadata:
  name: podgroup-1
spec:
  numMember: 5
```
