#!/bin/bash
docker run  -p 3306:3306 -v "$PWD/schema/mysql:/docker-entrypoint-initdb.d/" -e  MYSQL_ALLOW_EMPTY_PASSWORD=yes -e MYSQL_DATABASE=passwdpot -d mysql:5.7
