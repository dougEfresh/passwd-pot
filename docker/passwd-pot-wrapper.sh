#!/bin/bash
set -x
nohup su -s /bin/bash nobody -c "/bin/passwd-pot potter $PASSWD_POT_OPTS" > /passwd-pot.log &
nohup su -s /bin/bash nobody -c "/bin/passwd-pot socket $PASSWD_POT_SOCKET_OPTS" > /passwd-pot-socket.log &
