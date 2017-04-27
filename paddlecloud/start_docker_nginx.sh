#!/bin/bash

docker run \
-v $PWD/nginx.conf:/etc/nginx/nginx.conf \
-v $PWD/paddlecloud.crt:/etc/nginx/paddlecloud.crt \
-v $PWD/paddlecloud.key:/etc/nginx/paddlecloud.key \
-d -p 443:443 -p 80:80 nginx
