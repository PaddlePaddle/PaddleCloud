# PaddlePaddle EDL: Elastic Deep Learning

While many hardware and software manufacturers are working on improving the running time of deep learning jobs, EDL optimizes (1) the global utilization of the cluster and (2) the waiting time of job submitters.

For more about the project EDL, please refer to this [invited blog post](http://blog.kubernetes.io/2017/12/paddle-paddle-fluid-elastic-learning.html) on the Kubernetes official blog.

EDL includes two parts:

1. a Kubernetes controller for the elastic scheduling of distributed deep learning jobs, and

1. making PaddlePaddle a fault-tolerable deep learning framework.  This directory contains the Kubernetes controller.  For more information about fault-tolerance, please refer to this [PaddlePaddle design doc](https://github.com/PaddlePaddle/Paddle/tree/develop/doc/design/cluster_train).

We deployed EDL on a real Kubernetes cluster, dlnel.com, opened for graduate students of Tsinghua University.  The performance test report of EDL on this cluster is [here](https://github.com/PaddlePaddle/cloud/blob/develop/doc/autoscale/experiment/README.md).


## Build

```bash
glide install --strip-vendor
go build -o path/to/output github.com/PaddlePaddle/cloud/go/cmd/edl
```
