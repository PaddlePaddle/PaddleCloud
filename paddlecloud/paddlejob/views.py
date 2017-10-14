from django.http import HttpResponseRedirect, HttpResponse, JsonResponse, HttpResponseNotFound, HttpResponseForbidden
from django.contrib import messages
from django.conf import settings
from django.utils.encoding import smart_str
from django.contrib.auth.decorators import login_required
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
import os
import copy
from notebook.models import FilePublish
import uuid
from cloudprovider.k8s_provider import K8sProvider
from paddle_job import PaddleJob


def file_publish_view(request):
    """
        view for download published files
    """
    username = request.user.username
    publish_uuid = request.GET.get("uuid")
    if not publish_uuid:
        return HttpResponseNotFound()
    record = FilePublish.objects.get(uuid=publish_uuid)
    if not record:
        return HttpResponseNotFound()
    # FIXME(typhoonzero): not support folder currently
    if record.path.endswith("/"):
        return HttpResponseNotFound()

    real_path = "/".join([settings.STORAGE_PATH] + record.path.split("/")[4:])
    logging.info("downloading file from: %s, record(%s)", real_path, record.path)

    # mimetype is replaced by content_type for django 1.7
    response = HttpResponse(open(real_path), content_type='application/force-download') 
    response['Content-Disposition'] = 'attachment; filename=%s' % os.path.basename(record.path)
    # It's usually a good idea to set the 'Content-Length' header too.
    # You can also set any other required headers: Cache-Control, etc.
    return response

class FilePublishAPIView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
            return a list of published files for current user
        """
        record = FilePublish.objects.filter(user=request.user)
        file_list = [rec.path for rec in record]
        url_list = [rec.url for rec in record]
        return Response({"files": file_list, "urls": url_list})

    def post(self, request, format=None):
        """
            given a pfs path generate a uniq sharing url for the path
        """
        post_body = json.loads(request.body)
        file_path = post_body.get("path")
        publish_uuid = uuid.uuid4()
        publish_url = "http://%s/filepub/?uuid=%s" % (request.META["HTTP_HOST"], publish_uuid)
        # save publish_url to mysql
        publish_record = FilePublish()
        publish_record.url = publish_url
        publish_record.user = request.user
        publish_record.path = file_path
        publish_record.uuid = publish_uuid
        publish_record.save()
        return Response({"url": publish_url})


class JobsView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        username = request.user.username
        p = K8sProvider()
        ret_dict = p.get_jobs(username)
        return Response(ret_dict)

    def post(self, request, format=None):
        """
        Submit the PaddlePaddle job
        """
        username = request.user.username
        paddlejob = json.loads(request.body)
        # ========== submit master ReplicaSet if using fault_tolerant feature ==
        p = K8sProvider()
        try:
            p.submit_job(paddlejob, username)
        except Exception, e:
            return utils.error_message_response(str(e))

        return utils.simple_response(200, "")

    def delete(self, request, format=None):
        """
        Kill a job
        """
        username = request.user.username
        obj = json.loads(request.body)
        jobname = obj.get("jobname")
        p = K8sProvider()
        retcode, status = p.delete_job(jobname, username)
        return utils.simple_response(retcode, "\n".join(status))

class PserversView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        List all pservers
        """
        username = request.user.username
        p = K8sProvider()
        return Response(p.get_pservers(username))

class LogsView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        username = request.user.username
        jobname = request.query_params.get("jobname")
        num_lines = request.query_params.get("n")
        worker = request.query_params.get("w")

        total_job_log = K8sProvider().get_logs(jobname, num_lines, worker, username)
        return utils.simple_response(200, total_job_log)

class WorkersView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        username = request.user.username
        jobname = request.query_params.get("jobname")
        ret = K8sProvider().get_workers(jobname, username)
        return Response(ret)

class QuotaView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        username = request.user.username
        ret = K8sProvider().get_quotas(username)
        return Response(ret)

class GetUserView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        """
        Get user name
        """
        content = {
            'user': request.user.username,  # `django.contrib.auth.User` instance.
        }
        return Response(content)

class SimpleFileView(APIView):
    permission_classes = (permissions.IsAuthenticated,)
    parser_classes = (FormParser, MultiPartParser,)

    def __validate_path(self, request, file_path):
        """
        returns error_msg. error_msg will be empty if there's no error.
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
        Simple get file.
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
        Simple put file.
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


class SimpleFileList(APIView):
    permission_classes = (permissions.IsAuthenticated,)
    parser_classes = (FormParser, MultiPartParser,)

    def get(self, request, format=None):
        """
        Simple list files.
        """
        file_path = request.query_params.get("path")
        dc = request.query_params.get("dc")
        # validate list path must be under user's dir
        path_parts = file_path.split(os.path.sep)
        msg = ""
        if len(path_parts) < 5:
            msg = "path must like /pfs/[dc]/home/[user]"
        else:
            if path_parts[1] != "pfs":
                msg = "path must start with /pfs"
            if path_parts[2] not in settings.DATACENTERS.keys():
                msg = "no datacenter "+path_parts[2]
            if path_parts[3] != "home":
                msg = "path must like /pfs/[dc]/home/[user]"
            if path_parts[4] != request.user.username:
                msg = "not a valid user: " + path_parts[4]
        if msg:
            return Response({"msg": msg})

        real_path = file_path.replace("/pfs/%s/home/%s"%(dc, request.user.username), "/pfs/%s"%request.user.username)
        if not os.path.exists(real_path):
            return Response({"msg": "dir not exist"})

        return Response({"msg": "", "items": os.listdir(real_path)})
