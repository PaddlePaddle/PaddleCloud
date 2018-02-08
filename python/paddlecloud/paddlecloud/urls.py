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

from django.conf import settings
from django.conf.urls import include, url
from django.conf.urls.static import static
from django.views.generic import TemplateView

from django.contrib import admin
import account.urls

import notebook.views
import paddlejob.views
from rest_framework.authtoken import views
from rest_framework import routers

urlpatterns = [
    url(r"^$",
        TemplateView.as_view(template_name="homepage.html"),
        name="home"),
    url(r"^healthz/", notebook.views.healthz),
    url(r"^admin/", include(admin.site.urls)),
    url(r"^account/signup/$",
        notebook.views.SignupView.as_view(),
        name="account_signup"),
    url(r"^account/login/$",
        notebook.views.LoginView.as_view(),
        name="account_login"),
    url(r"^account/settings/$",
        notebook.views.SettingsView.as_view(),
        name="account_settings"),
    url(r"^account/certs/$",
        notebook.views.user_certs_view,
        name="account_certs"),
    url(r"^account/", include("account.urls")),
    url(r"^notedash/", notebook.views.notebook_view),
    url(r"^notestop/", notebook.views.stop_notebook_backend),
    url(r"^certsdown/", notebook.views.user_certs_download),
    url(r"^certsgen/", notebook.views.user_certs_generate),
    url(r'^api-token-auth/', views.obtain_auth_token),
    url(r'^api/sample/$', notebook.views.SampleView.as_view()),
    url(r"^api/v1/jobs/", paddlejob.views.JobsView.as_view()),
    url(r"^api/v1/trainingjobs/", paddlejob.views.TrainingJobsView.as_view()),
    url(r"^api/v1/pservers/", paddlejob.views.PserversView.as_view()),
    url(r"^api/v1/logs/", paddlejob.views.LogsView.as_view()),
    url(r"^api/v1/workers/", paddlejob.views.WorkersView.as_view()),
    url(r"^api/v1/quota/", paddlejob.views.QuotaView.as_view()),
    url(r"^api/v1/file/", paddlejob.views.SimpleFileView.as_view()),
    url(r"^api/v1/token2user/", paddlejob.views.GetUserView.as_view()),
    url(r"^api/v1/filelist/", paddlejob.views.SimpleFileList.as_view()),
    url(r"^api/v1/registry/", paddlejob.registry.RegistryView.as_view()),
    url(r"^api/v1/publish/", paddlejob.views.FilePublishAPIView.as_view()),
    url(r"^filepub/", paddlejob.views.file_publish_view),
]

urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
urlpatterns += static(settings.STATIC_URL, document_root=settings.STATIC_ROOT)
