import time
import collector
import sys
import utils
import os

JOB_NAME_PREFIX='mnist'
COLLECTION_INTERVAL=5
REPORT_SEPARATOR="|"
JOB_COUNT = int(os.getenv("JOB_COUNT", 1))
PASSES = int(os.getenv("PASSES", 1))
PASSE_NUM = int(os.getenv("PASSE_NUM", 1))
DETAILS = os.getenv("DETAILS", "ON")

class StatInfo(object):
    def __init__(self, 
                 pass_num,
                 average_running_time,
                 average_pending_time,
                 jobs_running_time,
                 cpu_utils):
        self._pass_num = pass_num
        self._average_runnint_time = average_running_time
        self._average_pending_time = average_pending_time
        self._jobs_running_time = jobs_running_time
        self._cpu_utils = cpu_utils

    def to_str(self):
        return REPORT_SEPARATOR.join([
            str(self._pass_num),
            str(self._average_runnint_time),
            str(self._average_pending_time),
            ','.join(self._jobs_running_time),
            str(self._cpu_utils)])

def load_stat_info_from_file(fn):
    with open(fn, 'r') as f:
        d = f.read().strip().split(REPORT_SEPARATOR)
        return StatInfo(d[0], d[1], d[2], d[3].split(','), d[4])

def generate_report():
    stats = [load_stat_info_from_file('./out/%s-pass%d' %(JOB_NAME_PREFIX, i) )\
        for i in xrange(PASSES)]
    avg_pending_time = 0
    avg_running_time = 0
    avg_cpu_utils = 0.0
    with open('./out/%s.csv'%JOB_NAME_PREFIX, 'w') as f:
        f.write(REPORT_SEPARATOR.join(['PASS', 'AVG_RUNNINT_TIME', \
            'AVG_PENDING_TIME', 'JOB_RUNNING_TIME', 'CPU_UTILS']) + '\n')

        for stat in stats:
            f.write(stat.to_str() + '\n')
            avg_pending_time += int(stat._average_pending_time)
            avg_running_time += int(stat._average_runnint_time)
            avg_cpu_utils += float(stat._cpu_utils)
        avg_pending_time = int(avg_pending_time / len(stats))
        avg_running_time = int(avg_running_time / len(stats))
        avg_cpu_utils = avg_cpu_utils / len(stats)
        f.write(REPORT_SEPARATOR.join(['AVG', str(avg_running_time), \
            str(avg_pending_time), 'N/A', '%0.2f' % avg_cpu_utils]) + '\n')

        
def wait_for_finished(c):
    jobs = utils.get_jobs(JOB_NAME_PREFIX, JOB_COUNT)
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

def case1(c):
    avg_running_time = 0
    avg_pending_time = 0
    avg_cpu_utils = 0.0
    if DETAILS == "ON":
        print 'Times\tName\tStatus\tCPU\tGPU\tPARALLELISM'
    jobs = utils.get_jobs(JOB_NAME_PREFIX, JOB_COUNT)

    times = 0
    while True:
        c.run_once()
        avg_cpu_utils += float(c.cpu_utils())
        for job in jobs:
            c.update_job(job, times)
            if DETAILS == "ON": 
                print '%d\t%s\t%s\t%s\t%s\t%d' % (times, \
                    job.name, job.status_str(), c.cpu_utils(), c.gpu_utils(),\
                    job.parallelism)
        if utils.is_jobs_finished(jobs):
            stat = StatInfo(
                PASSE_NUM,
                utils.avg_running_time(jobs),
                utils.avg_pending_time(jobs),
                [str(job.end_time - job.start_time) for job in jobs],
                '%0.2f' % ((avg_cpu_utils) / ((times / COLLECTION_INTERVAL) + 1))
            )
            # create out folder is not exists
            try:
                os.stat('./out')
            except:
                os.mkdir('./out')

            with open('./out/%s-pass%d' % (JOB_NAME_PREFIX, PASSE_NUM), 'w') as f:
                f.write(stat.to_str())
            break
        time.sleep(COLLECTION_INTERVAL)
        times += COLLECTION_INTERVAL

def usage():
    print 'Usage python main.py [run_case1|run_case2|wait_for_finished|wait_for_cleaned]'

if __name__=="__main__":
    if len(sys.argv) != 2:
        usage()
        exit(0)

    c = collector.Collector()
    if sys.argv[1] == 'run_case1':
        case1(c)
    elif sys.argv[1] == 'wait_for_finished':
        wait_for_finished(c)
    elif sys.argv[1] == 'wait_for_cleaned':
        wait_for_cleaned(c)
    elif sys.argv[1] == 'generate_report':
        generate_report()
    else:
        usage()