#   Copyright (c) 2018 PaddlePaddle Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from kubernetes import client, config
from kubernetes.client.rest import ApiException
from pprint import pprint
import time

JOB_STATUS_NOT_EXISTS = 0
JOB_STATUS_PENDING = 1
JOB_STATUS_RUNNING = 2
JOB_STATUS_FINISHED = 3
JOB_STSTUS_KILLED = 4


class JobInfo(object):
    def __init__(self, name):
        self.name = name
        self.started = False
        self.status = JOB_STATUS_NOT_EXISTS
        self.submit_time = -1
        self.start_time = -1
        self.end_time = -1
        self.parallelism = 0
        self.cpu_utils = ''

    def status_str(self):
        if self.status == JOB_STATUS_FINISHED:
            return 'FINISH'
        elif self.status == JOB_STATUS_PENDING:
            return 'PENDING'
        elif self.status == JOB_STATUS_NOT_EXISTS:
            return 'N/A'
        elif self.status == JOB_STATUS_RUNNING:
            return 'RUNNING'
        elif self.status == JOB_STSTUS_KILLED:
            return 'KILLED'


class Collector(object):
    '''
    Collector monitor data from Kubernetes API
    '''

    def __init__(self):
        config.load_kube_config()
        self.namespace = config.list_kube_config_contexts()[1]['context'][
            'namespace']
        self.cpu_allocatable = 0
        self.gpu_allocatable = 0
        self.cpu_requests = 0
        self.gpu_requests = 0
        self._namespaced_pods = []
        # Collect cluster wide resource
        self._init_allocatable()

        self._pods = []

    def _init_allocatable(self):
        api_instance = client.CoreV1Api()
        try:
            api_response = api_instance.list_node()
            cpu = 0
            gpu = 0
            for item in api_response.items:
                allocate = item.status.allocatable
                cpu += int(allocate.get('cpu', 0))
                gpu += int(allocate.get('gpu', 0))
            self.cpu_allocatable = cpu
            self.gpu_allocatable = gpu
        except ApiException as e:
            print("Exception when calling CoreV1Api->list_node: %s\n" % e)

    def _real_cpu(self, cpu):
        if cpu:
            if cpu.endswith('m'):
                return 0.001 * int(cpu[:-1])
            else:
                return int(cpu)
        return 0

    def run_once(self):
        api_instance = client.CoreV1Api()
        config.list_kube_config_contexts()
        try:
            api_response = api_instance.list_pod_for_all_namespaces()
            self._pods = api_response.items
            self._namespaced_pods = []
            for pod in self._pods:
                if pod.metadata.namespace == self.namespace:
                    self._namespaced_pods.append(pod)

        except ApiException as e:
            print(
                "Exception when calling CoreV1Api->list_pod_for_all_namespaces: %s\n"
                % e)
        return int(time.time())

    def cpu_utils(self):
        cpu = 0
        for item in self._pods:
            if item.status.phase != 'Running':
                continue
            for container in item.spec.containers:
                requests = container.resources.requests
                if requests:
                    cpu += self._real_cpu(requests.get('cpu', None))

        return '%0.2f' % ((100.0 * cpu) / self.cpu_allocatable)

    def gpu_utils(self):
        gpu = 0
        for item in self._pods:
            if item.status.phase != 'Running':
                continue
            for container in item.spec.containers:
                limits = container.resources.limits
                if limits:
                    gpu += int(limits.get('alpha.kubernetes.io/nvidia-gpu', 0))
        if not self.gpu_allocatable:
            return '0'
        return '%0.2f' % ((100.0 * gpu) / self.gpu_allocatable)

    def get_paddle_pods(self):
        pods = []
        for item in self._namespaced_pods:
            if not item.metadata.labels:
                continue
            for k, v in item.metadata.labels.items():
                if k.startswith('paddle-job'):
                    pods.append((item.metadata.name, item.status.phase))
        return pods

    def get_running_trainers(self):
        cnt = 0
        for item in self._namespaced_pods:
            if not item.metadata.labels:
                continue
            for k, v in item.metadata.labels.items():
                if k == 'paddle-job' and item.status.phase == 'Running':
                    cnt += 1

        return cnt

    def update_job(self, job, times):
        phases = set()
        parallelism = 0
        cpu = 0
        running_trainers = 0
        for item in self._namespaced_pods:
            if item.metadata.labels:
                for k, v in item.metadata.labels.items():
                    # All PaddleCloud jobs has the label key: paddle-job-*
                    if k == 'paddle-job' and v == job.name:
                        parallelism += 1
                        if job.submit_time == -1:
                            job.submit_time = times
                        phases.add(item.status.phase)
                        if item.status.phase == 'Running':
                            running_trainers += 1

                    if k.startswith(
                            'paddle-job'
                    ) and v == job.name and item.status.phase == 'Running':
                        for container in item.spec.containers:
                            requests = container.resources.requests
                            if requests:
                                cpu += self._real_cpu(
                                    requests.get('cpu', None))

        job.parallelism = parallelism
        job.running_trainers = running_trainers
        job.cpu_utils = '%0.2f' % ((100.0 * cpu) / self.cpu_allocatable)
        if len(phases) == 0:
            if job.submit_time != -1:
                job.status = JOB_STSTUS_KILLED
                if job.end_time == -1:
                    job.end_time = times
        elif 'Running' in phases:
            if job.start_time == -1:
                job.start_time = times
            job.status = JOB_STATUS_RUNNING
        elif ('Failed' in phases or \
            (len(phases) == 1 and 'Succeeded' in phases)) and \
            job.end_time == -1:
            job.end_time = times
            job.status = JOB_STATUS_FINISHED
        elif 'Pending' in phases:
            job.status = JOB_STATUS_PENDING

    def get_running_pods(self, labels):
        pods = 0
        for item in self._namespaced_pods:
            if item.metadata.labels:
                for k, v in item.metadata.labels.items():
                    if k in labels and labels[k] == v and \
                            item.status.phase == 'Running':
                        pods += 1

        return pods
