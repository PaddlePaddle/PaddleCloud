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

#Build Python Package
docker run --rm -it -v $PWD:/cloud $base_image \
  bash -c "cd /cloud/python && python setup.py bdist_wheel"

#Build Docker Image
cat > Dockerfile <<EOF
FROM ${base_image}
RUN pip install -U kubernetes && apt-get install -y iputils-ping
ADD ./paddle_k8s /usr/bin
ADD ./k8s_tools.py /root/
ADD ./python/dist/pcloud-0.1.1-py2-none-any.whl /tmp/
#RUN pip install /tmp/pcloud-0.1.1-py2-none-any.whl && \
#  rm /tmp/pcloud-0.1.1-py2-none-any.whl
RUN pip install /tmp/pcloud-0.1.1-py2-none-any.whl 

CMD ["paddle_k8s"]
EOF

docker build -t $pcloudjob_image .
