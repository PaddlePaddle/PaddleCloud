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
    conf_obj.cert_file = os.path.join(settings.USER_CERTS_PATH, username, "%s.pem"%username)
    conf_obj.key_file = os.path.join(settings.USER_CERTS_PATH, username, "%s-key.pem"%username)
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
    has_cert = os.path.isfile(os.path.join(settings.USER_CERTS_PATH, username, "%s.pem"%username))
    has_key = os.path.isfile(os.path.join(settings.USER_CERTS_PATH, username, "%s-key.pem"%username))
    if has_cert and has_key:
        return True
    else:
        return False
