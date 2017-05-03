FROM python:2.7.13-alpine
RUN apk add --update nodejs openssl

ADD requirements.txt package.json /pcloud/
RUN cd /pcloud && npm install && pip install -r requirements.txt && npm build
WORKDIR /pcloud

CMD ["sh", "-c", "./manage.py migrate; ./manage.py loaddata sites; npm run dev"]
