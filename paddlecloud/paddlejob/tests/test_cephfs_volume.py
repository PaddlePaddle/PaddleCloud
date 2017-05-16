import unittest
from paddlejob import CephFSVolume
class CephFSVolumeTest(unittest.TestCase):
    def test_get_volume(self):
        cephfs_volume = CephFSVolume(
            monitors_addr="192.168.1.123:6789",
            user="admin",
            secret_name="cephfs-secret",
            mount_path="/mnt/cephfs",
            cephfs_path="/")
        self.assertEqual(cephfs_volume.volume["name"] , "cephfs")
if __name__=="__main__":
    unittest.main()
