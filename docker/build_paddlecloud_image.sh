#!/bin/bash 

set -x

if [ -n $1 ]; then
  IMAGE_TAG="paddlepaddle/paddlecloud"
else
  IMAGE_TAG=$1
fi

docker build -t $IMAGE_TAG ../paddlecloud