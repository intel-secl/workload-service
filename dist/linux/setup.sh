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

# assert that installer is ran as root
if [[ $EUID -ne 0 ]]; then
   echo "This installer must be run as root" 
   exit 1
fi

echo Installing Workload Service ... 
cp workload-service /usr/local/bin && chmod +x /usr/local/bin/workload-service

echo Creating Workload Service User ...
id -u somename 2> /dev/null || useradd wls
# make /etc/workload-service and /var/run/workload-service
mkdir -p /etc/workload-service 
chown wls:wls /etc/workload-service
mkdir -p /var/run/workload-service
chown wls:wls /var/run/workload-service

# switch context to wls user, then run workload-service 

echo Running setup tasks ...
su wls
workload-service setup

# install system service only if not in a docker container

echo Installation complete!
