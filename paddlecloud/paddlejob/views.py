from django.http import HttpResponseRedirect, HttpResponse, JsonResponse
from django.contrib import messages
from django.conf import settings
from kubernetes import client, config
from . import PaddleJob, CephFSVolume
from rest_framework.authtoken.models import Token
from rest_framework import viewsets, generics, permissions
from rest_framework.response import Response
from rest_framework.views import APIView
import json

class JobsView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        List all jobs
        """
        return Response(utils.simple_content(200, "OK"))

    def post(self, request, format=None):
        """
        Submit the PaddlePaddle job
        """
        #submit parameter server, it's Kubernetes ReplicaSet
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
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
                namespace,
                paddle_job.new_pserver_job(),
                pretty=True)
        except ApiException, e:
            logging.error("Exception when submit pserver job: %s " % e)
            return utils.simple_response(500, str(e))

        #submit trainer job, it's Kubernetes Job
        try:
            ret = client.BatchV1Api().create_namespaced_job(
                namespace,
                paddle_job.new_trainer_job(),
                pretty=True)
        except ApiException, e:
            logging.err("Exception when submit trainer job: %s" % e)
            return utils.simple_response(500, str(e))
        return utils.simple_response(200, "OK")

    def delete(self, request, format=None):
        """
        Kill a job
        """
        return utils.simple_response(200, "OK")
