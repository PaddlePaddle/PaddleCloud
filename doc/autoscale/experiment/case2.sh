#!/bin/bash

function usage() {
    echo "usage: control_case1.sh <action>"
    echo "  action[required]: str[start|stop], will start or stop all the jobs."
    echo "env var:"
    echo "  JOB_COUNT[optional]:        int, The number of submiting jobs, defualt is 1."
    echo "  PASSES[optional]:           int, The times of the experiment."
    echo "  DETAILS[optional]:          str[ON|OFF], print detail monitor information."
    echo "  NGINX_REPLICAS[optional]:   int, The number of Nginx pods, default is 10."
}

function start() {
    # submit Nginx deployment
    cat k8s/nginx_deployment.yaml.tmpl | sed "s/<nginx-replicas>/$NGINX_REPLICAS/g" | kubectl create -f -
    sleep 5
    PASSE_NUM=0 AUTO_SCALING=$AUTO_SCALING JOB_COUNT=$JOB_COUNT JOB_NAME=$JOB_NAME\
            stdbuf -oL nohup python python/main.py run_case2 &> ./out/${JOB_NAME}-case2.log &

    # submit the auto-scaling training jobs
    for ((j=0; j<$JOB_COUNT; j++))
    do
        if [ "$AUTO_SCALING" == "ON" ]
        then
            submit_ft_job $JOB_NAME$j 15
            cat k8s/trainingjob.yaml.tmpl | sed "s/<jobname>/$JOB_NAME$j/g" | kubectl create -f - 
        else
            submit_ft_job $JOB_NAME$j 15
        fi
    sleep 5
    done
    kubectl scale deployment/nginx --replicas=400
    sleep 30
    kubectl scale deployment/nginx --replicas=200
    sleep 30
    kubectl scale deployment/nginx --replicas=100
    sleep 30
    kubectl scale deployment/nginx --replicas=200
    sleep 30
    kubectl scale deployment/nginx --replicas=400
    # waiting for all jobs finished
    python python/main.py wait_for_finished
    # stop all jobs
    stop
    # waiting for all jobs have been cleaned
    python python/main.py wait_for_cleaned
    # waiting for the data collector exit
    while true
    do
        FILE=./out/$JOB_NAME-case1-pass0.csv
        if [ ! -f $FILE ]; then
            echo "waiting for collector exit, generated file " $FILE
            sleep 5
            continue
        fi
        break
    done
    python python/main.py merge_case1_reports
}

function stop() {
    for ((i=0; i<$JOB_COUNT; i++))
    do
        echo "kill" $JOB_NAME$i
        paddlecloud kill $JOB_NAME$i
        if [ "$AUTO_SCALING" == "ON" ]
        then
            cat k8s/trainingjob.yaml.tmpl | sed "s/<jobname>/$JOB_NAME$i/g" | kubectl delete -f - 
        fi
    done
    cat k8s/nginx_deployment.yaml.tmpl | sed "s/<nginx_replicas>/$NGINX_REPLICAS/g" | kubectl delete -f -
}

