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

from __future__ import unicode_literals
from django import template
from django.contrib.messages.utils import get_level_tags
from django.utils.encoding import force_text
from notebook.utils import email_escape, UserNotebook, user_certs_exist
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
    if user_certs_exist(username):
        namespace = email_escape(user.email)
        ub = UserNotebook()
        return ub.status(username, namespace)
    else:
        return "unknown"
