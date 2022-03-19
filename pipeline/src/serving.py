import argparse
import logging
import launch_crd

from kubernetes import client as k8s_client
from kubernetes import config


class Serving(launch_crd.K8sCR):

    def __init__(self, client=None):
        super(Serving, self).__init__("elasticserving.paddlepaddle.org", "paddleservices", "v1", client)

    def get_action_status(self, action=None):
        return None, None

    def get_spec(self, spec):
        endpoint = spec.get("endpoint")
        model_name = spec.get('model_name')
        model_version = spec.get('model_version')
        image_list = spec.get("image").split(":")
        if len(image_list) != 2:
            raise Exception("The format of image is not valid, it should contains tag")
        image, tag = image_list[0], image_list[1]

        arg = f"wget {endpoint}/model-center/{model_name}/{model_version}/{model_name}.tar.gz && "
        arg += f"tar xzf {model_name}.tar.gz && rm -rf {model_name}.tar.gz && "
        arg += f"python3 -m paddle_serving_server.serve --model {model_name}/server --port 9292"

        return {
            "apiVersion": "%s/%s" % (self.group, self.version),
            "kind": "PaddleService",
            "metadata": {
                "name": spec.get("name"),
                "namespace": spec.get("namespace"),
            },
            "spec": {
                "default": {
                    "arg": arg,
                    "port": spec.get("port"),
                    "tag": tag,
                    "containerImage": image,
                },
                "runtimeVersion": "paddleserving",
                "service": {
                    "minScale": 1,
                }
            },
        }


class KService(launch_crd.K8sCR):
    def __init__(self, client=None):
        super(KService, self).__init__("serving.knative.dev", "services", "v1", client)


def main(argv=None):
    parser = argparse.ArgumentParser(description='Serving launcher')
    parser.add_argument('--name', type=str,
                        help='PaddleService name.')
    parser.add_argument('--namespace', type=str,
                        default='kubeflow',
                        help='The namespace of PaddleService.')
    parser.add_argument('--action', type=str, default="apply",
                        help='Action to serving execute on ElasticServing.')

    parser.add_argument('--image', type=str, required=True,
                        help='The image of the Paddle Serving.')
    parser.add_argument('--model_name', type=str, required=True,
                        help='The storage uri of model.')
    parser.add_argument('--endpoint', type=str, default="http://minio-service.kubeflow:9000",
                        help='The endpoint uri of minio object storage service.')
    parser.add_argument('--port', type=int, default=9292,
                        help='The port of Paddle Serving.')
    parser.add_argument('--model_version', type=str, default="latest",
                        help='The version of model.')

    args = parser.parse_args()

    logging.getLogger().setLevel(logging.INFO)
    logging.info('Generating PaddleService template.')

    serving_spec = {
        "name": args.name,
        "namespace": args.namespace,
        "image": args.image,
        "port": args.port,
        "model_name": args.model_name,
        "endpoint": args.endpoint,
        "model_version": args.model_version,
    }

    config.load_incluster_config()
    api_client = k8s_client.ApiClient()
    serving = Serving(client=api_client)
    serving.run(serving_spec, action=args.action)

    kservice = KService(client=api_client)
    expected_conditions = ["RoutesReady", "Ready"]
    kservice.wait_for_condition(args.namespace, args.name,
                                expected_conditions, wait_created=True)


if __name__ == "__main__":
    main()
