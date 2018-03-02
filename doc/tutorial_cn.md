# 提交第一个训练任务

---

## 下载并配置paddlectl

`paddlectl`是提交PaddlePaddleCloud分布式训练任务的命令行工具。

- 步骤1

  下载`paddlectl`客户端，并将二进制文件拷贝到环境变量 `$PATH` 所指向的目录下，例如
  `/usr/local/bin`, 然后使用命令 `chmod +x /usr/local/bin/paddlectl` 赋予
  可执行权限。

  我们推荐用户优先从[Release Page](https://github.com/PaddlePaddle/cloud/releases)中下载最新
  版本的客户端，同时您也可以在我们CI系统上下载develop分支最新编译的二进制文件。

  操作系统 | 下载链接
  -- | --
  Mac OSX| [paddlecloud.darwin](http://guest:@paddleci.ngrok.io/repository/download/PaddleCloud_Client/.lastSuccessful/paddlecloud.darwin)
  Windows| [paddlecloud.exe](http://guest:@paddleci.ngrok.io/repository/download/PaddleCloud_Client/.lastSuccessful/paddlecloud.exe)
  Linux | [paddlecloud.x86_64](http://guest:@paddleci.ngrok.io/repository/download/PaddleCloud_Client/.lastSuccessful/paddlecloud.x86_64)

- 步骤2
  创建`~/.paddle/config`文件(windows系统创建当前用户目录下的`.paddle\config`文件)，下面是示例文件：

  ```bash
  datacenters:
  - name: production
    username: paddlepaddle
    password: paddlecloud
    endpoint: http://production.paddlecloud.com
  - name: experimentation
    username: paddlepaddle
    password: paddlecloud
    endpoint: http://experimentation.paddlecloud.com
  current-datacenter: production
  ```

  假设您有两个data-center的权限，其中一个是 `production`, 另外一个是 `experimentation`,
  您可以通过 `current-datacenter` 字段来指定当前使用哪以为data-center.

  `username`, `password` 和 `endpoint` 字段是您在data-center中的认证信息，这些信息会包含在
  管理员发给您的邮件中。

以上两步骤完成后，执行 `paddelctl` 可以打印出帮助信息：

```bash
> paddlectl
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

## 下载Demo代码并尝试提交任务

我们准备了一些样例代码帮助理解集群训练任务的提交方法，这些示例代码都是基于
[Paddle Book](https://github.com/PaddlePaddle/book) 编写的， 您可以在每个章节中找到说明文档。

您可以使用下面命令获取Demo代码并提交任务：

```bash
> mkdir fit_a_line
> cd fit_a_line
> wget https://raw.githubusercontent.com/PaddlePaddle/cloud/develop/demo/fit_a_line/train.py
> cd ..
> paddlecloud submit -jobname fit-a-line -cpu 1 -gpu 1 -parallelism 1 -entry "python train.py train" fit_a_line/
```

参数说明：

- `-jobname`, STRING，任务名称，您需要指定一个唯一的任务名字。
- `-cpu`, INT，Trainer节点使用的CPU核心数。
- `-gpu 1`, INT，使用的GPU资源（卡数），若集群无GPU资源，可以去掉这个配置
- `-parallelism`, INT, 并行度， 训练节点的个数。
- `-entry`, STRING, 训练脚本的启动命令。
- `fit_a_line` STRING, 任务程序目录。

***说明1：*** 如果希望查看完整的任务提交参数说明，可以执行`paddlecloud submit -h`。

***说明2：*** 每个任务推荐使用不同的jobname提交，这样之前的任务的代码和执行结果都会保存在云端。

***说明3：*** 如果提交的节点比较多，可以先修改启动命令中的参数`-entry "python train.py prepare"`，先将数据集下载到PFS上，再提交训练任务。

## 查看任务运行状态和日志

任务启动之后，可以用过命令`paddlecloud get jobs`查看正在运行的任务：

```bash
> paddlecloud get jobs
NUM   NAME         SUCC    FAIL    START                  COMP                   ACTIVE
0     fit-a-line   <nil>   <nil>   2017-06-26T08:41:01Z   <nil>                  1
```

- `ACTIVE`, 正在运行的节点数量。
- `SUCC`，运行结束并成功的节点数量。
- `FAIL`，运行失败的节点数量。

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

```bash
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

```bash
paddlecloud kill fit-a-line
```

---
详细使用文档见：[中文使用文档](./usage_cn.md)
