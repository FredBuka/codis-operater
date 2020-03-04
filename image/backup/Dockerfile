# Custom redis rdb backup script, here implemented in interpreted language python
FROM python:2.7.17

RUN pip install redis

COPY backup.py /data/service/codis-backup/backup.py

WORKDIR /data/service/codis-backup/
