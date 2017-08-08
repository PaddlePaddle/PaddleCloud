#!/bin/bash
paddlecloud submit -jobname fitaline-ft \
  -cpu 1 \
  -gpu 0 \
  -memory 1Gi \
  -parallelism 2 \
  -pscpu 1 \
  -pservers 2 \
  -psmemory 1Gi \
  -faulttolerant \
  -entry "python ./train.py" \
  ./fit_a_line

