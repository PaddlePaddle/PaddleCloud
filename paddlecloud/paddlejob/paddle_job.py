import kubernetes
from kubernetes import client, config
import os

from specs import spec_master, spec_pserver, spec_trainer

DEFAULT_PADDLE_PORT=7164
DEFAULT_MASTER_PORT=8080
DEFAULT_ETCD_PORT=2379

class UniversionedAPI(object):
    """
        Base defination for Paddle Cloud API fields.
    """
    required = {
        "name": "The name of the job.",
        "job_package": "The folder containing the job programs.",
        "parallelism": "The number of trainers to launch.",
        "cpu": "The CPU resource for each trainer.",
        "memory": "The memory resource for each trainer.",
        "pservers": "The number of pservers to launch.",
        "pscpu": "The CPU resource for each pserver.",
        "psmemory": "The memory resouce for each pserver.",
        "topology": "The Paddle V1 config file.",
        "entry": "The command to run.",
        "dc": "The datacenter specs."
    }
    optional = {
        "image": "The docker image to use",
        "passes": "The number of passes to run.",
        "gpu": "The number of GPU for each trainer.",
        "fault_tolerant": "Whether using the new fault-tolerant mode.",
        "volumes": "The data volumes to mount on pod.",
        "registry_secret": "The secret for reading registry.",
        "envs": "The environment variables for all pods",
        "etcd_image": "The etcd docker image.",
        "min_instance": "The minimum number of trainers to launch, only used for faulttolerant.",
        "max_instance": "The maximum number of trainers to launch, only used for faulttolerant."
    }
    optional_defaults = {
        "image": "",
        "passes": 1,
        "gpu": 0,
        "fault_tolerant": False,
        "volumes": [],
        "registry_secret": "",
        "envs": {},
        "etcd_image": "quay.io/coreos/etcd:v3.2.1"
    }
    # do not expose to user attributes.
    internal = {
        "ports_num": "ports_num argument for trainer.",
        "ports_num_for_sparse": "ports_num_for_sparse argument for trainer.",
        "mastercpu": "master process cpu resource.",
        "mastermemory": "master process memory resource."
    }
    # internal_defaults may be changed during setup.
    internal_defaults = {
        "ports_num": 1,
        "ports_num_for_sparse": 1,
        "mastercpu": 1,
        "mastermemory": "300Mi",
        "num_gradient_servers": 1
    }

class APIV1(UniversionedAPI):
    """
        For v1 implementation
    """
    pass

class PaddleJob(object):
    """
        PaddleJob Abstraction.
        A job can be submited to any cluster environment
        using one submit engine.
    """
    def __init__(self, **kwargs):
        self.apiv1 = APIV1()
        # setup required
        for k in self.apiv1.required:
            if k not in kwargs:
                raise Exception("Field required: %s" % k)
            setattr(self, k, kwargs[k])
        for k in self.apiv1.optional:
            if k in kwargs:
                setattr(self, k, kwargs[k])
            else:
                setattr(self, k, self.apiv1.optional_defaults[k])
        for k in self.apiv1.internal:
            setattr(self, k, self.apiv1.internal_defaults[k])

        self.num_gradient_servers = self.parallelism

    def get_master_name(self):
        return "%s-master" % self.name

    def get_pserver_name(self):
        return "%s-pserver" % self.name

    def get_trainer_name(self):
        return "%s-trainer" % self.name

    def get_env(self):
        envs = []
        envs.append({"name":"PADDLE_JOB_NAME",      "value":self.name})
        envs.append({"name":"TRAINERS",             "value":str(self.parallelism)})
        envs.append({"name":"PSERVERS",             "value":str(self.pservers)})
        envs.append({"name":"TOPOLOGY",             "value":self.topology})
        envs.append({"name":"ENTRY",                "value":self.entry})
        envs.append({"name":"TRAINER_PACKAGE",      "value":self.job_package})
        envs.append({"name":"PADDLE_INIT_PORT",     "value":str(DEFAULT_PADDLE_PORT)})
        if self.gpu > 0:
            envs.append({"name":"PADDLE_INIT_TRAINER_COUNT", "value":str(self.gpu)})
        else:
            envs.append({"name":"PADDLE_INIT_TRAINER_COUNT", "value":str(self.cpu)})
        envs.append({"name":"PADDLE_INIT_PORTS_NUM",            "value":str(self.ports_num)})
        envs.append({"name":"PADDLE_INIT_PORTS_NUM_FOR_SPARSE", "value":str(self.ports_num_for_sparse)})
        envs.append({"name":"PADDLE_INIT_NUM_GRADIENT_SERVERS", "value":str(self.num_gradient_servers)})
        envs.append({"name":"PADDLE_INIT_NUM_PASSES",           "value":str(self.passes)})
        if self.gpu:
            envs.append({"name":"PADDLE_INIT_USE_GPU", "value":str("1")})
            # HACK: add nvidia lib LD_LIBRARY_PATH for all pods
            envs.append({"name":"LD_LIBRARY_PATH",                  "value":"/usr/local/nvidia/lib64"})
        else:
            envs.append({"name":"PADDLE_INIT_USE_GPU", "value":str("0")})
        envs.append({"name":"NAMESPACE", "valueFrom":{
            "fieldRef":{"fieldPath":"metadata.namespace"}}})
        if self.envs:
            for k, v in self.envs.items():
                envs.append({"name": k, "value": v})
        return envs

    def get_pserver_container_ports(self):
        ports = []
        port = DEFAULT_PADDLE_PORT
        for i in xrange(self.ports_num + self.ports_num_for_sparse):
            ports.append({"containerPort":port, "name":"jobport-%d" % i})
            port += 1
        return ports

    def get_master_container_ports(self):
        ports = []
        port = DEFAULT_MASTER_PORT
        ports.append({"containerPort": DEFAULT_MASTER_PORT, "name":"master-port"})
        ports.append({"containerPort": DEFAULT_ETCD_PORT, "name":"etcd-port"})
        return ports

    def get_master_labels(self):
        return {"paddle-job-master": self.name}

    def get_pserver_labels(self):
        return {"paddle-job-pserver": self.name}

    def get_master_entrypoint(self):
        return ["paddle_k8s", "start_master"]

    def get_pserver_entrypoint(self):
        if not self.fault_tolerant:
            return ["paddle_k8s", "start_pserver"]
        else:
            return ["paddle_k8s", "start_new_pserver"]

    def get_trainer_entrypoint(self):
        if self.entry:
            if self.fault_tolerant:
                return ["paddle_k8s", "start_new_trainer"]
            return ["paddle_k8s", "start_trainer", "v2"]
        return ["paddle_k8s", "start_trainer", "v1"]

    def get_trainer_labels(self):
        return {"paddle-job": self.name}


    def get_trainer_volumes(self):
        volumes = []
        for item in self.volumes:
            volumes.append(item["volume"])
        return volumes

    def get_trainer_volume_mounts(self):
        volume_mounts = []
        for item in self.volumes:
            volume_mounts.append(item["volume_mount"])
        return volume_mounts
    
    def new_master_job(self):
        return spec_master.get_spec_master(self)

    def new_pserver_job(self):
        return spec_pserver.get_spec_pserver(self)

    def new_trainer_job(self):
        return spec_trainer.get_spec_trainer(self)
