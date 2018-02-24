# Build Components

This article contains instructions of build all the components
of PaddlePaddle Cloud and how to pack them into Docker images
so that server-side components can run in the Kubernetes cluster.

- Server-side components:
  - Cloud Server (written in Python, only need to pack to image)
  - EDL Controller
  - PaddleFS (PFS) Server
  - PaddlePaddle Cloud Job runtime Docker image
- Client side component:
  - Command line client

Before starting, you have to setup [Go development environment](https://golang.org/doc/install#install) and install
[glide](https://github.com/Masterminds/glide).

## Build EDL Controller

Run the following commands to finish the build:

```bash
cd go
glide install --strip-vendor
cd go/cmd/edl
go build
```

The above step will generate a binary file named `edl` which should
run as a daemon process on the Kubernetes cluster.


## Build paddlectl client

Run the following command to build paddlectl binary.

```bash
cd go/cmd/paddlectl
go build
```

Then file `paddlectl` will be generated under the current directory.


# Build Docker Images for Server side

## EDL Controller Image

After you've built edl binary, run the following command to build the
corresponding Docker image.

```bash
cd go/cmd/edl
docker build -t [your image tag] .
```

## Cloud Server Image

This image is used to start the Cloud Server in Kubernetes cluster. To
build, just run:

```bash
cd python/paddlecloud
docker build -t [your image tag] .
```

## PaddleFS (PFS) Server Image

To build PaddleFS image, just run:

```bash
cd docker/pfs
sh build.sh
```

## Cloud Job runtime Docker image

To build job runtime image which do the actual cloud job running, run:

```bash
cd docker
sh build_docker.sh [base paddlepaddle image] [target image]
```

- base paddlepaddle image is PaddlePaddle docker runtime image, like
  paddlepaddle/paddle:latest-gpu
- target image is the cloud job image name you want to build.
