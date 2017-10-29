#!/bin/bash
DEFAULT_JOBNAME_PREFIX="mnist"
CPU="10"
MEMORY="8Gi"
PSCPU="6"
PSMEMORY="5Gi"
JOB_COUNT=${JOB_COUNT:-1}
FAULT_TOLERANT=${FAULT_TOLERANT:-OFF}
PASSES=${PASSES:-1}
DETAILS=${DETAILS:-OFF}

function submit_general_job() {
    paddlecloud submit -jobname $1 \
        -cpu $CPU \
        -gpu 0 \
        -memory $MEMORY \
        -parallelism 20 \
        -pscpu $PSCPU \
        -pservers 10 \
        -psmemory $PSMEMORY \
        -entry "python ./train.py train" \
        ./mnist
}

function submit_ft_job() {
   paddlecloud submit -jobname $1 \
        -cpu $CPU \
        -gpu 0 \
        -memory $MEMORY \
        -parallelism 2 \
        -pscpu $PSCPU \
        -pservers 10 \
        -psmemory $PSMEMORY \
        -entry "python ./train_ft.py train" \
        -faulttolerant \
        ./mnist
    sleep 2
    cat k8s/trainingjob.yaml.tmpl | sed "s/<jobname>/$1/g" | kubectl create -f -
}

function usage() {
    echo "usage: control_case1.sh <action>"
    echo "  action[required]: str[start|stop], will start or stop all the jobs."
    echo "env var:"
    echo "  JOB_COUNT[optional]:             int, The number of submiting jobs, defualt is 1."
    echo "  FAULT_TOLERANT[optional]:   str[ON|OFF], whether a fault-tolerant job,\
default is OFF."
    echo "  PASSES[optional]:           int, The number of run passes."
    echo "  DETAILS[optional:           str[ON|OFF], print detail monitor information."
}

function start() {
    echo "JOB_COUNT: "$JOB_COUNT
    echo "FAULT_TOLERANT: "$FAULT_TOLERANT
    echo "PASSES: "$PASSES
    echo "DETAILS: "$DETAILS
    # Following https://apple.stackexchange.com/a/193156,
    # we need to set the envrionment var for MacOS
    if [ $(uname) == "Darwin" ]
    then
        export PATH=/usr/local/opt/coreutils/libexec/gnubin:$PATH
    fi
    rm -rf ./out > /dev/null
    mkdir ./out > /dev/null
    rm -f ./experiment.log > /dev/null
    for ((pass=0; pass<$PASSES; pass++))
    do
        echo "Run pass "$pass
        PASSE_NUM=$pass FAULT_TOLERANT=$FAULT_TOLERANT JOB_COUNT=$JOB_COUNT \
            stdbuf -oL nohup python python/main.py run_case1 &> ./out/pass$pass.log &

        for ((j=0; j<$JOB_COUNT; j++)) 
        do 
            if [ "$FAULT_TOLERANT" == "ON" ]
            then
                submit_ft_job $DEFAULT_JOBNAME_PREFIX$j $JOB_COUNT
            else
                submit_general_job $DEFAULT_JOBNAME_PREFIX$j $JOB_COUNT
            fi
            sleep 2
        done
        # waiting for all jobs finished
        python python/main.py wait_for_finished
        # stop all jobs
        stop
        # waiting for all jobs have been cleaned
        python python/main.py wait_for_cleaned
        # waiting for the data collector exit
        while true
        do
            FILE=./out/$DEFAULT_JOBNAME_PREFIX-pass$pass
            if [ ! -f $FILE ]; then
                echo "waiting for collector exit, generated file " $FILE
                sleep 5
            fi
            break
        done
    done
    python python/main.py generate_report
    rm -f ./out/%DEFAULT_JOBNAME_PREFIX-pass*
    
}

function stop() {
    for ((i=0; i<$JOB_COUNT; i++))
    do
        echo "kill" $DEFAULT_JOBNAME_PREFIX$i
        paddlecloud kill $DEFAULT_JOBNAME_PREFIX$i
        if [ "$FAULT_TOLERANT" == "ON" ]
        then
           cat k8s/trainingjob.yaml.tmpl | sed "s/<jobname>/$DEFAULT_JOBNAME_PREFIX$i/g" | kubectl delete -f - 
        fi
        sleep 2
    done
}

ACTION=${1}

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