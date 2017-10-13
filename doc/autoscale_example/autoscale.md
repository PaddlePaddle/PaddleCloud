# Run Autoscaling job on your local machine

This documentation shows an example to run two jobs on a local kubernetes cluster and see the job scaling status.

## Prepare

- install minikube
- install kubectl

## Run local Autoscaling job

```bash
minikube start --kubernetes-version v1.6.4
kubectl create -f ../../k8s/controller.yaml
kubectl create -f autoscale_job/
kubectl get po
kubectl create -f autoscale_load/
kubectl get po
```
