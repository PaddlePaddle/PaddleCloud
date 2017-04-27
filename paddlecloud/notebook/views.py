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

# a class for creating jupyter notebook resources
class UserNotebook():
    dep_body = {
        "apiVersion": "extensions/v1beta1",
        "kind": "Deployment",
        "metadata": {
            "name": "paddle-book-deployment"
        },
        "spec": {
            "replicas": 1,
            "template": {
                "metadata": {
                    "labels": {
                        "app": "paddle-book"
                    }
                },
                "spec": {
                    "containers": [{
                        "name": "paddle-book",
                        "image": settings.PADDLE_BOOK_IMAGE,
                        "ports": [{
                            "containerPort": settings.PADDLE_BOOK_PORT
                        }]
                    }]
                }
            }
        }
    }
    service_body = {
        "apiVersion": "v1",
        "kind": "Service",
        "metadata": {
            "name": "paddle-book-service"
        },
        "spec": {
            "selector": {
                "app": "paddle-book"
            },
            "ports": [{
                "protocol": "TCP",
                "port": 8888,
                "targetPort": 8888
            }]
        }
    }
    ing_body = {
        "apiVersion": "extensions/v1beta1",
        "kind": "Ingress",
        "metadata": {
            "name": "paddle-book-ingress"
        },
        "spec": {
            "rules": [{
                "host": "notebook.paddlepaddle.org",
                "http": {
                    "paths": [{
                        "path": "/",
                        "backend": {
                            "serviceName": "paddle-book-service",
                            "servicePort": 8888
                        }
                    }]
                }
            }]
        }
    }
    def get_notebook_id(self, username):
        # notebook id is md5(username)
        m = hashlib.md5()
        m.update(username)

        return m.hexdigest()[:8]

    def __wait_api_response(self, resp):
        print resp.status

    def __find_item(self, resource_list, match_name):
        item_found = False
        for item in resource_list.items:
            if item.metadata.name == match_name:
                item_found = True
        return item_found

    def __create_deployment(self, namespace):
        v1beta1api = kubernetes.client.ExtensionsV1beta1Api()
        dep_list = v1beta1api.list_namespaced_deployment(namespace)
        if not self.__find_item(dep_list, "paddle-book-deployment"):
            resp = v1beta1api.create_namespaced_deployment(namespace, body=self.dep_body, pretty=True)
            self.__wait_api_response(resp)

    def __create_service(self, namespace):
        v1api = kubernetes.client.CoreV1Api()
        service_list = v1api.list_namespaced_service(namespace)
        if not self.__find_item(service_list, "paddle-book-service"):
            resp = v1api.create_namespaced_service(namespace, body=self.service_body)
            self.__wait_api_response(resp)

    def __create_ingress(self, username, namespace):
        v1beta1api = kubernetes.client.ExtensionsV1beta1Api()
        ing_list = v1beta1api.list_namespaced_ingress(namespace)
        if not self.__find_item(ing_list, "paddle-book-ingress"):
            # FIXME: must split this for different users
            #self.ing_body["spec"]["rules"][0]["http"]["paths"][0]["path"] = "/" + self.get_notebook_id(username)
            resp = v1beta1api.create_namespaced_ingress(namespace, body=self.ing_body)
            self.__wait_api_response(resp)

    def start_all(self, username, namespace):
        self.__create_deployment(namespace)
        self.__create_service(namespace)
        self.__create_ingress(username, namespace)

def email_escape(email):
    """
        escape email to a safe string of kubernetes namespace
    """
    safe_email = email.replace("@", "-")
    safe_email = safe_email.replace(".", "-")
    safe_email = safe_email.replace("_", "-")
    return safe_email

@login_required
def notebook_view(request):
    """
        call kubernetes client to create a Deployment of jupyter notebook,
        mount user's default volume in the pod, then jump to the notebook webpage.
    """
    # NOTICE: username is the same to user's email
    # NOTICE: escape the username to safe string to create namespaces
    username = request.user.username
    conf_obj = kubernetes.client.Configuration()
    conf_obj.host = settings.K8S_HOST
    conf_obj.ssl_ca_cert = os.path.join(settings.CA_PATH)
    conf_obj.cert_file = os.path.join(settings.USER_CERTS_PATH, username, username, ".pem")
    conf_obj.key_file = os.path.join(settings.USER_CERTS_PATH, username, username, "-key.pem")
    kubernetes.config.load_kube_config(client_configuration=conf_obj)
    # FIXME: notebook must be started under username's namespace
    v1api = kubernetes.client.CoreV1Api()
    namespaces = v1api.list_namespace()
    user_namespace_found = False
    user_namespace = email_escape(username)
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

    ub = UserNotebook()
    ub.start_all(username, user_namespace)

    return render(request, "notebook.html", context={"notebook_id": ub.get_notebook_id(username)})
