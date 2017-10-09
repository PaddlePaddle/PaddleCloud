# Web Interface design

This design doc will talk about features and web pages needed to let users manage cloud paddle jobs.

## Feature List

- Account Management
    - Registration, send email to inform if registration succeeded
    - Account Login/Logout
    - Password changing, find back
    - Download SSL keys
- Jupiter Notebook
    - Private Jupiter Notebook environment to run Python scripts
    - Private workspace
    - Submit job from Jupiter Notebook
- Job Dashboard
    - Job history and currently running jobs
    - Performance Monitoring
    - Quota Monitoring
- Datasets
    - Public Dataset viewing
    - Upload/Download private datasets
    - Share datasets
- Models
    - Upload/Download models file
    - Share/Publish Models
- Paddle Board
    - Training metrics visualization
        - cost
        - evaluator
        - user-defined metrics
- Serving
    - Submit serving instances
    - Deactivate serving
    - Serving performance monitoring

## Account Management

Account management page is designed to satisfy multi-tenant use cases. One account should have a unique account ID for login, and this account owns one access key to one unique [Kubernetes namespace](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/) cluster. Multiple users can log in to this account ID and operate jobs and data files. The only "master user" can do modifications like increase quota or manage account settings.

One example is [AWS IAM](https://aws.amazon.com/iam/?nc2=h_m1), but we can do more simpler than that.

The current implementation under this repo can only have one user for one Kubernetes namespace. We can implement multi-tenant in the near future.

Once a user logged in, s/he will be redirected to the "Job Dashboard" page.

## Jupiter Notebook

Start a [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) using image `docker.paddlepaddle.org/book` in the Kubernetes cluster and add an [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) endpoint when a user first enters the notebook page. This is already implemented.

<img src="pictures/notebook.png" width="500px" align="center">

Users can write a program in python in the web page and save their programs, which will be saved at cloud storage. Users also can run a script like below to submit a cluster training job:

```python
create_cloud_job(
    name,
    num_trainer,
    mem_per_trainer,
    gpu_per_trainer,
    cpu_per_trainer,
    num_ps,
    mem_per_ps,
    cpu_per_ps,
)
```

After this, there will be a job description and performance monitoring pages able to view at "Job Dashboard"

## Job Dashboard

### Job history and currently running jobs

A web page containing a table to list jobs satisfying user's filters. The user can only list jobs that were submitted by themselves.

| jobname | start time | age      | success | fails | actions |
| ------- | ---------- | -------- | ------- | ----- | ------- |
| test1   | 2017-01-01 | 17m      |    0    |   0   | stop/log/perf |

Users can filter the list using:

- status:  running/stopped/failed
- time:    job start time
- jobname: search by jobname

Viewing job logs:

Click the "log" button on the right of the job will pop up a console frame at the bottom of the page showing the tail of the job log, here shows the first pod's log. On the left side of pop up console frame, there should be a vertical list containing the list of pods, then click one of the pod, the console will show it's log.

### Performance Monitoring

A web page containing graphs monitoring job's resource usages according to time change:

- CPU usage
- GPU usage
- memory usage
- network bandwidth
- disk I/O

### Quota Monitoring

A web page displaying total quota and quota currently in use.

Also display total CPU time, GPU time in latest 1day, 1week, and 1month.

## Datasets and Models

Datasets and Models are quite the same, both like a simple file management and sharing service.

- file listing and viewing page
- Upload/Download page
- file sharing page

## Paddle Board

A web page containing graphs showing the internal status when the job is training, metrics like:

- cost(can be multiple costs)
- evaluator output
- user-defined metric

User can caculate metrics and define the graph like:

```python
cost = my_train_network()
evaluator = my_evaluator(output, label)
def my_metric_graph(output, label):
    metric = paddle.auc(output, label)
my_metric = my_metric_graph(output, label)
my_metric_value = output

draw_board(cost, evaluator)
draw_board(my_metric)
draw_board(my_metric_value)
```

Calling `draw_board` will output graph files on the distributed storage, and then the web page can load the data and refresh the graph.

## Serving

After training or uploading pre-trained models, a user can start a serving instance to serve the model as an inference HTTP service:

The Serving web page contains a table listing currently running serving instance and a "Launch" button to configure and start the serving program.

Click the "Launch" button in this web page will pop up a modal dialogue to configure the job:

1. model `tar.gz` files to the cloud.
1. inference network configuration in `.proto` format or user can also define the network in Python in the web page.
1. number of CPU/GPU resource in total to use for serving the model, the more resource there is, the more concurrent calls can be served.

Then click the "Launch" button on the pop-up dialogue, a "Kubernetes Deployment" will be created to serve the model. The current serving instances will be listed at the current page.

Users can also scale/shrink the resource used for the serving instances.