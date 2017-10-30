import collector

def is_jobs_finished(jobs):
    for job in jobs:
        if job.status != collector.JOB_STATUS_FINISHED:
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
    return sum  / len(jobs)

def get_jobs(jobname_prefix, jobs):
    return [collector.JobInfo('%s%d' % (jobname_prefix, idx)) \
            for idx in xrange(jobs)]