#!/usr/bin/env bash
yum update -y aws-cfn-bootstrap
yum install -y docker rsyslog
echo 'OPTIONS="-p 2222"' > /etc/sysconfig/sshd
systemctl restart sshd
systemctl restart docker
systemctl stop rpcbind
systemctl disable rpcbind
systemctl mask rpcbind
systemctl stop rpcbind.socket
systemctl disable rpcbind.socket
echo -e 'module(load="imtcp")\ninput(type="imtcp" port="514" address="172.17.0.1")' > /etc/rsyslog.d/99_listen.conf
systemctl daemon-reload
systemctl restart rsyslog;

wget -O /usr/bin/systemd-docker https://github.com/ibuildthecloud/systemd-docker/releases/download/v0.2.1/systemd-docker
chmod 755 /usr/bin/systemd-docker

echo "[Unit]
Description=docker-passwd-pot
After=network.target auditd.service docker.service
Requires=docker.service

[Service]
EnvironmentFile=/etc/default/docker-passwd-pot
TimeoutStartSec=0
Restart=always
ExecStart=/usr/bin/systemd-docker run \$DOCKER_OPTS

[Install]
WantedBy=multi-user.target
" > /etc/systemd/system/docker-passwd-pot.service

if [ -z "$API_SERVER" ]; then
API_SERVER="https://api.passwd-pot.io"
fi
dh=`curl -s http://169.254.169.254/latest/meta-data/public-hostname`
PORTS="-p 22:2222 -p 127.0.0.1:6161:6161  -p 127.0.0.1:6060:6060 -p 80:8000 -p 21:2121 -p 8080:8000 -p 8000:8000 -p 8888:8000 -p 110:1110 -p 5432:5432"

echo "SSHD_OPTS=\"-o Audit=yes -o MaxAuthTries=200 -o AuditSocket=/tmp/passwd.socket -o AuditUrl=${API_SERVER}\"" > /etc/default/docker-passwd-pot
echo "PASSWD_POT_OPTS=\" --all --bind 0.0.0.0  --syslog 172.17.0.1:514 --server $API_SERVER --logz $LOGZ\"" >> /etc/default/docker-passwd-pot
echo "PASSWD_POT_SOCKET_OPTS=\"--duration 30m --syslog 172.17.0.1:514 --server https://$API_SERVER --socket /tmp/passwd.socket --logz $LOGZ\"" >> /etc/default/docker-passwd-pot
echo "DOCKER_OPTS=\"-e SSHD_OPTS -e PASSWD_POT_OPTS -e PASSWD_POT_SOCKET_OPTS  $PORTS  --hostname=$db  --rm --name docker-passwd-pot $IMAGE\"" >> /etc/default/docker-passwd-pot

systemctl daemon-reload
systemctl start docker-passwd-pot
sleep 10
systemctl status docker-passwd-pot