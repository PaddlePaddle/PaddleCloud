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

import collector
REPORT_SEPARATOR = "|"


def is_jobs_finished(jobs):
    for job in jobs:
        if job.status != collector.JOB_STATUS_FINISHED:
            return False
    return True


def is_jobs_killed(jobs):
    for job in jobs:
        if job.status != collector.JOB_STSTUS_KILLED:
            return False
    return True


def avg_running_time(jobs):
    sum = 0
    for job in jobs:
        sum += job.end_time - job.start_time
    return sum / len(jobs)


def avg_pending_time(jobs):
    sum = 0
    for job in jobs:
        sum += job.start_time - job.submit_time
    return sum / len(jobs)


def get_jobs(jobname_prefix, jobs):
    return [collector.JobInfo('%s%d' % (jobname_prefix, idx)) \
            for idx in xrange(jobs)]
