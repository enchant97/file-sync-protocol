FROM alpine:3.17

RUN apk add --no-cache rsync

COPY rsyncd.conf /etc/rsyncd.conf

CMD ["rsync", "--no-detach", "--daemon", "--config", "/etc/rsyncd.conf"]

EXPOSE 873
