from django.conf import settings
from django.conf.urls import include, url
from django.conf.urls.static import static
from django.views.generic import TemplateView

from django.contrib import admin

import notebook.views

urlpatterns = [
    url(r"^$", TemplateView.as_view(template_name="homepage.html"), name="home"),
    url(r"^admin/", include(admin.site.urls)),
    url(r"^account/signup/$", notebook.views.SignupView.as_view(), name="account_signup"),
    url(r"^account/login/$", notebook.views.LoginView.as_view(), name="account_login"),
    url(r"^account/", include("account.urls")),
    url(r"^notebook/", notebook.views.notebook_view),
]

urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
