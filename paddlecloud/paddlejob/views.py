from django.http import HttpResponseRedirect, HttpResponse, JsonResponse
from django.contrib import messages
from django.conf import settings
from kubernetes import client, config
from . import PaddleJob, CephFSVolume

import json

@login_required
def jobs_handler(request):
    if request.method == "POST":
        submit_paddle_job(request)
    elif request.method == "GET":
        list_jobs(request)
    elif request.method == "DELETE":
        delete_job(request)
    else:
        return utils.simple_response(404, "Does not supports method: " % request.method)

def submit_job(request):
    #submit parameter server, it's Kubernetes ReplicaSet
    username = request.user.username
    obj = json.loads(request.body)
    paddle_job = PaddleJob(
        name = obj.get("name", ""),
        job_package = obj.get("jobPackage", ""),
        parallelism = obj.get("parallelism", ""),
        cpu = obj.get("cpu", 1),
        gpu = obj.get("gpu", 0),
        memory = obj.get("memory", "1Gi")
        topology = obj["topology"]
        image = obj.get("image", "yancey1989/paddle-job")
    )
    try:
        ret = client.ExtensionsV1beta1Api().create_namespaced_replica_set(
            username,
            paddle_job.new_pserver_job(),
            pretty=True)
    except ApiException, e:
        logging.error("Exception when submit pserver job: %s " % e)
        return utils.simple_response(500, str(e))

    #submit trainer job, it's Kubernetes Job
    try:
        ret = client.BatchV1Api().create_namespaced_job(
            username,
            paddle_job.new_trainer_job(),
            pretty=True)
    except ApiException, e:
        logging.err("Exception when submit trainer job: %s" % e)
        return utils.simple_response(500, str(e))
    return utils.simple_response(200, "OK")

def list_jobs(request):
    return utils.simple_response(200,
                                 [{
                                 "name": "paddle-job-b82x",
                                 "jobPackage": "/pfs/datacenter1/home/user1/job_word_emb",
                                 "parallelism": 3,
                                 "cpu": 1,
                                 "gpu": 1,
                                 "memory": "1Gi",
                                 "pservers": 3,
                                 "pscpu": 1,
                                 "psmemory": "1Gi",
                                 "topology": "train.py",
                                 "status": {
                                   "active": 0,
                                   "completionTime": "2017-05-15 13:33:23",
                                   "succeeded": 3,
                                   "startTime": "2017-05-15 12:33:53"
                                 }
                               }])

def delete_job(request):
    return utils.simple_response(200, "OK")
