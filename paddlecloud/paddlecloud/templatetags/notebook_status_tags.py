from __future__ import unicode_literals
from django import template
from django.contrib.messages.utils import get_level_tags
from django.utils.encoding import force_text
from notebook.utils import email_escape, UserNotebook, get_user_api_client
import kubernetes

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
    namespace = email_escape(user.email)
    ub = UserNotebook()

    return ub.status(username, namespace)
