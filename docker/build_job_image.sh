#!/bin/bash
cur_path="$(cd "$(dirname "$0")" && pwd -P)"
WHL_URL="https://pypi.python.org/packages/3b/74/2ae16bfac4bc3947f20347852e111a64758947dcbd24b2bdb8540319624c/paddlepaddle-0.10.0-cp27-cp27mu-manylinux1_x86_64.whl#md5=7550b01af5f75939ea72346de8735c6d"

#PaddleCloud job Docker image
if [ ! -n "$1" ]; then
  pcloudjob_image=paddlepaddle/cloud-job
else
  pcloudjob_image=$1
fi

# find if local whl package exists
WHL_PACKAGE=$(ls paddlepaddle-*.whl)
if [ $? -ne 0 ]; then
  wget -q $WHL_URL
  if [ $? -ne 0 ]; then
    exit 1
  fi
  WHL_PACKAGE=$(ls paddlepaddle-*.whl)
fi

echo "pcloudjob_image": $pcloudjob_image

#Build Docker Image
cat > Dockerfile <<EOF
FROM python:2.7-stretch
ADD $WHL_PACKAGE /
RUN pip install -U kubernetes opencv-python && \
  pip install /$WHL_PACKAGE && \
  apt-get update -y && \
  apt-get install -y iputils-ping libgtk2.0-dev 
ADD ./paddle_k8s /usr/bin
ADD ./k8s_tools.py /root/

CMD ["paddle_k8s"]
EOF

docker build -t $pcloudjob_image .
