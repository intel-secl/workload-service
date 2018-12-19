#!/bin/bash

# Postconditions:
# * exit with error code 1 only if there was a fatal error:
#   functions.sh not found (must be adjacent to this file in the package)
#   

# WORKLOAD_SERVICE install script
# Outline:
# 1. load application environment variables if already defined from env directory
# 2. load installer environment file, if present
# 3. source the utility script file "functions.sh":  workloadservice-linux-util-3.0-SNAPSHOT.sh
# 4. source the version script file "version"
# 5. define application directory layout
# 6. install pre-required packages
# 7. determine if we are installing as root or non-root, create groups and users accordingly
# 10. backup current configuration and data, if they exist
# 11. store directory layout in env file
# 12. store workloadservice username in env file
# 13. store log level in env file, if it's set
# 14. If VIRSH_DEFAULT_CONNECT_URI is defined in environment copy it to env directory
# 15. extract workloadservice zip
# 16. symlink workloadservice

# 18. migrate any old data to the new locations (v1 - v3)
# 19. setup authbind to allow non-root workloadservice to listen on ports 80 and 443
# 20. create tpm-tools and additional binary symlinks
# 21. copy utilities script file to application folder
# 22. delete existing dependencies from java folder, to prevent duplicate copies
# 23. fix_libcrypto for RHEL
# 24. create workloadservice-version file
# 25. fix_existing_aikcert
# 27. create WORKLOAD_SERVICE_TLS_CERT_IP list of system host addresses
# 28. update the extensions cache file
# 29. ensure the workloadservice owns all the content created during setup

# 31. workloadservice start
# 32. workloadservice setup
# 33. register tpm password with mtwilson
# 35. config logrotate


#####


# WARNING:
# *** do NOT use TABS for indentation, use SPACES
# *** TABS will cause errors in some linux distributions

# application defaults (these are not configurable and used only in this script so no need to export)
DEFAULT_WORKLOAD_SERVICE_HOME=/opt/workloadservice
DEFAULT_WORKLOAD_SERVICE_USERNAME=workloadservice
if [[ ${container} == "docker" ]]; then
    DOCKER=true
else
    DOCKER=false
fi

echo "Dockerized install is: $DOCKER"

# default settings
export WORKLOAD_SERVICE_ADMIN_USERNAME=${WORKLOAD_SERVICE_ADMIN_USERNAME:-workloadservice-admin}
export WORKLOAD_SERVICE_LOGIN_REGISTER=${WORKLOAD_SERVICE_LOGIN_REGISTER:-true}
export WORKLOAD_SERVICE_HOME=${WORKLOAD_SERVICE_HOME:-$DEFAULT_WORKLOAD_SERVICE_HOME}
WORKLOAD_SERVICE_LAYOUT=${WORKLOAD_SERVICE_LAYOUT:-home}

export WORKLOAD_SERVICE_PORTNUM=${WORKLOAD_SERVICE_PORTNUM:-8444}
export DATABASE_HOSTNAME=${DATABASE_HOSTNAME:-127.0.0.1}
export DATABASE_PORTNUM=${DATABASE_PORTNUM:-5432}
export DATABASE_SCHEMA=${DATABASE_SCHEMA:-mw_ws}
export DATABASE_USERNAME=${DATABASE_USERNAME:-root}
export DATABASE_PASSWORD=${DATABASE_PASSWORD:-dbpassword}
export DATABASE_VENDOR=postgres
export POSTGRES_HOSTNAME=${DATABASE_HOSTNAME}
export POSTGRES_PORTNUM=${DATABASE_PORTNUM}
export POSTGRES_DATABASE=${DATABASE_SCHEMA}
export POSTGRES_USERNAME=${DATABASE_USERNAME}
export POSTGRES_PASSWORD=${DATABASE_PASSWORD}
export POSTGRESQL_KEEP_PGPASS=${POSTGRESQL_KEEP_PGPASS:-true}
export POSTGRES_REQUIRED_VERSION=${POSTGRES_REQUIRED_VERSION:-9.4}
export DATABASE_VENDOR=${DATABASE_VENDOR:-postgres}
export ADD_POSTGRESQL_REPO=${ADD_POSTGRESQL_REPO:-yes}

# the env directory is not configurable; it is defined as WORKLOAD_SERVICE_HOME/env.d and the
# administrator may use a symlink if necessary to place it anywhere else
export WORKLOAD_SERVICE_ENV=$WORKLOAD_SERVICE_HOME/env.d

# 1. load application environment variables if already defined from env directory
if [ -d $WORKLOAD_SERVICE_ENV ]; then
  WORKLOAD_SERVICE_ENV_FILES=$(ls -1 $WORKLOAD_SERVICE_ENV/*)
  for env_file in $WORKLOAD_SERVICE_ENV_FILES; do
    . $env_file
    env_file_exports=$(cat $env_file | grep -E '^[A-Z0-9_]+\s*=' | cut -d = -f 1)
    if [ -n "$env_file_exports" ]; then eval export $env_file_exports; fi
  done
fi

# Deployment phase
# 2. load installer environment file, if present
if [ -f ~/workloadservice.env ]; then
  echo "Loading environment variables from $(cd ~ && pwd)/workloadservice.env"
  . ~/workloadservice.env
  env_file_exports=$(cat ~/workloadservice.env | grep -E '^[A-Z0-9_]+\s*=' | cut -d = -f 1)
  if [ -n "$env_file_exports" ]; then eval export $env_file_exports; fi
else
  echo "No environment file"
fi


# 3. source the utility script file "functions.sh":  mtwilson-linux-util-3.0-SNAPSHOT.sh
# FUNCTION LIBRARY
## TODO - no function file exists - so don't exits for now
if [ -f functions ]; then . functions; else echo "Missing file: functions"; exit 1; fi

# 4. source the version script file "version"
# VERSION INFORMATION
if [ -f version ]; then . version; else echo_warning "Missing file: version"; fi
# The version script is automatically generated at build time and looks like this:
#ARTIFACT=mtwilson-workloadservice-installer
#VERSION=3.0
#BUILD="Fri, 5 Jun 2015 15:55:20 PDT (release-3.0)"


# LOCAL CONFIGURATION
directory_layout() {
if [ "$WORKLOAD_SERVICE_LAYOUT" == "linux" ]; then
  export WORKLOAD_SERVICE_CONFIGURATION=${WORKLOAD_SERVICE_CONFIGURATION:-/etc/workloadservice}
  export WORKLOAD_SERVICE_REPOSITORY=${WORKLOAD_SERVICE_REPOSITORY:-/var/opt/workloadservice}
  export WORKLOAD_SERVICE_LOGS=${WORKLOAD_SERVICE_LOGS:-/var/log/workloadservice}
elif [ "$WORKLOAD_SERVICE_LAYOUT" == "home" ]; then
  export WORKLOAD_SERVICE_CONFIGURATION=${WORKLOAD_SERVICE_CONFIGURATION:-$WORKLOAD_SERVICE_HOME/configuration}
  export WORKLOAD_SERVICE_REPOSITORY=${WORKLOAD_SERVICE_REPOSITORY:-$WORKLOAD_SERVICE_HOME/repository}
  export WORKLOAD_SERVICE_LOGS=${WORKLOAD_SERVICE_LOGS:-$WORKLOAD_SERVICE_HOME/logs}
fi
export WORKLOAD_SERVICE_VAR=${WORKLOAD_SERVICE_VAR:-$WORKLOAD_SERVICE_HOME/var}
export WORKLOAD_SERVICE_SHARE=${WORKLOAD_SERVICE_SHARE:-$WORKLOAD_SERVICE_HOME/share}
export WORKLOAD_SERVICE_BIN=${WORKLOAD_SERVICE_BIN:-$WORKLOAD_SERVICE_HOME/bin}
export WORKLOAD_SERVICE_BACKUP=${WORKLOAD_SERVICE_BACKUP:-$WORKLOAD_SERVICE_REPOSITORY/backup}
export INSTALL_LOG_FILE=$WORKLOAD_SERVICE_LOGS/install.log
}

# 5. define application directory layout
directory_layout

# 6. install pre-required packages
if [ "${WORKLOAD_SERVICE_SETUP_PREREQS:-yes}" == "yes" ]; then
  chmod +x install_prereqs.sh
  ./install_prereqs.sh
  ipResult=$?

  # set WORKLOAD_SERVICE_REBOOT=no (in workloadservice.env) if you want to ensure it doesn't reboot
  # set WORKLOAD_SERVICE_SETUP_PREREQS=no (in workloadservice.env) if you want to skip this step 
  chmod +x setup_prereqs.sh
  ./setup_prereqs.sh
  spResult=$?
fi

# 7. determine if we are installing as root or non-root, create groups and users accordingly
if [ "$(whoami)" == "root" ]; then
  # create a workloadservice user if there isn't already one created
  WORKLOAD_SERVICE_USERNAME=${WORKLOAD_SERVICE_USERNAME:-$DEFAULT_WORKLOAD_SERVICE_USERNAME}
  if ! getent passwd $WORKLOAD_SERVICE_USERNAME 2>&1 >/dev/null; then
    useradd --comment "ISecL Workload Agent" --home $WORKLOAD_SERVICE_HOME --system --shell /bin/false $WORKLOAD_SERVICE_USERNAME
    usermod --lock $WORKLOAD_SERVICE_USERNAME
    # note: to assign a shell and allow login you can run "usermod --shell /bin/bash --unlock $WORKLOAD_SERVICE_USERNAME"
  fi
else
  # already running as workloadservice user
  WORKLOAD_SERVICE_USERNAME=$(whoami)
  if [ ! -w "$WORKLOAD_SERVICE_HOME" ] && [ ! -w $(dirname $WORKLOAD_SERVICE_HOME) ]; then
    WORKLOAD_SERVICE_HOME=$(cd ~ && pwd)
  fi
  echo_warning "Installing as $WORKLOAD_SERVICE_USERNAME into $WORKLOAD_SERVICE_HOME"  
fi
directory_layout


# before we start, clear the install log (directory must already exist; created above)
mkdir -p $(dirname $INSTALL_LOG_FILE)
if [ $? -ne 0 ]; then
  echo_failure "Cannot write to log directory: $(dirname $INSTALL_LOG_FILE)"
  exit 1
fi
date > $INSTALL_LOG_FILE
if [ $? -ne 0 ]; then
  echo_failure "Cannot write to log file: $INSTALL_LOG_FILE"
  exit 1
fi
chown $WORKLOAD_SERVICE_USERNAME:$WORKLOAD_SERVICE_USERNAME $INSTALL_LOG_FILE
logfile=$INSTALL_LOG_FILE

# 8. create application directories (chown will be repeated near end of this script, after setup)
for directory in $WORKLOAD_SERVICE_HOME $WORKLOAD_SERVICE_BIN $WORKLOAD_SERVICE_CONFIGURATION $WORKLOAD_SERVICE_ENV $WORKLOAD_SERVICE_REPOSITORY $WORKLOAD_SERVICE_VAR $WORKLOAD_SERVICE_LOGS $WORKLOAD_SERVICE_SHARE; do
  # mkdir -p will return 0 if directory exists or is a symlink to an existing directory or directory and parents can be created
  mkdir -p $directory
  if [ $? -ne 0 ]; then
    echo_failure "Cannot create directory: $directory"
    exit 1
  fi
  chown -R $WORKLOAD_SERVICE_USERNAME:$WORKLOAD_SERVICE_USERNAME $directory
  chmod 700 $directory
done

# ensure we have our own workloadservice programs in the path
export PATH=$WORKLOAD_SERVICE_BIN:$PATH

profile_dir=$HOME
if [ "$(whoami)" == "root" ] && [ -n "$WORKLOAD_SERVICE_USERNAME" ] && [ "$WORKLOAD_SERVICE_USERNAME" != "root" ]; then
  profile_dir=$WORKLOAD_SERVICE_HOME
fi
profile_name=$profile_dir/$(basename $(getUserProfileFile))

appendToUserProfileFile "export WORKLOAD_SERVICE_HOME=$WORKLOAD_SERVICE_HOME" $profile_name

# if an existing workloadservice is already running, stop it while we install
existing_workloadservice=`which workloadservice 2>/dev/null`
if [ -f "$existing_workloadservice" ]; then
  $existing_workloadservice stop
fi

workloadservice_backup_configuration() {
  if [ -n "$WORKLOAD_SERVICE_CONFIGURATION" ] && [ -d "$WORKLOAD_SERVICE_CONFIGURATION" ]; then
    mkdir -p $WORKLOAD_SERVICE_BACKUP
    if [ $? -ne 0 ]; then
      echo_warning "Cannot create backup directory: $WORKLOAD_SERVICE_BACKUP"
      echo_warning "Backup will be stored in /tmp"
      WORKLOAD_SERVICE_BACKUP=/tmp
    fi
    datestr=`date +%Y%m%d.%H%M`
    backupdir=$WORKLOAD_SERVICE_BACKUP/workloadservice.configuration.$datestr
    cp -r $WORKLOAD_SERVICE_CONFIGURATION $backupdir
  fi
}
workloadservice_backup_repository() {
  if [ -n "$WORKLOAD_SERVICE_REPOSITORY" ] && [ -d "$WORKLOAD_SERVICE_REPOSITORY" ]; then
    mkdir -p $WORKLOAD_SERVICE_BACKUP
    if [ $? -ne 0 ]; then
      echo_warning "Cannot create backup directory: $WORKLOAD_SERVICE_BACKUP"
      echo_warning "Backup will be stored in /tmp"
      WORKLOAD_SERVICE_BACKUP=/tmp
    fi
    datestr=`date +%Y%m%d.%H%M`
    backupdir=$WORKLOAD_SERVICE_BACKUP/workloadservice.repository.$datestr
    cp -r $WORKLOAD_SERVICE_REPOSITORY $backupdir
  fi
}

# 10. backup current configuration and data, if they exist
workloadservice_backup_configuration
#workloadservice_backup_repository

# 11. store directory layout in env file
echo "# $(date)" > $WORKLOAD_SERVICE_ENV/workloadservice-layout
echo "WORKLOAD_SERVICE_HOME=$WORKLOAD_SERVICE_HOME" >> $WORKLOAD_SERVICE_ENV/workloadservice-layout
echo "WORKLOAD_SERVICE_CONFIGURATION=$WORKLOAD_SERVICE_CONFIGURATION" >> $WORKLOAD_SERVICE_ENV/workloadservice-layout
echo "WORKLOAD_SERVICE_JAVA=$WORKLOAD_SERVICE_JAVA" >> $WORKLOAD_SERVICE_ENV/workloadservice-layout
echo "WORKLOAD_SERVICE_BIN=$WORKLOAD_SERVICE_BIN" >> $WORKLOAD_SERVICE_ENV/workloadservice-layout
echo "WORKLOAD_SERVICE_REPOSITORY=$WORKLOAD_SERVICE_REPOSITORY" >> $WORKLOAD_SERVICE_ENV/workloadservice-layout
echo "WORKLOAD_SERVICE_LOGS=$WORKLOAD_SERVICE_LOGS" >> $WORKLOAD_SERVICE_ENV/workloadservice-layout
echo "WORKLOAD_SERVICE_SHARE=$WORKLOAD_SERVICE_SHARE" >> $WORKLOAD_SERVICE_ENV/workloadservice-layout

# 12. store config in env file # move to config
echo "# $(date)" > $WORKLOAD_SERVICE_ENV/config
echo "DATABASE_SCHEMA=$DATABASE_SCHEMA" >> $WORKLOAD_SERVICE_ENV/config
echo "DATABASE_USERNAME=$DATABASE_USERNAME" >> $WORKLOAD_SERVICE_ENV/config
echo "DATABASE_PASSWORD=$DATABASE_PASSWORD" >> $WORKLOAD_SERVICE_ENV/config
echo "DATABASE_HOSTNAME=$DATABASE_HOSTNAME" >> $WORKLOAD_SERVICE_ENV/config
echo "DATABASE_PORTNUM=$DATABASE_PORTNUM" >> $WORKLOAD_SERVICE_ENV/config
echo "WORKLOAD_SERVICE_PORTNUM=$WORKLOAD_SERVICE_PORTNUM" >> $WORKLOAD_SERVICE_ENV/config

# 12. store workloadservice username in env file
echo "# $(date)" > $WORKLOAD_SERVICE_ENV/workloadservice-username
echo "WORKLOAD_SERVICE_USERNAME=$WORKLOAD_SERVICE_USERNAME" >> $WORKLOAD_SERVICE_ENV/workloadservice-username

# 13. store the auto-exported environment variables in temporary env file
# to make them available after the script uses sudo to switch users;
# we delete that file later
echo "# $(date)" > $WORKLOAD_SERVICE_ENV/workloadservice-setup
for env_file_var_name in $env_file_exports
do
  eval env_file_var_value="\$$env_file_var_name"
  echo "export $env_file_var_name='$env_file_var_value'" >> $WORKLOAD_SERVICE_ENV/workloadservice-setup
done

cp version $WORKLOAD_SERVICE_CONFIGURATION/workloadservice-version

# 15. extract workloadservice zip  (workloadservice-0.1.zip)
echo "Extracting application..."
WORKLOAD_SERVICE_ZIPFILE=`ls -1 workloadservice-*.zip 2>/dev/null | head -n 1`
unzip -oq $WORKLOAD_SERVICE_ZIPFILE -d $WORKLOAD_SERVICE_HOME
cp workloadservice.sh $WORKLOAD_SERVICE_BIN/

# add bin and sbin directories in workloadservice home directory to path
bin_directories=$(find_subdirectories ${WORKLOAD_SERVICE_HOME} bin; find_subdirectories ${WORKLOAD_SERVICE_HOME} sbin)
bin_directories_path=$(join_by : ${bin_directories[@]})
for directory in ${bin_directories[@]}; do
  chmod -R 700 $directory
done
export PATH=$bin_directories_path:$PATH
appendToUserProfileFile "export PATH=${bin_directories_path}:\$PATH" $profile_name

# add lib directories in workloadservice home directory to LD_LIBRARY_PATH variable env file
lib_directories=$(find_subdirectories ${WORKLOAD_SERVICE_SHARE} lib)
lib_directories_path=$(join_by : ${lib_directories[@]})
export LD_LIBRARY_PATH=$lib_directories_path
echo "export LD_LIBRARY_PATH=${lib_directories_path}" > $WORKLOAD_SERVICE_ENV/workloadservice-lib

##TODO - rework as appropriate when we identify the logging mechanism
# update logback.xml with configured workloadservice log directory
if [ -f "$WORKLOAD_SERVICE_CONFIGURATION/logback.xml" ]; then
  sed -e "s|<file>.*/workloadservice.log</file>|<file>$WORKLOAD_SERVICE_LOGS/workloadservice.log</file>|" $WORKLOAD_SERVICE_CONFIGURATION/logback.xml > $WORKLOAD_SERVICE_CONFIGURATION/logback.xml.edited
  if [ $? -eq 0 ]; then
    mv $WORKLOAD_SERVICE_CONFIGURATION/logback.xml.edited $WORKLOAD_SERVICE_CONFIGURATION/logback.xml
  fi
#else
#  echo_warning "Logback configuration not found: $WORKLOAD_SERVICE_CONFIGURATION/logback.xml"
fi

chown -R $WORKLOAD_SERVICE_USERNAME:$WORKLOAD_SERVICE_USERNAME $WORKLOAD_SERVICE_HOME
chmod 755 $WORKLOAD_SERVICE_BIN/*

# 16. symlink workloadservice
# if prior version had control script in /usr/local/bin, delete it
if [ "$(whoami)" == "root" ] && [ -f /usr/local/bin/workloadservice ]; then
  rm /usr/local/bin/workloadservice
fi
EXISTING_WORKLOAD_SERVICE_COMMAND=`which workloadservice 2>/dev/null`
if [ -n "$EXISTING_WORKLOAD_SERVICE_COMMAND" ]; then
  rm -f "$EXISTING_WORKLOAD_SERVICE_COMMAND"
fi
ln -s $WORKLOAD_SERVICE_BIN/workloadservice.sh $WORKLOAD_SERVICE_BIN/workloadservice
# link /usr/local/bin/workloadservice -> /opt/workloadservice/bin/workloadservice
ln -s $WORKLOAD_SERVICE_BIN/workloadservice.sh /usr/local/bin/workloadservice

# Redefine the variables to the new locations
package_config_filename=$WORKLOAD_SERVICE_CONFIGURATION/workloadservice.properties

# 19. setup authbind to allow non-root workloadservice to listen on ports 80, 443 and 8444
# setup authbind to allow non-root workloadservice to listen on port 8444
mkdir -p /etc/authbind/byport
if [ ! -f /etc/authbind/byport/8444 ]; then
  if [ "$(whoami)" == "root" ]; then
    if [ -n "$WORKLOAD_SERVICE_USERNAME" ] && [ "$WORKLOAD_SERVICE_USERNAME" != "root" ] && [ -d /etc/authbind/byport ]; then
      touch /etc/authbind/byport/8444
      chmod 500 /etc/authbind/byport/8444
      chown $WORKLOAD_SERVICE_USERNAME /etc/authbind/byport/8444
    fi
  else
    echo_warning "You must be root to setup authbind configuration"
  fi
fi


# 21. copy utilities script file to application folder
mkdir -p "$WORKLOAD_SERVICE_HOME"/share/scripts
cp version "$WORKLOAD_SERVICE_HOME"/share/scripts/version.sh
cp functions "$WORKLOAD_SERVICE_HOME"/share/scripts/functions.sh
chmod -R 700 "$WORKLOAD_SERVICE_HOME"/share/scripts
chown -R $WORKLOAD_SERVICE_USERNAME:$WORKLOAD_SERVICE_USERNAME "$WORKLOAD_SERVICE_HOME"/share/scripts
chmod +x $WORKLOAD_SERVICE_BIN/*


# 24. create workloadservice-version file
package_version_filename=$WORKLOAD_SERVICE_ENV/workloadservice-version
datestr=`date +%Y-%m-%d.%H%M`
touch $package_version_filename
chmod 600 $package_version_filename
chown $WORKLOAD_SERVICE_USERNAME:$WORKLOAD_SERVICE_USERNAME $package_version_filename
echo "# Installed Trust Agent on ${datestr}" > $package_version_filename
echo "WORKLOAD_SERVICE_VERSION=${VERSION}" >> $package_version_filename
echo "WORKLOAD_SERVICE_RELEASE=\"${BUILD}\"" >> $package_version_filename

# during a Docker image build, we don't know if 1.2 is going to be used, defer this until Docker startup script.
if [[ "$(whoami)" == "root" && ${DOCKER} == "false" ]]; then
  echo "Registering workloadservice in start up"
  register_startup_script $WORKLOAD_SERVICE_BIN/workloadservice workloadservice 21 >>$logfile 2>&1
  # trousers has N=20 startup number, need to lookup and do a N+1
else
  echo_warning "Skipping startup script registration"
fi

# Ensure we have given workloadservice access to its files
for directory in $WORKLOAD_SERVICE_HOME $WORKLOAD_SERVICE_CONFIGURATION $WORKLOAD_SERVICE_ENV $WORKLOAD_SERVICE_REPOSITORY $WORKLOAD_SERVICE_VAR $WORKLOAD_SERVICE_LOGS; do
  echo "chown -R $WORKLOAD_SERVICE_USERNAME:$WORKLOAD_SERVICE_USERNAME $directory" >>$logfile
  chown -R $WORKLOAD_SERVICE_USERNAME:$WORKLOAD_SERVICE_USERNAME $directory 2>>$logfile
done

##TODO - do we need update any system info related to workloadservice. 
##if [[ "$(whoami)" == "root" && ${DOCKER} != "true" ]]; then
##  echo "Updating system information"
##  workloadservice update-system-info 2>/dev/null
##else
##  echo_warning "Skipping updating system information"
##fi

# Make the logs dir owned by workloadservice user
chown -R $WORKLOAD_SERVICE_USERNAME:$WORKLOAD_SERVICE_USERNAME $WORKLOAD_SERVICE_LOGS/


# 29. ensure the workloadservice owns all the content created during setup
for directory in $WORKLOAD_SERVICE_HOME $WORKLOAD_SERVICE_CONFIGURATION $WORKLOAD_SERVICE_JAVA $WORKLOAD_SERVICE_BIN $WORKLOAD_SERVICE_ENV $WORKLOAD_SERVICE_REPOSITORY $WORKLOAD_SERVICE_LOGS; do
  chown -R $WORKLOAD_SERVICE_USERNAME:$WORKLOAD_SERVICE_USERNAME $directory
done


# exit workloadservice setup if WORKLOAD_SERVICE_NOSETUP is set
if [ -n "$WORKLOAD_SERVICE_NOSETUP" ]; then
  echo "WORKLOAD_SERVICE_NOSETUP value is set. So, skipping the workloadservice setup task."
  exit 0;
fi

# 33. workloadservice setup
workloadservice setup 

# 34. workloadservice setup
workloadservice start

##TODO - any sort of setup tasks after setup
# 35. workloadservice post-setup
# workloadservice post set up tasks
