"""
WSGI config for paddlecloud project.

It exposes the WSGI callable as a module-level variable named ``application``.

For more information on this file, see
https://docs.djangoproject.com/en/1.11/howto/deployment/wsgi/
"""

import os
os.environ.setdefault('DJANGO_SETTINGS_MODULE', 'myapp.settings')
from django.conf import settings
from django.core.wsgi import get_wsgi_application


_django_app = get_wsgi_application()

def uwsgi_ws_proxy_application(env, start_response):
    # complete the handshake
    uwsgi.websocket_handshake(env['HTTP_SEC_WEBSOCKET_KEY'], env.get('HTTP_ORIGIN', ''))
    while True:
        msg = uwsgi.websocket_recv()
        #TODO: send to notebook and get response here
        uwsgi.websocket_send(msg)


def application(environ, start_response):
    # proxy notbook websocket to the backend service
    if environ.get('PATH_INFO').startswith("/notebook") and \
       environ.get('PATH_INFO').find("channels?session_id")>0:
        return uwsgi_ws_proxy_application(environ, start_response)
    return _django_app(environ, start_response)
