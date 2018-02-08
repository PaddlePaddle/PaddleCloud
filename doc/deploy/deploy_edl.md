# Deploy EDL on Kubernetes Cluster

[Build EDL and it's Docker image](../build/build_edl_controller.md) first.

Make sure you have `kubectl`
[configured](https://kubernetes.io/docs/tasks/tools/install-kubectl/) properly
before running the below commands:

NOTE: `trainingjob_resource.yaml` is only used when you are using EDL with TPR.

```bash
cd k8s/edl
kubectl create -f trainingjob_resource.yaml
kucectl create -f controller.yaml
```