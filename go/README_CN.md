1. 如何构建PFSserver的DockerImage
  - 构建PFSserver的编译环境

  ```
  cd docker
  bash build.sh
  ```

  - 编译PFSserver
 
  ```
  cd ..
  docker run  --rm -v  $(pwd):/root/gopath/src/github.com/PaddlePaddle/cloud  PFSserver:dev
  ```
  
  - 创建PFSserver的DockerImage
  
  ```
  docker build . -t pfsserver:latest
  ```

2. 如何部署PFSserver
3. 如何使用PFSclient
	- cp
	- ls
	- rm
	- mkdir
