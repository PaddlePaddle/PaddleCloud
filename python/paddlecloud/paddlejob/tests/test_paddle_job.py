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

import unittest
from paddle_job import PaddleJob


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
            volumes=[])

    def test_runtime_image(self):
        paddle_job = self.__new_paddle_job()
        self.assertEqual(paddle_job.pservers, 3)

    def test_new_pserver_job(self):
        paddle_job = self.__new_paddle_job()
        pserver_job = paddle_job.new_pserver_job()
        self.assertEqual(pserver_job["metadata"]["name"], "paddle-job-pserver")

    def test_new_trainer_job(self):
        paddle_job = self.__new_paddle_job()
        pserver_job = paddle_job.new_trainer_job()
        self.assertEqual(pserver_job["metadata"]["name"], "paddle-job-trainer")


if __name__ == "__main__":
    unittest.main()
