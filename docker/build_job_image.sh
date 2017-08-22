#!/bin/bash
cur_path="$(cd "$(dirname "$0")" && pwd -P)"
WHL_URL="https://pypi.python.org/packages/3b/74/2ae16bfac4bc3947f20347852e111a64758947dcbd24b2bdb8540319624c/paddlepaddle-0.10.0-cp27-cp27mu-manylinux1_x86_64.whl#md5=7550b01af5f75939ea72346de8735c6d"

USAGE="$0 cpu|gpu [paddle cloud job image name]"

if [ -n "$1" ]; then
  if [ "$1" == "cpu" ]; then
    BASE_IMAGE="ubuntu:16.04"
  elif [ "$1" == "gpu" ]; then
    BASE_IMAGE="nvidia/cuda:8.0-cudnn5-runtime-ubuntu16.04"
  else
    echo "first argument must be cpu or gpu"
    exit 1
  fi
else
  echo "usage: " $USAGE
  exit 1
fi

echo ${BASE_IMAGE}

#PaddleCloud job Docker image
if [ ! -n "$2" ]; then
  pcloudjob_image=paddlepaddle/cloud-job
else
  pcloudjob_image=$2
fi

# find if local whl package exists
WHL_PACKAGE=$(ls paddlepaddle*.whl)
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
FROM $BASE_IMAGE
RUN /bin/bash -c 'sed -i 's#http://archive.ubuntu.com/ubuntu#http://mirrors.ustc.edu.cn/ubuntu#g' /etc/apt/sources.list;'
RUN apt-get update -y && \
  apt-get install -y python-pip iputils-ping libgtk2.0-dev && \
  apt-get install -f -y && apt-get clean -y && \
  pip install -U kubernetes opencv-python
ADD $WHL_PACKAGE /
RUN pip install /$WHL_PACKAGE
ADD ./paddle_k8s /usr/bin
ADD ./k8s_tools.py /root/

CMD ["paddle_k8s"]
EOF

docker build -t $pcloudjob_image .
