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


class CaseTwoItem(object):
    def __init__(self, ts, nginx_pods, running_trainers, collector):
        self.ts = ts
        self.nginx_pods = nginx_pods
        self.running_trainers = running_trainers
        self.cpu_utils = collector.cpu_utils()
        self.gpu_utils = collector.gpu_utils()

    #def values(self):
    #    return [str(self.times), str(self.nginx_pods), str(self.running_trainers), str(self.cpu_utils)]


class CaseTwoReport(object):
    def __init__(self):
        self.items = []
        self.avg_cpu_utils = 0.0
        self.cnt = 0

    def append_item(self, item):
        # append new item only if Nginx pods changed
        if not self.items or \
            self.items[-1].nginx_pods != item.nginx_pods or \
            self.items[-1].running_trainers != item.running_trainers:
            self.items.append(item)
        self.avg_cpu_utils += float(item.cpu_utils)
        self.cnt += 1

    def title(self):
        return ['TIME', 'NGINX PODS', 'RUNNING TRAINERS', 'CLUSTER CPU UTILS']

    def to_csv(self, fn):
        self.avg_cpu_utils /= self.cnt
        with open(fn, 'w') as f:
            f.write('|'.join(self.title()) + '\n')
            for item in self.items:
                f.write('|'.join([
                    str(item.ts), str(item.nginx_pods), str(
                        item.running_trainers), item.cpu_utils
                ]) + '\n')
