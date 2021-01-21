FROM python:3.8

RUN python3 -m pip install paddlepaddle==0.0.0 -f https://paddlepaddle.org.cn/whl/stable.html

RUN /bin/bash -c 'pip3 install kubernetes'

COPY . /paddle/

CMD ["python3", "/paddle/start.py"]