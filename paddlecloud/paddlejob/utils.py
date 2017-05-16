from django.http import HttpResponse, JsonResponse
import json
import re
first_cap_re = re.compile('(.)([A-Z][a-z]+)')
all_cap_re = re.compile('([a-z0-9])([A-Z])')
def simple_response(code, msg):
    body = {
        "code":code,
        "msg": msg
    }
    return JsonResponse(body)

def convert_camel2snake(data):
    s1 = first_cap_re.sub(r'\1_\2', name)
    return all_cap_re.sub(r'\1_\2', s1).lower()
    
