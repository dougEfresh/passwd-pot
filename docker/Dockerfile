FROM dougefresh/sshd-passwd-pot:af58885ee54e745edc2563e86969ac9bd18757ed

ENV PASSWD_POT_OPTS --bind 0.0.0.0 --all --dry-run --debug
ENV PASSWD_POT_PROXY_OPTS --debug --bind localhost:8889 --dry-run
ENV SSHD_OPTS -o Audit=yes -o AuditUrl=http://localhost:8889/

EXPOSE 2222
EXPOSE 8000
EXPOSE 2121
EXPOSE 1110
EXPOSE 5432

COPY *wrapper.sh /docker-entrypoint.d/
COPY passwd-pot /bin
RUN mkdir -p /var/log/passwd-pot
RUN chown nobody /var/log/passwd-pot
