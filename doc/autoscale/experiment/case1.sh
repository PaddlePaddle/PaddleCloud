#!/bin/bash
function start() {
    for ((pass=0; pass<$PASSES; pass++))
    do
        echo "Run pass "$pass
        PASSE_NUM=$pass AUTO_SCALING=$AUTO_SCALING JOB_COUNT=$JOB_COUNT JOB_NAME=$JOB_NAME \
            stdbuf -oL nohup python python/main.py run_case1 &> ./out/${JOB_NAME}-case1-pass$pass.log &
        sleep 5
        for ((j=0; j<$JOB_COUNT; j++)) 
        do 
            if [ "$AUTO_SCALING" == "ON" ]
            then
                submit_ft_job $JOB_NAME$j 5
                cat k8s/trainingjob.yaml.tmpl | sed "s/<jobname>/$JOB_NAME$j/g" | kubectl create -f -
            else
                submit_ft_job $JOB_NAME$j 30
            fi
            sleep 5
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
            FILE=./out/${JOB_NAME}-case1-pass$pass.csv
            if [ ! -f $FILE ]; then
                echo "waiting for collector exit, generated file " $FILE
                sleep 5
            fi
            break
        done
    done
    python python/main.py merge_case1_reports
    rm -f ./out/${JOB_NAME}-pass*
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
        sleep 2
    done
}