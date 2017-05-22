FROM typhoon1986/paddle:0.10.0
RUN pip install -U kubernetes && apt-get install iputils-ping
ADD ./paddle_k8s /usr/bin
ADD ./k8s_tools.py /root/

CMD ["paddle_k8s"]
