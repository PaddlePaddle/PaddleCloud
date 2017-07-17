from django.conf import settings
import os
import kubernetes
import hashlib
import copy
import logging

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

# a class for creating jupyter notebook resources
class UserNotebook():
    dep_body = {
        "apiVersion": "extensions/v1beta1",
        "kind": "Deployment",
        "metadata": {
            "name": "cloud-notebook-deployment"
        },
        "spec": {
            "replicas": 1,
            "template": {
                "metadata": {
                    "labels": {
                        "app": "cloud-notebook"
                    }
                },
                "spec": {
                    "containers": [{
                        "name": "cloud-notebook",
                        "image": settings.PADDLE_BOOK_IMAGE,
                        "command": ["sh", "-c",
                         "mkdir -p /root/.jupyter; echo \"c.NotebookApp.base_url = '/notebook/%s'\" > /root/.jupyter/jupyter_notebook_config.py; echo \"c.NotebookApp.allow_origin = '*'\" >> /root/.jupyter/jupyter_notebook_config.py; jupyter notebook --ip=0.0.0.0 --no-browser --allow-root --NotebookApp.token='' --NotebookApp.disable_check_xsrf=True /book/"],
                        "ports": [{
                            "containerPort": settings.PADDLE_BOOK_PORT
                        }],
                        "resources": {
                            "requests": {
                                "memory": "4Gi",
                                "cpu": "1",
                            },
                            "limits": {
                                "memory": "4Gi",
                                "cpu": "1",
                            }
                        },
                        "env" : [
                            {"name": "USER_NAMESPACE", "valueFrom": {"fieldRef": {"fieldPath": "metadata.namespace"}}},
                            {"name": "USER_POD_NAME", "valueFrom": {"fieldRef": {"fieldPath": "metadata.name"}}},
                            {"name": "USER_POD_IP", "valueFrom": {"fieldRef": {"fieldPath": "status.podIP"}}},
                            {"name": "USER_POD_SERVICE_ACCOUNT", "valueFrom": {"fieldRef": {"fieldPath": "spec.serviceAccountName"}}},
                        ]
                    }]
                }
            }
        }
    }
    service_body = {
        "apiVersion": "v1",
        "kind": "Service",
        "metadata": {
            "name": "cloud-notebook-service"
        },
        "spec": {
            "selector": {
                "app": "cloud-notebook"
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
            "name": "cloud-notebook-ingress"
        },
        "spec": {
            "rules": [{
                "host": "cloud.paddlepaddle.org",
                "http": {
                    "paths": [{
                        "path": "/",
                        "backend": {
                            "serviceName": "cloud-notebook-service",
                            "servicePort": 8888
                        }
                    },
                    ]
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

    def __create_deployment(self, username, namespace):
        v1beta1api = kubernetes.client.ExtensionsV1beta1Api(api_client=get_user_api_client(username))
        dep_list = v1beta1api.list_namespaced_deployment(namespace)
        if not self.__find_item(dep_list, "cloud-notebook-deployment"):
            dep_body = copy.deepcopy(self.dep_body)
            logging.info("command: %s, userid: %s", self.dep_body["spec"]["template"]["spec"]["containers"][0]["command"][2], self.get_notebook_id(username))
            dep_body["spec"]["template"]["spec"]["containers"][0]["command"][2] = \
                dep_body["spec"]["template"]["spec"]["containers"][0]["command"][2] % (self.get_notebook_id(username))
            resp = v1beta1api.create_namespaced_deployment(namespace, body=dep_body, pretty=True)
            self.__wait_api_response(resp)

    def __create_service(self, username, namespace):
        v1api = kubernetes.client.CoreV1Api(api_client=get_user_api_client(username))
        service_list = v1api.list_namespaced_service(namespace)
        if not self.__find_item(service_list, "cloud-notebook-service"):
            resp = v1api.create_namespaced_service(namespace, body=self.service_body)
            self.__wait_api_response(resp)

    def __create_ingress(self, username, namespace):
        v1beta1api = kubernetes.client.ExtensionsV1beta1Api(api_client=get_user_api_client(username))
        ing_list = v1beta1api.list_namespaced_ingress(namespace)
        if not self.__find_item(ing_list, "cloud-notebook-ingress"):
            # FIXME: must split this for different users
            ing_body = copy.deepcopy(self.ing_body)
            ing_body["spec"]["rules"][0]["http"]["paths"][0]["path"] = "/notebook/" + self.get_notebook_id(username)
            resp = v1beta1api.create_namespaced_ingress(namespace, body=ing_body)
            self.__wait_api_response(resp)

    def start_all(self, username, namespace):
        """
            start deployment, service, ingress to start a notebook service for current user
        """
        self.__create_deployment(username, namespace)
        self.__create_service(username, namespace)
        self.__create_ingress(username, namespace)

    def stop_all(self, username, namespace):
        v1beta1api = kubernetes.client.ExtensionsV1beta1Api(api_client=get_user_api_client(username))
        v1api = kubernetes.client.CoreV1Api(api_client=get_user_api_client(username))
        v1beta1api.delete_namespaced_deployment("cloud-notebook-deployment", namespace)
        v1beta1api.delete_namespaced_ingress("cloud-notebook-ingress", namespace)
        v1api.delete_namespaced_service("cloud-notebook-service", namespace)

    def status(self, username, namespace):
        """
            check notebook deployment status
            @return: running starting stopped
        """
        v1api = kubernetes.client.CoreV1Api(api_client=get_user_api_client(username))
        v1beta1api = kubernetes.client.ExtensionsV1beta1Api(api_client=get_user_api_client(username))
        d, s, i = (True, True, True)
        # -------------------- deployment status --------------------
        dep_list = v1beta1api.list_namespaced_deployment(namespace)
        if not self.__find_item(dep_list, "cloud-notebook-deployment"):
            d = False
        else:
            # notebook must have at least one replica running
            for i in dep_list.items:
                if i.status.ready_replicas < 1:
                    d = False
        # -------------------- service status --------------------
        service_list = v1api.list_namespaced_service(namespace)
        if not self.__find_item(service_list, "cloud-notebook-service"):
            s = False
        else:
            # service is ready when the endpoints to pods has been found
            endpoints_list = v1api.list_namespaced_endpoints(namespace)
            if not self.__find_item(endpoints_list, "cloud-notebook-service"):
                s = False
        # -------------------- ingress status --------------------
        ing_list = v1beta1api.list_namespaced_ingress(namespace)
        if not self.__find_item(ing_list, "cloud-notebook-ingress"):
            i = False
        else:
            i = False
            try:
                # ingress is ready when the remote ip is assigned
                for i in ing_list.items:
                    if i:
                        for ing in i.status.load_balancer.ingress:
                            if not ing.ip:
                                i = False
            except:
                pass

        if d and s and i:
            return "running"
        elif d or s or i:
            return "starting"
        else:
            return "stopped"
