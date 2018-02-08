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


def get_spec_trainer(paddlejob):
    job = {
        "apiVersion": "batch/v1",
        "kind": "Job",
        "metadata": {
            "name": paddlejob.get_trainer_name(),
        },
        "spec": {
            "parallelism": paddlejob.parallelism,
            "completions": paddlejob.parallelism,
            "template": {
                "metadata": {
                    "labels": paddlejob.get_trainer_labels()
                },
                "spec": {
                    "volumes": paddlejob.get_trainer_volumes(),
                    "containers": [{
                        "name": "trainer",
                        "image": paddlejob.image,
                        "imagePullPolicy": "Always",
                        "command": paddlejob.get_trainer_entrypoint(),
                        "env": paddlejob.get_env(),
                        "volumeMounts": paddlejob.get_trainer_volume_mounts(),
                        "resources": {
                            "requests": {
                                "memory": str(paddlejob.memory),
                                "cpu": str(paddlejob.cpu)
                            },
                            "limits": {
                                "memory": str(paddlejob.memory),
                                "cpu": str(paddlejob.cpu * 1.5)
                            }
                        }
                    }],
                    "restartPolicy": "Never"
                }
            }
        }
    }
    if paddlejob.registry_secret:
        job["spec"]["template"]["spec"].update({
            "imagePullSecrets": [{
                "name": paddlejob.registry_secret
            }]
        })
    if paddlejob.gpu > 0:
        job["spec"]["template"]["spec"]["containers"][0]["resources"][
            "limits"]["alpha.kubernetes.io/nvidia-gpu"] = str(paddlejob.gpu)
    return job
