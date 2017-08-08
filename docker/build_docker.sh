#!/bin/bash
cur_path="$(cd "$(dirname "$0")" && pwd -P)"

#base Docker image
if [ ! -n "$1" ]; then
  base_image=paddlepaddle/paddle:latest
else
  base_image=$1
fi

#PaddleCloud job Docker image
if [ ! -n "$2" ]; then
  pcloudjob_image=paddlepaddle/cloud-job
else
  pcloudjob_image=$2
fi

echo "base_image": $base_image
echo "pcloudjob_image": $pcloudjob_image

#Build Docker Image
cat > Dockerfile <<EOF
FROM ${base_image}
RUN pip install -U kubernetes opencv-python && \
  apt-get update -y && \
  apt-get install -y iputils-ping libgtk2.0-dev 
ADD ./paddle_k8s /usr/bin
ADD ./k8s_tools.py /root/

CMD ["paddle_k8s"]
EOF

docker build -t $pcloudjob_image .
