from kubernetes import client, config
from kubernetes.client.rest import ApiException
from pprint import pprint
import time

JOB_STATUS_NOT_EXISTS = 0
JOB_STATUS_PENDING = 1
JOB_STATUS_RUNNING = 2
JOB_STATUS_FINISHED = 3

class JobInfo(object):
    def __init__(self, name):
        self.name = name
        self.started = False
        self.status = JOB_STATUS_NOT_EXISTS
        self.submit_time = 0
        self.start_time = 0
        self.end_time = 0
        self.parallelism = 0

    def status_str(self):
        if self.status == JOB_STATUS_FINISHED:
            return 'FINISH'
        elif self.status == JOB_STATUS_PENDING:
            return 'PENDING'
        elif self.status == JOB_STATUS_NOT_EXISTS:
            return 'N/A'
        elif self.status == JOB_STATUS_RUNNING:
            return 'RUNNING'


class Collector(object):
    '''
    Collector monitor data from Kubernetes API
    '''
    def __init__(self):
        # TODO(Yancey1989): 
        #   init kubernetes configuration 
        #   from settings.py
        config.load_kube_config()
        self.cpu_allocatable = 0
        self.gpu_allocatable = 0
        self.cpu_requests = 0
        self.gpu_requests = 0

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
        try:
            api_response = api_instance.list_pod_for_all_namespaces(pretty=True)
            self._pods = api_response.items
        except ApiException as e:
            print("Exception when calling CoreV1Api->list_pod_for_all_namespaces: %s\n" % e)
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
        
        return '%0.2f%%' % ((100.0 * cpu) / self.cpu_allocatable)

    def gpu_utils(self):
        gpu = 0
        for item in self._pods:
            if item.status.phase != 'Running':
                continue
            for container in item.spec.containers:
                limits = container.resources.limits
                if limits:
                    gpu += int(limits.get('alpha.kubernetes.io/nvidia-gpu',0))
        if not self.gpu_allocatable:
            return '0'
        return '%0.2f%%' % ((100.0 * gpu) / self.gpu_allocatable)

    def update_job(self, job, times):
        phases = set()
        parallelism = 0
        for item in self._pods:
            if item.metadata.labels: 
                for k, v in item.metadata.labels.items():
                    # All PaddleCloud jobs has the label key: paddle-job-*
                    if k == 'paddle-job' and v == job.name:
                        parallelism += 1
                        if not job.submit_time:
                            job.submit_time = times
                        phases.add(item.status.phase)

        job.parallelism = parallelism
        if phases and not job.submit_time:
            job.submit_time = times

        if len(phases) == 0:
            # The job has not been submited
            return
        elif len(phases) == 1 and 'Running' in phases:
        #elif 'Running' in phases:
            # If all pods is Running phase, the job is running
            # TODO(Yancey1989): If fault-tolerant job, only need 
            # one of the Pod running
            if not job.start_time:
                job.start_time = times
            job.status = JOB_STATUS_RUNNING
        elif ('Failed' in phases or \
            (len(phases) == 1 and 'Succeeded' in phases)) and \
            job.end_time == 0:
            job.end_time = times
            job.status = JOB_STATUS_FINISHED
        elif 'Pending' in phases:
            job.status = JOB_STATUS_PENDING