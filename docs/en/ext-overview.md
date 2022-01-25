# Cache Component Overview

English | [简体中文](../zh_CN/ext-overview.md)

Due to the high scalability of cloud platforms, more and more machine learning tasks run in Kubernetes clusters. However, in the Kubernetes architecture, computing and storage are separated, and model training jobs need to access remote storage to obtain sample data, which brings high network IO overhead. Inspired by the [Fluid](https://github.com/fluid-cloudnative/fluid), we have implemented cache component in this project, which aims to accelerate the distributed training operations of Paddle on the cloud by caching sample data locally in the Kubernetes cluster.

## Features

1. The sample data and it's operations are abstracted into two custom resource: SampleSet and SampleJob. Users can simply manipulate the CRD to manage sample data without worrying about details such as remote storage and distributed cached data in cluster.
2. The cache component of Paddle Operator provides data acceleration for cloud applications by using [JuiceFS](https://github.com/juicedata/juicefs) as the cache engine, especially in the scene of massive and small files, which can be significantly improved.
3. The cache component can automatically warm up the sample data to the Kubernetes cluster, and it can schedule the training jobs to nodes with cache, this is greatly shortening the PaddleJob Execution time, and sometimes GPU resource utilization will be improved.

## Architecture

<div align="center">
  <img src="http://paddleflow-public.hkg.bcebos.com/Static/ext-arch.png" title="architecture" width="60%" height="60%" alt="">
</div>

The figure above is the architecture of Paddle Operator, which is built on Kubernetes and contains the following three main parts:

1. The first part is custom resource. Paddle Operator defines three CRDs. Users can write and modify the YAML files to manage training jobs and sample data sets.

- **PaddleJob**: is an abstraction of Paddle's distributed training job. It unifies the two distributed deep learning architecture of Parameter Server and Collective into one CRD. Users can easily run distributed training jobs in Kubernetes clusters by creating PaddleJob.
- **SampleSet**: is an abstraction of sample data sets. Data can come from distributed file systems such as OSS, HDFS or Ceph. Users can specify the number of partitions for caching data, the cache engine, the multi-level cache directory, and the configuration of cache nodes by SampleSet.
- **SampleJob**: defines multiple data operations for SampleSet, such as data synchronization, data prefetch, clearing cache, and remove expired data. Users can set the parameters of each data management command, and those tasks can run as cron job.

2. The second part is custom controller. In the operator framework of Kubernetes, the controller is used to monitor API object changes (such as creation, modification, deletion, etc.), and then determine which work should be performed.

- **PaddleJob Controller**: is used to manage the life cycle of PaddleJob, such as creating Pods for parameter servers or workers, and maintaining the replicas of each role.
- **SampleSet Controller**: is used to manage the life cycle of the SampleSet, such as create resource objects, create cache runtime services, and label the cache nodes.
- **SampleJob Controller**: is used to manage the life cycle of the SampleJob. It triggers the cache engine to perform data management operations asynchronously by request the cache runtime server.

3. The third part is cache engine. The cache engine consists of two components: the Cache Runtime Server and the JuiceFS CSI Driver. It provides the functions of data storage, caching, and management.

- **Cache Runtime Server**: is used to manager sample data, it receives data operation requests from SampleSet Controller and SampleJob Controller, and calls JuiceFS command line to finish data operations.
- **JuiceFS CSI Driver**: is used to storage and cache sample data, it caches the sample data locally in the cluster and mounts the data into the training worker of PaddleJob.

## Quick Start
Refer to the document [Quick Start for Cache Component](./ext-get-start.md) and try.

## Benchmark
Please refer to the [Benchmark of Cache Component](./ext-benchmark.md) for more information about benchmark.

## More Information
Please refer to the [API docs](./api_doc.md) for more information about custom resource definition.
