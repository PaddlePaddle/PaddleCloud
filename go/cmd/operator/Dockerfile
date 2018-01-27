FROM golang:1.9
RUN go get github.com/Masterminds/glide
COPY cloud $GOPATH/src/github.com/PaddlePaddle/cloud 
WORKDIR $GOPATH/src/github.com/PaddlePaddle/cloud/go
RUN glide install --strip-vendor
RUN go build -o /usr/local/bin/operator github.com/PaddlePaddle/cloud/go/cmd/operator
CMD ["operator"]
