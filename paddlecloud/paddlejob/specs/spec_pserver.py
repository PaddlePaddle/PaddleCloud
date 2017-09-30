def get_spec_pserver(paddlejob):
    rs = {
        "apiVersion": "extensions/v1beta1",
        "kind": "ReplicaSet",
        "metadata":{
            "name": paddlejob.get_pserver_name(),
        },
        "spec":{
            "replicas": paddlejob.pservers,
            "template": {
                "metadata": {
                    "labels": paddlejob.get_pserver_labels()
                },
                "spec": {
                    "volumes": paddlejob.get_trainer_volumes(),
                    "containers":[{
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
        rs["spec"]["template"]["spec"].update({"imagePullSecrets": [{"name": paddlejob.registry_secret}]})
    return rs