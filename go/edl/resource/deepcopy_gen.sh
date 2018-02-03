#!/bin/bash

# FIXME: should run this script before build, when using api >= 1.7
#        api == 1.6 is not compatible with this deep copy code generations.

go get -u github.com/kubernetes/gengo/examples/deepcopy-gen &&
go run $GOPATH/src/github.com/kubernetes/gengo/examples/deepcopy-gen/main.go -i github.com/PaddlePaddle/cloud/go/edl/resource -O generated.deepcopy
