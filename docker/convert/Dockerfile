FROM paddlepaddle/paddle
ADD ./convert.py ./run.sh /convert/
RUN chmod +x /convert/run.sh
ADD .cache/paddle/dataset /root/.cache/paddle/dataset
CMD ["/convert/run.sh"]
