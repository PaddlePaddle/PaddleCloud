- [macOS 安装 microk8s](#macos-安装-microk8s)
  - [安装 microk8s](#安装-microk8s)
    - [安装虚拟机和 microk8s 服务：](#安装虚拟机和-microk8s-服务)
      - [官方方法](#官方方法)
      - [multipass方法（建议）：](#multipass方法建议)
    - [检查安装情况](#检查安装情况)
  - [解决安装失败问题](#解决安装失败问题)
    - [查看问题](#查看问题)
    - [处理缺失的image](#处理缺失的image)
    - [打开 dns 服务](#打开-dns-服务)
    - [查看是否成功](#查看是否成功)
# macOS 安装 microk8s

## 安装 microk8s

### 安装虚拟机和 microk8s 服务：

#### 官方方法

ref：https://microk8s.io/docs/install-alternatives#heading--macos

```bash
$ brew install ubuntu/microk8s/microk8s
$ microk8s install
$ microk8s enable dns
```

#### multipass方法（建议）：

```bash
$ brew install ubuntu/microk8s/microk8s
$ brew cask install multipass
# 创建虚拟机，名字一定要是这个，不然 microk8s 的命令不能用，--mem 建议 >= 6G, --cpus 建议 > 1
$ multipass launch --name microk8s-vm --mem 8G --disk 40G --cpus 4
```

在虚拟机内安装 microk8s

```bash
$ multipass shell microk8s-vm
######## 根据网络情况选择镜像源
$ sudo sed -i s/archive.ubuntu.com/mirrors.aliyun.com/g /etc/apt/sources.list
$ sudo sed -i s/security.ubuntu.com/mirrors.aliyun.com/g /etc/apt/sources.list
$ sudo apt-get update
########
$ sudo snap install microk8s --classic --channel=1.21/stable		# 建议选择 kubernates v1.21
```

更改默认用户权限

```bash
# 默认 ubuntu 账号无权限操作集群，均需要 sudo
# 可将 ubuntu 账号加入 microk8s 用户组以便简化访问
$ sudo usermod -a -G microk8s ubuntu
$ sudo chown -f -R ubuntu ~/.kube
```

将kubernates 的 config 文件写入本地（在本机终端执行）

```bash
$ multipass exec microk8s-vm -- /snap/bin/microk8s.config > ~/.kube/config
```

完成这步后，可直接在本机使用 kubectl, helm等命令访问集群。

如果有多个 k8s 集群，可通过 `kubectl --kubeconfig="$path_to_config"` 的形式指定访问某一集群

### 检查安装情况

使用 `microk8s kubectl get pods -A`查看 pods 运行情况，如果下面这种情况，恭喜您，microk8s 安装成功

```
NAMESPACE     NAME                                       READY   STATUS    RESTARTS       AGE
kube-system   calico-kube-controllers-6fdcdb4bb8-msg2r   1/1     Running   0              10s
kube-system   calico-node-rdmgf                          1/1     Running   0              10s
```

如果发现`calico-node-*****` 的 STATUS 卡在了 `Init:0/3` 状态，则按照教程继续安装

## 解决安装失败问题 

### 查看问题

查看 calico-node-rdmgf pod 信息（注意 *calico-node-rdmgf* 中的后5位为随机字符，请自行更换）

```bash
$ microk8s -n kube-system kubectl describe pod calico-node-rdmgf 			
```

查看报错信息，发现卡在了`pull "k8s.gcr.io/pause:3.1"`这一步，原因是国内网络连接不到此镜像网站，接下来手动将镜像下载到本地

### 处理缺失的image

这里先使用主机将国内镜像pull下来，再传到虚拟机。

[常用镜像仓库](https://gist.github.com/qwfys/aec4d2ab79281aeafebdb40b22d0b748)

```bash
# pull 国内镜像
$ docker pull mirrorgooglecontainers/pause:3.1
# registry.cn-beijing.aliyuncs.com/zhoujun/pause:3.1
# k8s以image的名字为label，因此将名字改为原始名字
$ docker tag mirrorgooglecontainers/pause:3.1 k8s.gcr.io/pause:3.1	
$ docker save k8s.gcr.io/pause:3.1 > pause.tar
# 将image从本机发送到k8s，或者通过mount，共享存储空间
$ multipass transfer pause.tar microk8s-vm:
# import image
$ microk8s ctr image import pause.tar							# microk8s v1.18+
$ microk8s.ctr -n k8s.io image import pause.tar		# microk8s v1.17-
```

更一般性的写法：

```bash
$ docker pull mirrorgooglecontainers/$imageName:$imageVersion
$ docker tag  mirrorgooglecontainers/$imageName:$imageVersion k8s.gcr.io/$imageName:$imageVersion
$ docker save k8s.gcr.io/$imageName:$imageVersion > $imageName.tar
$ multipass transfer $source $destination
$ microk8s ctr image import $imageName.tar							# docker v1.18+
$ microk8s.ctr -n k8s.io image import $imageName.tar		# docker v1.17-
```

一般下载过这个镜像后，其他的都可以 pull 成功，如有不成功的，可用同样的方法处理

### 打开 dns 服务

```
microk8s enable dns
```

### 查看是否成功

```bash
$ microk8s kubectl get pods -A
NAMESPACE     NAME                                       READY   STATUS    RESTARTS       AGE
kube-system   calico-kube-controllers-6fdcdb4bb8-msg2r   1/1     Running   0              10s
kube-system   calico-node-rdmgf                          1/1     Running   0              10s

$ microk8s status
microk8s is running
...
...
...
```

