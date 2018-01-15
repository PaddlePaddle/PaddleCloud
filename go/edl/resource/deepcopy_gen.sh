#!/bin/bash

# FIXME: should run this script before build, when using api >= 1.7
#        api == 1.6 is not compatible with this deep copy code generations.

go get -u k8s.io/gengo
go build -o /tmp/deepcopy-gen k8s.io/gengo/examples/deepcopy-gen
/tmp/deepcopy-gen -i github.com/PaddlePaddle/cloud/go/edl/resource -O zz_generated.deepcopy 2> /dev/null
