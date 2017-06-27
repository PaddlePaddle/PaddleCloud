FROM paddlepaddle/paddle
ADD ./convert.py ./logging.conf ./run.sh /convert/
RUN chmod +x /convert/run.sh
ADD .cache/paddle/dataset /root/.cache/paddle/dataset
