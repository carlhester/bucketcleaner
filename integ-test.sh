#!/bin/bash

# generate some unique values to use in the name
TS=`date +%s`
UUID=`uuidgen`

# Create the unique bucket
BUCKETNAME=bucketcleaner-${TS}-${UUID:0:16}
aws s3 mb s3://${BUCKETNAME}

# add bucket versioning
aws s3api put-bucket-versioning --bucket ${BUCKETNAME} --versioning-configuration Status=Enabled

# create and overwrite the same file in the bucket multiple times
uuidgen > outfile
aws s3 cp outfile s3://${BUCKETNAME}

uuidgen > outfile
aws s3 cp outfile s3://${BUCKETNAME}

uuidgen > outfile
aws s3 cp outfile s3://${BUCKETNAME}

# whats in the bucket
aws s3 ls s3://${BUCKETNAME}

export TESTBUCKET=${BUCKETNAME}