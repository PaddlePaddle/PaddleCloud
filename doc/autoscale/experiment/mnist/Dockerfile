FROM docker.paddlepaddle.org/paddle
RUN mkdir -p /data/mnist && \
    python -c "import paddle.v2.dataset as dataset; dataset.mnist.train(); dataset.mnist.test(); dataset.common.convert('/data/mnist', dataset.mnist.train(), 100, 'mnist-train')"
ADD ./train_ft.py /root
CMD ["paddle", "version"]
