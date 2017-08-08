#!/bin/bash
paddlecloud submit -jobname fitaline-ft \
  -cpu 1 \
  -gpu 0 \
  -memory 1Gi \
  -parallelism 2 \
  -pscpu 1 \
  -image bootstrapper:5000/yancey1989/paddlecloud-job-yx \
  -pservers 2 \
  -psmemory 1Gi \
  -faulttolerant \
  -entry "python ./train_ft.py" \
  ./fit_a_line

