#!/bin/bash

TS=`date +%s`
FILENAME=bucketcleaner-test-bucketname-${TS}-${TS}
aws s3 mb s3://${FILENAME}

aws s3 ls s3://${FILENAME}

uuidgen > outfile
aws s3 cp outfile s3://${FILENAME}

uuidgen > outfile
aws s3 cp outfile s3://${FILENAME}

uuidgen > outfile
aws s3 cp outfile s3://${FILENAME}

aws s3 ls s3://${FILENAME}