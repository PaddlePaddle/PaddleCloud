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

import json

__all__ = ["get_volume_config"]

tmpl_volume = {
    "hostpath": "{\"name\": $NAME, \"hostPath\":{\"path\": $HOST_PATH}}",
    "cephfs": "{\"name\": $NAME,\"cephfs\":{\"name\": \"cephfs\", \
               \"monitors\": $MONITORS_ADDR,\"path\": $CEPHFS_PATH, \
               \"readOnly\": $READ_ONLY, \"user\": $USER, \
               \"secretRef\": {\"name\": $SECRET}}}"
}
tmpl_volume_mount = {
    "hostpath": "{\"name\": $NAME, \"mountPath\":$MOUNT_PATH}",
    "cephfs": "{\"mountPath\": $MOUNT_PATH, \"name\": $NAME}"
}


def __render(tmpl, **kwargs):
    for k, v in kwargs.items():
        tmpl_k = "$%s" % k.upper()
        if tmpl.find(tmpl_k) != -1:
            if type(v) is str or type(v) is unicode:
                tmpl = tmpl.replace(tmpl_k, "\"%s\"" % v)
            elif type(v) is list or type(v) is bool:
                tmpl = tmpl.replace(tmpl_k, json.dumps(v))
            else:
                pass
    return tmpl


def __get_template(tmpls, fstype):
    if fstype in tmpls.keys():
        return tmpls[fstype]
    else:
        return ""


def get_volume_config(**kwargs):
    """
    :param fstype: which filesystem type
    :type fstype: str

    if fstype is host_path:

    :param name: a unique name for a Kubernetes job configuration
    :type name: str
    :param mount_path: path in pod
    :type mount_path: str
    :param host_path: path no the host
    :type host_path: str

    if fstype is cephfs:

    :param name: unique name for a Kubernetes Job configuration
    :type name: str
    :param monitors_addr: the CephFS monitors address
    :type monitors_addr: list
    :param cephfs_path: CephFS Path
    :type cephfs_path: str
    :param user: ceph cluster user
    :type user: str
    :param secret: Kubernetes Secret for Ceph secret
    :type secret: str
    :param mount_path: mount path in Pod
    :type mount_path: str
    """
    fstype = kwargs["fstype"]
    tmpl_v = __get_template(tmpl_volume, fstype)
    tmpl_vm = __get_template(tmpl_volume_mount, fstype)
    return {
        "volume": json.loads(__render(
            tmpl=tmpl_v, **kwargs)),
        "volume_mount": json.loads(__render(
            tmpl=tmpl_vm, **kwargs))
    }
