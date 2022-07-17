#!/bin/bash

# generate some unique values to use in the name
TS=`date +%s`
UUID=`uuidgen`

# Create the unique bucket
BUCKETNAME=bucketcleaner-${TS}-${UUID:0:16}
aws s3 mb s3://${BUCKETNAME}
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

# add bucket versioning
aws s3api put-bucket-versioning --bucket ${BUCKETNAME} --versioning-configuration Status=Enabled
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

# create and overwrite the same file in the bucket multiple times
uuidgen > outfile
aws s3 cp outfile s3://${BUCKETNAME}
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

uuidgen > outfile
aws s3 cp outfile s3://${BUCKETNAME}
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

uuidgen > outfile
aws s3 cp outfile s3://${BUCKETNAME}
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

# whats in the bucket
aws s3 ls s3://${BUCKETNAME}
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

go mod tidy
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

go build -o bin ./main.go
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

./bin ${BUCKETNAME} us-west-1 really
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

aws s3 ls s3://${BUCKETNAME}
result=$?
if [ $result -eq 1 ]; then
    exit 0
fi