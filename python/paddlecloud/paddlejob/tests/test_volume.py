import unittest
from volume import get_volume_config
class GetVolumeConfigTest(unittest.TestCase):
    def test_hostpath(self):
        volume = get_volume_config(
            fstype="hostpath", name="abc",mount_path="/pfs/dc1/home/yanxu05",
            host_path= "/mnt/hdfs_mulan")
        self.assertEqual(volume["volume"]["name"], "abc")
    def test_cephfs(self):
        volume = get_volume_config(
            fstype="cephfs", name="cephfs",
            monitors_addr="192.168.2.1:6789,182.68.2.2:6789".split(","),
            cephfs_path="/a/d", user="admin", secret="ceph-secret",
            mount_path="/pfs/dc1/home/yanxu05")
        self.assertEqual(volume["volume"]["name"], "cephfs")
