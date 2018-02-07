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
import os
import base64

def docker_cfg(username, password, email, server):
    auth = "%s:%s" % (username, password)
    auth_encode = base64.b64encode(auth)
    return json.dumps({server:
                       {"username": username,
                        "password": password,
                        "email": email,
                        "auth": auth_encode}})

class RegistryView(APIView):
    permission_classes = (permissions.IsAuthenticated,)
    def post(self, request):
        """
        Cretea a registry secret
        """
        username = request.user.username
        user_namespace = notebook.utils.email_escape(username)
        api_client = notebook.utils.get_user_api_client(username)
        obj = json.loads(request.body)
        name = obj.get("name")
        docker_username = obj.get("username")
        docker_password = obj.get("password")
        docker_server = obj.get("server")
        cfg = docker_cfg(docker_username,
                                docker_password,
                                username,
                                docker_server)
        try:
            ret = client.CoreV1Api(
                api_client=api_client).create_namespaced_secret(
                    namespace = user_namespace,
                    body = {
                        "apiVersion": "v1",
                        "kind": "Secret",
                        "metadata": {
                            "name": name
                        },
                        "data": {
                            ".dockerconfigjson": base64.b64encode(cfg)
                        },
                        "type": "kubernetes.io/dockerconfigjson"})
        except ApiException, e:
            logging.error("Failed when create secret.")
            return utils.simple_response(500, str(e))
        return utils.simple_response(200, "")

    def delete(self, request):
        """
        Delete a registry secret
        """
        username = username = request.user.username
        user_namespace = notebook.utils.email_escape(username)
        api_client = notebook.utils.get_user_api_client(username)
        obj = json.loads(request.body)
        name = obj.get("name")
        try:
            ret = client.CoreV1Api(api_client=api_client).delete_namespaced_secret(
                name = name,
                namespace = user_namespace,
                body = client.V1DeleteOptions())
        except ApiException, e:
            logging.error("Failed when delete secret.")
            return utils.simple_response(500, str(e))
        return utils.simple_response(200, "")

    def get(self, request):
        """
        Get registrys
        """
        username = username = request.user.username
        user_namespace = notebook.utils.email_escape(username)
        api_client = notebook.utils.get_user_api_client(username)
        try:
            secretes_list = client.CoreV1Api(api_client=api_client).list_namespaced_secret(
                namespace=user_namespace)
            return utils.simple_response(200, secretes_list.to_dict())
        except ApiException, e:
            logging.error("Failed when list secrets.")
            return utils.simple_response(500, str(e))
