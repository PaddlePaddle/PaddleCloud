# Data Dispatch

## Local Data

Normally training data can directly be downloaded into container, or one can use Kubernetes PVC to mount the disk into container.

## Integration with Baidu Object Store

Baidu Object Store(BOS) can be used for data and model storage. PaddlePaddle job can use BOS client to access BOS.

The BOS client requires authentication before data get accessed. The client support using Baidu service account for authentication.

The service account should use the necessary IAM roles for accessing BOS. The easiest way for getting the service account is to mount the key file into the Kubernetes secret volume.

The following is an example of how to integrate with BOS.

```

apiVersion: paddlepaddle.org/v1
kind: TrainingJob
metadata:
  name: examplejob
spec:
  image: "paddlepaddle/paddlecloud-job"
  port: 7164
  ports_num: 1
  ports_num_for_sparse: 1
  fault_tolerant: true
  trainer:
    entrypoint: "python /workspace/vgg16_v2.py"
    workspace: "/workspace"
    passes: 50
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
    secrets:
    - name: "bos-bq"
      path: "/mnt/secrets"
      secretType: BOSServiceAccount
    serviceAccount: paddlejob
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
    secrets:
    - name: "bos-bq"
      path: "/mnt/secrets"
      secretType: BOSServiceAccount
    serviceAccount: paddlejob

```