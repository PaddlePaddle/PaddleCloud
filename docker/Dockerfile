FROM paddlepaddle/paddle
ADD ./paddle_k8s /usr/bin
ADD ./k8s_tools.py /root/
RUN pip install -U kubernetes
CMD ["paddle_k8s"]
