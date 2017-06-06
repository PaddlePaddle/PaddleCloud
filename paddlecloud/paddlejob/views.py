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
from rest_framework.parsers import MultiPartParser, FormParser, FileUploadParser
import json
import utils
import notebook.utils
import logging
import volume
import os

class JobsView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        List all jobs
        """
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
        api_instance = client.BatchV1Api(api_client=notebook.utils.get_user_api_client(username))
        job_list = api_instance.list_namespaced_job(namespace)
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
        api_client = notebook.utils.get_user_api_client(username)
        if not topology and not entry:
            return utils.simple_response(500, "no topology or entry specified")
        if not obj.get("datacenter"):
            return utils.simple_response(500, "no datacenter specified")
        cfgs = {}
        dc = obj.get("datacenter")

        volumes = []
        for k, cfg in settings.DATACENTERS.items():
            if k != dc and k != "public":
                continue
            fstype = cfg["fstype"]
            if fstype == settings.FSTYPE_CEPHFS:
                if k == "public":
                    mount_path = cfg["mount_path"] % dc
                    cephfs_path = cfg["cephfs_path"]
                else:
                    mount_path = cfg["mount_path"] % (dc, username)
                    cephfs_path = cfg["cephfs_path"] % username
                volumes.append(volume.get_volume_config(
                    fstype = fstype,
                    name = k.replace("_", "-"),
                    monitors_addr = cfg["monitors_addr"],
                    secret = cfg["secret"],
                    user = cfg["user"],
                    mount_path = mount_path,
                    cephfs_path = cephfs_path,
                    admin_key = cfg["admin_key"],
                    read_only = cfg.get("read_only", False)
                ))
            elif fstype == settings.FSTYPE_HOSTPATH:
                if k == "public":
                    mount_path = cfg["mount_path"] % dc
                    host_path = cfg["host_path"]
                else:
                    mount_path = cfg["mount_path"] % (dc, username)
                    host_path = cfg["host_path"] % username

                volumes.append(volume.get_volume_config(
                    fstype = fstype,
                    name = k.replace("_", "-"),
                    mount_path = mount_path,
                    host_path = host_path
                ))
            else:
                pass

        registry_secret = settings.JOB_DOCKER_IMAGE.get("registry_secret", None)
        # get user specified image
        job_image = obj.get("image", None)
        gpu_count = obj.get("gpu", 0)
        # jobPackage validation: startwith /pfs
        # NOTE: always overwrite the job package when the directory exists
        job_package =obj.get("jobPackage", "")
        if not job_package.startswith("/pfs"):
            # add /pfs... cloud path
            if job_package.startswith("/"):
                # use last dirname as package name
                package_in_pod = os.path.join("/pfs/%s/home/%s"%(dc, username), os.path.basename(job_package))
            else:
                package_in_pod = os.path.join("/pfs/%s/home/%s"%(dc, username), job_package)
        else:
            package_in_pod = job_package
        # package must be ready before submit a job
        current_package_path = package_in_pod.replace("/pfs/%s/home"%dc, settings.STORAGE_PATH)
        if not os.path.exists(current_package_path):
            return utils.error_message_response("error: package not exist in cloud")

        # use default images
        if not job_image :
            if gpu_count > 0:
                job_image = settings.JOB_DOCKER_IMAGE["image_gpu"]
            else:
                job_image = settings.JOB_DOCKER_IMAGE["image"]

        # add Nvidia lib volume if training with GPU
        if gpu_count > 0:
            volumes.append(volume.get_volume_config(
                fstype = settings.FSTYPE_HOSTPATH,
                name = "nvidia-libs",
                mount_path = "/usr/local/nvidia/lib64",
                host_path = settings.NVIDIA_LIB_PATH
            ))

        paddle_job = PaddleJob(
            name = obj.get("name", "paddle-cluster-job"),
            job_package = package_in_pod,
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
            ret = client.ExtensionsV1beta1Api(api_client=api_client).create_namespaced_replica_set(
                namespace,
                paddle_job.new_pserver_job(),
                pretty=True)
        except ApiException, e:
            logging.error("error submitting pserver job: %s " % e)
            return utils.simple_response(500, str(e))

        #submit trainer job, it's Kubernetes Job
        try:
            ret = client.BatchV1Api(api_client=api_client).create_namespaced_job(
                namespace,
                paddle_job.new_trainer_job(),
                pretty=True)
        except ApiException, e:
            logging.error("error submitting trainer job: %s" % e)
            return utils.simple_response(500, str(e))
        return utils.simple_response(200, "")

    def delete(self, request, format=None):
        """
        Kill a job
        """
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
        obj = json.loads(request.body)
        jobname = obj.get("jobname")
        api_client = notebook.utils.get_user_api_client(username)
        if not jobname:
            return utils.simple_response(500, "must specify jobname")
        # FIXME: options needed: grace_period_seconds, orphan_dependents, preconditions
        # FIXME: cascade delteing
        delete_status = []
        # delete job
        trainer_name = jobname + "-trainer"
        try:
            u_status = client.BatchV1Api(api_client=api_client)\
                .delete_namespaced_job(trainer_name, namespace, {})
        except ApiException, e:
            logging.error("error deleting job: %s, %s", jobname, str(e))
            delete_status.append(str(e))

        # delete job pods
        try:
            job_pod_list = client.CoreV1Api(api_client=api_client)\
                .list_namespaced_pod(namespace,
                                     label_selector="paddle-job=%s"%jobname)
            for i in job_pod_list.items:
                u_status = client.CoreV1Api(api_client=api_client)\
                    .delete_namespaced_pod(i.metadata.name, namespace, {})
        except ApiException, e:
            logging.error("error deleting job pod: %s", str(e))
            delete_status.append(str(e))

        # delete pserver rs
        pserver_name = jobname + "-pserver"
        try:
            u_status = client.ExtensionsV1beta1Api(api_client=api_client)\
                .delete_namespaced_replica_set(pserver_name, namespace, {})
        except ApiException, e:
            logging.error("error deleting pserver: %s" % str(e))
            delete_status.append(str(e))

        # delete pserver pods
        try:
            # pserver replica set has label with jobname
            job_pod_list = client.CoreV1Api(api_client=api_client)\
                .list_namespaced_pod(namespace,
                                     label_selector="paddle-job-pserver=%s"%jobname)
            for i in job_pod_list.items:
                u_status = client.CoreV1Api(api_client=api_client)\
                    .delete_namespaced_pod(i.metadata.name, namespace, {})
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
        api_client = notebook.utils.get_user_api_client(username)
        jobname = request.query_params.get("jobname")
        num_lines = request.query_params.get("n")
        worker = request.query_params.get("w")
        job_pod_list = client.CoreV1Api(api_client=api_client)\
            .list_namespaced_pod(namespace, label_selector="paddle-job=%s"%jobname)
        total_job_log = ""
        if not worker:
            for i in job_pod_list.items:
                total_job_log = "".join((total_job_log, "==========================%s==========================" % i.metadata.name))
                if num_lines:
                    pod_log = client.CoreV1Api(api_client=api_client)\
                        .read_namespaced_pod_log(
                            i.metadata.name, namespace, tail_lines=int(num_lines))
                else:
                    pod_log = client.CoreV1Api(api_client=api_client)\
                        .read_namespaced_pod_log(i.metadata.name, namespace)
                total_job_log = "\n".join((total_job_log, pod_log))
        else:
            if num_lines:
                pod_log = client.CoreV1Api(api_client=api_client)\
                    .read_namespaced_pod_log(worker, namespace, tail_lines=int(num_lines))
            else:
                pod_log = client.CoreV1Api(api_client=api_client)\
                    .read_namespaced_pod_log(worker, namespace)
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
        api_client = notebook.utils.get_user_api_client(username)
        if not jobname:
            job_pod_list = client.CoreV1Api(api_client=api_client)\
                .list_namespaced_pod(namespace)
        else:
            selector = "paddle-job=%s"%jobname
            job_pod_list = client.CoreV1Api(api_client=api_client)\
                .list_namespaced_pod(namespace, label_selector=selector)
        return Response(job_pod_list.to_dict())

class QuotaView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        Get user quotas
        """
        username = request.user.username
        namespace = notebook.utils.email_escape(username)
        api_client = notebook.utils.get_user_api_client(username)
        quota_list = api_client.CoreV1Api(api_cilent=api_client)\
            .list_namespaced_resource_quota(namespace)
        return Response(quota_list.to_dict())


class SimpleFileView(APIView):
    permission_classes = (permissions.IsAuthenticated,)
    parser_classes = (FormParser, MultiPartParser,)

    def __validate_path(self, request, file_path):
        """
        returns error_msg. error_msg will be empty if there's no error
        """
        path_parts = file_path.split(os.path.sep)

        assert(path_parts[1]=="pfs")
        assert(path_parts[2] in settings.DATACENTERS.keys())
        assert(path_parts[3] == "home")
        assert(path_parts[4] == request.user.username)

        server_file = os.path.join(settings.STORAGE_PATH, request.user.username, *path_parts[5:])

        return server_file

    def get(self, request, format=None):
        """
        Simple down file
        """
        file_path = request.query_params.get("path")
        try:
            write_file = self.__validate_path(request, file_path)
        except Exception, e:
            return utils.error_message_response("file path not valid: %s"%str(e))

        if not os.path.exists(os.sep+write_file):
            return Response({"msg": "file not exist"})

        response = HttpResponse(open(write_file), content_type='application/force-download')
        response['Content-Disposition'] = 'attachment; filename="%s"' % os.path.basename(write_file)

        return response

    def post(self, request, format=None):
        """
        Simple up file
        """
        file_obj = request.data['file']
        file_path = request.query_params.get("path")
        if not file_path:
            return utils.error_message_response("must specify path")
        try:
            write_file = self.__validate_path(request, file_path)
        except Exception, e:
            return utils.error_message_response("file path not valid: %s"%str(e))

        if not os.path.exists(os.path.dirname(write_file)):
            try:
                os.makedirs(os.path.dirname(write_file))
            except OSError as exc: # Guard against race condition
                if exc.errno != errno.EEXIST:
                    raise
        # FIXME: always overwrite package files
        with open(write_file, "w") as fn:
            while True:
                data = file_obj.read(4096)
                if not data:
                    break
                fn.write(data)

        return Response({"msg": ""})
