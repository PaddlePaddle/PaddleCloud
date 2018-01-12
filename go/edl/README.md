# TrainingJob controller and Autoscaler

Currently support only TPR, if update to CRD support need to run `deepcopy_gen.sh` to generate code for CRD struct deepcopy interface.

## Build

```bash
# must use --strip-vendor to avoid recursive vendoring
glide install --strip-vendor
go build -o path/to/output github.com/PaddlePaddle/cloud/go/cmd/autoscaler
```
