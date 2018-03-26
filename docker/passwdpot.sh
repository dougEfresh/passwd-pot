#!/usr/bin/env bash

image=${1:-"dougefresh/docker-passwd-pot:dev"}

dockerId=`docker ps -q -f name=docker_passwd_pot`

[ -n "$dockerId" ] &&  echo "Stopping existing docker $dockerId" && \
    docker stop $dockerId > /dev/null  && docker rm $dockerId > /dev/null

dh=`curl -s http://169.254.169.254/latest/meta-data/public-hostname`
PORTS="-p 22:2222 -p 127.0.0.1:6161:6161  -p 127.0.0.1:6060:6060 -p 80:8000 -p 21:2121 -p 8080:8000 -p 8000:8000 -p 8888:8000 -p 110:1110 -p 5432:5432"

docker rm docker_passwd_pot > /dev/null 2>&1
set -x
docker run $PORTS -d --name docker_passwd_pot --hostname=${dh:-"passwdpot"} -e SSHD_OPTS -e PASSWD_POT_OPTS -e PASSWD_POT_SOCKET_OPTS $image