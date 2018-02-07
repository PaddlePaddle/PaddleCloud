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

import time
import collector
import sys
import utils
import os
from case1 import CaseOneReport, merge_case_one_reports
from case2 import CaseTwoItem, CaseTwoReport
COLLECTION_INTERVAL = 2

JOB_NAME = os.getenv("JOB_NAME", "mnist")
JOB_COUNT = int(os.getenv("JOB_COUNT", 1))
PASSES = int(os.getenv("PASSES", 1))
PASSE_NUM = int(os.getenv("PASSE_NUM", 1))
DETAILS = os.getenv("DETAILS", "ON")

outdir = os.environ['OUTDIR']


class StatInfo(object):
    def __init__(self, pass_num, average_running_time, average_pending_time,
                 jobs_running_time, cpu_utils):
        self._pass_num = pass_num
        self._average_runnint_time = average_running_time
        self._average_pending_time = average_pending_time
        self._jobs_running_time = jobs_running_time
        self._cpu_utils = cpu_utils

    def to_str(self):
        return utils.REPORT_SEPARATOR.join([
            str(self._pass_num), str(self._average_runnint_time),
            str(self._average_pending_time), ','.join(self._jobs_running_time),
            str(self._cpu_utils)
        ])


def wait_for_finished(c):
    jobs = utils.get_jobs(JOB_NAME, JOB_COUNT)
    while True:
        c.run_once()
        for job in jobs:
            c.update_job(job, 0)
        if utils.is_jobs_finished(jobs):
            print 'All the jobs have already finished'
            return
        print 'Waiting for all the jobs finsihed for 5 seconds...'
        time.sleep(5)


def wait_for_cleaned(c):
    while True:
        c.run_once()
        pods = c.get_paddle_pods()
        if not pods:
            print 'All the jobs have been cleaned.'
            return
        print 'Waiting for all the jobs cleaned for 5 seconds...'
        time.sleep(5)


def print_title():
    print utils.REPORT_SEPARATOR.join([
        'TIME', 'JOB NAME:STATUS:RUNNING TRAINERS:CPU UTILS', 'NGINX PODS',
        'CLUSTER CPU UTILS', 'CLUSTER GPU UTILS'
    ])


def print_info(ts, c, jobs, nginx_pods):

    running_job_count = 0
    pending_job_count = 0
    finished_job_count = 0
    waiting_job_count = 0
    jobs_running_trainers = []
    jobs_cpu_utils = []
    for job in jobs:
        if job.status == collector.JOB_STATUS_RUNNING:
            running_job_count += 1
        if job.status == collector.JOB_STATUS_PENDING:
            pending_job_count += 1
        if job.status == collector.JOB_STATUS_NOT_EXISTS:
            waiting_job_count += 1
        if job.status == collector.JOB_STATUS_FINISHED:
            finished_job_count += 1
        jobs_running_trainers.append(str(job.running_trainers))
        jobs_cpu_utils.append(str(job.cpu_utils))

    print ','.join([
        str(ts),
        str(c.cpu_utils()),  # cluster CPU utils
        str(c.get_running_trainers()),  # total running trainers count
        str(waiting_job_count),
        str(pending_job_count),
        str(running_job_count),
        str(finished_job_count),
        str(nginx_pods),
        '|'.join(jobs_running_trainers
                 ),  # Running trainers count for each job, separator by '|'
        '|'.join(
            jobs_cpu_utils)  # the CPU utils for each job, separator by '|' 
    ])


def run_case2(c):
    r1 = CaseOneReport()
    r2 = CaseTwoReport()
    jobs = utils.get_jobs(JOB_NAME, JOB_COUNT)
    start = int(time.time())

    while True:
        c.run_once()
        ts = int(time.time()) - start
        for job in jobs:
            c.update_job(job, ts)

        running_trainers = c.get_running_trainers()
        nginx_pods = c.get_running_pods({'app': 'nginx'})

        item = CaseTwoItem(ts, nginx_pods, running_trainers, c)

        if DETAILS:
            print_info(ts, c, jobs, nginx_pods)

        r1.update_cluster_utils(c)
        r2.append_item(item)

        if utils.is_jobs_killed(jobs):
            r1.update_jobs(jobs)
            r1.run()
            r2.to_csv('%s/%s-case2-result.csv' % (outdir, JOB_NAME))
            r1.to_csv('%s/%s-case1-pass%d.csv' % (outdir, JOB_NAME, PASSE_NUM))
            break


def run_case1(c):
    report = CaseOneReport()
    jobs = utils.get_jobs(JOB_NAME, JOB_COUNT)
    start = int(time.time())

    while True:
        ts = int(time.time()) - start
        c.run_once()
        for job in jobs:
            c.update_job(job, ts)
        report.update_cluster_utils(c)

        if DETAILS:
            print_info(ts, c, jobs, 0)

        if utils.is_jobs_killed(jobs):
            report.update_jobs(jobs)
            report.run()
            report.to_csv('%s/%s-case1-pass%d.csv' %
                          (outdir, JOB_NAME, PASSE_NUM))
            break


def usage():
    print 'Usage python main.py [run_case1|run_case2|wait_for_finished|wait_for_cleaned]'


if __name__ == "__main__":
    if len(sys.argv) != 2:
        usage()
        exit(0)

    c = collector.Collector()
    if sys.argv[1] == 'run_case1':
        run_case1(c)
    elif sys.argv[1] == 'run_case2':
        run_case2(c)
    elif sys.argv[1] == 'wait_for_finished':
        wait_for_finished(c)
    elif sys.argv[1] == 'wait_for_cleaned':
        wait_for_cleaned(c)
    elif sys.argv[1] == 'merge_case1_reports':
        merge_case_one_reports(JOB_NAME, PASSES)
    else:
        usage()
