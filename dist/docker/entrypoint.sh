#!/bin/bash 

SECRETS=/etc/secrets
IFS=$'\r\n' GLOBIGNORE='*' command eval 'secretFiles=($(ls  $SECRETS))'
for i in "${secretFiles[@]}"; do
    export $i=$(cat $SECRETS/$i)
done

USER_ID=$(id -u)
WORKLOAD_SERVICE_CONFIGURATION=/etc/workload-service
WORKLOAD_SERVICE_LOGS=/var/log/workload-service
WORKLOAD_SERVICE_TRUSTEDCA_DIR=${WORKLOAD_SERVICE_CONFIGURATION}/certs/trustedca
WORKLOAD_SERVICE_JWT_DIR=${WORKLOAD_SERVICE_CONFIGURATION}/certs/trustedjwt

# Create application directories (chown will be repeated near end of this script, after setup)
if [ ! -f $WORKLOAD_SERVICE_CONFIGURATION/.setup_done ]; then
  for directory in $WORKLOAD_SERVICE_CONFIGURATION $WORKLOAD_SERVICE_LOGS $WORKLOAD_SERVICE_TRUSTEDCA_DIR $WORKLOAD_SERVICE_JWT_DIR; do
    mkdir -p $directory
    if [ $? -ne 0 ]; then
      echo "Cannot create directory: $directory"
      exit 1
    fi
    chown -R $USER_ID:$USER_ID $directory
    chmod 700 $directory
  done
  workload-service setup all
  if [ $? -ne 0 ]; then
    exit 1
  fi
  touch $WORKLOAD_SERVICE_CONFIGURATION/.setup_done
fi

if [ ! -z "$SETUP_TASK" ]; then
  cp $WORKLOAD_SERVICE_CONFIGURATION/config.yml /tmp/config.yml
  IFS=',' read -ra ADDR <<< "$SETUP_TASK"
  for task in "${ADDR[@]}"; do
    workload-service setup $task --force
    if [ $? -ne 0 ]; then
      cp /tmp/config.yml $WORKLOAD_SERVICE_CONFIGURATION/config.yml
      exit 1
    fi
  done
  rm -rf /tmp/config.yml
fi

for i in "${secretFiles[@]}"; do
    unset $i
done

workload-service startServer
