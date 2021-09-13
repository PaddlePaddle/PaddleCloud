# Build the manager binary
FROM golang:1.15 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

ENV GOPROXY="https://goproxy.cn,direct"
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

COPY hack/install-go-licenses.sh hack/
COPY go-licenses.yaml $WORKDIR
RUN bash ./hack/install-go-licenses.sh 

COPY third_party/licenses/licenses.csv third_party/licenses/licenses.csv 
RUN go-licenses save third_party/licenses/licenses.csv --save_path /tmp/NOTICES 

FROM bitnami/minideb:stretch
WORKDIR /
COPY --from=builder /workspace/manager .
COPY third_party/licenses/licenses.csv /workspace/licenses.csv
COPY --from=builder /tmp/NOTICES /third_party/NOTICES
USER 65532:65532

ENTRYPOINT ["/manager"]
