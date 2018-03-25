#!/bin/bash -xe
version=${1:-"stg"}
[ "$TRAVIS_BRANCH" == "master" ] && version="prod"
[ "$TRAVIS_BRANCH" == "dev" ] && version="dev"
f="passwdpot-${version}.zip"
REGION=us-east-2
bucket="passwdpot-$REGION"
dockerRun="docker run -v $HOME/.aws:/root/.aws -v $PWD:/tmp/aws -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -t -i mesosphere/aws-cli"

rm -f handler
go build handler.go
rm -f passwdpot.zip 2> /dev/null
zip passwdpot.zip  handler
$dockerRun --region $REGION  s3 cp /tmp/aws/passwdpot.zip s3://$bucket/$f
v=`$dockerRun --region $REGION lambda update-function-code --publish --function-name passwdpot-create-event --s3-bucket $bucket --query Version  --s3-key $f`
awsversion=`echo $v | tr -d $'\r' |bc`

$dockerRun lambda update-alias \
    --function-name passwdpot-create-event \
    --name $version \
    --function-version $awsversion \
    --region $REGION

v=`$dockerRun --region $REGION lambda update-function-code --publish --function-name passwdpot-create-batch-events --s3-bucket $bucket --query Version  --s3-key $f`
awsversion=`echo $v | tr -d $'\r' |bc`

$dockerRun lambda update-alias \
    --function-name passwdpot-create-batch-events \
    --name $version \
    --function-version $awsversion \
    --region $REGION

$dockerRun --region $REGION lambda invoke /dev/stdout  --function-name passwdpot-create-event --qualifier $version --invocation-type RequestResponse  \
    --payload '{ "originAddr": "203.116.142.113", "time": 2, "user": "admin", "passwd": "12345678", "remoteAddr": "158.69.243.135", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51", "application": "OpenSSH", "protocol": "ssh"}' && \
$dockerRun --region $REGION lambda invoke /dev/stdout  --function-name passwdpot-create-batch-events --qualifier $version --invocation-type RequestResponse  \
    --payload '{ "originAddr": "203.116.142.113", "events":[{ "time": 20000000, "user": "admin", "passwd": "12345678", "remoteAddr": "158.69.243.135", "remotePort": 63185, "remoteName": "203.116.142.113", "remoteVersion": "SSH-2.0-JSCH-0.1.51", "application": "OpenSSH", "protocol": "ssh"}]}'

