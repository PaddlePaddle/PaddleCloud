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

import unittest
from volume import get_volume_config


class GetVolumeConfigTest(unittest.TestCase):
    def test_hostpath(self):
        volume = get_volume_config(
            fstype="hostpath",
            name="abc",
            mount_path="/pfs/dc1/home/yanxu05",
            host_path="/mnt/hdfs_mulan")
        self.assertEqual(volume["volume"]["name"], "abc")

    def test_cephfs(self):
        volume = get_volume_config(
            fstype="cephfs",
            name="cephfs",
            monitors_addr="192.168.2.1:6789,182.68.2.2:6789".split(","),
            cephfs_path="/a/d",
            user="admin",
            secret="ceph-secret",
            mount_path="/pfs/dc1/home/yanxu05")
        self.assertEqual(volume["volume"]["name"], "cephfs")
