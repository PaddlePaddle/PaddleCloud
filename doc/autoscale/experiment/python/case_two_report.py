class CaseTwoItem(object):
    def __init__(self, times, nginx_pods, running_trainers, cpu_utils):
        self.nginx_pods = nginx_pods
        self.running_trainers = running_trainers
        self.times = times
        self.cpu_utils = cpu_utils

    def values(self):
        return [str(self.times), str(self.nginx_pods), str(self.running_trainers), str(self.cpu_utils)]
    
class CaseTwoReport(object):
    def __init__(self):
        self.items = []

    def append_item(self, item):
        # append new item only if Nginx pods changed
        if not self.items or self.items[-1].nginx_pods != item.nginx_pods:
            self.items.append(item)
    def title(self):
        return ['TIME', 'NGINX_PODS', 'RUNNING_TRAINERS', 'CPU_UTILS']

    def to_csv(self, f):
        f.write('|'.join(self.title()) + '\n')
        for item in self.items:
            f.write('|'.join(item.values()) + '\n')