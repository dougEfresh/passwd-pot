#!/bin/bash
./build-alpine.sh
docker build -f Dockerfile -t ${1:?} .
