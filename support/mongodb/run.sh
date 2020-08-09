#!/bin/bash

set -x

BACKUP_NAME=${HOST}-$(date  +"%Y_%m_%d-%H_%M").dump.gz
S3PATH="s3://$S3_BUCKET/$S3_FOLDER"

mongodump --host ${HOST} --port ${PORT} --archive=${BACKUP_NAME} --gzip
aws s3 cp ${BACKUP_NAME} ${S3PATH}/${BACKUP_NAME} --storage-class REDUCED_REDUNDANCY
rm ${BACKUP_NAME}
