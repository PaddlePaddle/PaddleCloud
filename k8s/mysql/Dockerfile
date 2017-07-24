FROM mysql:5.7

COPY paddlecloud.sql /docker-entrypoint-initdb.d
ENTRYPOINT ["docker-entrypoint.sh"]

EXPOSE 3306
CMD ["mysqld"]
