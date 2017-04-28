# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.shortcuts import render
from django.dispatch import receiver
from django.http import HttpResponseRedirect, HttpResponse, JsonResponse
from django.db.models.signals import post_save
from django.contrib.auth.models import User
from django.contrib.auth.decorators import login_required
from django.conf import settings
# local imports
from notebook.models import PaddleUser
import notebook.forms
import account.views
import tls
# libraries
import os
import json
import logging
import hashlib
import kubernetes
import utils

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
        try:
            logging.info("creating default user certs...")
            tls.create_user_cert(settings.CA_PATH, form.get_user())
        except Exception, e:
            logging.error("create user certs error: %s", str(e))

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


@login_required
def notebook_view(request):
    """
        call kubernetes client to create a Deployment of jupyter notebook,
        mount user's default volume in the pod, then jump to the notebook webpage.
    """
    # NOTICE: username is the same to user's email
    # NOTICE: escape the username to safe string to create namespaces
    username = request.user.username
    utils.update_user_k8s_config(username)

    # FIXME: notebook must be started under username's namespace
    v1api = kubernetes.client.CoreV1Api()
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
        context={"notebook_id": ub.get_notebook_id(username)})

@login_required
def stop_notebook_backend(request):
    username = request.user.username
    utils.update_user_k8s_config(username)
    ub = utils.UserNotebook()
    ub.start_all(username, user_namespace)
    return HttpResponseRedirect("/")
