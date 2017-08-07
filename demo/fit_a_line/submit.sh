#!/bin/bash
paddlecloud submit -jobname fitaline \
  -cpu 1 \
  -gpu 0 \
  -memory 1Gi \
  -parallelism 2 \
  -pscpu 1 \
  -pservers 2 \
  -psmemory 1Gi \
  -entry "python ./train.py" \
  ./

