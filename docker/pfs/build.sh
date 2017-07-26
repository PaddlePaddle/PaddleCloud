cat > ./Dockerfile << EOF
FROM ubuntu:16.04

RUN apt-get update && \
    apt-get install -y wget git && \
    wget -O go.tgz https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go.tgz && \
    mkdir /root/gopath && \
    rm go.tgz

ENV GOROOT=/usr/local/go GOPATH=/root/gopath 
ENV PATH=\${PATH}:\${GOROOT}/bin

CMD ["sh", "-c", "cd /root/gopath/src/github.com/PaddlePaddle/cloud/go/cmd/pfsserver && go get ./... && go build"]
EOF

docker build .  -t  pfsserver:dev

rm -f Dockerfile
