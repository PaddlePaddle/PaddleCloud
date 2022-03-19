import argparse
import logging
import launch_crd

from kubernetes import client as k8s_client
from kubernetes import config


class SampleSet(launch_crd.K8sCR):

    def __init__(self, client=None):
        super(SampleSet, self).__init__("batch.paddlepaddle.org", "samplesets", "v1alpha1", client)

    def is_expected_conditions(self, inst, expected_conditions):
        conditions = inst.get('status', {}).get("phase")
        if not conditions:
            return False, ""
        if conditions in expected_conditions:
            return True, conditions
        else:
            return False, conditions

    def get_spec(self, spec):
        return {
            "apiVersion": "%s/%s" % (self.group, self.version),
            "kind": "SampleSet",
            "metadata": {
                "name": "data-center",
                "namespace": spec.get("namespace"),
            },
            "spec": {
                "partitions": spec.get("partitions"),
                "noSync": True,
                "secretRef": {
                    "name": "data-center"
                }
            },
        }

    def get_action_status(self, action=None):
        return ["Ready"], ["SyncFailed", "PartialReady"]


class SampleJob(launch_crd.K8sCR):

    def __init__(self, client=None):
        super(SampleJob, self).__init__("batch.paddlepaddle.org", "samplejobs", "v1alpha1", client)

    def is_expected_conditions(self, inst, expected_conditions):
        conditions = inst.get('status', {}).get("phase")
        if not conditions:
            return False, ""
        if conditions in expected_conditions:
            return True, conditions
        else:
            return False, conditions

    def get_action_status(self, action=None):
        return ["Succeeded"], ["Failed"]


class SyncSampleJob(SampleJob):

    def get_spec(self, spec):
        return {
            "apiVersion": "%s/%s" % (self.group, self.version),
            "kind": "SampleJob",
            "metadata": {
                "name": f"{spec.get('name')}-sync",
                "namespace": spec.get("namespace"),
            },
            "spec": {
                "type": "sync",
                "sampleSetRef": {
                    "name": "data-center"
                },
                "secretRef": {
                    "name": spec.get("source_secret")
                },
                "syncOptions": {
                    "source": spec.get("source_uri"),
                    "destination": spec.get("name"),
                }
            },
        }


class WarmupSampleJob(SampleJob):

    def get_spec(self, spec):
        return {
            "apiVersion": "%s/%s" % (self.group, self.version),
            "kind": "SampleJob",
            "metadata": {
                "name": f"{spec.get('name')}-warmup",
                "namespace": spec.get("namespace"),
            },
            "spec": {
                "type": "warmup",
                "sampleSetRef": {
                    "name": "data-center"
                },
                "warmupOptions": {
                    "paths": [spec.get("name")]
                }
            },
        }


def main():
    parser = argparse.ArgumentParser(description='PaddleJob launcher')
    parser.add_argument('--name', type=str, required=True,
                        help='The name of DataSet.')
    parser.add_argument('--namespace', type=str, default='kubeflow',
                        help='The namespace of DataSet.')
    parser.add_argument('--action', type=str, default='apply',
                        help='Action to execute on PaddleJob.')

    parser.add_argument('--partitions', type=int, default=1,
                        help='Partitions is the number of SampleSet partitions, partition means cache node.')
    parser.add_argument('--source_uri', type=str, required=True,
                        help='Source describes the information of data source uri and secret name.')
    parser.add_argument('--source_secret', type=str, required=True,
                        help='SecretRef is reference to the authentication secret for source storage and cache engine.')

    args = parser.parse_args()

    logging.getLogger().setLevel(logging.INFO)
    logging.info('Generating DataSet template.')

    sample_set_spec = {
        "name": "data-center",
        "namespace": args.namespace,
        "partitions": args.partitions
    }
    sync_sample_job_spec = {
        "name": args.name,
        "namespace": args.namespace,
        "source_uri": args.source_uri,
        "source_secret": args.source_secret
    }
    warmup_sample_job_spec = {
        "name": args.name,
        "namespace": args.namespace,
    }

    config.load_incluster_config()
    api_client = k8s_client.ApiClient()
    sample_set = SampleSet(client=api_client)
    sync_job = SyncSampleJob(client=api_client)
    warmup_job = WarmupSampleJob(client=api_client)

    sample_set.run(sample_set_spec, action=args.action)
    sync_job.run(sync_sample_job_spec, action=args.action)
    warmup_job.run(warmup_sample_job_spec, action=args.action)

    if args.action != "delete":
        sync_job.run(sync_sample_job_spec, action="delete")
        warmup_job.run(warmup_sample_job_spec, action="delete")


if __name__ == "__main__":
    main()
