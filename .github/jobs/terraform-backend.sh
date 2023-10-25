#!/bin/bash

BUCKET_NAME="$PROJECT_NAME-tf-state-lock-$ENVIRONMENT-$AWS_REGION"
DYNAMODB_TABLE="$PROJECT_NAME-tf-state-$ENVIRONMENT-$AWS_REGION"
echo " validate if the bucket $BUCKET_NAME exist"

bucketExists="$(aws s3api list-buckets --output text --query "Buckets[?Name=='$BUCKET_NAME'].Name")"

if [ "$bucketExists" == "$BUCKET_NAME" ]; then
    echo " bucket exists"
else
   if [ "$AWS_REGION" != 'us-east-1' ]; then
        aws s3api create-bucket --bucket $BUCKET_NAME --region $AWS_REGION --create-bucket-configuration LocationConstraint=$AWS_REGION

   else 
        aws s3api create-bucket --bucket $BUCKET_NAME --region $AWS_REGION #--create-bucket-configuration LocationConstraint=$Region
    
   aws s3api put-bucket-versioning --bucket $BUCKET_NAME --versioning-configuration Status=Enabled --region $AWS_REGION
   fi
fi

echo " validate if the DynamoDB table exist"
dynamodbTables="$(aws dynamodb list-tables --region $AWS_REGION --output json --query 'TableNames[]' | grep $DYNAMODB_TABLE )"

if [ -z $dynamodbTables ]; then 
    echo "table doesn't exist, creating table"
    aws dynamodb create-table --region $AWS_REGION \
                              --table-name $DYNAMODB_TABLE \
                              --attribute-definitions AttributeName=LockID,AttributeType=S \
                              --key-schema AttributeName=LockID,KeyType=HASH \
                              --billing-mode PAY_PER_REQUEST
else
    echo "table exist nothing to create"     
fi