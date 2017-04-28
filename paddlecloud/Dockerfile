FROM python:2.7.13-alpine
RUN apk add --update -y nodejs

ADD ./ /pcloud/
RUN cd /pcloud && npm install && pip install -r requirements.txt && npm build
WORKDIR /pcloud

CMD ["sh", "-c", "./manage.py migrate; ./manage.py loaddata sites; ./manage.py runserver 0.0.0.0:8000"]
