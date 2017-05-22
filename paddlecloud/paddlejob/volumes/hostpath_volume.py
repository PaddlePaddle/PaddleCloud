from volume import Volume
__all__ = ["HostPathVolume"]
class HostPathVolume(Volume):
    def __init__(self, name, host_path, mount_path):
        self._name = name
        self._host_path = host_path
        self._mount_path = mount_path

    @property
    def volume(self):
        return {
            "name": self._name,
            "hostPath": {
                "path": self._host_path
            }
        }

    @property
    def volume_mount(self):
        return {
            "name": self._name,
            "mountPath": self._mount_path
        }
