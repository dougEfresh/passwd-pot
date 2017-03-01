#!/bin/bash
docker run  -p 5432:5432 -v "$PWD/schema:/docker-entrypoint-initdb.d/" -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD="" -d postgres
