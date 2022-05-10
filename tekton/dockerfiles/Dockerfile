ARG IMAGE_TAG=2.2.2
ARG PADDLE_TOOLKIT=PaddleOCR
FROM paddlepaddle/paddle:${IMAGE_TAG}

ARG PADDLE_TOOLKIT
COPY ${PADDLE_TOOLKIT} ./${PADDLE_TOOLKIT}

WORKDIR /home/${PADDLE_TOOLKIT}

RUN pip3.7 install --upgrade pip
RUN pip3.7 install -r requirements.txt
RUN python3.7 setup.py install

CMD ["sleep", "infinity"]
