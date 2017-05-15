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
from wsgiref.util import FileWrapper

def healthz(request):
    return HttpResponse("OK")

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
    v1api = kubernetes.client.CoreV1Api(utils.get_user_api_client(username))
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
