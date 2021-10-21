#!/bin/bash

SERVICE_NAME=workload-service
CURRENT_VERSION=v4.0.1
BACKUP_PATH=${BACKUP_PATH:-"/tmp/"}
LOG_FILE=${LOG_FILE:-"/tmp/$SERVICE_NAME-upgrade.log"}
echo "" > $LOG_FILE
./upgrade.sh -s $SERVICE_NAME -v $CURRENT_VERSION -b $BACKUP_PATH |& tee -a $LOG_FILE
