#!/bin/bash
cur_path="$(cd "$(dirname "$0")" && pwd -P)"

#base Docker image
if [ ! -n "$3" ]; then
  base_image=paddlepaddle/paddle:latest
else
  base_image=$3
fi

#PaddleCloud job Docker image
if [ ! -n "$4" ]; then
  pcloudjob_image=paddlepaddle/cloud-job
else
  pcloudjob_image=$4
fi

echo "base_image": $base_docker_image
echo "pcloudjob_image": $pcloudjob_image

#Build Docker Image
cat > Dockerfile <<EOF
FROM ${base_image}
ADD ./paddle_k8s /usr/bin
ADD ./k8s_tools.py /root/
RUN pip install -U kubernetes
CMD ["paddle_k8s"]
EOF

docker build -t $pcloudjob_image .
