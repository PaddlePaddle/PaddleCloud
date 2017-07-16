"""
WSGI config for paddlecloud project.

It exposes the WSGI callable as a module-level variable named ``application``.

For more information on this file, see
https://docs.djangoproject.com/en/1.11/howto/deployment/wsgi/
"""

import os
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'paddlecloud.settings')
from django.conf import settings
from django.core.wsgi import get_wsgi_application
import uwsgi
from kubernetes import client, config
from kubernetes.client.rest import ApiException
import ssl
from websocket import create_connection
import websocket
websocket.enableTrace(True)
# import tornado.websocket
import socket


_django_app = get_wsgi_application()

def uwsgi_ws_proxy_application(env, start_response):
    """
        websockt proxy application for uwsgi
        So that jupyter notebook's kernel call can pass through the server
    """
    # complete the handshake
    print env
    uwsgi.websocket_handshake()
    while True:
        try:
            msg = uwsgi.websocket_recv()
            print msg
            # TODO: send to notebook and get response here
            # FIXME: short conn for simplicity
            URI = env.get("REQUEST_URI")
            namespace_code = URI.split("/")[2]
            service_name_cross_namespace = "cloud-notebook-service.%s.svc.cluster.local" % namespace_code
            header = {"Accept-Encoding": env.get("HTTP_ACCEPT_ENCODING"),
                    "Accept-Language:": env.get("HTTP_ACCEPT_LANGUAGE"),
                    #"Connection": "Upgrade",
                    "Pragma": env.get("HTTP_PRAGMA", ""),
                    "Sec-WebSocket-Extensions": env.get("HTTP_SEC_WEBSOCKET_EXTENSIONS", ""),
                    "Sec-WebSocket-Key": env.get("HTTP_SEC_WEBSOCKET_KEY", ""),
                    "Sec-WebSocket-Version": env.get("HTTP_SEC_WEBSOCKET_VERSION", ""),
                    #"Upgrade": "websocket",
                    "User-Agent": env.get("HTTP_USER_AGENT", "")}
            ip = socket.gethostbyname(service_name_cross_namespace)
            print "connect to ws://"+ip+":8888"+URI+" header: ", header, " sending: ", msg
            ws = create_connection("ws://%s:8888%s"%(ip, URI),
                                   #sslopt={"cert_reqs": ssl.CERT_NONE, "check_hostname": False},
                                   #subprotocols=["binary", "base64"],
                                   #host=env.get("HTTP_HOST", ""),
                                   #origin=env.get("HTTP_ORIGIN", ""),
                                   cookie=env.get("HTTP_COOKIE", ""),
                                   header=header)
            print "sending ", msg
            ws.send(msg)
            print "receiving..."
            result =  ws.recv()
            ws.close()
            print "return result: ", result
            uwsgi.websocket_send(result)
        except Exception, e:
            print "error: ", str(e)

def application(environ, start_response):
    # proxy notbook websocket to the backend service
    if environ.get('PATH_INFO').startswith("/notebook") and \
       environ.get('REQUEST_URI').find("channels?session_id")>0:
        return uwsgi_ws_proxy_application(environ, start_response)
    return _django_app(environ, start_response)
