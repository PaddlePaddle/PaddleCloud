## Docker 基本使用

使用前请确保您已经安装 [Docker](https://docs.docker.com/desktop/)，如果想要使用 GPU 版本镜像，还需要安装 [nvidia-docker](https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/install-guide.html#supported-platforms)。

### 下载镜像

选择需要的[镜像版本](https://hub.docker.com/repositories)，使用 docker pull 命令下载镜像，这里以 PaddleOCR 5月9日的镜像为例：

```
docker pull paddlecloud/paddleocr:2.5-cpu-efbb0a
```

### 启动容器

CPU 版本

```
docker run --name dev -v $PWD:/paddle -p 8888:8888 -it paddlecloud/paddleocr:2.5-cpu-efbb0a /bin/bash
```

GPU 版本

```
docker run --name dev --runtime=nvidia -v $PWD:/paddle -p 8888:8888 -it paddlecloud/paddleocr:2.5-gpu-cuda10.2-cudnn7-efbb0a /bin/bash
```

参数说明：

- --name dev：设定 container 的名称为 dev，dev 可随意替换。
- --v $PWD:/paddle：$PWD 是当前执行程序的路径（可通过 echo $PWD 查看），/paddle 是启动容器内的路径，此条命令的含义是把本地 $PWD 目录挂载到容器 /paddle 目录。例如你想要使用的代码目录是 /home/ubuntu/code，想要把它挂载到容器内的 /paddle 位置，即加上命令 `-v /home/ubuntu/code:/paddle`
- -p 8888:8888：将容器的 8888 端口发布到主机 8888 端口。
- -it：开启容器和本机的交互式运行
- --rm：退出容器后，自动删除容器，只有当您仅想试用一次时，加上此参数
- --runtime=nvidia：表示使用 nvidia-docker 环境，及挂载相应的 cuda 环境和 GPU。

注：一般情况下，推荐将代码和数据集保存在本地，通过 -v 挂载到容器内，后续即使将 container 删除，代码和数据集不会受到影响。

### 查看运行中的容器

```
docker container ls
```

### 进入容器

```
docker exec -it [container_name] /bin/bash
```

这里的 name 在上述例子中，即 `--name dev` 参数设置的 dev，也可以通过`docker container ls`查看

### 关闭容器

- 从容器中退出，回到本机 shell：

```
ctrl + d
```

- 停止容器，可通过 `docker start [container_name]` 再次开启容器

```
docker stop [container_name]
```

- 删除容器，将容器在本地删除

```
docker rm [container_name]
```

