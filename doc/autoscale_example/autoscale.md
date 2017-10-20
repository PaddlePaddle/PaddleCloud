# Run Autoscaling job on your local machine

This documentation shows an example to run two jobs on a local kubernetes cluster and see the job scaling status.

## Prerequisites

- [install minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
- [install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

## Run local Autoscaling job

1. Start a local minikube cluster.

  ```bash
  minikube start --kubernetes-version v1.6.4
  ```

1. Run the following commands to create sample training workspace and
data.

  ```bash
  mkdir /path/to/workspace
  cp $REPO_PATH/doc/autoscale_example/*.py /path/to/workspace
  mkdir -p /path/to/workspace/data/
  cp -r $REPO_PATH/doc/autoscale_example/uci_housing/ /path/to/workspace/data/
  ```

1. Mount the workspace folder into Minikube:

  ```bash
  minikube mount /path/to/workspace:/workspace
  ```

  The `minikube mount` command will block, so start a new terminal to
  continue the tutorial.

1. Start controller and a example job:

  ```bash
  cd $REPO_PATH/k8s/controller
  kubectl create -f controller.yaml
  kubectl create -f trainingjob_resource.yaml
  kubectl create -f autoscale_job/
  kubectl get pods
  ```

1. Start another job simulating cluster load, then you can observe the
scale process using `kubectl get pods`:

  ```bash
  kubectl create -f autoscale_load/
  kubectl get pods
  ```
