#!/bin/bash

REGION=${1:?}
PASS=${2:?}
stackName=passwdpot-db
awsRun="aws --region $REGION cloudformation"

$awsRun create-stack --stack-name $stackName --template-body file://passwdpot-db-template.json --capabilities CAPABILITY_IAM --parameters ParameterKey=PasswdPotDBPassword,ParameterValue=$PASS 
$awsRun wait stack-create-complete  --stack-name $stackName
$awsRun describe-stacks --stack-name $stackName  --query 'Stacks[*].Outputs' 
