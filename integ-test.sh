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

# get the go dependencies
go mod tidy
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

# build the go binary
go build -o bin ./main.go
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

# run the binary and delete the bucket that was created above
./bin ${BUCKETNAME} us-west-1 really
result=$?
if [ $result -ne 0 ]; then
    exit 1
fi

# try to list the bucket (which should be deleted)
aws s3 ls s3://${BUCKETNAME}
result=$?
# the expected result code is a 1; the bucket should not exist
if [ $result -eq 1 ]; then
    exit 0
fi