from setuptools import setup

packages=[
  'paddle',
  'paddle.cloud',
  'paddle.cloud.dataset']

setup(name='pcloud',
      version='0.1.1',
      description="PaddlePaddle Cloud",
      packages=packages
)
