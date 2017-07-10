from __future__ import unicode_literals
from django import template
from django.contrib.messages.utils import get_level_tags
from django.utils.encoding import force_text
from notebook.utils import email_escape, UserNotebook, user_certs_exist
import kubernetes
import logging

LEVEL_TAGS = get_level_tags()

register = template.Library()


def _get_notebook_id(self, username):
    # notebook id is md5(username)
    m = hashlib.md5()
    m.update(username)

    return m.hexdigest()[:8]

@register.simple_tag()
def get_user_notebook_status(user):
    if not user.is_authenticated:
        return ""
    username = user.username
    if user_certs_exist(username):
        namespace = email_escape(user.email)
        try:
            ub = UserNotebook()
            return ub.status(username, namespace)
        except Exception, e:
            logging.error("error get notebook status: %s", str(e))
            return "unknown"
    else:
        return "unknown"
