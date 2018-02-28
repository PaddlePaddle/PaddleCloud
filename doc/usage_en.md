# Submit PaddlePaddle Cluster Jobs

---

## Donwload And Configure paddlecloud client

`paddlecloud` is the command line tool for PaddlePaddleCloud distributed training.

### Steps

1. Download latest `paddlecloud` client binary from https://github.com/PaddlePaddle/cloud/releases, copy `paddlecloud` to system $PATH directory, I.E. `/usr/local/bin`

1. Create `~/.paddle/config` (`.paddle\config` under current user directory for Windows users) and fill it with the following content:

```yaml
datacenters:
- name: datacenter1
  username: [your user name]
  password: [secret]
  endpoint: http://cloud.paddlepaddle.org
current-datacenter: datacenter1
```
Above file configures PaddlePaddleCloud cluster endpoints with user credentials.

- name: self defined datacenter name, can be any string.
- username: username for PaddlepaddleCloud, usually this is a email address, created by system administrator
- password: the password for above account.
- endpoint: URL for PaddlePaddleCloud cluster API. This info will be provided by system administrator along with username&password
- current-datacenter: To identify current datacenter

After finish editing the config file, run `paddlecloud` will show help message as follows:

``` bash
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

## Preparing Training Data

Different PaddlePaddle Cloud environment may provide different storage services. PaddlePaddle Cloud currently works with HDFS and CephFS.

### Manually upload training data to HDFS

You can login to public data server via `ssh` to upload, download or update data. You can find cluster HDFS directory under `/mnt` and access the directories you are authorized. Data uploading is accomplished via common Linux file managing fashion, `scp` you local data to public data server's `/mnt` directory, for example:

```bash
scp -r my_training_data_dir/ user@tunnel-server:/mnt/hdfs_mulan/idl/idl-dl/mydir/
```

***Please note：You might need to contact system administrator to find the public server login address and credential***

When training job is submitted, each training node will mount HDFS under `/pfs/[datacenter_name]/home/[username]/`, so that the training program can access the data and start training.

### Preprocess training data with [RecordIO](https://github.com/PaddlePaddle/recordio)

User need to preprocess and convert local data to the format of RecordIO before uploading to training cluster.

- Preprocessing data with RecordIO library
```python
import paddle.v2.dataset as dataset
dataset.common.convert(output_path = "./dataset",
                reader = dataset.uci_housing.train(),
                line_count = 10,
                name_prefix = "uci_housing_train")
```
  - `output_path` output path
  - `reader` user defined [reader](https://github.com/PaddlePaddle/Paddle/tree/develop/doc/design/reader), please refer to [paddle.v2.dataset.uci_housing.train()] for implementation(https://github.com/PaddlePaddle/Paddle/blob/develop/python/paddle/v2/dataset/uci_housing.py#L74)
  - `num_shards` number of files to generate
  - `num_prefix` prefix of file names

files will be generated as follows：
```bash
.
./dataset
./dataset/uci_houseing_train-00000-of-00009
./dataset/uci_houseing_train-00001-of-00009
./dataset/uci_houseing_train-00002-of-00009
./dataset/uci_houseing_train-00003-of-00009
...
```

- Implement RecordIO reader
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

### Upload training data using paddlecloud

`paddlecloud` command line is able to upload data training cluster to CephFS.

```bash
paddlecloud file src dest
```

- `src` must be child directory of current directory，`../` is not allowed.
- if `src` ends with `/`, that means uploading files under `src`, no new directories will be created under `des`.
- if `src` does NOT end with `/`, that means `src` directory will be uploaded, which will be a sub directory of `dest`.
- `dest` must contain `/pfs/{datacenter}/user/{username}` as part of the path.


### Using public dataser

In either HDFS or CephFS, user training task is able to access `/pfs/public` for public dataset.
In distributed training environment, if each trainer just need to access a subset of the data, try the following reader to access the sliced dataset

```python
TRAIN_FILES_PATTERN = os.path.join(common.DATA_HOME,
                                   "uci_housing/train-*.pickle")
def train():
    return common.cluster_files_reader(
        TRAIN_FILES_PATTERN,
        trainer_count = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS", "1")),
        trainer_id = int(os.getenv("PADDLE_INIT_TRAINER_ID", "0")))

```

## Training Program Package

Training program package is a directory which contains training program, dependencies and config files. This directory must contain all the dependencies, or it will not work properly in cluster.

### Define the training data distribution in cluster

Each training job will launch multiple training instances in multiple cluster nodes, each training instance will take a subset of the training task. To evenly distribute the large amount of data to each of the nodes, the following python program is an example of doing so:

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
Above program will distribute files to training nodes in order, according to the training data path in HDFS, and create `train.list`,  `test.list` to store training data file path for current node. Pass the path to these 2 files when you call `define_py_data_sources2` to define training data.


### Upload training program package to HDFS

Uploading training program to HDFS is the same as uploading training data. Upload training program package via public server. for example: 

```bash
scp -r my_training_package/ user@tunnel-server:/mnt/hdfs_mulan/idl/idl-dl/mypackage/
```

In the following steps of submitting training jobs, you will need to set the directory of training program package with `/pfs/[datacenter_name]/home/[username]/idl/idl-dl/mypackage/` to run it.


***NOTE: This method will be deprecated.***


## Submitting A Job

Submitting training jobs with the following command:

- Submitting jobs utilizing v1 API

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

- Submitting jobs utilizing V2 API 

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

Parameters：
- `jobname`：training job's name，paddlecloud use `jobname` as unique identifier for jobs.
    - ***Note：*** jobname can only contain alphabet character, number, `-` `.`, and must end with combination of character or number，*** NO *** `_` is allowed.
- `-cpu`：CPU resource for each trainer, in units of `core`
- `-gpu`：CPU resource for each trainer, in units of `card`
- `-memory`：Memory resource for each trainer, value can be a integer with a `Ki，Mi，Gi` as suffix.
- `-parallelism`：trainer/parallelism node count.
- `-pscpu`：CPU resource for parameter server, in unites of `core`
- `-pservers`：number of parameter servers
- `-psmemory`：Memory resource for parameter server, value can be a integer with a `Ki，Mi，Gi` as suffix.
- `-topology`：PaddlePaddle v1 training config file path
- `-entry`: PaddlePaddle v2 training program launch command
- `-passes`：number of passes
- `package`：path to HDFS training job


### Using customized Runtime Docker Image

If your training task need to be wrapped in a docker image to be scheduled by kubernetes, here is how:

- define Runtime Docker Image
  ```bash
  git clone https://github.com/PaddlePaddle/cloud.git && cd cloud/docker
  ./build_docker.sh {PaddlePaddle production image} {runtime Docker image}
  docker push {runtime Docker image}
  ```
- define runtime Docker Image to run training jobs
  ```bash
  paddlecloud submit -image {runtime Docker image} -jobname ...
  ```

- if your runtime Docker image is submitted to a private registry
  - add registry credentials in PaddleCloud
    ```bash
    paddlecloud registry \
      -username {your username}
      -password {your password}
      -server {your registry server}
      -name {your registry name}
    ```
  - submit training job via private registry
    ```bash
    paddlecloud submit \
      -image {runtime Docker image} \
      -registry {your registry name}
    ```
  - list all registries
    ```bash
    paddlecloud get registry
    ```
  - delete a registry
    ```bash
    paddlecloud delete registry
    ```


## View Job Status

User can check the status of training jobs, nodes and disk quota with the following command:

``` bash
paddlecloud get jobs


NUM	NAME	SUCC	FAIL	START	COMP	ACTIVE
0	paddle-cluster-job	<nil>	1	2017-05-24T07:52:45Z	<nil>	1
```

To check the status of workers, run the following command:

``` bash
paddlecloud get workers paddle-cluster-job

NAME	STATUS	START
paddle-cluster-job-trainer-3s4nz	Running	2017-05-24T07:53:41Z
paddle-cluster-job-trainer-6sc4q	Running	2017-05-24T07:53:03Z
...
```

## View Job Logs

Run following command to check current training job's log from all workers:

``` bash
paddlecloud logs paddle-cluster-job

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



## Terminating Jobs

Run `paddlecloud kill paddle-cluster-job` to terminate the training job.

When above command is successful, cluster will try to terminate all the workers process in background, this procedure might take some time and works might not be terminated immediately. If you need to check if your work has been cleared, run `paddlecloud get workers paddle-cluster-job`

*** When submitting a new job after terminating one, make sure -name is different to prevent name conflicting ***

## To prepare a dataset for distributed training

Since distributed training will start multiple training instance, to ensure data is evenly distributed and delivered to each trainer, we need to split dataset into small pieces, each trainer will decide files to read based on its runtime state. [here](../demo/fit_a_line/train.py) is an example training program, [here](https://github.com/PaddlePaddle/Paddle/blob/develop/python/paddle/v2/dataset/uci_housing.py) is an example of dataset.


### Preprocessing training data.

You can utilize PaddlePaddle's [paddle.v2.dataset.common.split]((https://github.com/PaddlePaddle/Paddle/blob/develop/python/paddle/v2/dataset/common.py#L129)) to chop reader's data into small pieces, or you can define your own as follows:

```python
import paddle.v2.dataset.uci_housing as uci_housing
import paddle.v2.dataset.common as common
common.split(reader = uci_housing.train(),   // Your reader instance
            line_count = 500,       // line count for each file
            suffix = "./uci_housing/train-%05d.pickle")              // filename suffix for each file, must contain %05d
```

`split` uses [cPickle](https://docs.python.org/2/library/pickle.html#module-cPickle) to serialize pythonb objects to local file by default. Above program will split uci_housing dataset into multipule cPcikle file. You can use PaddlePaddle production docker image to split dataset locally.

```bash
docker run --rm -it -v $PWD:/work paddlepaddle/paddle:latest python /work/run.py
```


- Customize serialization function

  You can set your own serialization function by passing it with `dumper` parameter. Dumper's interface spec is as follows:

  ```python
  dumper(obj=<data object>, file=<open file object>)
  ```

  For an example，use [marshal.dump](https://docs.python.org/2.7/library/marshal.html#marshal.dump) instead

  ```python
  common.split(reader = uci_housing.train(),   // Your reader instance
              line_count = 500,       // reader iterator count for each file
              suffix="./uci_housing/train-%05d.pickle",              // filename suffix for each file
              dumper=marshal.dump)      // using pickle.dump instead of the default function: cPickle.dump
  ```

### Reader to read from distributed files

Training program need to decide files to read based on it's runtime role, PaddlePaddle provides default file reader [paddle.v2.dataset.common.cluster_files_reader](https://github.com/PaddlePaddle/Paddle/blob/develop/python/paddle/v2/dataset/common.py#L167) to read these files, or you can customize it with your own function utilizing following environment variables.
- `PADDLE_INIT_NUM_GRADIENT_SERVERS`: number of trainers
- `PADDLE_INIT_TRAINER_ID`: current trainer id, starts from 0
- `PADDLE_CLOUD_CURRENT_DATACENTER`: current data center.

code example:
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

- Customize file loading function.
  
  You can also customize `loader` to define how the the file is loaded, the interface spec for loader is as follows:

  ```python
  d = loader(f = <open file object>)
  ```

  For an example, use [marshal.load](https://docs.python.org/2.7/library/marshal.html#marshal.load)to replace `cPickle.load`:

  ```python
  def train():
    return common.cluster_files_reader(
      "/pfs/%s/public/dataset/uci_housing/train-*.pickle" % dc,
      trainer_count = int(os.getenv("PADDLE_INIT_NUM_GRADIENT_SERVERS")),
      train_id = int(os.getenv("PADDLE_INIT_TRAINER_ID")),
      loader = marshal.load
    )
  ```

*Please Node*: `"/pfs/%s/public" % dc` is the default path for all the public training datasets, which is *READ ONLY* for all jobs.