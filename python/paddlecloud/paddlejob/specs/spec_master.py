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


def get_spec_master(paddlejob):
    return {
        "apiVersion": "extensions/v1beta1",
        "kind": "ReplicaSet",
        "metadata": {
            "name": paddlejob.get_master_name(),
        },
        "spec": {
            "replicas": 1,  # NOTE: always 1 replica of master
            "template": {
                "metadata": {
                    "labels": paddlejob.get_master_labels()
                },
                "spec": {
                    # mount trainer volumes to dispatch datasets
                    "volumes": paddlejob.get_trainer_volumes(),
                    "containers": [{
                        "name": paddlejob.name,
                        "image": paddlejob.image,
                        "ports": paddlejob.get_master_container_ports(),
                        "env": paddlejob.get_env(),
                        "volumeMounts": paddlejob.get_trainer_volume_mounts(),
                        "command": paddlejob.get_master_entrypoint(),
                        "resources": {
                            "requests": {
                                "memory": str(paddlejob.mastermemory),
                                "cpu": str(paddlejob.mastercpu)
                            },
                            "limits": {
                                "memory": str(paddlejob.mastermemory),
                                "cpu": str(paddlejob.mastercpu)
                            }
                        }
                    }, {
                        "name": paddlejob.name + "-etcd",
                        "image": paddlejob.etcd_image,
                        "command": [
                            "etcd", "-name", "etcd0", "-advertise-client-urls",
                            "http://$(POD_IP):2379,http://$(POD_IP):4001",
                            "-listen-client-urls",
                            "http://0.0.0.0:2379,http://0.0.0.0:4001",
                            "-initial-advertise-peer-urls",
                            "http://$(POD_IP):2380", "-listen-peer-urls",
                            "http://0.0.0.0:2380", "-initial-cluster",
                            "etcd0=http://$(POD_IP):2380",
                            "-initial-cluster-state", "new"
                        ],
                        "env": [{
                            "name": "POD_IP",
                            "valueFrom": {
                                "fieldRef": {
                                    "fieldPath": "status.podIP"
                                }
                            }
                        }]
                    }]
                }
            }
        }
    }
