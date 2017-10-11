go get -u k8s.io/gengo
go build -o /tmp/deepcopy-gen k8s.io/gengo/examples/deepcopy-gen
/tmp/deepcopy-gen -i github.com/PaddlePaddle/cloud/go/controller/k8s -O zz_generated.deepcopy 2> /dev/null
