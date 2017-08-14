import kubernetes
from kubernetes import client, config
import os
__all__ = ["PaddleJob"]
DEFAULT_PADDLE_PORT=7164
DEFAULT_MASTER_PORT=8080
DEFAULT_ETCD_PORT=2379

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
                 entry,
                 image,
                 passes,
                 gpu=0,
                 volumes=[],
                 registry_secret=None,
                 envs = {},
                 fault_tolerant=False,
                 etcd_image="quay.io/coreos/etcd:v3.2.1"):

        self._ports_num=1
        self._ports_num_for_sparse=1
        self._num_gradient_servers=parallelism

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
        self._entry = entry
        self._image = image
        self._volumes = volumes
        self._registry_secret = registry_secret
        self._passes = passes
        self._usr_envs = envs
        # master resources are static
        self._mastercpu = 1
        self._mastermemory = "300Mi"
        # use new pserver for tolerant
        self._fault_tolerant = fault_tolerant
        self._etcd_image = etcd_image

    @property
    def pservers(self):
        return self._pservers

    @property
    def parallelism(self):
        return self._parallelism

    @property
    def runtime_image(self):
        return self._image

    def _get_master_name(self):
        return "%s-master" % self._name

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
        envs.append({"name":"ENTRY",                "value":self._entry})
        envs.append({"name":"TRAINER_PACKAGE",      "value":self._job_package})
        envs.append({"name":"PADDLE_INIT_PORT",     "value":str(DEFAULT_PADDLE_PORT)})
        if self._gpu > 0:
            envs.append({"name":"PADDLE_INIT_TRAINER_COUNT", "value":str(self._gpu)})
        else:
            envs.append({"name":"PADDLE_INIT_TRAINER_COUNT", "value":str(self._cpu)})
        envs.append({"name":"PADDLE_INIT_PORTS_NUM",            "value":str(self._ports_num)})
        envs.append({"name":"PADDLE_INIT_PORTS_NUM_FOR_SPARSE", "value":str(self._ports_num_for_sparse)})
        envs.append({"name":"PADDLE_INIT_NUM_GRADIENT_SERVERS", "value":str(self._num_gradient_servers)})
        envs.append({"name":"PADDLE_INIT_NUM_PASSES",           "value":str(self._passes)})
        if self._gpu:
            envs.append({"name":"PADDLE_INIT_USE_GPU", "value":str("1")})
            # HACK: add nvidia lib LD_LIBRARY_PATH for all pods
            envs.append({"name":"LD_LIBRARY_PATH",                  "value":"/usr/local/nvidia/lib64"})
        else:
            envs.append({"name":"PADDLE_INIT_USE_GPU", "value":str("0")})
        envs.append({"name":"NAMESPACE", "valueFrom":{
            "fieldRef":{"fieldPath":"metadata.namespace"}}})
        if self._usr_envs:
            for k, v in self._usr_envs.items():
                envs.append({"name": k, "value": v})
        return envs

    def _get_pserver_container_ports(self):
        ports = []
        port = DEFAULT_PADDLE_PORT
        for i in xrange(self._ports_num + self._ports_num_for_sparse):
            ports.append({"containerPort":port, "name":"jobport-%d" % i})
            port += 1
        return ports

    def _get_master_container_ports(self):
        ports = []
        port = DEFAULT_MASTER_PORT
        ports.append({"containerPort": DEFAULT_MASTER_PORT, "name":"master-port"})
        ports.append({"containerPort": DEFAULT_ETCD_PORT, "name":"etcd-port"})
        return ports

    def _get_master_labels(self):
        return {"paddle-job-master": self._name}

    def _get_pserver_labels(self):
        return {"paddle-job-pserver": self._name}

    def _get_master_entrypoint(self):
        return ["paddle_k8s", "start_master"]

    def _get_pserver_entrypoint(self):
        if not self._fault_tolerant:
            return ["paddle_k8s", "start_pserver"]
        else:
            return ["paddle_k8s", "start_new_pserver"]

    def _get_trainer_entrypoint(self):
        if self._entry:
            if self._fault_tolerant:
                return ["paddle_k8s", "start_new_trainer"]
            return ["paddle_k8s", "start_trainer", "v2"]
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

    def new_master_job(self):
        """
        return: Master ReplicaSet
        """
        rs = {
            "apiVersion": "extensions/v1beta1",
            "kind": "ReplicaSet",
            "metadata":{
                "name": self._get_master_name(),
            },
            "spec":{
                "replicas": 1, # NOTE: always 1 replica of master
                "template": {
                    "metadata": {
                        "labels": self._get_master_labels()
                    },
                    "spec": {
                        # mount trainer volumes to dispatch datasets
                        "volumes": self._get_trainer_volumes(),
                        "containers":[{
                            "name": self._name,
                            "image": self._image,
                            "ports": self._get_master_container_ports(),
                            "env": self.get_env(),
                            "volumeMounts": self._get_trainer_volume_mounts(),
                            "command": self._get_master_entrypoint(),
                            "resources": {
                                "requests": {
                                    "memory": str(self._mastermemory),
                                    "cpu": str(self._mastercpu)
                                },
                                "limits": {
                                    "memory": str(self._mastermemory),
                                    "cpu": str(self._mastercpu)
                                }
                            }
                        }, {
                            "name": self._name + "-etcd",
                            "image": self._etcd_image,
                            "command": ["etcd", "-name", "etcd0", "-advertise-client-urls", "http://$(POD_IP):2379,http://$(POD_IP):4001", "-listen-client-urls", "http://0.0.0.0:2379,http://0.0.0.0:4001", "-initial-advertise-peer-urls", "http://$(POD_IP):2380", "-listen-peer-urls", "http://0.0.0.0:2380", "-initial-cluster", "etcd0=http://$(POD_IP):2380", "-initial-cluster-state", "new"],
                            "env": [{
                                "name": "POD_IP",
                                "valueFrom": {"fieldRef": {"fieldPath": "status.podIP"}}
                            }]

                        }]
                    }
                }
            }
        }
        return rs

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
                            "volumeMounts": self._get_trainer_volume_mounts(),
                            "resources": {
                                "requests": {
                                    "memory": str(self._memory),
                                    "cpu": str(self._cpu)
                                },
                                "limits": {
                                    "memory": str(self._memory),
                                    "cpu" : str(self._cpu*1.5)
                                }
                            }
                        }],
                        "restartPolicy": "Never"
                    }
                }
            }
        }
        if self._registry_secret:
            job["spec"]["template"]["spec"].update({"imagePullSecrets": [{"name": self._registry_secret}]})
        if self._gpu > 0:
            job["spec"]["template"]["spec"]["containers"][0]["resources"]["limits"]["alpha.kubernetes.io/nvidia-gpu"] = str(self._gpu)
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
                        "volumes": self._get_trainer_volumes(),
                        "containers":[{
                            "name": self._name,
                            "image": self._image,
                            "ports": self._get_pserver_container_ports(),
                            "env": self.get_env(),
                            "volumeMounts": self._get_trainer_volume_mounts(),
                            "command": self._get_pserver_entrypoint(),
                            "resources": {
                                "requests": {
                                    "memory": str(self._psmemory),
                                    "cpu": str(self._pscpu)
                                },
                                "limits": {
                                    "memory": str(self._psmemory),
                                    "cpu": str(self._pscpu*1.5)
                                }
                            }
                        }]
                    }
                }
            }
        }
        if self._registry_secret:
            rs["spec"]["template"]["spec"].update({"imagePullSecrets": [{"name": self._registry_secret}]})
        return rs
