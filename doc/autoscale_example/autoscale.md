# Run Autoscaling job on your local machine

This documentation shows an example to run two jobs on a local kubernetes cluster and see the job scaling status.

## Prepare

- install minikube
- install kubectl

## Run local Autoscaling job

Start a local minikube cluster.

```bash
minikube start --kubernetes-version v1.6.4
```

Run the following commands to create sample training workspace and data.

```bash
# please ensure your workspace directory is mounted in minikube VM.
mkdir /path/to/workspace
cp doc/autoscale_example/*.py /path/to/workspace
mkdir -p /path/to/workspace/data/uci_housing
cd /path/to/workspace && python convert.py
```

Start controller and a example job. Then start another job simulating cluster load, then you can observe the scale process.

```bash
kubectl create -f k8s/controller.yaml
cd k8s/controller/
kubectl create -f autoscale_job/
kubectl get po
kubectl create -f autoscale_load/
kubectl get po
```
