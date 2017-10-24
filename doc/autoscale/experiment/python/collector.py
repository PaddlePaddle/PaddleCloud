from kubernetes import client, config
from kubernetes.client.rest import ApiException
from pprint import pprint
import time

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

    def run_once(self, labels=None):
        api_instance = client.CoreV1Api()
        try:
            api_response = api_instance.list_pod_for_all_namespaces(pretty=True)
            cpu = 0
            gpu = 0
            for item in api_response.items:
                for container in item.spec.containers:
                    requests = container.resources.requests
                    if requests:
                        cpu += self._real_cpu(requests.get('cpu', None))
                        gpu += int(requests.get('gpu', 0))
            self.cpu_requests = cpu
            self.gpu_requests = gpu

        except ApiException as e:
            print("Exception when calling CoreV1Api->list_pod_for_all_namespaces: %s\n" % e)
        return int(time.time())

    def cpu_utils(self):
        return '%0.2f' % ((100.0 * self.cpu_requests) / self.cpu_allocatable)

    def gpu_utils(self):
        if not self.gpu_allocatable:
            return '0'
        return '%0.2f' % ((1.0 * self.gpu_requests) / self.gpu_allocatable)