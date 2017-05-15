# Design Doc: PaddleCloud Client

A Command Line Interface for PaddlePaddle Cloud

---

# Goals:

Developers using PaddlePadle Cloud can use this command-line client for the convenience of managing cloud Deep-Learning jobs, including:

- Submitting a PaddlePaddle cluster training job
- List jobs that are currently running.
- List all job history that has been submitted.
- Fetch logs of the jobs.
- Download output(model data) of the completed job.
- View user's quota usages in their spaces.

PaddleCloud Client is written in Go, which simplifies building for different platforms like MacOS/Linux/Windows. The client is a binary program named "paddlecloud".

# Client Configurations

Client configurations stores in a file: `~/.paddle/config` in the below `yaml` form:
```yaml
datacenter1:
  username: user
  usercert: /path/to/user.pem
  userkey:  /path/to/user-key.pem
  endpoint: dc1.paddlepaddle.org
```

# Client commands

## Reference

- `paddlecloud submit [options] <package path>`: submit job to PaddlePaddle Cloud
    - `<package path>`: ***Required*** Job package to submit. Including user training program and it's dependencies.
    - `-parallelism`: Number of parallel trainers. Defaults to 1.
    - `-cpu`: CPU resource each trainer will use. Defaults to 1.
    - `-gpu`: GPU resource each trainer will use. Defaults to 0.
    - `-memory`: Memory resource each trainer will use. Defaults to 1Gi. Memory ammounts consists a plain integer using one of these suffixes: Ei, Pi, Ti, Gi, Mi, Ki
    - `-pservers`: Number of parameter servers. Defaults equal to `-p`
    - `-pscpu`: Parameter server CPU resource. Defaults to 1.
    - `-psmemory`: Parameter server memory resource. Defaults to 1Gi.
    - `-entry`: Command of starting trainer process. Defaults to `paddle train`
    - `-topology`: ***Will Be Deprecated*** `.py` file contains paddle v1 job configs
- `paddlecloud kill [-rm] <job name>`: Stop the job. `-rm` will remove the job from history.
- `paddlecloud get [options] <resource>`: Print resources
    - `jobs`: List jobs. List only running jobs if no `-a` specified.
    - `workers`: List detailed job worker nodes.
    - `quota`: Print quota usages
    - `-a`: List all resources.
- `paddlecloud logs [-n lines] <job>`: Print logs of the job.
- `paddlecloud pfs ...`: PaddlePaddle Cloud data management.
    `<dest>` is the path on the cloud. The form must be like `/pfs/$DATACENTER/home/$USER`. `$DATACENTER` is the from configuration where you setup at `~/.paddle/config`
    - `paddlecloud pfs cp <local src> [<local src> ... ] <remote dest>`: Upload a file
    - `paddlecloud pfs cp <remote src> [<remote src> ... ] <local dest>`: Download a file
    - `paddlecloud pfs ls <remote dir>`: List files under `<remote dir>`.
    - `paddlecloud pfs rm <remote> ...`: Delete remote files

## Examples

A Sample job package may contain files like below:

```
job_word_emb/
  |-- train.py
  |-- dict1.pickle
  |-- dict2.pickle
  |-- my_topo.py
data/
  |-- train/
  |   |-- train.txt-00000
  |   |-- train.txt-00001
  |   ...
  `-- test/
      |-- test.txt-00000
      |-- test.txt-00001
      ...
```

Run the following command to submit the job to the cloud:

```bash
# upload training data to cloud, which may be very large
$ paddlecloud pfs cp -r ./job_word_emb/data /pfs/datacenter1/home/user1/job_word_emb
# submit a v1 paddle training job
$ paddlecloud submit ./job_word_emb -p 4 -c 2 -m 10Gi -t modules/train.py
Collecting package ... Done
Uploading package ... Done
Starting kuberntes job ... Done
# list running jobs
$ paddlecloud jobs
NAMESPACE           NAME                      DESIRED   SUCCESSFUL   AGE
user1-gmail-com     paddle-job-trainer-x63f   3         0            15s
# get job logs
$ paddlecloud logs
running pod list:  [('Running', '10.1.9.3'), ('Running', '10.1.32.9'), ('Running', '10.1.18.7')]
I0515 05:44:19.106667    21 Util.cpp:166] commandline:  --ports_num_for_sparse=1 --use_gpu=False --trainer_id=0 --pservers=10.1.9.3,10.1.32.9,10.1.18.7 --trainer_count=1 --num_gradient_servers=1 --ports_num=1 --port=7164
[INFO 2017-05-15 05:44:19,123 networks.py:1482] The input order is [firstw, secondw, thirdw, fourthw, fifthw]
[INFO 2017-05-15 05:44:19,123 networks.py:1488] The output order is [__classification_cost_0__]
[INFO 2017-05-15 05:44:19,126 networks.py:1482] The input order is [firstw, secondw, thirdw, fourthw, fifthw]
[INFO 2017-05-15 05:44:19,126 networks.py:1488] The output order is [__classification_cost_0__]
I0515 05:44:19.131026    21 GradientMachine.cpp:85] Initing parameters..
I0515 05:44:19.161273    21 GradientMachine.cpp:92] Init parameters done.
I0515 05:44:19.161551    41 ParameterClient2.cpp:114] pserver 0 10.1.9.3:7165
I0515 05:44:19.161573    21 ParameterClient2.cpp:114] pserver 0 10.1.9.3:7164
I0515 05:44:19.161813    21 ParameterClient2.cpp:114] pserver 1 10.1.32.9:7164
I0515 05:44:19.161854    41 ParameterClient2.cpp:114] pserver 1 10.1.32.9:7165
I0515 05:44:19.162405    21 ParameterClient2.cpp:114] pserver 2 10.1.18.7:7164
I0515 05:44:19.162410    41 ParameterClient2.cpp:114] pserver 2 10.1.18.7:7165
I0515 05:44:21.187485    48 ParameterClient2.cpp:114] pserver 0 10.1.9.3:7165
I0515 05:44:21.187595    48 ParameterClient2.cpp:114] pserver 1 10.1.32.9:7165
I0515 05:44:21.189729    48 ParameterClient2.cpp:114] pserver 2 10.1.18.7:7165
I0515 05:44:21.242624    49 ParameterClient2.cpp:114] pserver 0 10.1.9.3:7164
I0515 05:44:21.242717    49 ParameterClient2.cpp:114] pserver 1 10.1.32.9:7164
I0515 05:44:21.243191    49 ParameterClient2.cpp:114] pserver 2 10.1.18.7:7164
$ paddlecloud pfs cp /pfs/datacenter1/home/user1/job_word_emb/output ./output
Downloading /pfs/datacenter1/home/user1/job_word_emb/output ... Done
```

# API Definition

PaddleCloud Client calls a remote "RESTful server" to accomplish the goals. This "RESTful server" is deployed in the cloud and serves all client calls for all users.

We have multiple API versions, currently, it is "v1" and "v2". "v1" stands for submitting a paddle job which is written with paddle v1 API. All endpoints start with the version path, like "/v1/submit"

We define the APIs below:

| Endpoint   | method | arguments |
| --------   | ------ | --------- |
| /api/v1/jobs   | POST   | see [above](#submit-job-api) |
| /api/v1/jobs   | GET    | see [above](#get-jobs) |
| /api/v1/quota  | GET    | see [above](#client-commands) |
| /api/v1/pfs/*  |  -     | see [here](https://github.com/gongweibao/Paddle/blob/filemanager2/doc/design/file_manager/README.md#pfsserver) |

## Submit Job
- HTTP Request

`POST /api/v1/jobs`

- Body Parameters
```json
{
  "name": "paddle-job",
  "job_package": "/pfs/datacenter1/home/user1/job_word_emb",
  "parallelism": 3,
  "cpu": 1,
  "gpu": 1,
  "memory": "1Gi",
  "pservers": 3,
  "pscpu": 1,
  "psmemory": "1Gi",
  "topology": "train.py"
}
```

- HTTP Response
```json
{
    "code":200,
    "msg":"OK"
}
```

## Get Jobs
- HTTP Request

`GET /api/v1/jobs`

- HTTP Response

```json
"code":200,
"msg":[
  {
    "name": "paddle-job-b82x",
    "job_package": "/pfs/datacenter1/home/user1/job_word_emb",
    "parallelism": 3,
    "cpu": 1,
    "gpu": 1,
    "memory": "1Gi",
    "pservers": 3,
    "pscpu": 1,
    "psmemory": "1Gi",
    "topology": "train.py"
  },
  {
    "name": "paddle-yx-02c2",
    "job_package": "/pfs/datacenter1/home/user2/job_word_emb",
    "parallelism": 3,
    "cpu": 1,
    "gpu": 1,
    "memory": "1Gi",
    "pservers": 3,
    "pscpu": 1,
    "psmemory": "1Gi",
    "topology": "train.py"
  }
]
```
