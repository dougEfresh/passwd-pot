#!/bin/bash -xe
cd ../
CC=$(which musl-gcc) go build --ldflags '-w -linkmode external -extldflags "-static"' passwd-pot.go
cd -
cp ../passwd-pot .
