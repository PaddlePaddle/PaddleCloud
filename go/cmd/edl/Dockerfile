FROM    golang:1.8
RUN     go get github.com/Masterminds/glide
RUN     apt-get update && apt-get install -y git
WORKDIR $GOPATH/src/github.com/PaddlePaddle
RUN     git clone https://github.com/PaddlePaddle/cloud.git
WORKDIR $GOPATH/src/github.com/PaddlePaddle/cloud/go
RUN     glide install --strip-vendor
RUN     go build -o /usr/local/bin/edl github.com/PaddlePaddle/cloud/go/cmd/edl
CMD     ["edl"]
