import os
from kubernetes import config

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), os.pardir))
PACKAGE_ROOT = os.path.abspath(os.path.dirname(__file__))
BASE_DIR = PACKAGE_ROOT

DEBUG = True

DATABASES = {
    "default": {
        "ENGINE": "django.db.backends.mysql",
        "NAME": "paddlecloud",
        'USER': 'root',
        'PASSWORD': 'root',
        'HOST': '127.0.0.1',   # Or an IP Address that your DB is hosted on
        'PORT': '3306',
    }
}

ALLOWED_HOSTS = [
    "127.0.0.1",
    "cloud.paddlepaddle.org",
]

POD_IP = os.getenv("POD_IP")
if POD_IP:
    ALLOWED_HOSTS.append(POD_IP)

REST_FRAMEWORK = {
    # Use Django's standard `django.contrib.auth` permissions,
    # or allow read-only access for unauthenticated users.
    'DEFAULT_AUTHENTICATION_CLASSES': (
        'rest_framework.authentication.BasicAuthentication',
        'rest_framework.authentication.SessionAuthentication',
        'rest_framework.authentication.TokenAuthentication',
    ),
    'DEFAULT_PERMISSION_CLASSES': [
        'rest_framework.permissions.DjangoModelPermissionsOrAnonReadOnly'
    ]
}

# Local time zone for this installation. Choices can be found here:
# http://en.wikipedia.org/wiki/List_of_tz_zones_by_name
# although not all choices may be available on all operating systems.
# On Unix systems, a value of None will cause Django to use the same
# timezone as the operating system.
# If running in a Windows environment this must be set to the same as your
# system time zone.
TIME_ZONE = "UTC"

# Language code for this installation. All choices can be found here:
# http://www.i18nguy.com/unicode/language-identifiers.html
LANGUAGE_CODE = "en-us"

SITE_ID = int(os.environ.get("SITE_ID", 1))

# If you set this to False, Django will make some optimizations so as not
# to load the internationalization machinery.
USE_I18N = True

# If you set this to False, Django will not format dates, numbers and
# calendars according to the current locale.
USE_L10N = True

# If you set this to False, Django will not use timezone-aware datetimes.
USE_TZ = True

# Absolute filesystem path to the directory that will hold user-uploaded files.
# Example: "/home/media/media.lawrence.com/media/"
MEDIA_ROOT = os.path.join(PACKAGE_ROOT, "site_media", "media")

# URL that handles the media served from MEDIA_ROOT. Make sure to use a
# trailing slash.
# Examples: "http://media.lawrence.com/media/", "http://example.com/media/"
MEDIA_URL = "/site_media/media/"

# Absolute path to the directory static files should be collected to.
# Don"t put anything in this directory yourself; store your static files
# in apps" "static/" subdirectories and in STATICFILES_DIRS.
# Example: "/home/media/media.lawrence.com/static/"
STATIC_ROOT = os.path.join(PACKAGE_ROOT, "site_media", "static")

# URL prefix for static files.
# Example: "http://media.lawrence.com/static/"
STATIC_URL = "/site_media/static/"

# Additional locations of static files
STATICFILES_DIRS = [
    os.path.join(PROJECT_ROOT, "static", "dist"),
]

STATICFILES_STORAGE = "django.contrib.staticfiles.storage.ManifestStaticFilesStorage"

# List of finder classes that know how to find static files in
# various locations.
STATICFILES_FINDERS = [
    "django.contrib.staticfiles.finders.FileSystemFinder",
    "django.contrib.staticfiles.finders.AppDirectoriesFinder",
]

# Make this unique, and don't share it with anybody.
SECRET_KEY = "vpu^(5mjr)*tloao^m$wlh)oc(fn1yoiqoq@m0$er((qlocq1k"

TEMPLATES = [
    {
        "BACKEND": "django.template.backends.django.DjangoTemplates",
        "DIRS": [
            os.path.join(PACKAGE_ROOT, "templates"),
        ],
        "APP_DIRS": True,
        "OPTIONS": {
            "debug": DEBUG,
            "context_processors": [
                "django.contrib.auth.context_processors.auth",
                "django.template.context_processors.debug",
                "django.template.context_processors.i18n",
                "django.template.context_processors.media",
                "django.template.context_processors.static",
                "django.template.context_processors.tz",
                "django.template.context_processors.request",
                "django.contrib.messages.context_processors.messages",
                "account.context_processors.account",
                "pinax_theme_bootstrap.context_processors.theme",
            ],
        },
    },
]

MIDDLEWARE = [
    "django.contrib.sessions.middleware.SessionMiddleware",
    "django.middleware.common.CommonMiddleware",
    "django.middleware.csrf.CsrfViewMiddleware",
    "django.contrib.auth.middleware.AuthenticationMiddleware",
    "django.contrib.auth.middleware.SessionAuthenticationMiddleware",
    "django.contrib.messages.middleware.MessageMiddleware",
    "django.middleware.clickjacking.XFrameOptionsMiddleware",
    "account.middleware.ExpiredPasswordMiddleware",
]

ROOT_URLCONF = "paddlecloud.urls"

# Python dotted path to the WSGI application used by Django's runserver.
WSGI_APPLICATION = "paddlecloud.wsgi.application"

INSTALLED_APPS = [
    "django.contrib.admin",
    "django.contrib.auth",
    "django.contrib.contenttypes",
    "django.contrib.messages",
    "django.contrib.sessions",
    "django.contrib.sites",
    "django.contrib.staticfiles",
    # token auth
    "rest_framework",
    "rest_framework.authtoken",
    # paddlecloud apps
    # NOTE: load before pinax_theme_bootstrap to customize the theme
    "notebook",

    # theme
    "bootstrapform",
    "pinax_theme_bootstrap",

    # external
    "account",
    "pinax.eventlog",
    "pinax.webanalytics",

    # project
    "paddlecloud",
]

# A sample logging configuration. The only tangible logging
# performed by this configuration is to send an email to
# the site admins on every HTTP 500 error when DEBUG=False.
# See http://docs.djangoproject.com/en/dev/topics/logging for
# more details on how to customize your logging configuration.
LOGGING = {
    "version": 1,
    "disable_existing_loggers": False,
    "filters": {
        "require_debug_false": {
            "()": "django.utils.log.RequireDebugFalse"
        }
    },
    'formatters': {
        'verbose': {
            'format': '[%(levelname)s %(asctime)s @ %(process)d] - %(message)s'
        },
        'simple': {
            'format': '%(levelname)s %(message)s'
        },
    },
    "handlers": {
        "mail_admins": {
            "level": "ERROR",
            "filters": ["require_debug_false"],
            "class": "django.utils.log.AdminEmailHandler"
        },
        "stdout": {
            "level": "INFO",
            "class": "logging.StreamHandler",
            "formatter": "verbose"
        },
    },
    "loggers": {
        "": {
            "handlers": ["stdout"],
            "level": "ERROR",
            "propagate": True,
        },
        "django.request": {
            "handlers": ["mail_admins"],
            "level": "ERROR",
            "propagate": True,
        },
    }
}

FIXTURE_DIRS = [
    os.path.join(PROJECT_ROOT, "fixtures"),
]

EMAIL_BACKEND = "django.core.mail.backends.console.EmailBackend"

LOGIN_URL="/account/login"

ACCOUNT_OPEN_SIGNUP = True
ACCOUNT_EMAIL_UNIQUE = True
ACCOUNT_EMAIL_CONFIRMATION_REQUIRED = False
ACCOUNT_LOGIN_REDIRECT_URL = "home"
ACCOUNT_LOGOUT_REDIRECT_URL = "home"
ACCOUNT_EMAIL_CONFIRMATION_EXPIRE_DAYS = 2
ACCOUNT_USE_AUTH_AUTHENTICATE = True
ACCOUNT_USER_DISPLAY = lambda user: user.email

ACCOUNT_PASSWORD_EXPIRY = 60*60*24*5  # seconds until pw expires, this example shows five days
ACCOUNT_PASSWORD_USE_HISTORY = True

AUTHENTICATION_BACKENDS = [
    "account.auth_backends.UsernameAuthenticationBackend",
]

# secret places to store ca and users keys
CA_PATH = "/certs/ca.pem"
CA_KEY_PATH = "/certs/ca-key.pem"
USER_CERTS_PATH="/certs"

K8S_HOST = "https://%s:%s" % (os.getenv("KUBERNETES_SERVICE_HOST"),
    os.getenv("KUBERNETES_SERVICE_PORT_HTTPS"))
# PADDLE_BOOK_IMAGE="docker.paddlepaddle.org/book:0.10.0rc2"
PADDLE_BOOK_IMAGE="yancey1989/book-cloud"
PADDLE_BOOK_PORT=8888

# ============== Datacenter Storage Config Samples ==============
#if Paddle cloud use CephFS as backend storage, configure CEPHFS_CONFIGURATION
#the following is an example:

#DATACENTERS = {
#   "datacenter1":{
#       "fstype": "cephfs",
#       "monitors_addr": "172.19.32.166:6789",
#       "secret": "ceph-secret",
#       "user": "admin",
#       "mount_path": "/pfs/%s/home/%s/", # mount_path % ( dc, username )
#       "cephfs_path": "/%s" # cephfs_path % username
#       "admin_key": "/certs/admin.secret"
#   }
#}
#for HostPath example:
#DATACENTERS = {
#   ...
#   "dc1":{
#       "fstype": "hostpath",
#       "host_path": "/mnt/hdfs/",
#       "mount_path" "/pfs/%s/home/%s/" # mount_path % ( dc, username )
#    }
#}
FSTYPE_CEPHFS = "cephfs"
FSTYPE_HOSTPATH = "hostpath"
DATACENTERS = {
    "meiyan":{
        "fstype": FSTYPE_CEPHFS,
        "monitors_addr": ["172.19.32.166:6789"],  # must be a list
        "secret": "ceph-secret",
        "user": "admin",
        "mount_path": "/pfs/%s/home/%s/", # mount_path % ( dc, username )
        "cephfs_path": "/%s", # cephfs_path % username
        "admin_key": "/certs/admin.secret",
    },
    "public": {
        "fstype": FSTYPE_CEPHFS,
        "monitors_addr": ["172.19.32.166:6789"],  # must be a list
        "secret": "ceph-secret",
        "user": "admin",
        "mount_path": "/pfs/%s/public/", # mount_path % ( dc, username )
        "cephfs_path": "/public", # cephfs_path % username
        "admin_key": "/certs/admin.secret",
        "read_only": True
    }
}
# where cephfs root is mounted when using cephfs storage service
STORAGE_PATH="/pfs"
# HACK: define use HDFS or CEPHFS, in cephfs mode jobpath will be /pfs/jobs/[jobname]
STORAGE_MODE="HDFS"

# ===================== Docker image registry =====================
JOB_DOCKER_IMAGE = {
    # These images are built by `docker/build_docker.sh` under this repo.
    "image": "typhoon1986/paddlecloud-job",
    "image_gpu": "typhoon1986/paddlecloud-job:gpu",
    # docker registry credentials
    "registry_secret": "job-registry-secret", # put this to None if not using registry login
    "docker_config":{"auths":
                     {"registry.baidu.com":
                      {"auth": "eWFueHUwNTpRTndVSGV1Rldl"}}}
}

# Path store all cuda, nvidia driver libs
NVIDIA_LIB_PATH="/usr/local/nvidia/lib64"
# etcd image for fault-tolerant jobs
ETCD_IMAGE="quay.io/coreos/etcd:v3.2.1"

# domains that allow notebook to enter
NOTEBOOK_DOMAINS=["www.paddlepaddle.org"]

# GPU limit for users
# TODO(Yancey1989): 
# 1. Implement 
# 2. Move GPU quota to Kubernetes
GPU_QUOTA={
    "DEFAULT": {
        "limit": 2
    },
    "yanxu05@baidu.com": {
        "limit": 5
    }
}