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

def get_spec_pserver(paddlejob):
    rs = {
        "apiVersion": "extensions/v1beta1",
        "kind": "ReplicaSet",
        "metadata": {
            "name": paddlejob.get_pserver_name(),
        },
        "spec": {
            "replicas": paddlejob.pservers,
            "template": {
                "metadata": {
                    "labels": paddlejob.get_pserver_labels()
                },
                "spec": {
                    "volumes": paddlejob.get_trainer_volumes(),
                    "containers": [{
                        "name": paddlejob.name,
                        "image": paddlejob.image,
                        "ports": paddlejob.get_pserver_container_ports(),
                        "env": paddlejob.get_env(),
                        "volumeMounts": paddlejob.get_trainer_volume_mounts(),
                        "command": paddlejob.get_pserver_entrypoint(),
                        "resources": {
                            "requests": {
                                "memory": str(paddlejob.psmemory),
                                "cpu": str(paddlejob.pscpu)
                            },
                            "limits": {
                                "memory": str(paddlejob.psmemory),
                                "cpu": str(paddlejob.pscpu * 1.5)
                            }
                        }
                    }]
                }
            }
        }
    }
    if paddlejob.registry_secret:
        rs["spec"]["template"]["spec"].update({
            "imagePullSecrets": [{
                "name": paddlejob.registry_secret
            }]
        })
    return rs
