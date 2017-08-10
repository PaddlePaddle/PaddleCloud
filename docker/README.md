# Build PaddlePaddle Cloud Docker Images

## paddlecloud

This image is the paddlecloud REST API server image which serves API for submitting jobs, getting logs etc.

To build, simply run:

```bash
sh build_paddlecloud_image.sh
```

## paddlecloud-job
For the distributed training job on Kubernetes, we package Paddle binary files and some tools for Kubernetes into a runtime Docker image, the runtime Docker image gets scheduled by Kubernetes to run during training.

You can build CPU and GPU Docker image which based on different PaddlePaddle product Docker image:

```bash
./build_docker.sh <base-docker_image> <runtime-docker-image>
```

- Build CPU runtime Docker image

```bash
./build_docker.sh paddlepaddle/paddle:0.10.0 paddlepaddle/paddlecloud-job:0.10.0
```

- Build GPU runtime Docker image

```bash
./build_docker.sh paddlepaddle/paddle:0.10.0-gpu paddlepaddle/paddlecloud-job:0.10.0-gpu
```

## pfsserver

The "pfsserver" image contains paddle cloud filesystem server. To build, simply run:

```bash
sh build_pfs_image.sh
```