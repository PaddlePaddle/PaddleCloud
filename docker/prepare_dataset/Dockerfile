FROM paddlepaddle/paddle:latest
RUN python -c "import paddle.v2.dataset.common as common; common.fetch_all()"
ADD ./prepare.py /root/
CMD ["python", "/root/prepare.py"]
