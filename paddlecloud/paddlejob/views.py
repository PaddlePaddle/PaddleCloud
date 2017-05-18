from django.http import HttpResponseRedirect, HttpResponse, JsonResponse
from django.contrib import messages
from django.conf import settings
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from . import PaddleJob, CephFSVolume
from rest_framework.authtoken.models import Token
from rest_framework import viewsets, generics, permissions
from rest_framework.response import Response
from rest_framework.views import APIView
import json
import utils
import notebook.utils
import logging

class JobsView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        List all jobs
        """
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
        job_list = client.BatchV1Api().list_namespaced_job(namespace)
        print job_list
        return Response(job_list.to_dict())

    def post(self, request, format=None):
        """
        Submit the PaddlePaddle job
        """
        #submit parameter server, it's Kubernetes ReplicaSet
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
        obj = json.loads(request.body)
        if not obj.get("topology"):
            return utils.simple_response(500, "no topology specified")

        paddle_job = PaddleJob(
            name = obj.get("name", "paddle-cluster-job"),
            job_package = obj.get("jobPackage", ""),
            parallelism = obj.get("parallelism", 1),
            cpu = obj.get("cpu", 1),
            memory = obj.get("memory", "1Gi"),
            pservers = obj.get("pservers", 1),
            pscpu = obj.get("pscpu", 1),
            psmemory = obj.get("psmemory", "1Gi"),
            topology = obj["topology"],
            gpu = obj.get("gpu", 0),
            image = obj.get("image", "yancey1989/paddle-job")
        )
        try:
            print paddle_job.new_pserver_job()
            ret = client.ExtensionsV1beta1Api().create_namespaced_replica_set(
                namespace,
                paddle_job.new_pserver_job(),
                pretty=True)
        except ApiException, e:
            logging.error("error submitting pserver job: %s " % e)
            return utils.simple_response(500, str(e))

        #submit trainer job, it's Kubernetes Job
        try:
            ret = client.BatchV1Api().create_namespaced_job(
                namespace,
                paddle_job.new_trainer_job(),
                pretty=True)
        except ApiException, e:
            logging.err("error submitting trainer job: %s" % e)
            return utils.simple_response(500, str(e))
        return utils.simple_response(200, "OK")

    def delete(self, request, format=None):
        """
        Kill a job
        """
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
        obj = json.loads(request.body)
        jobname = obj.get("jobname")
        if not jobname:
            return utils.simple_response(500, "must specify jobname")
        # FIXME: options needed: grace_period_seconds, orphan_dependents, preconditions
        # FIXME: cascade delteing
        delete_status = ""
        # delete job
        try:
            u_status = client.BatchV1Api().delete_namespaced_job(jobname, namespace, {})
            delete_status += u_status.status
        except ApiException, e:
            logging.error("error deleting job: %s, %s", jobname, str(e))

        # delete job pods
        try:
            job_pod_list = client.CoreV1Api().list_namespaced_pod(namespace, label_selector="job-name=%s"%jobname)
            logging.error("jobpodlist: %s", job_pod_list)
            for i in job_pod_list.items:
                u_status = client.CoreV1Api().delete_namespaced_pod(i.metadata.name, namespace, {})
                delete_status += u_status.status
        except ApiException, e:
            logging.error("error deleting job pod: %s", str(e))

        # delete pserver rs
        pserver_name = "-".join(jobname.split("-")[:-1])
        pserver_name += "-pserver"
        try:
            u_status = client.ExtensionsV1beta1Api().delete_namespaced_replica_set(pserver_name, namespace, {})
            delete_status += u_status.status
        except ApiException, e:
            logging.error("error deleting pserver: %s" % str(e))

        # delete pserver pods
        try:
            pserver_pod_label = "-".join(jobname.split("-")[:-1])
            job_pod_list = client.CoreV1Api().list_namespaced_pod(namespace, label_selector="paddle-job-pserver=%s"%pserver_pod_label)
            logging.error("pserver podlist: %s", job_pod_list)
            for i in job_pod_list.items:
                u_status = client.CoreV1Api().delete_namespaced_pod(i.metadata.name, namespace, {})
                delete_status += u_status.status
        except ApiException, e:
            logging.error("error deleting pserver pods: %s" % str(e))
        return utils.simple_response(200, delete_status)
