import subprocess
import os
from django.conf import settings

def __check_cert_requirements__(program):
    def is_exe(fpath):
        return os.path.isfile(fpath) and os.access(fpath, os.X_OK)

    fpath, fname = os.path.split(program)
    if fpath:
        if is_exe(program):
            return program
    else:
        for path in os.environ["PATH"].split(os.pathsep):
            path = path.strip('"')
            exe_file = os.path.join(path, program)
            if is_exe(exe_file):
                return exe_file

    return None

def create_user_cert(ca_path, username):
    """
        @ca_path directory that contains ca.pem and ca-key.pem
    """
    if not username:
        raise AttributeError("username must be specified!")
    if not __check_cert_requirements__("openssl"):
        raise AssertionError("create user key depends on openssl command!")
    user_cert_cmds = []
    user_cert_dir = os.path.join(settings.USER_CERTS_PATH, username)
    user_cert_cmds.append("mkdir -p %s" % user_cert_dir)
    user_cert_cmds.append("openssl genrsa -out \
        %s/%s-key.pem 2048"%(user_cert_dir, username))
    user_cert_cmds.append("openssl req -new -key %s/%s-key.pem -out\
        %s/%s.csr -subj \"/CN=%s\""%\
        (user_cert_dir, username,
        user_cert_dir, username, username))
    # FIXME(gongwb):why need delete ca.srl when mount afs path while not need
    # when mount hostpath?
    user_cert_cmds.append("rm -f %s/ca.srl" % settings.USER_CERTS_PATH)
    user_cert_cmds.append("openssl x509 -req -in %s/%s.csr -CA %s -CAkey %s \
        -CAcreateserial -out %s/%s.pem -days 365"% \
        (user_cert_dir, username,
        settings.CA_PATH, settings.CA_KEY_PATH,
        user_cert_dir, username))
    for cmd in user_cert_cmds:
        process = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE)
        process.wait()
        out, err = process.communicate()
        if process.returncode != 0:
            raise RuntimeError("%s error with: (%d) - %s" % (cmd, process.returncode, err))
