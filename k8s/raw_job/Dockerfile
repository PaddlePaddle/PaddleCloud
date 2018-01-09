FROM paddlepaddle/paddlecloud-job

RUN mkdir -p /workspace
ADD prepare_data.py train.py /workspace/
RUN python /workspace/prepare_data.py

