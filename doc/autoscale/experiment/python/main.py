import time
import collector
import sys
import utils

JOB_NAME_PREFIX='mnist'
COLLECTION_INTERVAL=5

def case1(c, jobs):
    print 'Times\tName\tStatus\tCPU\tGPU\tPARALLELISM'
    times = 0
    while True:
        c.run_once()
        for job in jobs:
            c.update_job(job, times)
            
            print '%d\t%s\t%s\t%s\t%s\t%d' % (times,\
                job.name, job.status_str(), c.cpu_utils(), c.gpu_utils(), job.parallelism)

        if utils.is_all_jobs_finished(jobs):
            print 'Average running time:', utils.avg_running_time(jobs)
            print 'Average pending time:', utils.avg_pending_time(jobs)
            print 'Cluster wide CPU:', c.cpu_allocatable
            print 'Cluster wide GPU:', c.gpu_allocatable
            for job in jobs:
                print '%s runnint time:' % job.name, (job.end_time - job.start_time)
            sys.exit(0)

        # TODO(Yancey1989): draw the figure with Ploter

        time.sleep(COLLECTION_INTERVAL)
        times += COLLECTION_INTERVAL

if __name__=="__main__":
    if len(sys.argv) != 3:
        print 'Usage python main.py [case1|case2] <job count>'
        exit(0)

    c = collector.Collector()
    if sys.argv[1] == 'case1':
        case1(c, [collector.JobInfo('%s%d' % (JOB_NAME_PREFIX, idx)) \
            for idx in xrange(int(sys.argv[2]))])