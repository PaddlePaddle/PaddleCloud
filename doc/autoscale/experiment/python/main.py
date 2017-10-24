import time
import settings
from collector import Collector

def run(c):
    while True:
        t = c.run_once()
        print 'timestmap: %d, cpu utils: %s, gpu utils: %s' % \
            (t, c.cpu_utils(), c.gpu_utils())
        # TODO(Yancey1989): draw the figure with Ploter
        time.sleep(settings.COLLECTION_INTERVAL)

if __name__=="__main__":
    c = Collector()
    run(c)