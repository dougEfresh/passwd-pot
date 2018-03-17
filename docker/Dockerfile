FROM dougefresh/sshd-passwd-pot:72dad114d97d6ce5a955c527d440977b8eb17ee0
ENV PASSWD_POT_OPTS --bind 0.0.0.0 --all --dry-run --debug --syslog 172.17.0.1:514
ENV PASSWD_POT_SOCKET_OPTS --socket /tmp/pot.socket --dry-run --debug  --syslog 172.17.0.1:514
ENV SSHD_OPTS -o Audit=yes -o AuditSocket=/tmp/pot.socket -o AuditUrl=http://localhost/

EXPOSE 2222
EXPOSE 8000
EXPOSE 2121
EXPOSE 1110
EXPOSE 5432

COPY *wrapper.sh /docker-entrypoint.d/
COPY passwd-pot /bin

