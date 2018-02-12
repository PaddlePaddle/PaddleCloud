#   Copyright (c) 2018 PaddlePaddle Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import kubernetes
from kubernetes import client, config
from kubernetes.client.rest import ApiException
import os
# FIXME(typhoonzero): still need to import settings
from django.conf import settings


def email_escape(email):
    """
        Escape email to a safe string of kubernetes namespace
    """
    safe_email = email.replace("@", "-")
    safe_email = safe_email.replace(".", "-")
    safe_email = safe_email.replace("_", "-")
    return safe_email


def get_user_api_client(username):
    """
        Update kubernetes client to use current logined user's crednetials
    """

    conf_obj = kubernetes.client.Configuration()
    conf_obj.host = settings.K8S_HOST
    conf_obj.ssl_ca_cert = os.path.join(settings.CA_PATH)
    conf_obj.cert_file = os.path.join(settings.USER_CERTS_PATH, username,
                                      "%s.pem" % username)
    conf_obj.key_file = os.path.join(settings.USER_CERTS_PATH, username,
                                     "%s-key.pem" % username)
    api_client = kubernetes.client.ApiClient(config=conf_obj)
    return api_client


def get_admin_api_client():
    """
        Update kubernetes client to use admin user to create namespace and authorizations
    """

    conf_obj = kubernetes.client.Configuration()
    conf_obj.host = settings.K8S_HOST
    conf_obj.ssl_ca_cert = os.path.join(settings.CA_PATH)
    conf_obj.cert_file = os.path.join(settings.USER_CERTS_PATH, "admin.pem")
    conf_obj.key_file = os.path.join(settings.USER_CERTS_PATH, "admin-key.pem")
    api_client = kubernetes.client.ApiClient(config=conf_obj)
    return api_client


def user_certs_exist(username):
    """
        Return True if the user's certs already generated. User's keys are of pairs.
    """
    has_cert = os.path.isfile(
        os.path.join(settings.USER_CERTS_PATH, username, "%s.pem" % username))
    has_key = os.path.isfile(
        os.path.join(settings.USER_CERTS_PATH, username, "%s-key.pem" %
                     username))
    if has_cert and has_key:
        return True
    else:
        return False
