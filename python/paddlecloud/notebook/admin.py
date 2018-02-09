# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.contrib import admin
from django.contrib.auth.admin import UserAdmin as BaseUserAdmin
from django.contrib.auth.models import User

from notebook.models import PaddleUser


# Define an inline admin descriptor for PaddleUser model
# which acts a bit like a singleton
class PaddleUserInline(admin.StackedInline):
    model = PaddleUser
    can_delete = False
    verbose_name_plural = 'PaddleUser'


# Define a new User admin
class UserAdmin(BaseUserAdmin):
    inlines = (PaddleUserInline, )


# Re-register UserAdmin
admin.site.unregister(User)
admin.site.register(User, UserAdmin)
