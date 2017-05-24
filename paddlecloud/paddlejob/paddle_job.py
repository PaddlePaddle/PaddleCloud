import kubernetes
from kubernetes import client, config
import os
__all__ = ["PaddleJob"]
DEFAULT_PADDLE_PORT=7164

class PaddleJob(object):
    """
        PaddleJob
    """
    def __init__(self,
                 name,
                 job_package,
                 parallelism,
                 cpu,
                 memory,
                 pservers,
                 pscpu,
                 psmemory,
                 topology,
                 image,
                 passes,
                 gpu=0,
                 volumes=[],
                 registry_secret=None):

        self._ports_num=1
        self._ports_num_for_sparse=1
        self._num_gradient_servers=1

        self._name = name
        self._job_package = job_package
        self._parallelism = parallelism
        self._cpu = cpu
        self._gpu = gpu
        self._memory = memory
        self._pservers = pservers
        self._pscpu = pscpu
        self._psmemory = psmemory
        self._topology = topology
        self._image = image
        self._volumes = volumes
        self._registry_secret = registry_secret
        self._passes = passes

    @property
    def pservers(self):
        return self._pservers

    @property
    def parallelism(self):
        return self._parallelism

    @property
    def runtime_image(self):
        return self._image

    def _get_pserver_name(self):
        return "%s-pserver" % self._name

    def _get_trainer_name(self):
        return "%s-trainer" % self._name

    def get_env(self):
        envs = []
        envs.append({"name":"PADDLE_JOB_NAME",      "value":self._name})
        envs.append({"name":"TRAINERS",             "value":str(self._parallelism)})
        envs.append({"name":"PSERVERS",             "value":str(self._pservers)})
        envs.append({"name":"TOPOLOGY",             "value":self._topology})
        envs.append({"name":"TRAINER_PACKAGE",      "value":self._job_package})
        envs.append({"name":"PADDLE_INIT_PORT",     "value":str(DEFAULT_PADDLE_PORT)})
        envs.append({"name":"PADDLE_INIT_TRAINER_COUNT",        "value":str(self._parallelism)})
        envs.append({"name":"PADDLE_INIT_PORTS_NUM",            "value":str(self._ports_num)})
        envs.append({"name":"PADDLE_INIT_PORTS_NUM_FOR_SPARSE", "value":str(self._ports_num_for_sparse)})
        envs.append({"name":"PADDLE_INIT_NUM_GRADIENT_SERVERS", "value":str(self._num_gradient_servers)})
        envs.append({"name":"PADDLE_INIT_NUM_PASSES",           "value":str(self._passes)})
        if self._gpu:
            envs.append({"name":"PADDLE_INIT_USE_GPU", "value":str("true")})
        else:
            envs.append({"name":"PADDLE_INIT_USE_GPU", "value":str("false")})
        envs.append({"name":"NAMESPACE", "valueFrom":{
            "fieldRef":{"fieldPath":"metadata.namespace"}}})
        return envs

    def _get_pserver_container_ports(self):
        ports = []
        port = DEFAULT_PADDLE_PORT
        for i in xrange(self._ports_num + self._ports_num_for_sparse):
            ports.append({"containerPort":port, "name":"jobport-%d" % i})
            port += 1
        return ports

    def _get_pserver_labels(self):
        return {"paddle-job-pserver": self._name}

    def _get_pserver_entrypoint(self):
        return ["paddle_k8s", "start_pserver"]

    def _get_trainer_entrypoint(sefl):
        return ["paddle_k8s", "start_trainer", "v1"]

    def _get_trainer_labels(self):
        return {"paddle-job": self._name}


    def _get_trainer_volumes(self):
        volumes = []
        for item in self._volumes:
            volumes.append(item["volume"])
        return volumes

    def _get_trainer_volume_mounts(self):
        volume_mounts = []
        for item in self._volumes:
            volume_mounts.append(item["volume_mount"])
        return volume_mounts

    def new_trainer_job(self):
        """
        return: Trainer job, it's a Kubernetes Job
        """
        job = {
            "apiVersion": "batch/v1",
            "kind": "Job",
            "metadata": {
                "name": self._get_trainer_name(),
            },
            "spec": {
                "parallelism": self._parallelism,
                "completions": self._parallelism,
                "template": {
                    "metadata":{
                        "labels": self._get_trainer_labels()
                    },
                    "spec": {
                        "volumes": self._get_trainer_volumes(),
                        "containers":[{
                            "name": "trainer",
                            "image": self._image,
                            "imagePullPolicy": "Always",
                            "command": self._get_trainer_entrypoint(),
                            "env": self.get_env(),
                            "volumeMounts": self._get_trainer_volume_mounts()
                        }],
                        "restartPolicy": "Never"
                    }
                }
            }
        }
        if self._registry_secret:
            job["spec"]["template"]["spec"].update({"imagePullSecrets": [{"name": self._registry_secret}]})
        return job
    def new_pserver_job(self):
        """
        return: PServer job, it's a Kubernetes ReplicaSet
        """
        rs = {
            "apiVersion": "extensions/v1beta1",
            "kind": "ReplicaSet",
            "metadata":{
                "name": self._get_pserver_name(),
            },
            "spec":{
                "replicas": self._pservers,
                "template": {
                    "metadata": {
                        "labels": self._get_pserver_labels()
                    },
                    "spec": {
                        "containers":[{
                            "name": self._name,
                            "image": self._image,
                            "ports": self._get_pserver_container_ports(),
                            "env": self.get_env(),
                            "command": self._get_pserver_entrypoint()
                        }]
                    }
                }
            }
        }
        if self._registry_secret:
            rs["spec"]["template"]["spec"].update({"imagePullSecrets": [{"name": self._registry_secret}]})
        return rs
