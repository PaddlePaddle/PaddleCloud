import json
import re
from rest_framework.authtoken.models import Token
from rest_framework import viewsets, generics, permissions
from rest_framework.response import Response
from rest_framework.views import APIView
first_cap_re = re.compile('(.)([A-Z][a-z]+)')
all_cap_re = re.compile('([a-z0-9])([A-Z])')
def simple_response(code, msg):
    return Response({
        "code":code,
        "msg": msg
    })

def error_message_response(msg):
    logging.error("error: %s", msg)
    return Response({"msg": msg})

def convert_camel2snake(data):
    s1 = first_cap_re.sub(r'\1_\2', name)
    return all_cap_re.sub(r'\1_\2', s1).lower()
