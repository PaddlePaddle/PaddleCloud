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

## 准备训练数据

不同的PaddlePaddleCloud集群环境会提供不同的分布式存储服务。目前PaddlePaddleCloud支持HDFS和CephFS。

### HDFS环境下训练数据准备

使用`ssh`登录到集群中的公用数据中转服务器上，进行数据上传，下载，更新等操作。您可以在中转服务器的`/mnt`路径下找到集群HDFS的目录，并可以访问当前有权限的目录。上传数据则可以使用管理普通Linux文件的方式，将数据`scp`到中转服务器`/mnt`下的用户数据目录。比如：

```bash
scp -r my_training_data_dir/ user@tunnel-server:/mnt/hdfs_mulan/idl/idl-dl/mydir/
```

***说明：您可能需要联系集群管理员以获得数据中转服务器的登录地址和权限。***

在训练任务提交后，每个训练节点会把HDFS挂载在`/pfs/[datacenter_name]/home/[username]/`目录下这样训练程序即可使用这个路径读取训练数据并开始训练。

### 使用[RecordIO](https://github.com/PaddlePaddle/recordio)对训练数据进行预处理
用户需要在本地将数据预先处理为RecordIO的格式，再上传至集群进行训练。
- 使用RecordIO库进行数据预处理
```python
import paddle.v2.dataset as dataset
dataset.convert(output_path = "./dataset",
                reader = dataset.uci_housing.train(),
                num_shards = 10,
                name_prefix = "uci_housing_train")
```
  - `output_path` 输出路径
  - `reader` 用户自定义的[reader](https://github.com/PaddlePaddle/Paddle/tree/develop/doc/design/reader),实现方法可以参考[paddle.v2.dataset.uci_housing.train()](https://github.com/PaddlePaddle/Paddle/blob/develop/python/paddle/v2/dataset/uci_housing.py#L74)
  - `num_shards` 生成的文件数量
  - `num_prefix` 生成的文件名前缀

执行成功后会在本地生成如下文件：
```bash
.
./dataset
./dataset/uci_houseing_train-00000-of-00009
./dataset/uci_houseing_train-00001-of-00009
./dataset/uci_houseing_train-00002-of-00009
./dataset/uci_houseing_train-00003-of-00009
...
```

- 编写reader来读取RecordIO格式的文件
```python
import cPickle as pickle
import recordio
import glob
import sys
def recordio_reader(filepath, parallelism, trainer_id):
    # sample filepath as "/pfs/dlnel/home/yanxu05@baidu.com/dataset/uci_housing/uci_housing_train*"
    def reader():
        if trainer_id >= parallelism:
            sys.stdout.write("invalied trainer_id: %d\n" % trainer_id)
            return
        files = glob.glob(filepath)
        files.sort()
        my_file_list = []
        for idx, f in enumerate(files):
            if idx % parallelism == trainer_id:
                my_file_list.append(f)

        for fn in my_file_list:
            r = recordio.reader(fn)
            while True:
                d = r.read()
                if not d:
                    break
                yield pickle.loads(d)

    return reader
```

### 使用paddlecloud上传训练数据

paddlecloud命令集成了上传数据的功能，目前仅针对存储系统是CephFS的环境。如果希望上传，执行：

```bash
paddlecloud file put src dest
```
- `src` 必须是当前目录的子目录，`../`是不允许的。
- `src` 如果以'/'结尾，则表示上传`src`目录下的文件，不会在`dest`下创建新的目录。
- `src` 如果没有以`/`结尾，则表示上传`src`目录，会在`dest`下创建一个新的目录。
- `dest` 必须包含`/pfs/{datacenter}/home/{username}`目录。



### 使用公共数据集

不论是在HDFS环境还是CephFS环境，用户提交的任务中都可以访问目录`/pfs/public`获得公开数据集的访问。
在分布式环境中，每个trainer希望访问一部分数据，则可以编写如下的reader代码访问已经拆分好的数据集：

```python
TRAIN_FILES_PATTERN = os.path.join(common.DATA_HOME,
                                   "uci_housing/train-*.pickle")
def train():
    return common.cluster_files_reader(
        TRAIN_FILES_PATTERN,
        trainer_count = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS", "1")),
        trainer_id = int(os.getenv("PADDLE_INIT_TRAINER_ID", "0")))

```

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

### 上传训练程序包到HDFS（仅HDFS存储下需要）

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
paddlecloud submit -jobname my-paddlecloud-job \
  -cpu 1 \
  -gpu 0 \
  -memory 1Gi \
  -parallelism 10 \
  -pscpu 1 \
  -pservers 3 \
  -psmemory 1Gi \
  -entry "python trainer_config.py" /pfs/[datacenter_name]/home/[username]/ctr_demo_package
```

- 提交基于V2 API的训练任务

```bash
paddlecloud submit -jobname my-paddlecloud-job \
  -cpu 1 \
  -gpu 0 \
  -memory 1Gi \
  -parallelism 10 \
  -pscpu 1 \
  -pservers 3 \
  -psmemory 1Gi \
  -passes 1 \
  -entry "python trainer_config.py" \
  /pfs/[datacenter_name]/home/[username]/ctr_demo_package
```

参数说明：
- `jobname`：提交任务的名称，paddlecloud使用`jobname`唯一标识一个任务
    - ***注意：*** jobname必须由字母、数字、“-”和“.”组成，并且以字母数字组合结尾，***不能*** 包含下划线“_”
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

### 使用自定义的Runtime Docker Image
runtime Docker Image是实际被Kubernetes调度的Docker Image，如果在某些情况下需要自定义属于某个任务的Docker Image可以通过以下方式
- 自定义Runtime Docker Image
  ```bash
  git clone https://github.com/PaddlePaddle/cloud.git && cd cloud/docker
  ./build_docker.sh {PaddlePaddle production image} {runtime Docker image}
  docker push {runtime Docker image}
  ```
- 使用自定义的runtime Docker Image来运行Job
  ```bash
  paddlecloud submit -image {runtime Docker image} -jobname ...
  ```

- 使用私有registry的runtime Docker image
  - 在PaddleCloud上添加registry认证信息
    ```bash
    paddlecloud registry \
      -username {your username}
      -password {your password}
      -server {your registry server}
      -name {your registry name}
    ```
  - 使用私有registry提交任务
    ```bash
    paddlecloud submit \
      -image {runtime Docker image} \
      -registry {your registry name}
    ```
  - 查看所有的registry
    ```bash
    paddlecloud get registry
    ```
  - 删除指定的registry
    ```bash
    paddlecloud delete registry
    ```


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

## 如何准备一个支持分布式的dataset
由于分布式训练会同时启动多个trainer实例，为了保证每个trainer实例能够获取到同等规模的数据集，我们需要对单机dataset拆分为多个小文件, 每个trainer根据自己的运行时信息来决定读取哪些具体的文件。[这里](../demo/fit_a_line/train.py)是训练程序的样例，[这里](../docker/python/paddle/cloud/dataset/uci_housing.py)是dataset的样例。

### 预处理训练数据
您可以使用PaddlePaddle提供的默认函数[paddle.v2.dataset.common.split](https://github.com/PaddlePaddle/Paddle/blob/develop/python/paddle/v2/dataset/common.py#L81)将reader的数据切分为多个小文件，当然您也可以自定义一个预处理函数来切分数据：

```python
import paddle.v2.dataset.uci_housing as uci_housing
import paddle.v2.dataset.common as common
common.split(reader = uci_housing.train(),   // Your reader instance
            line_count = 500,       // line count for each file
            suffix = "./uci_housing/train-%05d.pickle")              // filename suffix for each file, must contain %05d
```

`split`默认会使用[cPickle](https://docs.python.org/2/library/pickle.html#module-cPickle)函数将Python对象序列化到本地文件, 上述代码会将uci_housing的数据集切分成成多个cPickle格式的小文件，您可以使用PaddlePaddle的生产环境镜像在本地运行切分数据的代码：

```bash
docker run --rm -it -v $PWD:/work paddlepaddle/paddle:latest python /work/run.py
```
执行成功后可以通过公用的数据中转机将数据上传至集群。

- 自定义序列化函数

  您可以用过`dumper`参数来指定序列化的函数，dumper的接口格式为

  ```python
  dumper(obj=<data object>, file=<open file object>)
  ```

  例如，使用[marshal.dump](https://docs.python.org/2.7/library/marshal.html#marshal.dump)替换默认的`cPickle.dump`

  ```python
  common.split(reader = uci_housing.train(),   // Your reader instance
              line_count = 500,       // reader iterator count for each file
              suffix="./uci_housing/train-%05d.pickle",              // filename suffix for each file
              dumper=marshal.dump)      // using pickle.dump instead of the default function: cPickle.dump
  ```

### 读取分布式文件的reader
训练代码需要在运行时判断自己身份并决定读取哪些文件,PaddlePaddle同样提供了默认函数[paddle.v2.dataset.common.cluster_files_reader](https://github.com/PaddlePaddle/Paddle/blob/develop/python/paddle/v2/dataset/common.py#L119)来读取这些文件，您也可以定义自己的函数来读取文件。通过以下环境变量可以获取到一些有用的运行时信息：
- `PADDLE_INIT_NUM_GRADIENT_SERVERS`: trainer实例的数量
- `PADDLE_INIT_TRAINER_ID`: 当前trainer的ID,从0开始到`$TRAINERS-1`
- `PADDLE_CLOUD_CURRENT_DATACENTER`: 当前的datacenter

样例代码：
```python
import paddle.v2.dataset.common as common

dc = os.getenv("PADDLE_CLOUD_CURRENT_DATACENTER")

def train():
  return common.cluster_files_reader(
    "/pfs/%s/public/dataset/uci_housing/train-*.pickle" % dc,
    trainer_count = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS")),
    train_id = int(os.getenv("PADDLE_INIT_TRAINER_ID"))
  )
```

- 自定义文件加载函数
  同样您也可以通过`loader`参数来指定如何加载文件,`loader`的接口格式:

  ```python
  d = loader(f = <open file object>)
  ```

  例如,使用[marshal.load](https://docs.python.org/2.7/library/marshal.html#marshal.load)替换默认的`cPickle.load`:

  ```python
  def train():
    return common.cluster_files_reader(
      "/pfs/%s/public/dataset/uci_housing/train-*.pickle" % dc,
      trainer_count = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS")),
      train_id = int(os.getenv("PADDLE_INIT_TRAINER_ID")),
      loader = marshal.load
    )
  ```

*注意*: `"/pfs/%s/public" % dc`是公用数据的默认访问路径，所有Job对此目录具有*只读*权限。
