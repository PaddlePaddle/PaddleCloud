# 使用命令提交集群训练任务

---

## 下载并配置paddlecloud

`paddlecloud`是提交PaddlePaddleCloud分布式训练任务的命令行工具。

步骤1: 访问链接 https://github.com/PaddlePaddle/cloud/releases 下载最新的`paddlecloud`二进制客户端，并把`paddlecloud`拷贝到环境变量$PATH中的路径下，比如：`/usr/local/bin`

步骤2: 创建`~/.paddle/config`文件(windows系统创建当前用户目录下的`.paddle\config`文件)，并写入下面内容，

```yaml
datacenters:
- name: datacenter1
  username: [your user name]
  password: [secret]
  endpoint: http://cloud.paddlepaddle.org
current-datacenter: datacenter1
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
	get              Print resources
	help             describe subcommands and their syntax
	kill             Stop the job. -rm will remove the job from history.
	logs             Print logs of the job.
	submit           Submit job to PaddlePaddle Cloud.


Use "paddlecloud flags" for a list of top-level flags
```

## 准备训练数据

不同的PaddlePaddleCloud集群环境会提供不同的分布式存储服务。目前PaddlePaddleCloud支持HDFS和CephFS。

### 手动上传训练数据到HDFS

使用`ssh`登录到集群中的公用数据中转服务器上，进行数据上传，下载，更新等操作。您可以在中转服务器的`/mnt`路径下找到集群HDFS的目录，并可以访问当前有权限的目录。上传数据则可以使用管理普通Linux文件的方式，将数据`scp`到中转服务器`/mnt`下的用户数据目录。比如：

```bash
scp -r my_training_data_dir/ user@tunnel-server:/mnt/hdfs_mulan/idl/idl-dl/mydir/
```

***说明：您可能需要联系集群管理员以获得数据中转服务器的登录地址和权限。***

在训练任务提交后，每个训练节点会把HDFS挂载在`/pfs/[datacenter_name]/home/[username]/`目录下这样训练程序即可使用这个路径读取训练数据并开始训练。


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

上传训练程序包到HDFS的方式和上传训练数据方式相同。使用公用数据中转服务器，将训练程序包上传到HDFS。比如：

```bash
scp -r my_training_package/ user@tunnel-server:/mnt/hdfs_mulan/idl/idl-dl/mypackage/
```

在下面提交任务的步骤中，需要指定集群上的程序包的位置：`/pfs/[datacenter_name]/home/[username]/idl/idl-dl/mypackage/`即可在集群中执行这个程序包中的训练程序。

***注意: 此方式会逐步淘汰***


## 提交任务

执行下面的命令提交准备好的任务:

- 提交基于V1 API的训练任务

```bash
paddlecloud submit -jobname my-paddlecloud-job -cpu 1 -gpu 0 -memory 1Gi -parallelism 10 -pscpu 1 -pservers 3 -psmemory 1Gi -passes 1 -topology trainer_config.py /pfs/[datacenter_name]/home/[username]/ctr_demo_package
```

- 提交基于V2 API的训练任务

```bash
paddlecloud submit -jobname my-paddlecloud-job -cpu 1 -gpu 0 -memory 1Gi -parallelism 10 -pscpu 1 -pservers 3 -psmemory 1Gi -passes 1 -entry "python trainer_config.py" /pfs/[datacenter_name]/home/[username]/ctr_demo_package
```

参数说明：
- `jobname`：提交任务的名称，paddlecloud使用`jobname`唯一标识一个任务
- `-cpu`：每个trainer进程使用的CPU资源，单位是“核”
- `-gpu`：每个trainer进程使用的GPU资源，单位是“卡”
- `-memory`：每个trainer进程使用的内存资源，格式为“数字+单位”，单位可以是：Ki，Mi，Gi
- `-parallelism`：启动trainer的个数／并发节点数
- `-pscpu`：parameter server占用的CPU资源，单位是“核”
- `-pservers`：parameter server的节点个数
- `-psmemory`：parameter server占用的内存资源，格式为“数字+单位”，单位可以是：Ki，Mi，Gi
- `-topology`：指定PaddlePaddle v1训练的模型配置python文件
- `-entry`: 指定PaddlePaddle v2训练程序的启动命令
- `-passes`：执行训练的pass个数
- `package`：HDFS 训练任务package的路径

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
label selector: paddle-job-pserver=paddle-cluster-job, desired: 3
running pod list:  [('Running', '172.17.29.47'), ('Running', '172.17.37.46'), …, ('Running', '172.17.28.244')]
sleep for 10 seconds...
running pod list:  [('Running', '172.17.29.47'), ('Running', '172.17.37.46'), …, ('Running', '172.17.28.244')]
label selector: paddle-job=paddle-job-yanxu, desired: 10
running pod list:  [('Running', '172.17.31.182’),…(‘Running', '172.17.12.234'), ('Running', '172.17.22.238')]
Starting training job:  /pfs/***/home/***/***/ctr_package_cloud, num_gradient_servers: 200, trainer_id:  102, version:  v1
I0524 12:00:31.511015    43 Util.cpp:166] commandline: /usr/bin/../opt/paddle/bin/paddle_trainer --port=7164 --nics= --ports_num=1 --ports_num_for_sparse=1 --num_passes=1 --trainer_count=1 --saving_period=1 --log_period=20 --local=0 --config=trainer_config.py --use_gpu=0 --trainer_id=102 --save_dir= --pservers=172.17.29.47,,172.17.28.244 --num_gradient_servers=200
[INFO 2017-05-24 12:00:39,316 networks.py:1482] The input order is [....]
[INFO 2017-05-24 12:00:39,319 networks.py:1488] The output order is [__cost_0__]
I0524 12:00:39.330195    43 Trainer.cpp:165] trainer mode: Normal
I0524 12:00:39.514008    43 PyDataProvider2.cpp:243] loading dataprovider dataprovider::process_deep
[INFO 2017-05-24 12:00:39,814 dataprovider.py:21] hook
[INFO 2017-05-24 12:00:39,883 dataprovider.py:33] dict_size is 5231
[INFO 2017-05-24 12:00:39,883 dataprovider.py:34] schema_pos_size is 552
[INFO 2017-05-24 12:00:39,883 dataprovider.py:35] schema_output_size is 51
I0524 12:00:39.884352    43 PyDataProvider2.cpp:243] loading dataprovider dataprovider::process_deep
[INFO 2017-05-24 12:00:39,884 dataprovider.py:21] hook
[INFO 2017-05-24 12:00:39,914 dataprovider.py:33] dict_size is 5231
[INFO 2017-05-24 12:00:39,914 dataprovider.py:34] schema_pos_size is 552
[INFO 2017-05-24 12:00:39,914 dataprovider.py:35] schema_output_size is 51
I0524 12:00:39.915364    43 GradientMachine.cpp:86] Initing parameters..
I0524 12:00:39.924811    43 GradientMachine.cpp:93] Init parameters done.
I0524 12:00:39.924881    43 ParameterClient2.cpp:114] pserver 0 172.17.29.47:7164
I0524 12:00:39.925227    43 ParameterClient2.cpp:114] pserver 1 172.17.37.46:7164
I0524 12:00:39.925472    43 ParameterClient2.cpp:114] pserver 2 172.17.55.171:7164
I0524 12:00:39.925714    43 ParameterClient2.cpp:114] pserver 3 172.17.35.175:7164
```

如果任务workers较多，可以指定查看某个worker的单独的日志：`paddlecloud logs -w paddle-cluster-job-trainer-3s4nz paddle-cluster-job`。

## 终止任务
执行`paddlecloud kill paddle-cluster-job`即可停止训练任务的所有节点和进程。

kill命令执行成功后，集群会在后台终止集群作业的workers进程，workers并不会在kill命令之后全部停止。如果需要查看kill掉的任务正在清理的workers，可以使用命令`paddlecloud get workers paddle-cluster-job`查看。

***所以在kill之后提交新的任务时，要记得更改提交时的`-name`参数，防止任务名称冲突。***
