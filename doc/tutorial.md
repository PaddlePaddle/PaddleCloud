# Getting Started with Submitting Job

## Download and Configure paddlectl

`paddlectl` is a command-line tool that submits distributed training job to the cloud.

- Step1

  Download `paddlectl` and copy the binary file to the folder which under
  environment variable `$PATH`, such as `/usr/local/bin`, then set the
  executable access by command `chmod +x /usr/local/bin/paddlectl`.

  We advise users to download the last tag `paddlectl` on
  the [Release Page](https://github.com/PaddlePaddle/cloud/releases),
  and you can also fetch the least version binary file from our CI system,
  it's built by the last code on develop branch.

  OS |  Link
  -- | --
  Mac OSX| [paddlecloud.darwin](http://guest:@paddleci.ngrok.io/repository/download/PaddleCloud_Client/.lastSuccessful/paddlecloud.darwin)
  Windows| [paddlecloud.exe](http://guest:@paddleci.ngrok.io/repository/download/PaddleCloud_Client/.lastSuccessful/paddlecloud.exe)
  Linux | [paddlecloud.x86_64](http://guest:@paddleci.ngrok.io/repository/download/PaddleCloud_Client/.lastSuccessful/paddlecloud.x86_64)

- Step2

  Configurate the configuration file `~/.paddle/config` (`./paddle\config`
  under current user folder for Windows),
  an example is as following:

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

  Supports you have two data-center's access, one for `production` and another one for
  `experimentation`, you can choose one data-center by the `current-datacenter` field
  for the current one.

  `username`, `password` and `endpoint` is your secret for the data-center, you will
  received an email including your secret information send by the administrator.

After finish the two steps above, execute `paddlectl` command will print the usage:

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


Use "paddlectl flags" for a list of top-level flags
```

## Download the Demo Code and Try to Submit it

We prepare some demo code to help users understanding how to submit
a distributed training job to PaddleCloud, these demo codes are based
on [Paddle Book](https://github.com/Paddlepaddle/book), you can find the
tutorials for each chapter.

You can fetch the demo code and submit the job as the following command:

```bash
> mkdir fit_a_line
> cd fit_a_line
> wget https://raw.githubusercontent.com/PaddlePaddle/cloud/develop/demo/fit_a_line/train.py
> cd ..
> paddlecloud submit -jobname fit-a-line -cpu 1 -gpu 1 -parallelism 1 -entry "python train.py train" fit_a_line/
```

For each submits arguments:

- `-jobname`, STRING, the job name, you should specify a unique name.
- `-cpu`, INT, CPU cores for each trainer node.
- `-gpu`, INT, GPU cards for each trainer node,
  if the cluster doesn't support GPU, please set `-gpu 0`.
- `-parallelism`, INT, the parallelism, means trainer node count.
- `-entry`, STRING, the entry point for the training job.
- `./fit_a_line`, STRING, the local directory including job package.

**NOTE1**: You can find the whole usage by `paddlectl submit -h`.
**NOTE2**: Submit the job by different jobname, so that the previous jobs data would be saved on the cloud.
**NOTE3**: If you want a higher parallelism, you could modify entry point by `-entry "python train.py prepare"`,
  to prepare the training data on PFS, and then submit the training job again.

## Check the Status and Logs

After submitting the job, you can check all the jobs by `paddlectl get jobs`

```bash
> paddlectl get jobs
NUM   NAME         SUCC    FAIL    START                  COMP                   ACTIVE
0     fit-a-line   <nil>   <nil>   2017-06-26T08:41:01Z   <nil>                  1
```

- `ACTIVE`, the number for running training node.
- `SUCC`, the number for finished training node.
- `FAIL`, the number for failed training node.

Then, you can view the logs of job(only running or finished status) by `paddelctl logs`

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
```

`paddlectl logs` will return the last 10 lines by default, you can also
use `-n <number>` argument to print the last number lines.

```bash
> paddlecloud logs -n 100 fit-a-line
...
```

## Download the Saved Models

When a training job finished, the output model would be saved on PFS, you can
check and fetch the output models by the following commands:

```bash
> paddlecloud file ls /pfs/dlnel/home/wuyi05@baidu.com/jobs/fit_a_line/
train.py
image
output
> paddlecloud file ls /pfs/dlnel/home/wuyi05@baidu.com/jobs/fit_a_line/output/
pass-0001.tar
...
> paddlecloud file get /pfs/dlnel/home/wuyi05@baidu.com/jobs/fit_a_line/output/pass-0001.tar ./
```

## Clean the Training Job

The following command will clean the training job, after that, you could check
the logs, but you can find the output on
`/pfs/<data-center/home/<username>/jobs/<job-name>`:

```bash
> paddlectl kill fit-a-line
```

---
More details about the usage: [usage_cn.md](./usage_cn.md)