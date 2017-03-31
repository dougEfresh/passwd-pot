#!/bin/bash
docker run  -p 5431:5432 -v "$PWD/schema:/docker-entrypoint-initdb.d/" -e POSTGRES_USER=postgres  -d postgres
