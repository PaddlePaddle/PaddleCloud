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
    #url(r"^$", TemplateView.as_view(template_name="homepage.html"), name="home"),
    url(r"^$", TemplateView.as_view(template_name="homepage.html"), name="home"),
    url(r"^healthz/", notebook.views.healthz),
    url(r"^admin/", include(admin.site.urls)),
    url(r"^account/signup/$", notebook.views.SignupView.as_view(), name="account_signup"),
    url(r"^account/login/$", notebook.views.LoginView.as_view(), name="account_login"),
    url(r"^account/settings/$", notebook.views.SettingsView.as_view(), name="account_settings"),
    url(r"^account/certs/$", notebook.views.user_certs_view, name="account_certs"),
    url(r"^account/", include("account.urls")),
    url(r"^notedash/", notebook.views.notebook_view),
    url(r"^notestop/", notebook.views.stop_notebook_backend),
    url(r"^certsdown/", notebook.views.user_certs_download),
    url(r"^certsgen/", notebook.views.user_certs_generate),
    url(r'^api-token-auth/', views.obtain_auth_token),
    url(r'^api/sample/$', notebook.views.SampleView.as_view()),

    url(r"^api/v1/jobs/", paddlejob.views.JobsView.as_view()),
    url(r"^api/v1/logs/", paddlejob.views.LogsView.as_view()),
    url(r"^api/v1/workers/", paddlejob.views.WorkersView.as_view()),
    url(r"^api/v1/quota/", paddlejob.views.QuotaView.as_view()),
    url(r"^api/v1/file/", paddlejob.views.SimpleFileView.as_view()),
    url(r"^api/v1/filelist/", paddlejob.views.SimpleFileList.as_view()),
    url(r"^api/v1/registry/", paddlejob.registry.RegistryView.as_view()),
]

urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
urlpatterns += static(settings.STATIC_URL, document_root=settings.STATIC_ROOT)
