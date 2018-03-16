#!/bin/bash

OP=${1:?}
REGION=${2:?}
PASS=${3:?}
DBHOST=${4:?}
stackName=passwdpot-app
awsRun="aws --region $REGION cloudformation"

$awsRun ${OP}-stack --stack-name $stackName --template-body file://passwdpot-template.yaml --capabilities CAPABILITY_IAM \
--parameters ParameterKey=PasswdPotDBHost,ParameterValue=$DBHOST  ParameterKey=PasswdPotDBPassword,ParameterValue=$PASS && \
$awsRun wait stack-${OP}-complete  --stack-name $stackName && \
$awsRun describe-stacks --stack-name $stackName  --query 'Stacks[*].Outputs' 
