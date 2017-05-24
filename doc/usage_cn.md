# 使用命令提交集群训练任务

---

## 下载并配置paddlecloud客户端

首先访问链接 https://github.com/PaddlePaddle/cloud/releases 下载最新的`paddlecloud`二进制客户端，并把`paddlecloud`拷贝到环境变量$PATH中的路径下，比如：`/usr/local/bin`

创建`~/.paddle/config`文件(windows系统创建当前用户目录下的`.paddle\config`文件)，并写入下面内容，

```yaml
datacenters:
- name: datacenter1
  active: true
  username: [your user name]
  password: [secret]
  endpoint: http://cloud.paddlepaddle.org
```

配置文件指定使用的PaddlePaddle Cloud服务器集群的接入地址，并需要配置用户的登录信息：
- name: 自定义的datacenter名称，可以是任意字符串
- active: 为true说明使用这个datacenter作为当前操作的datacenter，配置文件中只能有一个datacenter的配置为true
- username: PaddlePaddle Cloud的用户名，账号在未开放注册前需要联系管理员分配
- password: 账号对应的密码
- endpoint: PaddlePaddle Cloud集群API地址，可以从集群管理员处获得

配置文件创建完成后，执行`paddlecloud`会显示该客户端的帮助信息。

## 准备训练数据

不同的PaddlePaddle Cloud集群环境会提供不同的分布式存储服务。目前PaddlePaddle Cloud支持
HDFS和CephFS。

### 手动上传训练数据到HDFS

您需要使用`ssh`登录到集群中的公用数据中转服务器上，进行数据上传，下载，更新等操作。您可以在
中转服务器的`/mnt`路径下找到集群HDFS的目录，并可以访问当前有权限的目录。上传数据则可以使用管理
普通Linux文件的方式，将数据`scp`到中转服务器`/mnt`下的用户数据目录。比如：

```bash
scp -r my_training_data_dir/ user@tunnel-server:/mnt/hdfs_mulan/idl/idl-dl/mydir/
```

***说明：您可能需要联系集群管理员以获得数据中转服务器的登录地址和权限。***

在训练任务提交后，每个训练节点会把HDFS挂载在`/pfs/[datacenter_name]/home/[username]/`目录下
这样训练程序即可使用这个路径读取训练数据并开始训练。

### 使用paddlecloud命令上传训练数据（TODO）


## 训练程序包

训练程序包是指包含训练程序、依赖、配置文件的一个目录。这个目录必须完整的包含此训练程序的完整依赖，否则可能无法在集群中正常执行。

### 定义集群中的训练数据分发

每个集群训练任务都会在集群的多个节点上启动训练程序实例，每个训练程序实例会完成一部分训练数据的训练任务。为了能比较均匀的将大量训练数据分配在这些节点上，在编写训练程序时，通常会使用下面的方法：

```python
TRAIN_FILE_PATTERN = "/pfs/[datacenter_name]/home/[username]/user001_space/part-*.gz"
TEST_FILE_PATTERN  = "/pfs/[datacenter_name]/home/[username]/user001_space_test/part-*.gz"

def gen_train_list(data_dir):
    # write "/train.list" and "/test.list" for each node
    trainer_id = -1
    trainer_count = -1
    with open("/trainer_id", "r") as f:
        trainer_id = int(f.readline()[:-1])
    with open("/trainer_count", "r") as f:
        trainer_count = int(f.readline()[:-1])

    train_file_list = glob.glob(TRAIN_FILE_PATTERN)
    train_file_list.sort()
    my_train_list = []
    for idx, f in enumerate(train_file_list):
        if idx % trainer_count == trainer_id:
            my_train_list.append(f)
    with open("/train.list", "w") as f:
        f.write('\n'.join(my_train_list))

    test_file_list = glob.glob(TEST_FILE_PATTERN)
    test_file_list.sort()
    my_test_list = []
    for idx, f in enumerate(test_file_list):
        if idx % trainer_count == trainer_id:
            my_test_list.append(f)
    with open("/test.list", "w") as f:
        f.write('\n'.join(my_test_list))

```

这段代码会根据指定的HDFS中的训练数据路径，将文件顺序的分配给每个节点，并生成两个文件`/train.list`和`/test.list`保存分配给当前节点的训练数据文件的路径。在调用`define_py_data_sources2`定义训练数据时，传入这两个文件路径即可。

### 上传训练程序包到HDFS

上传训练程序包到HDFS的方式和上传训练数
据方式相同。您需要使用公用数据中转服务器，将训练程序包上传到HDFS。比如：

```bash
scp -r my_training_package/ user@tunnel-server:/mnt/hdfs_mulan/idl/idl-dl/mypackage/
```

在下面提交任务的步骤中，需要指定集群上的程序包的位置：
`/pfs/[datacenter_name]/home/[username]/idl/idl-dl/mypackage/`即可在集群中执行这个
程序包中的训练程序。

***注意: 此方式会逐步淘汰***


## 提交任务

## 查看任务状态

用户可以查看任务、任务节点、用户空间配额的当前状态。

执行`paddlecloud get jobs`命令查看任务运行状态，将显示：

```
NUM	NAME	SUCC	FAIL	START	COMP	ACTIVE
0	paddle-cluster-job	<nil>	1	2017-05-24T07:52:45Z	<nil>	1
```

执行`paddlecloud get workers paddle-cluster-job`查看worker运行状态，显示：

```
NAME	STATUS	START
paddle-cluster-job-trainer-3s4nz	Running	2017-05-24T07:53:41Z
paddle-cluster-job-trainer-6sc4q	Running	2017-05-24T07:53:03Z
...
```

## 查看任务日志

执行`paddlecloud logs paddle-cluster-job`显示当前任务的所有worker的日志：

```
```

如果任务workers较多，可以指定查看某个worker的单独的日志：`paddlecloud logs -w paddle-cluster-job-trainer-3s4nz paddle-cluster-job`。

## 终止任务
执行`paddlecloud kill paddle-cluster-job`即可停止训练任务的所有节点和进程。

kill命令执行成功后，集群会在后台终止集群作业的workers进程，workers并不会在kill命令之后全部停止。如果需要查看kill掉的任务正在清理的workers，可以使用命令`paddlecloud get workers paddle-cluster-job`查看。

***所以在kill之后提交新的任务时一定要更改提交时的`-name`参数，防止任务名称冲突。***

## 性能优化(TODO)
