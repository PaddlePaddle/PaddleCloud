from cephfs_volume import CephFSVolume
from hostpath_volume import HostPathVolume

class Volume(object):
    def __init__(self):
        pass

    @property
    def volume(self):
        return {}

    @property
    def volume_mount(self):
        return {}

__all__ = ["CephFSVolume", "HostPathVolume"]
