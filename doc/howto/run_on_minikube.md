# Run PaddlePaddle Cloud with minikube

This documentation explains how to run PaddlePaddle Cloud on the development computer using [minikube](https://kubernetes.io/docs/getting-started-guides/minikube/).

## Steps

1. Install [minikube and kubectl](https://kubernetes.io/docs/tasks/tools/install-minikube/).

1. Start a local minikube cluster. The reason we start a 1.6.4 in this case is paddle cloud currently has some dependency on [TPR](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-third-party-resource/), which will be deprecated from 1.7.

    ```bash
    minikube start --kubernetes-version v1.6.4
    ```
    
    If you can't connect to minikube distribution server,add https_proxy like that:
    
    ```bash
    https_proxy=YOURPROXY:PORT minikube start --kubernetes-version v1.6.4
    ```
1. Enable ingress addon:

 	```
 	minikube addons enable ingress
 	```
 	
1. Create workspace directory:

	```
	mkdir $HOME/workspace
	```  
	Mount this directory to `/workspce` 
	```
	minikube mount $HOME/workspace:/workspace
	```
	
1. Copy `ca` and generate `admin` certificate：    
	(We must use `ca` under `~/.minikube` rather than `~/.minikube/certs`.)
	
	```
	mkdir /workspace/certs && cd /workspace/certs
	openssl genrsa -out admin-key.pem 2048
	openssl req -new -key admin-key.pem -out admin.csr -subj "/CN=kube-admin"
	openssl x509 -req -in admin.csr -CA ~/.minikube/ca.crt -CAkey ~/.minikube/ca.key \
  		-CAcreateserial -out admin.pem -days 365
	cp ~/.minikube/ca.crt .
	cp ~/.minikube/ca.key .		
	```
	
1. Copy and update PaddlePaddle Cloud configurations:

	```
	git clone https://github.com/PaddlePaddle/cloud 
	cp cloud/k8s/minikube/* $HOME/workspace/
	sed -i'.bak' -e "s#<yourpath>#yourpath#g"  $HOME/workspace/*.yaml
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
1. To interact with PaddlePaddle Cloud instance, you need to use paddlecloud command line tool. Edit command line config file from `~/.paddle/config` like this:

```
datacenters:
- name: testpcloud
  username: <username>
  password: <password>
  endpoint: http://cloud.testpcloud.org
current-datacenter: testpcloud
```

Now PaddlePaddle Cloud command line is ready to use.

For further usage, please refer to the doc from [here](https://github.com/PaddlePaddle/cloud/blob/develop/doc/usage_cn.md) (in Chinese), and [here](https://github.com/PaddlePaddle/cloud/blob/develop/doc/usage_en.md) (comming soon)


## FAQ
1. There is nothing when open `cloud.testpcloud.org` in browser.  
   One possible cause is: `default-http-backend` is not ready yet. Run `minikube get po --all-namespaces` to check its status.
If it's stuck at pulling the image, one alternative is to manually download the image and load it with minikube's docker.
  - run `kubectrl describe po name --namespace=kube-system` to get docker image uri.
  - `docker save <docker-image-uri> > <tarname>.tar` to save the image to a tar file. You may need to use proxy with this step.
  - `minikube ssh` to login to Minikube's command line.
  - `docker load < <tarname>.tar` to load the image to Kubenetes' local docker image repo.
  
1. I edited a file in host file system, but it remained unchanged in Minikube virtual machine sometimes.  
    That might be due to the files cache mechanism of Minikube. You can try to restart the Minikube with `minikube stop` `minikube start --kubernetes-version v1.6.4` to fix it.

## TODO	

1. Currently the `mysql` docker runs `mysqld` under user `mysql` instead of `root`, which makes it difficult to save `mysql` data to hostpath.

1. Update TPR to CRD to be able to run on latest kubernetes
	
