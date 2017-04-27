# -*- coding: utf-8 -*-
from __future__ import unicode_literals

from django.db import models
from django.contrib.auth.models import User

class PaddleUser(models.Model):
    user = models.OneToOneField(User, on_delete=models.CASCADE)
    school = models.CharField(max_length=256)
    studentID = models.CharField(max_length=512)
    major = models.CharField(max_length=256)
