#PaddleCloud server
set -e 

cat > python/paddlecloud/Dockerfile << EOF
FROM python:2.7.13-alpine
RUN apk add --update nodejs openssl gcc mysql-dev musl-dev linux-headers mailx

ADD ./ /pcloud
RUN cd /pcloud && \
rm -rf node_modules && npm run clean && \
npm install && pip install -r requirements.txt && npm run build && \
npm run copy:fonts && npm run copy:images && npm run copy:fonts && npm run copy:images && \
npm run optimize
WORKDIR /pcloud

# TODO
CMD ["sh", "-c", "sleep 60 ; ./manage.py migrate; ./manage.py loaddata sites; ./manage.py runserver 0.0.0.0:$PORT"]
EOF
pushd python/paddlecloud
docker build . -t paddlepaddle/paddlecloud
pod

glide install -v

#pfserver
env GOOS=linux GOARCH=amd64 go build go/cmd/pserver
cat > go/Dockerfile << EOF
FROM ubuntu:16.04
ADD ./cmd/pfsserver/pfsserver /usr/local/bin/
CMD ["/pfsserver/pfsserver", "-tokenuri", "http://paddle-cloud-service:8000", "-logtostderr=true", "-v=4"]
EOF
pushd go
docker build . -t paddlepaddle/pfsserver
popd


#edl
env GOOS=linux GOARCH=amd64 go build go/cmd/edl
cat > go/Dockerfile << EOF
FROM ubuntu:16.04
ADD ./cmd/edl/edl /usr/local/bin/
CMD     ["edl"]
EOF
pushd go
docker build . -t paddlepaddle/edl
popd


#Cloud job runtime image
cat > docker/Dockerfile << EOF
FROM paddlepaddle/paddle:latest
RUN pip install -U kubernetes opencv-python &&   apt-get update -y &&   apt-get install -y iputils-ping libgtk2.0-dev 
ADD ./paddle_k8s /usr/local/bin
ADD ./k8s_tools.py /root/

CMD ["paddle_k8s"]
EOF
pushd docker
docker build . -t paddlepaddle/paddlecloud-job
popd


