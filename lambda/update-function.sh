#!/bin/bash -xe
go build handler.go
rm passwdpot.zip 2> /dev/null
zip passwdpot.zip geo*.gz  handler
aws --region eu-central-1  s3 cp passwdpot.zip s3://lambda-passwdpot-eu-central-1/passwdpot.zip && \
aws --region eu-central-1 lambda update-function-code --function-name passwdpot-create-event --s3-bucket lambda-passwdpot-eu-central-1  --s3-key passwdpot.zip && \
aws --region eu-central-1 lambda invoke output   --function-name passwdpot-create-event --invocation-type RequestResponse  --payload '{ "event":  { "time": 1487973301661, "user": "admin", "passwd": "12345678", "remoteAddr": "158.69.243.135", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51", "application": "OpenSSH", "protocol": "ssh" , "originAddr": "203.116.142.113" }}' && \
cat output
