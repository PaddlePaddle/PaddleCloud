#!/bin/bash
function submit_general_job() {
    paddlecloud submit -jobname $1 \
        -cpu 10 \
        -gpu 0 \
        -memory 8Gi \
        -parallelism 20 \
        -pscpu 6 \
        -pservers 10 \
        -psmemory 5Gi \
        -entry "python ./train.py train" \
        ./mnist
}

function submit_ft_job() {
   paddlecloud submit -jobname $1 \
        -cpu 10 \
        -gpu 0 \
        -memory 8Gi \
        -parallelism 20 \
        -pscpu 6 \
        -pservers 10 \
        -psmemory 5Gi \
        -entry "python ./train_ft.py train" \
        -faulttolerant \
        ./mnist 
}

function usage() {
    echo "usage: control_case1.sh <action> <jobs> <fault-tolerant>"
    echo "  action[required]:         str[start|stop], will start or stop the jobs."
    echo "  jobs[required]:           int, specify the job count that will be submited, \
default is 1."
    echo "  fault-tolerant[optional]  str[ON|OFF], whether submit a fault-tolerant \
mode job, default is OFF."
}

function start() {
    for ((i=0; i<$JOBS; i++)) 
    do 
        if [ "$FAULT_TOLERANT" == "ON" ]
        then
            #submit_ft_job mnist$i $JOBS
            echo mnist$i $JOBS
        else
            submit_general_job mnist$i $JOBS
        fi
        sleep 2
    done
}

function stop() {
    for ((i=0; i<$JOBS; i++))
    do
        echo "paddlecloud kill" mnist$i
        paddlecloud kill mnist$i
        sleep 2
    done
}

if [ -z $1 ] || [ -z $2 ]
then
    usage
    exit 0
fi

ACTION=${1}
JOBS=${2}
FAULT_TOLERANT=${3:-OFF}

case $ACTION in 
    start)
        start
        ;;
    stop)
        stop
        ;;
    --help)
        usage
        ;;
    *)
        usage
        ;;
esac