__all__ = ["CephFSVolume"]


class Volume(object):
    def __init__(self):
        pass

    @property
    def volume(self):
        return {}

    @property
    def volume_mount(self):
        return {}

class CephFSVolume(Volume):
    def __init__(self, monitors_addr, user, secret_name, mount_path, cephfs_path):
        self._monitors = monitors_addr
        self._user = user
        self._secret_name = secret_name
        self._mount_path = mount_path
        self._cephfs_path = cephfs_path

    @property
    def volume(self):
        return {
             "name": "cephfs",
             "cephfs":{
                "name": "cephfs",
                "monitors": self._monitors,
                "path": self._cephfs_path,
                "user": self._user,
                "secretRef": {
                    "name": self._secret_name
                }
             }
        }

    @property
    def volume_mount(self):
        return {
            "mountPath": self._mount_path,
            "name": "cephfs"
        }
