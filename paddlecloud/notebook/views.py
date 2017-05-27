# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.shortcuts import render
from django.dispatch import receiver
from django.http import HttpResponseRedirect, HttpResponse, JsonResponse
from django.db.models.signals import post_save
from django.contrib.auth.models import User
from django.contrib.auth.decorators import login_required
from django.contrib import messages
from django.conf import settings
# local imports
from notebook.models import PaddleUser
import notebook.forms
import account.views
import tls
import utils
# libraries
import os
import json
import logging
import hashlib
import kubernetes
import zipfile
import cStringIO as StringIO
import base64
from wsgiref.util import FileWrapper
from rest_framework.authtoken.models import Token
from rest_framework import viewsets, generics, permissions
from rest_framework.response import Response
from rest_framework.views import APIView

def healthz(request):
    return HttpResponse("OK")

class SampleView(APIView):
    permission_classes = (permissions.IsAuthenticated,)

    def get(self, request, format=None):
        content = {
            'user': unicode(request.user),  # `django.contrib.auth.User` instance.
            'auth': unicode(request.auth),  # None
            'result': "sample api result",
        }
        return Response(content)

@receiver(post_save, sender=settings.AUTH_USER_MODEL)
def create_auth_token(sender, instance=None, created=False, **kwargs):
    if created:
        Token.objects.create(user=instance)

@receiver(post_save, sender=User)
def handle_user_save(sender, instance, created, **kwargs):
    if created:
        PaddleUser.objects.create(user=instance)

class LoginView(account.views.LoginView):

    form_class = account.forms.LoginEmailForm

class SignupView(account.views.SignupView):
    form_class = notebook.forms.SignupForm
    identifier_field = "email"

    def after_signup(self, form):
        self.update_profile(form)
        logging.info("creating default user certs...")
        tls.create_user_cert(settings.CA_PATH, form.cleaned_data["email"])
        # HACK: username is the same as user email
        # create user's default RBAC permissions
        logging.info("creating default user namespace and RBAC...")
        create_user_namespace(form.cleaned_data["email"])
        create_user_RBAC_permissions(form.cleaned_data["email"])
        # create user's cephfs storage dir
        try:
            os.mkdir(os.path.join(settings.STORAGE_PATH, form.cleaned_data["email"]))
        except Exception, e:
            # FIXME: all exception is ignored
            logging.error("create user's storage path error: %s", e)

        super(SignupView, self).after_signup(form)

    def update_profile(self, form):
        profile = self.created_user.paddleuser  # replace with your reverse one-to-one profile attribute
        data = form.cleaned_data
        profile.school = data["school"]
        profile.studentID = data["studentID"]
        profile.major = data["major"]
        profile.save()

    def generate_username(self, form):
        # do something to generate a unique username (required by the
        # Django User model, unfortunately)
        username = form.cleaned_data["email"]
        return username

class SettingsView(account.views.SettingsView):
    form_class = notebook.forms.SettingsForm

@login_required
def user_certs_view(request):
    key_exist = utils.user_certs_exist(request.user.username)
    user_keys = ["%s.pem" % request.user.username, "%s-key.pem" % request.user.username]

    return render(request, "user_certs.html",
        context={"key_exist": key_exist, "user_keys": user_keys})

@login_required
def user_certs_download(request):
    certs_file = StringIO.StringIO()
    with zipfile.ZipFile(certs_file, mode='w', compression=zipfile.ZIP_DEFLATED) as zf:
        with open(os.path.join(settings.USER_CERTS_PATH, request.user.username, "%s.pem"%request.user.username), "r") as c:
            zf.writestr('%s.pem'%request.user.username, c.read())
        with open(os.path.join(settings.USER_CERTS_PATH, request.user.username, "%s-key.pem"%request.user.username), "r") as s:
            zf.writestr('%s-key.pem'%request.user.username, s.read())

    response = HttpResponse(certs_file.getvalue(), content_type='application/zip')
    response['Content-Disposition'] = 'attachment; filename=%s.zip' % request.user.username
    response['Content-Length'] = certs_file.tell()
    return response

@login_required
def user_certs_generate(request):
    logging.info("creating default user certs...")
    try:
        tls.create_user_cert(settings.CA_PATH, request.user.email)
        messages.success(request, "X509 certificate generated and updated.")
    except Exception, e:
        messages.error(request, str(e))
        logging.error(str(e))
    return HttpResponseRedirect(request.META.get('HTTP_REFERER'))

def create_user_RBAC_permissions(username):
    namespace = utils.email_escape(username)
    rbacapi = kubernetes.client.RbacAuthorizationV1beta1Api(utils.get_admin_api_client())
    body = {
    "apiVersion": "rbac.authorization.k8s.io/v1beta1",
    "kind": "RoleBinding",
    "metadata": {
        "name": "%s-admin-binding"%namespace,
        "namespace": namespace
        },
    "roleRef": {
        "apiGroup": "rbac.authorization.k8s.io",
        "kind": "ClusterRole",
        "name": "admin"
        },
    "subjects": [{
        "apiGroup": "rbac.authorization.k8s.io",
        "kind": "User",
        "name": username
        }]
    }
    rbacapi.create_namespaced_role_binding(namespace, body)
    # create service account permissions
    body = {
    "apiVersion": "rbac.authorization.k8s.io/v1beta1",
    "kind": "RoleBinding",
    "metadata": {
        "name": "%s-sa-view"%namespace,
        "namespace": namespace
        },
    "roleRef": {
        "apiGroup": "rbac.authorization.k8s.io",
        "kind": "ClusterRole",
        "name": "view"
        },
    "subjects": [{
        "kind": "ServiceAccount",
        "name": "default",
        "namespace": namespace
        }]
    }
    rbacapi.create_namespaced_role_binding(namespace, body)

def create_user_namespace(username):
    v1api = kubernetes.client.CoreV1Api(utils.get_admin_api_client())
    namespaces = v1api.list_namespace()
    user_namespace_found = False
    user_namespace = utils.email_escape(username)
    for ns in namespaces.items:
        # must commit to user's namespace
        if ns.metadata.name == user_namespace:
            user_namespace_found = True
    # Create user's namespace if it does not exist
    if not user_namespace_found:
        v1api.create_namespace({"apiVersion": "v1",
            "kind": "Namespace",
            "metadata": {
                "name": user_namespace
            }})
    #create DataCenter sercret if not exists
    secrets = v1api.list_namespaced_secret(user_namespace)
    secret_names = [item.metadata.name for item in secrets.items]
    for dc, cfg in settings.DATACENTERS.items():
        # create Kubernetes Secret for ceph admin key
        if cfg["fstype"] == "cephfs" and cfg["secret"] not in secret_names:
            with open(cfg["admin_key"], "r") as f:
                key = f.read()
                encoded = base64.b64encode(key)
                v1api.create_namespaced_secret(user_namespace, {
                    "apiVersion": "v1",
                    "kind": "Secret",
                    "metadata": {
                        "name": cfg["secret"]
                    },
                    "data": {
                        "key": encoded
                    }})
    # create docker registry secret
    registry_secret = settings.JOB_DOCKER_IMAGE.get("registry_secret", None)
    if registry_secret and registry_secret not in secret_names:
        docker_config = settings.JOB_DOCKER_IMAGE["docker_config"]
        encode = base64.b64encode(json.dumps(docker_config))
        v1api.create_namespaced_secret(user_namespace, {
            "apiVersion": "v1",
            "kind": "Secret",
            "metadata": {
                "name": registry_secret
            },
            "data": {
                ".dockerconfigjson": encode
            },
            "type": "kubernetes.io/dockerconfigjson"
        })
    return user_namespace


@login_required
def notebook_view(request):
    """
        call kubernetes client to create a Deployment of jupyter notebook,
        mount user's default volume in the pod, then jump to the notebook webpage.
    """
    # NOTICE: username is the same to user's email
    # NOTICE: escape the username to safe string to create namespaces
    username = request.user.username

    # FIXME: notebook must be started under username's namespace
    user_namespace = create_user_namespace(username)

    ub = utils.UserNotebook()
    ub.start_all(username, user_namespace)

    return render(request, "notebook.html",
        context={"notebook_id": ub.get_notebook_id(username),
                 "notebook_status": ub.status(username, user_namespace)})

@login_required
def stop_notebook_backend(request):
    username = request.user.username
    utils.update_user_k8s_config(username)
    ub = utils.UserNotebook()
    ub.stop_all(username, user_namespace)
    return HttpResponseRedirect("/")
