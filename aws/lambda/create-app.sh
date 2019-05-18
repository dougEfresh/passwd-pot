#!/bin/bash

while getopts ":o:p:d:t:" opt; do
  case ${opt} in
    t )
      TOKEN=$OPTARG
      ;;
    d)
     DSN=$OPTARG
     ;;
    \? )
      echo "Invalid option: $OPTARG" 1>&2
      ;;
    : )
      echo "Invalid option: $OPTARG requires an argument" 1>&2
      ;;
  esac
done

shift $((OPTIND -1))

REGION=us-east-2

stackName=passwdpot-app
awsRun="aws --region $REGION cloudformation"

$awsRun update-stack --stack-name $stackName --template-body file://passwdpot-template.yaml --capabilities CAPABILITY_IAM \
--parameters "ParameterKey=PasswdPotDsn,ParameterValue=$DSN" "ParameterKey=PasswdPotIpStackToken,ParameterValue=$TOKEN"
$awsRun wait stack-update-complete  --stack-name $stackName && \
$awsRun describe-stacks --stack-name $stackName  --query 'Stacks[*].Outputs' 
