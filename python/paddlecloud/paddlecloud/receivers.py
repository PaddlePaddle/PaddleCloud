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

from django.dispatch import receiver

from account.signals import password_changed
from account.signals import user_sign_up_attempt, user_signed_up
from account.signals import user_login_attempt, user_logged_in

from pinax.eventlog.models import log


@receiver(user_logged_in)
def handle_user_logged_in(sender, **kwargs):
    log(user=kwargs.get("user"), action="USER_LOGGED_IN", extra={})


@receiver(password_changed)
def handle_password_changed(sender, **kwargs):
    log(user=kwargs.get("user"), action="PASSWORD_CHANGED", extra={})


@receiver(user_login_attempt)
def handle_user_login_attempt(sender, **kwargs):
    log(user=None,
        action="LOGIN_ATTEMPTED",
        extra={
            "username": kwargs.get("username"),
            "result": kwargs.get("result")
        })


@receiver(user_sign_up_attempt)
def handle_user_sign_up_attempt(sender, **kwargs):
    log(user=None,
        action="SIGNUP_ATTEMPTED",
        extra={
            "username": kwargs.get("username"),
            "email": kwargs.get("email"),
            "result": kwargs.get("result")
        })


@receiver(user_signed_up)
def handle_user_signed_up(sender, **kwargs):
    log(user=kwargs.get("user"), action="USER_SIGNED_UP", extra={})
