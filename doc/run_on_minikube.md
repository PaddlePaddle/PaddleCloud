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
    https_proxy=<your's> minikube start --kubernetes-version v1.6.4
    ```
1. Enable ingress addon:

 	```
 	minikube addons enable ingress
 	```
 	
1. Create workspace directory:

	```
	mkdir <yourpath>
	```  
	Minikube mount `$Home` path default,and you'd better to create path under it or you may find that it can't be changed in minikube virtual mathine or in kubertes pod.
	
1. Copy `ca` and generate `admin` certificate：    
	(We must use `ca` under `~/.minikube` instead of under `~/.minikube/certs`.)
	
	```
	mkdir <yourpath>/certs && cd <yourpath>/certs
	openssl genrsa -out admin-key.pem 2048
	openssl req -new -key admin-key.pem -out admin.csr -subj "/CN=kube-admin"
	openssl x509 -req -in admin.csr -CA ~/.minikube/ca.crt -CAkey ~/.minikube/ca.key \
  		-CAcreateserial -out admin.pem -days 365
	cp ~/.minikube/ca.crt .
	cp ~/.minikube/ca.key .		
	```
	
1. Copy and configure paddlecloud configures:

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
1. Type `$(minikube ip)` into browser,and sign up a user.
1. Edit `~/.paddle/config` like this:

```
datacenters:
- name: testpcloud
  username: g1@163.com
  password: 1
  endpoint: http://cloud.testpcloud.org
current-datacenter: testpcloud
```

You can use PaddleCloud command line now.


## FAQ
1. We can't get anything when we type `$(minikube ip)` into browser.
You can use `minikube get po --all-namespaces` to check it,it should like:

```
NAMESPACE     NAME                                         READY     STATUS    RESTARTS   AGE
default       paddle-cloud-3328278932-pqtf0                1/1       Running   5          3h
default       paddle-cloud-mysql-2598050360-rdzpr          1/1       Running   4          3h
g1-163-com    cloud-notebook-deployment-2982752518-6gr0m   0/1       Pending   0          3h
kube-system   default-http-backend-cwqd4                   1/1       Running   4          4h
kube-system   kube-addon-manager-minikube                  1/1       Running   4          4h
kube-system   kube-dns-4149932207-0clkw                    3/3       Running   12         4h
kube-system   kubernetes-dashboard-tqrd7                   1/1       Running   4          4h
kube-system   nginx-ingress-controller-gxgl8               1/1       Running   4          4h
```

But docker images such as `default-http-backend` can't be downloaded sometimes , you can use 

  -  `kubectrl describe po name --namespace=kube-system` to get docker image uri.
  -  and then pull with proxy and then save to tar `docker save <docker-image-uri> > <tarname>.tar`.
  - `minikube ssh`.
  - `docker load < <tarname>.tar`.

## TODO	
1. The `mysql` docker runs `mysqld` under user `mysql` instead of `root`,so it's difficult to save `mysql` data to hostpath.
1. Fix bug: `kubernetes can't mount a file use host to volume a file` on linux. 
	
	
	