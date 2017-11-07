import utils
import os

outdir = os.environ['OUTDIR']

def merge_case_one_reports(jobname, passes):
    rs = [_load_case_one_from_file('%s/%s-case1-pass%d.csv' %(outdir, jobname, i) )\
        for i in xrange(passes)]
    avg_pending_time = 0
    avg_running_time = 0
    avg_cpu_utils = 0.0
    with open('%s/%s-case1-result.csv'%(outdir, jobname), 'w') as f:
        f.write(utils.REPORT_SEPARATOR.join(rs[0].title()) + '\n')
        for i in xrange(len(rs)):
            r = rs[i]
            avg_running_time += r.avg_running_time
            avg_pending_time += r.avg_pending_time
            avg_cpu_utils += r.avg_cpu_utils
            res = [str(i)]
            res.extend(r.values())
            f.write(utils.REPORT_SEPARATOR.join(res) + '\n')
        f.write(utils.REPORT_SEPARATOR.join([
            'AVG',
            str(avg_running_time),
            str(avg_pending_time),
            'N/A',
            '%0.2f' % avg_cpu_utils
        ]) + '\n') 

def _load_case_one_from_file(fn):
    with open(fn, 'r') as f:
        d = f.read().strip().split(utils.REPORT_SEPARATOR)
        r = CaseOneReport()
        r.avg_running_time = int(d[0])
        r.avg_pending_time = int(d[1])
        r.job_running_time = d[2].split(',')
        r.avg_cpu_utils = float(d[3])
        return r

class CaseOneReport(object):
    def __init__(self):
        self.avg_running_time = 0
        self.avg_pending_time = 0
        self.job_running_time = []
        self.job_pending_time = []
        self.avg_cpu_utils = 0.0
        self.avg_gpu_utils = 0.0
        self.cnt = 0

    def update_cluster_utils(self, collector):
        self.avg_cpu_utils += float(collector.cpu_utils())
        self.avg_gpu_utils += float(collector.gpu_utils())
        self.cnt += 1

    def update_jobs(self, jobs):
        for job in jobs:
            if job.start_time == -1:
                # job always pending
                running_time = 0
                pending_time = job.end_time - job.submit_time
            else:
                running_time = job.end_time - job.start_time
                pending_time = job.start_time - job.submit_time
            self.job_running_time.append(str(running_time))
            self.avg_running_time += running_time
            self.avg_pending_time += pending_time

        self.avg_running_time /= len(jobs)
        self.avg_pending_time /= len(jobs)

    def title(self):
        return ['PASS', 'AVG RUNNING TIME', 'AVG PENDING TIME', 'JOB RUNNING TIME', 'AVG CLUSTER CPU UTILS']

    def run(self):
        self.avg_cpu_utils /= self.cnt
        self.avg_gpu_utils /= self.cnt

    def values(self):
        return [
            str(self.avg_running_time),
            str(self.avg_pending_time),
            ','.join(self.job_running_time),
            '%0.2f' % self.avg_cpu_utils
        ]

    def to_csv(self, fn):
        with open(fn, 'w') as f:
            f.write(utils.REPORT_SEPARATOR.join(self.values()))

