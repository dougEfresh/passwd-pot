#!/bin/bash -xe
REGION=us-east-2
dockerRun="docker run -v $HOME/.aws:/root/.aws -v $PWD:/tmp/aws -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -t -i mesosphere/aws-cli"

go build handler.go
rm -f passwdpot.zip 2> /dev/null
zip passwdpot.zip  handler
$dockerRun --region $REGION  s3 cp /tmp/aws/passwdpot.zip s3://passwdpot-$REGION/passwdpot.zip && \
$dockerRun --region $REGION lambda update-function-code --publish --function-name passwdpot-create-event --s3-bucket passwdpot-$REGION  --s3-key passwdpot.zip && \
$dockerRun --region $REGION lambda invoke /dev/stdout  --function-name passwdpot-create-event --invocation-type RequestResponse  --payload '{ "originAddr": "203.116.142.113",  "event":  { "time": 2, "user": "admin", "passwd": "12345678", "remoteAddr": "158.69.243.135", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51", "application": "OpenSSH", "protocol": "ssh"}}'

#$dockerRun --region $REGION cloudformation update-stack --stack-name passwdpot-api-lambda  --template-body file:///tmp/aws/passwdpot-template.json  --capabilities CAPABILITY_IAM --parameters ParameterKey=PasswdPotDsn,ParameterValue="'$PASSWDPOT_DSN'" ParameterKey=S3Bucket,UsePreviousValue=true ParameterKey=S3Key,UsePreviousValue=true ParameterKey=Debug,UsePreviousValue=true
