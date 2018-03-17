#!/bin/bash
set -x
echo "Using rsyslog $RSYSLOG_SERVER"
echo "Starting $@"
export PATH=$PATH:/opt/ssh/bin

rm -f /opt/ssh/etc/ssh_host_dsa_key
rm -f /opt/ssh/etc/ssh_host_rsa_key 
rm -f /opt/ssh/etc/ssh_host_ed25519_key 
rm -f /opt/ssh/etc/ssh_host_ecdsa_key

ssh-keygen -t dsa -f /opt/ssh/etc/ssh_host_dsa_key -N ""
ssh-keygen -t rsa -f /opt/ssh/etc/ssh_host_rsa_key -N ""
ssh-keygen -t ed25519 -f /opt/ssh/etc/ssh_host_ed25519_key -N ""
ssh-keygen -t ecdsa -f /opt/ssh/etc/ssh_host_ecdsa_key -N ""

sed -i -e  "s/%RSYSLOG_SERVER%/$RSYSLOG_SERVER/g" /etc/rsyslog.d/10-sshd.conf

/etc/init.d/rsyslog start
exec "$@" "$SSHD_OPTS"
