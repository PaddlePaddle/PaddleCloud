import unittest
from paddle_job import PaddleJob
from paddlejob import CephFSVolume
class PaddleJobTest(unittest.TestCase):
    def __new_paddle_job(self):
        return PaddleJob(
            image="yancey1989/paddle-job",
            name="paddle-job",
            cpu=1,
            memory="1Gi",
            parallelism=3,
            job_package="/example/word2vec",
            pservers=3,
            pscpu=1,
            psmemory="1Gi",
            topology="train.py",
            volumes=[
                CephFSVolume(
                    monitors_addr="192.168.1.123:6789",
                    user="admin",
                    secret_name="cephfs-secret",
                    mount_path="/mnt/cephfs",
                    cephfs_path="/")
            ])
    def test_runtime_image(self):
        paddle_job=self.__new_paddle_job()
        self.assertEqual(paddle_job.pservers, 3)

    def test_new_pserver_job(self):
        paddle_job=self.__new_paddle_job()
        pserver_job = paddle_job.new_pserver_job()
        self.assertEqual(pserver_job["metadata"]["name"], "paddle-job-pserver")

    def test_new_trainer_job(self):
        paddle_job=self.__new_paddle_job()
        pserver_job = paddle_job.new_trainer_job()
        self.assertEqual(pserver_job["metadata"]["name"], "paddle-job-trainer")

if __name__=="__main__":
    unittest.main()
