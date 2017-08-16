# 提交第一个训练任务

---

## 下载并配置paddlecloud

`paddlecloud`是提交PaddlePaddleCloud分布式训练任务的命令行工具。

步骤1: 访问链接 https://github.com/PaddlePaddle/cloud/releases 根据操作系统下载最新的`paddlecloud`
二进制客户端，并把`paddlecloud`拷贝到环境变量$PATH中的路径下，比如：`/usr/local/bin`，然后增加可执行权限：
`chmod +x /usr/local/bin/paddlecloud`

|操作系统|二进制版本|
-- | --
Mac OSX| paddlecloud.dawin
Windows| paddlecloud.exe
Linux | paddlecloud.x86_64

步骤2: 创建`~/.paddle/config`文件(windows系统创建当前用户目录下的`.paddle\config`文件)，并写入下面内容，

```yaml
datacenters:
- name: dlnel
  username: [your user name]
  password: [secret]
  endpoint: http://cloud.dlnel.com
current-datacenter: dlnel
```

配置文件用于指定使用的PaddlePaddleCloud服务器集群的接入地址，并需要配置用户的登录信息：
- name: 自定义的datacenter名称，可以是任意字符串
- username: PaddlePaddleCloud的用户名，账号在未开放注册前需要联系管理员分配，通常用户名为邮箱地址
- password: 账号对应的密码
- endpoint: PaddlePaddleCloud集群API地址，可以从集群管理员处获得
- current-datacenter: 标明使用哪个datacenter作为当前操作的datacenter

配置文件创建完成后，执行`paddlecloud`会显示该客户端的帮助信息：

```
Usage: paddlecloud <flags> <subcommand> <subcommand args>

Subcommands:
	commands         list all command names
	delete           Delete the specify resource.
	file             Simple file operations.
	get              Print resources
	help             describe subcommands and their syntax
	kill             Stop the job. -rm will remove the job from history.
	logs             Print logs of the job.
	registry         Add registry secret on paddlecloud.
	submit           Submit job to PaddlePaddle Cloud.

Subcommands for PFS:
	cp               upload or download files
	ls               List files on PaddlePaddle Cloud
	mkdir            mkdir directoies on PaddlePaddle Cloud
	rm               rm files on PaddlePaddle Cloud


Use "paddlecloud flags" for a list of top-level flags
```

## 下载demo代码并提交运行

完成上面的配置之后，您可以马上提交一个示例的集群训练任务。我们准备了一些样例代码帮助理解集群训练
任务的提交方法，您可以使用下面的命令获取样例代码并提交任务：

这些示例都是基于[paddle book](https://github.com/PaddlePaddle/book)编写的，对应的每个示例
的解释可以参考paddle book。

```bash
mkdir fit_a_line
cd fit_a_line
wget https://raw.githubusercontent.com/PaddlePaddle/cloud/develop/demo/fit_a_line/train.py
cd ..
paddlecloud submit -jobname fit-a-line -cpu 1 -gpu 1 -parallelism 1 -entry "python train.py" fit_a_line/
```

可以看到在提交任务的时候，我们指定了以下参数:
- `-jobname fit-a-line`, 任务名称
- `-cpu 1`, 使用的CPU资源
- `-parallelism 1`, 并行度(训练节点个数)
- `-entry "python train.py"`, 启动命令
- `fit_a_line` 任务程序目录

***说明1：*** 如果希望查看完整的任务提交参数说明，可以执行`paddlecloud submit -h`。

***说明2：*** 每个任务推荐使用不同的jobname提交，这样之前的任务的代码和执行结果都会保存在云端。

## 查看任务运行状态和日志

任务启动之后，可以用过命令`paddlecloud get jobs`查看正在运行的任务：
```bash
paddlecloud get jobs
NUM   NAME         SUCC    FAIL    START                  COMP                   ACTIVE
0     fit-a-line   <nil>   <nil>   2017-06-26T08:41:01Z   <nil>                  1
```

其中， “ACTIVE”表示正在运行的节点个数，“SUCC”表示运行成功的节点个数，“FAIL”表示运行失败的节点个数。

然后，使用下面的命令可以查看正在运行或完成运行任务的日志：

```bash
 paddlecloud logs fit-a-line
Test 28, Cost 13.184950
append file: /pfs/dlnel/public/dataset/uci_housing/train-00000.pickle
append file: /pfs/dlnel/public/dataset/uci_housing/train-00001.pickle
append file: /pfs/dlnel/public/dataset/uci_housing/train-00002.pickle
append file: /pfs/dlnel/public/dataset/uci_housing/train-00003.pickle
append file: /pfs/dlnel/public/dataset/uci_housing/train-00004.pickle
Pass 28, Batch 0, Cost 9.695825
Pass 28, Batch 100, Cost 14.143484
Pass 28, Batch 200, Cost 11.380404
Test 28, Cost 13.184950
...
# logs命令默认返回10条末尾的日志，如果需要查看更多的日志，
# 也可以使用-n参数指定日志的条数
paddlecloud logs -n 100 fit-a-line
...
```

任务执行完成后，任务的状态会显示为如下状态：

```bash
paddlecloud get jobs
NUM   NAME         SUCC   FAIL    START                  COMP                   ACTIVE
0     fit-a-line   1      <nil>   2017-06-26T08:41:01Z   2017-06-26T08:41:29Z   <nil>
```

## 下载任务的模型输出

任务成功执行后，训练程序一般会将模型输出保存在云端文件系统中，可以用下面的命令查看，并下载模型的输出：

```
paddlecloud file ls /pfs/dlnel/home/wuyi05@baidu.com/jobs/fit_a_line/
train.py
image
output
paddlecloud file ls /pfs/dlnel/home/wuyi05@baidu.com/jobs/fit_a_line/output/
pass-0001.tar
...
paddlecloud file get /pfs/dlnel/home/wuyi05@baidu.com/jobs/fit_a_line/output/pass-0001.tar ./
```

模型下载之后，就可以把模型应用在预测服务等环境了。

## 清除任务

使用下面命令可以完全清除集群上的训练任务，清理之后，任务的历史日志将无法查看，但仍然可以在任务名的目录下找到之前的输出。

```back
paddlecloud kill fit-a-line
```

---
详细使用文档见：[中文使用文档](./usage_cn.md)
