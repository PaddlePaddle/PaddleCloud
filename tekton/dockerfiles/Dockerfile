ARG IMAGE_TAG=2.2.2
ARG PADDLE_TOOLKIT=PaddleOCR
FROM paddlepaddle/paddle:${IMAGE_TAG}

# add tini
ENV TINI_VERSION v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

ARG PADDLE_TOOLKIT
COPY ${PADDLE_TOOLKIT} ./${PADDLE_TOOLKIT}

# previous install

RUN pip3.7 install --upgrade pip
RUN pip3.7 install jupyter -U && pip3.7 install jupyterlab
RUN pip3.7 install --upgrade ipykernel ipython

WORKDIR /home/${PADDLE_TOOLKIT}

# toolkit install
RUN pip3.7 install -r requirements.txt
RUN python3.7 setup.py install

ENTRYPOINT ["/tini", "--"]
CMD ["jupyter", "lab", "--ip=0.0.0.0", "--port=8888", "--allow-root", "--notebook-dir=/home"]

EXPOSE 8888