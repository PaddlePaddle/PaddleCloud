## Build EDL Controller

To build EDL Controller you have to setup [Go development environment](https://golang.org/doc/install#install) and install
[glide](https://github.com/Masterminds/glide).

After that you can run the following commands to finish the build:

```bash
cd go
glide install --strip-vendor
cd go/cmd/edl
go build
```

## Build EDL Controller Docker Image

The above step will generate a binary file named `edl` which should
run as a daemon process on the Kubernetes cluster. In order to achieve
this, we have to build one Docker image containing `edl` binary:

```bash
cd go/cmd/edl
docker build -t [your image tag] .
```
