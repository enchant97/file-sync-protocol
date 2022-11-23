FROM alpine:3.17

RUN apk add --no-cache samba

COPY smb.conf /etc/samba/smb.conf

CMD ["smbd", "-F", "--configfile=/etc/samba/smb.conf", "--debug-stdout", "--no-process-group"]

EXPOSE 139 445
