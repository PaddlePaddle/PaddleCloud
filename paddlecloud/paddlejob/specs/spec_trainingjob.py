def get_trainingjob(paddlejob):
    spec = {
        "apiVersion": "paddlepaddle.org/v1",
        "kind": "TrainingJob",
        "metadata":{
            "name": paddlejob.name,
        },
        "spec": {
            "image": paddlejob.image,
            "port": paddlejob.port,
            "ports_num_for_sparse": paddlejob.ports_num_for_sparse,
            "fault_tolerant": paddlejob.fault_tolerant,
            "trainer": {
                "entrypoint": paddlejob.entry,
                "workspace": paddlejob.job_package,
                "passes": paddlejob.passes,
                "min-instance": paddlejob.min_instance,
                "max-instance": paddlejob.max_instance,
                "resources": {
                    "requests": {
                        "memory": str(paddlejob.memory),
                        "cpu": str(paddlejob.cpu)
                    },
                    "limits": {
                        "memory": str(paddlejob.memory),
                        "cpu" : str(paddlejob.cpu * 1.5)
                    }
                }
            },
            "pservser": {
                "min-instance": paddlejob.pservers,
                "max-instance": paddlejob.pservers,
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
            }
        }
    }
    if paddlejob.gpu > 0:
        spec["spec"]["trainer"]["resources"]["limits"]["alpha.kubernetes.io/nvidia-gpu"] = str(paddlejob.gpu)
    return spec