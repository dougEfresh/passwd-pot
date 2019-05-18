#!/bin/bash -ex

API_SERVER=api.passwd-pot.io
SSHD_PASSWD_POT_RPM=https://s3.eu-central-1.amazonaws.com/passwd-pot/sshd-passwd-pot-8.0p1-1.amzn2.x86_64.rpm
PASSWD_POT_RPM=https://s3.eu-central-1.amazonaws.com/passwd-pot/passwd-pot-1.0-1.amzn2.x86_64.rpm

export PATH=$PATH:/bin:/usr/bin:/sbin:/usr/sbin:/usr/local/bin

cd /tmp

yum install -y json-c
echo 'OPTIONS="-p 2222"' > /etc/sysconfig/sshd
systemctl restart sshd
rm -rf *.rpm 
wget ${SSHD_PASSWD_POT_RPM:?}
wget ${PASSWD_POT_RPM:?}
rpm -i *.rpm

echo 'OPTIONS="-p 22 -o Audit=yes -o MaxAuthTries=200 -o AuditUrl=http://localhost:8889/v1/event"' > /etc/sysconfig/sshd-passwd-pot
systemctl daemon-reload

systemctl restart sshd-passwd-pot
sleep 1
systemctl status sshd-passwd-pot


echo  "PASSWD_POT_OPTIONS=\"--all --bind 0.0.0.0  --server https://${API_SERVER:?}\"" > /etc/sysconfig/passwd-pot
echo "PASSWD_POT_PROXY_OPTIONS=\"--bind localhost:8889 --server https://${API_SERVER:?}\"" >> /etc/sysconfig/passwd-pot

systemctl restart passwd-pot
sleep 1
systemctl status passwd-pot

systemctl restart passwd-pot-proxy
sleep 1
systemctl status passwd-pot-proxy && exit 0











