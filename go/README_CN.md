1. 如何创建pfsserver的DockerImage
  - 创建pfsserver的编译环境

  ```
  cd docker
  bash build.sh
  ```

  - 编译pfsserver
 
  ```
  cd ..
  docker run  --rm -v  $(pwd):/root/gopath/src/github.com/PaddlePaddle/cloud/go  pfsserver:dev
  ```
  
  - 创建pfsserver的DockerImage
  
  ```
  docker build . -t pfsserver:latest
  ```
