FROM ubuntu:16.04
RUN wget -qO- https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz | \
    tar -xz -C /usr/local && \
    mkdir /root/gopath && \
    mkdir /root/gopath/bin && \
    mkdir /root/gopath/src
ENV GOROOT=/usr/local/go GOPATH=/root/gopath
# should not be in the same line with GOROOT definition, otherwise docker build could not find GOROOT.
ENV PATH=${PATH}:${GOROOT}/bin:${GOPATH}/bin

RUN     go get github.com/Masterminds/glide
RUN     glide install --strip-vendor
