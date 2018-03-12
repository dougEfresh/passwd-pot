#!/bin/bash -xe
REGION=eu-central-1
dockerRun="docker run -v $HOME/.aws:/root/.aws -v $PWD:/tmp/aws -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -t -i pebbletech/docker-aws-cli aws"

go build handler.go
rm -f passwdpot.zip 2> /dev/null
zip passwdpot.zip  handler
$dockerRun --region $REGION  s3 cp /tmp/aws/passwdpot.zip s3://lambda-passwdpot-eu-central-1/passwdpot.zip && \
$dockerRun --region $REGION lambda update-function-code --publish --function-name passwdpot-create-event --s3-bucket lambda-passwdpot-eu-central-1  --s3-key passwdpot.zip && \
$dockerRun --region $REGION lambda invoke /dev/stdout  --function-name passwdpot-create-event --invocation-type RequestResponse  --payload '{ "originAddr": "203.116.142.113",  "event":  { "time": 2, "user": "admin", "passwd": "12345678", "remoteAddr": "158.69.243.135", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51", "application": "OpenSSH", "protocol": "ssh"}}'



# aws cloudformation update-stack --stack-name passwdpot-api-lambda  --template-body file://passwdpot-template.json  --capabilities CAPABILITY_IAM --parameters ParameterKey=PasswdPotDsn,ParameterValue='root:!qazxsw#2@tcp(passwd-pot.ct9t9eoz6luj.eu-central-1.rds.amazonaws.com:3306)/passwdpot?tls=skip-verify&parseTime=true&loc=UTC&timeout=250ms' ParameterKey=S3Bucket,ParameterValue=lambda-passwdpot-eu-central-1 ParameterKey=S3Key,ParameterValue=passwdpot.zip ParameterKey=Debug,ParameterValue=0
