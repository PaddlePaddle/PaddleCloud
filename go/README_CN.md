1. 如何构建PFSServer的DockerImage
  - 构建PFSServer的编译环境

  ```
  cd docker
  bash build.sh
  ```

  - 编译PFSServer
 
  ```
  docker run  --rm -v  $(pwd):/root/gopath/src/github.com/PaddlePaddle/cloud/go pfsserver:dev
  ```
  
  - 构建PFSServer的DockerImage
  
  ```
  docker build . -t pfsserver:latest
  ```
  - PFSServer启动命令
  
  ```
  docker run pfsserver:latest /pfsserver/pfsserver -tokenuri http://cloud.paddlepaddle.org -logtostderr=false -log_dir=./log -v=3
  ```

2. 如何部署PFSServer

	```
	cd ../k8s
	kuberctl create -f cloud_pfsserver.yaml
	```
	 
3. 如何使用PFSClient
	- cp
	
	```
	upload:
		paddlecloud cp ./file /pfs/$DATACENTER/home/$USER/file
		
	download:
		paddlecloud cp /pfs/$DATACENTER/home/$USER/file ./file
	```
	- ls
	
	```
	paddlecloud ls  /pfs/$DATACENTER/home/$USER/folder
	```
	
	- rm
	
	```
	paddlecloud rm /pfs/$DATACENTER/home/$USER/file
	paddlecloud rm -r /pfs/$DATACENTER/home/$USER/folder
	```
	
	- mkdir
	
	```
	paddlecloud mkdir  /pfs/$DATACENTER/home/$USER/folder
	```
