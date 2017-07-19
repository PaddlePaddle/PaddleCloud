FROM python:2.7.13-alpine
RUN apk add --update nodejs openssl gcc mysql-dev musl-dev linux-headers mailx

ADD ./ /pcloud
RUN cd /pcloud && \
rm -rf node_modules && npm run clean && \
npm install && pip install -r requirements.txt && npm run build && \
npm run copy:fonts && npm run copy:images && npm run copy:fonts && npm run copy:images && \
npm run optimize
WORKDIR /pcloud

# TODO
CMD ["sh", "-c", "sleep 60 ; ./manage.py migrate; ./manage.py loaddata sites; ./manage.py runserver 0.0.0.0:$PORT"]
