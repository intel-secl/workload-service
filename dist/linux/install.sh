#!/bin/bash

# READ .env file from ~/wls.env or ../wls.env
if [ -f ~/wls.env ]; then 
    echo Reading Installation options from `realpath ~/wls.env`
    source ~/wls.env
elif [ -f ./wls.env ]; then
    echo Reading Installation options from `realpath ./wls.env`
        source ./wls.env
fi

# export all known variables from answer file 
export WLS_PORT
export WLS_NOSETUP
export WLS_DB_USERNAME
export WLS_DB_PASSWORD
export WLS_DB_HOSTNAME
export WLS_DB_PORT

# assert that installer is ran as root
if [[ $EUID -ne 0 ]]; then
   echo "This installer must be run as root" 
   exit 1
fi

echo Creating Workload Service User ...
id -u somename 2> /dev/null || useradd wls

echo Installing Workload Service ... 
mkdir -p /opt/workload-service/bin
cp workload-service /opt/workload-service/bin/workload-service
ln -s /opt/workload-service/bin/workload-service /usr/local/bin/workload-service
chmod +x /usr/local/bin/workload-service
chmod +s /usr/local/bin/workload-service 
chown wls:wls /usr/local/bin/workload-service

# Create Configuration directory in /etc
mkdir -p /etc/workload-service 
chown wls:wls /etc/workload-service
# Create PID file directory in /var/run
mkdir -p /var/run/workload-service
chown wls:wls /var/run/workload-service
# Create arbitrary data repository in /var/lib
mkdir -p /var/lib/workload-service
chown wls:wls /var/lib/workload-service

echo Running setup tasks ...
workload-service setup
SETUP_RESULT=$?

# install system service only if not in a docker container

echo Installation complete!

# now run it
if [ ${SETUP_RESULT} == 0 ]; then
    workload-service start
fi
