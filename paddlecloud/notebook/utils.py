from django.conf import settings
import os
import kubernetes
import hashlib
import copy

def email_escape(email):
    """
        escape email to a safe string of kubernetes namespace
    """
    safe_email = email.replace("@", "-")
    safe_email = safe_email.replace(".", "-")
    safe_email = safe_email.replace("_", "-")
    return safe_email


def get_user_api_client(username):
    """
        update kubernetes client to use current logined user's crednetials
    """

    conf_obj = kubernetes.client.Configuration()
    conf_obj.host = settings.K8S_HOST
    conf_obj.ssl_ca_cert = os.path.join(settings.CA_PATH)
    conf_obj.cert_file = os.path.join(settings.USER_CERTS_PATH, username, "%s.pem"%username)
    conf_obj.key_file = os.path.join(settings.USER_CERTS_PATH, username, "%s-key.pem"%username)
    #kubernetes.config.load_kube_config(client_configuration=conf_obj)
    api_client = kubernetes.client.ApiClient(config=conf_obj)
    return api_client

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
                        "command": ["sh", "-c",
                         "mkdir -p /root/.jupyter; echo \"c.NotebookApp.base_url = '/notebook/%s'\" > /root/.jupyter/jupyter_notebook_config.py; jupyter notebook --debug --ip=0.0.0.0 --no-browser --allow-root --NotebookApp.token='' --NotebookApp.disable_check_xsrf=True /book/"],
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
                "host": "cloud.paddlepaddle.org",
                "http": {
                    "paths": [{
                        "path": "/",
                        "backend": {
                            "serviceName": "paddle-book-service",
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
        if not self.__find_item(dep_list, "paddle-book-deployment"):
            self.dep_body["spec"]["template"]["spec"]["containers"][0]["command"][2] = \
                self.dep_body["spec"]["template"]["spec"]["containers"][0]["command"][2] % self.get_notebook_id(username)
            resp = v1beta1api.create_namespaced_deployment(namespace, body=self.dep_body, pretty=True)
            self.__wait_api_response(resp)

    def __create_service(self, username, namespace):
        v1api = kubernetes.client.CoreV1Api(api_client=get_user_api_client(username))
        service_list = v1api.list_namespaced_service(namespace)
        if not self.__find_item(service_list, "paddle-book-service"):
            resp = v1api.create_namespaced_service(namespace, body=self.service_body)
            self.__wait_api_response(resp)

        # service for notebook websocket
        service_ws = copy.deepcopy(self.service_body)
        service_ws["metadata"]["name"] = "paddle-book-service-ws"
        if not self.__find_item(service_list, "paddle-book-service-ws"):
            resp = v1api.create_namespaced_service(namespace, body=service_ws)
            self.__wait_api_response(resp)

    def __create_ingress(self, username, namespace):
        v1beta1api = kubernetes.client.ExtensionsV1beta1Api(api_client=get_user_api_client(username))
        ing_list = v1beta1api.list_namespaced_ingress(namespace)
        if not self.__find_item(ing_list, "paddle-book-ingress"):
            # FIXME: must split this for different users
            self.ing_body["spec"]["rules"][0]["http"]["paths"][0]["path"] = "/notebook/" + self.get_notebook_id(username)
            #self.ing_body["spec"]["rules"][0]["http"]["paths"][1]["path"] = "/notebook/" + self.get_notebook_id(username) + "/api/kernels/*"
            resp = v1beta1api.create_namespaced_ingress(namespace, body=self.ing_body)
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
        v1beta1api.delete_namespaced_deployment("paddle-book-deployment", namespace)
        v1beta1api.delete_namespaced_ingress("paddle-book-ingress", namespace)
        v1api.delete_namespaced_service("paddle-book-service", namespace)

    def status(self, username, namespace):
        """
            check notebook deployment status
            @return: running starting stopped
        """
        v1api = kubernetes.client.CoreV1Api(api_client=get_user_api_client(username))
        v1beta1api = kubernetes.client.ExtensionsV1beta1Api(api_client=get_user_api_client(username))
        d, s, i = (True, True, True)
        dep_list = v1beta1api.list_namespaced_deployment(namespace)
        if not self.__find_item(dep_list, "paddle-book-deployment"):
            d = False
        service_list = v1api.list_namespaced_service(namespace)
        if not self.__find_item(service_list, "paddle-book-service"):
            s = False
        ing_list = v1beta1api.list_namespaced_ingress(namespace)
        if not self.__find_item(ing_list, "paddle-book-ingress"):
            i = False

        if d and s and i:
            return "running"
        elif d or s or i:
            return "starting"
        else:
            return "stopped"
