#!/bin/bash
DEFAULT_JOBNAME_PREFIX="mnist"
CPU="10"
MEMORY="8Gi"
PSCPU="6"
PSMEMORY="5Gi"
JOB_COUNT=${JOB_COUNT:-1}
AUTO_SCALING=${AUTO_SCALING:-OFF}
PASSES=${PASSES:-1}
DETAILS=${DETAILS:-OFF}

function submit_ft_job() {
   paddlecloud submit -jobname $1 \
        -cpu $CPU \
        -gpu 0 \
        -memory $MEMORY \
        -parallelism $2 \
        -pscpu $PSCPU \
        -pservers 10 \
        -psmemory $PSMEMORY \
        -entry "python ./train_ft.py train" \
        -faulttolerant \
        -image registry.baidu.com/paddlepaddle/paddlecloud-job:yx_exp \
        ./mnist
    sleep 2
}

function usage() {
    echo "usage: control_case1.sh <action>"
    echo "  action[required]: str[start|stop], will start or stop all the jobs."
    echo "env var:"
    echo "  JOB_COUNT[optional]:             int, The number of submiting jobs, defualt is 1."
    echo "  AUTO_SCALING[optional]:   str[ON|OFF], whether a fault-tolerant job,\
default is OFF."
    echo "  PASSES[optional]:           int, The number of run passes."
    echo "  DETAILS[optional:           str[ON|OFF], print detail monitor information."
}

function start() {
    echo "JOB_COUNT: "$JOB_COUNT
    echo "AUTO_SCALING: "$AUTO_SCALING
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
        PASSE_NUM=$pass AUTO_SCALING=$AUTO_SCALING JOB_COUNT=$JOB_COUNT \
            stdbuf -oL nohup python python/main.py run_case1 &> ./out/pass$pass.log &

        for ((j=0; j<$JOB_COUNT; j++)) 
        do 
            if [ "$AUTO_SCALING" == "ON" ]
            then
                submit_ft_job $DEFAULT_JOBNAME_PREFIX$j 5
                cat k8s/trainingjob.yaml.tmpl | sed "s/<jobname>/$DEFAULT_JOBNAME_PREFIX$j/g" | kubectl create -f -
            else
                submit_ft_job $DEFAULT_JOBNAME_PREFIX$j 20
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
        if [ "$AUTO_SCALING" == "ON" ]
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