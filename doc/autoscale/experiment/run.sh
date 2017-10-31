#!/bin/bash
CPU="10"
MEMORY="8Gi"
PSCPU="6"
PSMEMORY="5Gi"
JOB_NAME=${JOB_NAME:-mnist}
JOB_COUNT=${JOB_COUNT:-1}
PASSES=${PASSES:-1}
DETAILS=${DETAILS:-OFF}
NGINX_REPLICAS=${NGINX_REPLICAS:-400}


ACTION=${1}
CASE=${2}

# job configuration
CPU="10"
MEMORY="8Gi"
PSCPU="6"
PSMEMORY="5Gi"

function submit_ft_job() {
   paddlecloud submit -jobname $1 \
        -cpu $CPU \
        -gpu 0 \
        -memory $MEMORY \
        -parallelism $2 \
        -pscpu $PSCPU \
        -pservers 10 \
        -psmemory $PSMEMORY \
        -entry "sleep 180" \
        -faulttolerant \
        -image registry.baidu.com/paddlepaddle/paddlecloud-job:yx_exp \
        ./mnist
    #-entry "python ./train_ft.py train" \
}

function print_env() {
    echo "JOB_NAME: "$JOB_NAME
    echo "JOB_COUNT: "$JOB_COUNT
    echo "DETAILS: "$DETAILS
    echo "NGINX_REPLICAS: "$NGINX_REPLICAS
    echo "AUTO_SCALING: "$AUTO_SCALING
    echo "PASSES: "$PASSES
}

function prepare() {
    print_env
    # Following https://apple.stackexchange.com/a/193156,
    # we need to set the envrionment var for MacOS
    if [ $(uname) == "Darwin" ]
    then
        export PATH=/usr/local/opt/coreutils/libexec/gnubin:$PATH
    fi
    rm -rf ./out > /dev/null
    mkdir ./out > /dev/null
}
function usage() {
    echo "Usage: run.sh <action> <case>"
    echo "  action[required]:   str[start|stop], will start or stop all the jobs."
    echo "  case[required]:     str[case1|case2], run or stop the specify case."
    echo "env var:"
    echo "  JOB_COUNT[optional]:        int, The number of submiting jobs, defualt is 1."
    echo "  JOB_NAME[optional]:         str, The job name."
    echo "  NGINX_REPLICAS[optional]    int, The replicas of Nginx Deployment, default is 10."
    echo "  AUTO_SCALING[optional]:     str[ON|OFF], whether a auto-scaling training job,\
default is OFF."
    echo "  PASSES[optional]:           int, The times of the experiment."
    echo "  DETAILS[optional:           str[ON|OFF], print detail monitor information."
}



if [ -z $1 ] || [ -z $2 ]; then
    usage
    exit 0
fi

if [ $CASE == "case1" ]; then
    source ./case1.sh
elif [ $CASE == "case2" ]; then
    source ./case2.sh
else
    usage
    exit 0
fi


case $ACTION in 
    start)
        prepare
        start
        ;;
    stop)
        prepare
        stop
        ;;
    --help)
        usage
        ;;
    *)
        usage
        ;;
esac