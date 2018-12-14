#!/bin/bash

# Preconditions:
# * http_proxy and https_proxy are already set, if required

# Postconditions:
# * All messages logged to stdout/stderr; caller redirect to logfile as needed

# NOTE:  \cp escapes alias, needed because some systems alias cp to always prompt before override

# Outline:
# 1. Install unzip packages
# 2. Install monit
# 3. Install postgres

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
WORKLOAD_SERVICEYUM_PACKAGES="unzip"
WORKLOAD_SERVICEAPT_PACKAGES="unzip"
WORKLOAD_SERVICEYAST_PACKAGES="unzip"
WORKLOAD_SERVICEZYPPER_PACKAGES="unzip"

##### install prereqs can only be done as root
if [ "$(whoami)" == "root" ]; then
  install_packages "Installer requirements" "WORKLOAD_SERVICE"
  if [ $? -ne 0 ]; then echo_failure "Failed to install prerequisites through package installer"; exit 1; fi
else
  echo_warning "Required packages:"
  auto_install_preview "Workload Service requirements" "WORKLOAD_SERVICE"
fi

# 2. Install monit
monit_required_version=5.5

# detect the packages we have to install
MONIT_PACKAGE=`ls -1 monit-*.tar.gz 2>/dev/null | tail -n 1`

# SCRIPT EXECUTION
monit_clear() {
  #MONIT_HOME=""
  monit=""
}

monit_detect() {
  local monitrc=`ls -1 /etc/monitrc 2>/dev/null | tail -n 1`
  monit=`which monit 2>/dev/null`
}

monit_install() {
if [ "$IS_RPM" != "true" ]; then
  MONIT_YUM_PACKAGES="monit"
fi
  MONIT_APT_PACKAGES="monit"
  MONIT_YAST_PACKAGES=""
  MONIT_ZYPPER_PACKAGES="monit"
  install_packages "Monit" "MONIT"
  if [ $? -ne 0 ]; then echo_failure "Failed to install monit through package installer"; return 1; fi
  monit_clear; monit_detect;
    if [[ -z "$monit" ]]; then
      echo_failure "Unable to auto-install Monit"
      echo "  Monit download URL:"
      echo "  http://www.mmonit.com"
    else
      echo_success "Monit installed in $monit"
    fi
}

monit_src_install() {
  local MONIT_PACKAGE="${1:-monit-5.5-linux-src.tar.gz}"
#  DEVELOPER_YUM_PACKAGES="make gcc openssl libssl-dev"
#  DEVELOPER_APT_PACKAGES="dpkg-dev make gcc openssl libssl-dev"
  DEVELOPER_YUM_PACKAGES="make gcc"
  DEVELOPER_APT_PACKAGES="dpkg-dev make gcc"
  install_packages "Developer tools" "DEVELOPER"
  if [ $? -ne 0 ]; then echo_failure "Failed to install developer tools through package installer"; return 1; fi
  monit_clear; monit_detect;
  if [[ -z "$monit" ]]; then
    if [[ -z "$MONIT_PACKAGE" || ! -f "$MONIT_PACKAGE" ]]; then
      echo_failure "Missing Monit installer: $MONIT_PACKAGE"
      return 1
    fi
    local monitfile=$MONIT_PACKAGE
    echo "Installing $monitfile"
    is_targz=`echo $monitfile | grep ".tar.gz$"`
    is_tgz=`echo $monitfile | grep ".tgz$"`
    if [[ -n "$is_targz" || -n "$is_tgz" ]]; then
      gunzip -c $monitfile | tar xf -
    fi
    local monit_unpacked=`ls -1d monit-* 2>/dev/null`
    local monit_srcdir
    for f in $monit_unpacked
    do
      if [ -d "$f" ]; then
        monit_srcdir="$f"
      fi
    done
    if [[ -n "$monit_srcdir" && -d "$monit_srcdir" ]]; then
      echo "Compiling monit..."
      cd $monit_srcdir
      ./configure --without-pam --without-ssl 2>&1 >/dev/null
      make 2>&1 >/dev/null
      make install  2>&1 >/dev/null
    fi
    monit_clear; monit_detect
    if [[ -z "$monit" ]]; then
      echo_failure "Unable to auto-install Monit"
      echo "  Monit download URL:"
      echo "  http://www.mmonit.com"
    else
      echo_success "Monit installed in $monit"
    fi
  else
    echo "Monit is already installed"
  fi
}

if [ "$(whoami)" == "root" ] && [ ${DOCKER} != "true" ]; then
  monit_install $MONIT_PACKAGE
else
  echo_warning "Skipping monit installation"
fi

if [ -n $result ]; then exit $result; fi



# 3. Install postgres
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