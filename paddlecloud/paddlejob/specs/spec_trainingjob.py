def get_trainingjob(paddlejob):
    trainer = get_spec_trainer(paddlejob)
    pserver = get_spec_pserver(paddlejob)
    master  = get_spec_master(paddlejob)

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
            "trainer": trainer["spec"],
            "pservser": pserver["spec"],
            "master": master["spec"]
        }
    }

    trainer["spec"]["min-instance"] = paddlejob.min_instance
    trainer["spec"]["max-instance"] = paddlejob.max_instance
    pserver["spec"]["min-instance"] = paddlejob.pservers
    pserver["spec"]["max-instance"] = paddlejob.pservers

    if paddlejob.gpu > 0:
        spec["spec"]["trainer"]["resources"]["limits"]["alpha.kubernetes.io/nvidia-gpu"] = str(paddlejob.gpu)

    return spec
