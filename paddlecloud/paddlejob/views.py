from django.http import HttpResponseRedirect, HttpResponse, JsonResponse
from django.contrib import messages
from django.conf import settings
from kubernetes import client, config
from kubernetes.client.rest import ApiException
from . import PaddleJob
from rest_framework.authtoken.models import Token
from rest_framework import viewsets, generics, permissions
from rest_framework.response import Response
from rest_framework.views import APIView
import json
import utils
import notebook.utils
import logging
import volume

class JobsView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        List all jobs
        """
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
        job_list = client.BatchV1Api().list_namespaced_job(namespace)
        return Response(job_list.to_dict())

    def post(self, request, format=None):
        """
        Submit the PaddlePaddle job
        """
        #submit parameter server, it's Kubernetes ReplicaSet
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
        obj = json.loads(request.body)
        topology = obj.get("topology", "")
        entry = obj.get("entry", "")
        if not topology and not entry:
            return utils.simple_response(500, "no topology or entry specified")
        if not obj.get("datacenter"):
            return utils.simple_response(500, "no datacenter specified")
        dc = obj.get("datacenter")
        volumes = []
        cfg = settings.DATACENTERS.get(dc, None)
        if cfg and cfg["fstype"] == "hostpath":
            volumes.append(volume.get_volume_config(
                fstype = "hostpath",
                name = dc.replace("_", "-"),
                mount_path = cfg["mount_path"] % (dc, username),
                host_path = cfg["host_path"]
            ))
        elif cfg and cfg["fstype"] == "cephfs":
            volumes.append(volume.get_volume_config(
                fstype = "cephfs",
                name = dc.replace("_", "-"),
                monitors_addr = cfg["monitors_addr"],
                secret = cfg["secret"],
                user = cfg["user"],
                mount_path = cfg["mount_path"] % (dc, username),
                cephfs_path = cfg["cephfs_path"] % username,
                admin_key = cfg["admin_key"]
            ))
        else:
            pass

        registry_secret = settings.JOB_DOCKER_IMAGE.get("registry_secret", None)
        # get user specified image
        job_image = obj.get("image", None)
        gpu_count = obj.get("gpu", 0)
        # use default images
        if not job_image :
            if gpu_count > 0:
                job_image = settings.JOB_DOCKER_IMAGE["image_gpu"]
            else:
                job_image = settings.JOB_DOCKER_IMAGE["image"]

        # add Nvidia lib volume if training with GPU
        if gpu_count > 0:
            volumes.append(volume.get_volume_config(
                fstype = "hostpath",
                name = "nvidia-libs",
                mount_path = "/usr/local/nvidia/lib64",
                host_path = settings.NVIDIA_LIB_PATH
            ))

        paddle_job = PaddleJob(
            name = obj.get("name", "paddle-cluster-job"),
            job_package = obj.get("jobPackage", ""),
            parallelism = obj.get("parallelism", 1),
            cpu = obj.get("cpu", 1),
            memory = obj.get("memory", "1Gi"),
            pservers = obj.get("pservers", 1),
            pscpu = obj.get("pscpu", 1),
            psmemory = obj.get("psmemory", "1Gi"),
            topology = topology,
            entry = entry,
            gpu = obj.get("gpu", 0),
            image = job_image,
            passes = obj.get("passes", 1),
            registry_secret = registry_secret,
            volumes = volumes
        )
        try:
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
            logging.error("error submitting trainer job: %s" % e)
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
        delete_status = []
        # delete job
        trainer_name = jobname + "-trainer"
        try:
            u_status = client.BatchV1Api().delete_namespaced_job(trainer_name, namespace, {})
        except ApiException, e:
            logging.error("error deleting job: %s, %s", jobname, str(e))
            delete_status.append(str(e))

        # delete job pods
        try:
            job_pod_list = client.CoreV1Api().list_namespaced_pod(namespace, label_selector="paddle-job=%s"%jobname)
            for i in job_pod_list.items:
                u_status = client.CoreV1Api().delete_namespaced_pod(i.metadata.name, namespace, {})
        except ApiException, e:
            logging.error("error deleting job pod: %s", str(e))
            delete_status.append(str(e))

        # delete pserver rs
        pserver_name = jobname + "-pserver"
        try:
            u_status = client.ExtensionsV1beta1Api().delete_namespaced_replica_set(pserver_name, namespace, {})
        except ApiException, e:
            logging.error("error deleting pserver: %s" % str(e))
            delete_status.append(str(e))

        # delete pserver pods
        try:
            # pserver replica set has label with jobname
            job_pod_list = client.CoreV1Api().list_namespaced_pod(namespace, label_selector="paddle-job-pserver=%s"%jobname)
            for i in job_pod_list.items:
                u_status = client.CoreV1Api().delete_namespaced_pod(i.metadata.name, namespace, {})
        except ApiException, e:
            logging.error("error deleting pserver pods: %s" % str(e))
            delete_status.append(str(e))
        if len(delete_status) > 0:
            retcode = 500
        else:
            retcode = 200
        return utils.simple_response(retcode, "\n".join(delete_status))


class LogsView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        Get logs for jobs
        """
        username = request.user.username
        namespace = notebook.utils.email_escape(username)

        jobname = request.query_params.get("jobname")
        num_lines = request.query_params.get("n")
        worker = request.query_params.get("w")
        job_pod_list = client.CoreV1Api().list_namespaced_pod(namespace, label_selector="paddle-job=%s"%jobname)
        total_job_log = ""
        if not worker:
            for i in job_pod_list.items:
                total_job_log = "".join((total_job_log, "==========================%s==========================" % i.metadata.name))
                if num_lines:
                    pod_log = client.CoreV1Api().read_namespaced_pod_log(i.metadata.name, namespace, tail_lines=int(num_lines))
                else:
                    pod_log = client.CoreV1Api().read_namespaced_pod_log(i.metadata.name, namespace)
                total_job_log = "\n".join((total_job_log, pod_log))
        else:
            if num_lines:
                pod_log = client.CoreV1Api().read_namespaced_pod_log(worker, namespace, tail_lines=int(num_lines))
            else:
                pod_log = client.CoreV1Api().read_namespaced_pod_log(worker, namespace)
            total_job_log = pod_log
        return utils.simple_response(200, total_job_log)

class WorkersView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        Get logs for jobs
        """
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
        jobname = request.query_params.get("jobname")
        job_pod_list = None
        if not jobname:
            job_pod_list = client.CoreV1Api().list_namespaced_pod(namespace)
        else:
            selector = "paddle-job=%s"%jobname
            job_pod_list = client.CoreV1Api().list_namespaced_pod(namespace, label_selector=selector)
        return Response(job_pod_list.to_dict())

class QuotaView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        Get user quotas
        """
        username = request.user.username
        namespace = notebook.utils.email_escape(username)

        quota_list = client.CoreV1Api().list_namespaced_resource_quota(namespace)
        return Response(quota_list.to_dict())
