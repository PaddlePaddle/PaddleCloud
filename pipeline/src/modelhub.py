import argparse
import logging
import launch_crd

from kubernetes import client as k8s_client
from kubernetes import config


class ModelHub(launch_crd.K8sCR):

    def __init__(self, client=None):
        super(ModelHub, self).__init__("batch", "jobs", "v1", client)

    def get_action_status(self, action=None):
        return ["Complete"], ["Failed"]

    def get_spec(self, spec):
        endpoint = spec.get("endpoint")
        model_name = spec.get('model_name')
        model_version = spec.get("model_version")
        model_file = f"/mnt/task-center/{model_name}.tar.gz"

        upload_shell = """
endpoint=$0
model_file=$1
model_name=$2
version=$3

# MINIO_ACCESS_ID and MINIO_SECRET_KEY is from environment variable
mc config host add minio ${endpoint} ${MINIO_ACCESS_KEY} ${MINIO_SECRET_KEY} --api s3v4

# make bucket of model-center
mc mb --ignore-existing minio/model-center

# change the policy to download of bucket model-center
mc policy set download minio/model-center/

# upload model file to model-center bucket of minio
mc cp $model_file minio/model-center/${model_name}/${version}/

echo "upload model to minio/model-center/${model_name}/${version}/"
"""

        container = {
            "name": "uploader",
            "image": "xiaolao/model-uploader:latest",
            "command": ["sh", "-exc", upload_shell],
            "args": [endpoint, model_file, model_name, model_version],
            "volumeMounts": [{
                "name": "task-center",
                "mountPath": "/mnt/task-center/",
            }],
            "env": [
                {
                    "name": "MINIO_ACCESS_KEY",
                    "valueFrom": {
                        "secretKeyRef": {
                            "key": "access-key",
                            "name": "data-center"
                        }
                    }
                },
                {
                    "name": "MINIO_SECRET_KEY",
                    "valueFrom": {
                        "secretKeyRef": {
                            "key": "secret-key",
                            "name": "data-center"
                        }
                    }
                },
            ]
        }

        job = {
            "apiVersion": "batch/v1",
            "kind": "Job",
            "metadata": {
                "name": f"{spec.get('name')}-model-hub",
                "namespace": spec.get("namespace"),
            },
            "spec": {
                "template": {
                    "spec": {
                        "containers": [container],
                        "restartPolicy": "Never",
                        "volumes": [{
                            "name": "task-center",
                            "persistentVolumeClaim": {
                                "claimName": spec.get("pvc_name"),
                            }
                        }]
                    }
                }
            }
        }

        return job


def main():
    parser = argparse.ArgumentParser(description='model-hub')
    parser.add_argument('--name', type=str,
                        help='The name of DataSet.')
    parser.add_argument('--namespace', type=str,
                        default='kubeflow',
                        help='The namespace of training task.')
    parser.add_argument('--action', type=str, default='apply',
                        help='Action to execute on model uploader job.')

    parser.add_argument('--endpoint', type=str, default="http://minio-service.kubeflow:9000",
                        help='The endpoint uri of minio object storage service.')
    parser.add_argument('--model_name', type=str, required=True,
                        help='The name of model.')
    parser.add_argument('--model_version', type=str, default="latest",
                        help='The version of model.')
    parser.add_argument('--pvc_name', type=str, required=True,
                        help='The persistent volume claim name of task-center.')

    args = parser.parse_args()

    logging.getLogger().setLevel(logging.INFO)
    logging.info('Generating DataSet template.')

    model_hub_spec = {
        "name": args.name,
        "namespace": args.namespace,
        "endpoint": args.endpoint,
        "model_name": args.model_name,
        "model_version": args.model_version,
        "pvc_name": args.pvc_name
    }

    config.load_incluster_config()
    api_client = k8s_client.ApiClient()
    model_hub = ModelHub(client=api_client)
    model_hub.run(model_hub_spec, action=args.action)
    if args.action != "delete":
        model_hub.run(model_hub_spec, action="delete")


if __name__ == "__main__":
    main()
