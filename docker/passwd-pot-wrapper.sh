#!/bin/bash
mkdir /tmp/logs
chown nobody /tmp/logs
nohup su -s /bin/bash nobody -c "/bin/passwd-pot potter $PASSWD_POT_OPTS" > /tmp/logs/passwd-pot.log &
nohup su -s /bin/bash nobody -c "/bin/passwd-pot socket $PASSWD_POT_SOCKET_OPTS" > /tmp/logs/passwd-pot-socket.log &
sleep 2
