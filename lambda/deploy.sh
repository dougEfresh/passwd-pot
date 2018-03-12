#!/bin/bash -xe
REGION=eu-central-1
dockerRun="docker run -v $PWD:/tmp/aws -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -t -i pebbletech/docker-aws-cli aws"

go build handler.go
rm -f passwdpot.zip 2> /dev/null
zip passwdpot.zip  handler
$dockerRun --region $REGION  s3 cp passwdpot.zip s3://lambda-passwdpot-eu-central-1/passwdpot.zip && \
$dockerRun  --region $REGION lambda update-function-code --function-name passwdpot-create-event --s3-bucket lambda-passwdpot-eu-central-1  --s3-key passwdpot.zip && \
$dockerRun --region $REGION lambda invoke /dev/stdout  --function-name passwdpot-create-event --invocation-type RequestResponse  --payload '{ "event":  { "time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "158.69.243.135", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51", "application": "OpenSSH", "protocol": "ssh" , "originAddr": "203.116.142.113" }}'
