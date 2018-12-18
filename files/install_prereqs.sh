#!/bin/bash

# Preconditions:
# * http_proxy and https_proxy are already set, if required

# Postconditions:
# * All messages logged to stdout/stderr; caller redirect to logfile as needed

# NOTE:  \cp escapes alias, needed because some systems alias cp to always prompt before override

# Outline:
# 1. Install unzip packages
# 2. Install postgres

if [[ ${container} == "docker" ]]; then
    DOCKER=true
else
    DOCKER=false
fi

# source functions file
if [ -f functions ]; then . functions; fi

WORKLOAD_SERVICE_HOME=${WORKLOAD_SERVICE_HOME:-/opt/workloadservice}
LOGFILE=${WORKLOAD_SERVICE_LOG_FILE:-$WORKLOAD_SERVICE_HOME/logs/install.log}
mkdir -p $(dirname $LOGFILE)

################################################################################

# 1. Install unzip packages
WORKLOAD_SERVICE_YUM_PACKAGES="unzip"
WORKLOAD_SERVICE_APT_PACKAGES="unzip"
WORKLOAD_SERVICE_YAST_PACKAGES="unzip"
WORKLOAD_SERVICE_ZYPPER_PACKAGES="unzip"

##### install prereqs can only be done as root
if [ "$(whoami)" == "root" ]; then
  install_packages "Installer requirements" "WORKLOAD_SERVICE"
  if [ $? -ne 0 ]; then echo_failure "Failed to install prerequisites through package installer"; exit 1; fi
else
  echo_warning "Required packages:"
  auto_install_preview "Workload Service requirements" "WORKLOAD_SERVICE"
fi

# 2. Install postgres
if [[ -z "$opt_postgres" && -z "$opt_mysql" ]]; then
 echo_warning "Relying on an existing database installation"
fi

# before database root portion of executed code
export POSTGRES_USERNAME=${DATABASE_USERNAME}
export POSTGRES_PASSWORD=${DATABASE_PASSWORD}
if using_postgres; then
  postgres_installed=1
  setup_pgpass
fi

# database root portion of executed code
if [ "$(whoami)" == "root" ]; then
  if using_postgres; then
    # Copy the www.postgresql.org PGP public key so add_postgresql_install_packages can add it later if needed
    if [ -d "/etc/apt" ]; then
      mkdir -p /etc/apt/trusted.gpg.d
      chmod 755 /etc/apt/trusted.gpg.d
      cp ACCC4CF8.asc "/etc/apt/trusted.gpg.d"
    fi
    POSTGRES_SERVER_APT_PACKAGES="postgresql-9.3"
    POSTGRES_SERVER_YUM_PACKAGES="postgresql93"
    if [ "$IS_RPM" != "true" ]; then
      add_postgresql_install_packages "POSTGRES_SERVER"
    fi
    if [ $? -ne 0 ]; then echo_failure "Failed to add postgresql repository to local package manager"; exit -1; fi

    postgres_userinput_connection_properties
    if [ -n "$opt_postgres" ]; then
      # Install Postgres server (if user selected localhost)
      if [[ "$POSTGRES_HOSTNAME" == "127.0.0.1" || "$POSTGRES_HOSTNAME" == "localhost" || -n `echo "$(hostaddress_list)" | grep "$POSTGRES_HOSTNAME"` ]]; then
        echo "Installing postgres server..."
        # when we install postgres server on ubuntu it prompts us for root pw
        # we preset it so we can send all output to the log
        aptget_detect; dpkg_detect; yum_detect;
        if [[ -n "$aptget" ]]; then
          echo "postgresql app-pass password $POSTGRES_PASSWORD" | debconf-set-selections
        fi
        postgres_installed=0 #postgres is being installed
        # don't need to restart postgres server unless the install script says we need to (by returning zero)
        postgres_server_install
        if [ $? -ne 0 ]; then echo_failure "Failed to install postgresql server"; exit -1; fi
        postgres_restart >> $INSTALL_LOG_FILE
        #sleep 10
        # postgres server end
      fi
      # postgres client install here
      echo "Installing postgres client..."
      if [ "$IS_RPM" != "true" ]; then
        postgres_install
      fi
      if [ $? -ne 0 ]; then echo_failure "Failed to install postgresql"; exit -1; fi
      echo "Installation of postgres client complete"
    else
      echo_warning "Relying on an existing Postgres installation"
    fi
  fi
fi

# after database root portion of executed code
if using_postgres; then
  if [ -z "$SKIP_DATABASE_INIT" ]; then
    # postgres db init here
    postgres_create_database
    if [ $? -ne 0 ]; then
      echo_failure "Cannot create database"
      exit 1
    fi
    #postgres_restart >> $INSTALL_LOG_FILE
    #sleep 10
    #export is_postgres_available postgres_connection_error
    if [ -z "$is_postgres_available" ]; then
      echo_warning "Run 'WORKLOAD_SERVICE setup' after a database is available";
    fi
    # postgress db init end
  else
    echo_warning "Skipping init of database"
  fi
  if [ $postgres_installed -eq 0 ]; then
    postgres_server_detect
    has_local_postgres_peer=`grep "^local.*all.*postgres.*peer" $postgres_pghb_conf`
    if [ -z "$has_local_postgres_peer" ]; then
      echo "Adding PostgreSQL local 'peer' authentication for 'postgres' user..."
      sed -i '/^.*TYPE.*DATABASE.*USER.*ADDRESS.*METHOD/a local all postgres peer' $postgres_pghb_conf
    fi
    has_local_peer=`grep "^local.*all.*all.*peer" $postgres_pghb_conf`
    if [ -n "$has_local_peer" ]; then
      echo "Replacing PostgreSQL local 'peer' authentication with 'password' authentication..."
      sed -i 's/^local.*all.*all.*peer/local all all password/' $postgres_pghb_conf
    fi
    has_max_connections=`grep "^max_connections" $postgres_conf`
    if [ -n "$has_max_connections" ]; then
      postgres_max_connections=$(cat "$postgres_conf" 2>/dev/null | grep "^max_connections" | head -n 1 | sed 's/#.*//' | awk -F '=' '{ print $2 }' | sed -e 's/^[ \t]*//' | sed -e 's/[ \t]*$//')
      if [ -z $postgres_max_connections ] || [ $postgres_max_connections -lt 400 ]; then
        echo "Changing postgresql configuration to set max connections to 400...";
        sed -i 's/^max_connections.*/max_connections = 400/' $postgres_conf
      fi
    else
      echo "Setting postgresql max connections to 400...";
      echo "max_connections = 400" >> $postgres_conf
    fi
    has_shared_buffers=`grep "^shared_buffers" $postgres_conf`
    if [ -n "$has_shared_buffers" ]; then
      echo "Changing postgresql configuration to set shared buffers size to 400MB...";
      sed -i 's/^shared_buffers.*/shared_buffers = 400MB/' $postgres_conf
    else
      echo "Setting postgresql shared buffers size to 400MB...";
      echo "shared_buffers = 400MB" >> $postgres_conf
    fi
    #if [ "$POSTGRESQL_KEEP_PGPASS" != "true" ]; then   # Use this line after 2.0 GA, and verify compatibility with other commands
    if [ "${POSTGRESQL_KEEP_PGPASS:-true}" == "false" ]; then
      if [ -f ${WORKLOAD_SERVICE_CONFIGURATION}/.pgpass ]; then
        echo "Removing .pgpass file to prevent insecure database password storage in plaintext..."
        rm -f ${WORKLOAD_SERVICE_CONFIGURATION}/.pgpass
        if [ $(whoami) == "root" ]; then rm -f ~/.pgpass; fi
      fi
    fi
    postgres_restart >> $INSTALL_LOG_FILE
  fi
fi