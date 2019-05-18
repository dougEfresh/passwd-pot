#!/bin/bash

REGION=${1:?}
stackName=passwdpot
dockerRun="docker run -v $HOME/.aws:/root/.aws -v $PWD:/tmp/aws -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY -t -i mesosphere/aws-cli"

$dockerRun --region $REGION cloudformation delete-stack   --stack-name $stackName && \
$dockerRun --region $REGION cloudformation wait stack-delete-complete  --stack-name $stackName
$dockerRun --region $REGION cloudformation create-stack --stack-name $stackName  --template-body file:///tmp/aws/passwdpot.yaml
$dockerRun --region $REGION cloudformation wait stack-create-complete   --stack-name $stackName && \
$dockerRun --region $REGION cloudformation describe-stacks --stack-name $stackName  --query 'Stacks[*].Outputs' 
