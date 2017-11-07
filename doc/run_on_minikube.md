# Run PaddleCloud on your local machine

This documentation shows how to run PaddleCloud on minikube.   
(This surports only mac now.On unbuntu-16.04,we met a bug of kubernetes when we use `hostpath` to volume a file.)

## Prerequisites

- [install minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
- [install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

1. Start a local minikube cluster:

    ```bash
    minikube start --kubernetes-version v1.6.4
    ```
    
    If you can't connect to minikube distribution server,add https_proxy like that:
    
    ```bash
    https_proxy=https://YOURPROXY:PORT minikube start --kubernetes-version v1.6.4
    ```
1. Enable ingress addon:

 	```
 	minikube addons enable ingress
 	```
 	
1. Create workspace directory:

	```
	mkdir <yourpath>
	```  
	Since Minikube mounts `$Home` path by default, we recommend creating the path under `$Home` which offers the flexibility of switching between directories in your deployment without stopping the MiniKube and mounting another one.
	
1. Copy `ca` and generate `admin` certificate：    
	(We must use `ca` under `~/.minikube` rather than `~/.minikube/certs`.)
	
	```
	mkdir <yourpath>/certs && cd <yourpath>/certs
	openssl genrsa -out admin-key.pem 2048
	openssl req -new -key admin-key.pem -out admin.csr -subj "/CN=kube-admin"
	openssl x509 -req -in admin.csr -CA ~/.minikube/ca.crt -CAkey ~/.minikube/ca.key \
  		-CAcreateserial -out admin.pem -days 365
	cp ~/.minikube/ca.crt .
	cp ~/.minikube/ca.key .		
	```
	
1. Copy and update paddlecloud configurations::

	```
	git clone https://github.com/PaddlePaddle/cloud 
	cp cloud/k8s/minikube/* <yourpath>/
	sed -i'.bak' -e "s#<yourpath>#yourpath#g"  <yourpath>/*.yaml
	```

1. Edit `/etc/hosts` and add `$(minikube ip) cloud.testpcloud.org` to it.
1. Start all jobs:
 
	```
	kubectl create -f cloud_ingress.yaml
	kubectl create -f cloud_service.yaml
	kubectl create -f mysql_deployment.yaml
	kubectl create -f mysql_service.yaml
	kubectl create -f pfs_deployment.yaml
	kubectl create -f pfs_service.yaml
	kubectl create -f cloud_deployment.yaml
	```
1. open `cloud.testpcloud.org` in your browser and sign up a user.
1. Edit `~/.paddle/config` like this:

```
datacenters:
- name: testpcloud
  username: <username>
  password: <password>
  endpoint: http://cloud.testpcloud.org
current-datacenter: testpcloud
```

You can use PaddleCloud command line now.


## FAQ
1. We can't get anything when we open `cloud.testpcloud.org` in browser.  
   One possible cause is: `default-http-backend` is not ready yet. Run `minikube get po --all-namespaces` to check its status.
If it's stuck at pulling the image, one alternative is to manually download the image and load it with minikube's docker.
  - run `kubectrl describe po name --namespace=kube-system` to get docker image uri.
  - `docker save <docker-image-uri> > <tarname>.tar` to save the image to a tar file. You may need to use proxy with this step.
  - `minikube ssh` to login to Minikube's command line.
  - `docker load < <tarname>.tar` to load the image to Kubenetes' local docker image repo.
  
1. I edited a file in host file system, but it remained unchanged in Minikube virtual machine sometimes.  
    That might be due to the files cache mechanism of Minikube. You can try to restart the Minikube with `minikube stop` `minikube start --kubernetes-version v1.6.4` to fix it.

## TODO	
1. The `mysql` docker runs `mysqld` under user `mysql` instead of `root`,so it's difficult to save `mysql` data to hostpath.
1. Fix bug: `Kubernetes can't mount the volume with hostpath to a single file` on linux. 	
	
	