#!/bin/bash -xe
for i in us-east-2 us-west-2 eu-central-1 ca-central-1; do
    bash ./release.sh $i ${1:?}
done
