def get_spec_master(paddlejob):
    rs = {
        "apiVersion": "extensions/v1beta1",
        "kind": "ReplicaSet",
        "metadata":{
            "name": paddlejob.get_master_name(),
        },
        "spec":{
            "replicas": 1, # NOTE: always 1 replica of master
            "template": {
                "metadata": {
                    "labels": paddlejob.get_master_labels()
                },
                "spec": {
                    # mount trainer volumes to dispatch datasets
                    "volumes": paddlejob.get_trainer_volumes(),
                    "containers":[{
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
                        "command": ["etcd", "-name", "etcd0", 
                                    "-advertise-client-urls", "http://$(POD_IP):2379,http://$(POD_IP):4001", 
                                    "-listen-client-urls", "http://0.0.0.0:2379,http://0.0.0.0:4001", 
                                    "-initial-advertise-peer-urls", "http://$(POD_IP):2380", 
                                    "-listen-peer-urls", "http://0.0.0.0:2380", 
                                    "-initial-cluster", "etcd0=http://$(POD_IP):2380", 
                                    "-initial-cluster-state", "new"],
                        "env": [{
                            "name": "POD_IP",
                            "valueFrom": {"fieldRef": {"fieldPath": "status.podIP"}}
                        }]

                    }]
                }
            }
        }
    }
    if paddlejob.registry_secret:
        rs["spec"]["template"]["spec"].update({"imagePullSecrets": [{"name": paddlejob.registry_secret}]})
    return rs