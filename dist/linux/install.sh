#!/bin/bash

# TERM_DISPLAY_MODE can be "plain" or "color"
TERM_DISPLAY_MODE=color
TERM_COLOR_GREEN="\\033[1;32m"
TERM_COLOR_CYAN="\\033[1;36m"
TERM_COLOR_RED="\\033[1;31m"
TERM_COLOR_YELLOW="\\033[1;33m"
TERM_COLOR_NORMAL="\\033[0;39m"

# Environment:
# - TERM_DISPLAY_MODE
# - TERM_DISPLAY_GREEN
# - TERM_DISPLAY_NORMAL
echo_success() {
  if [ "$TERM_DISPLAY_MODE" = "color" ]; then echo -en "${TERM_COLOR_GREEN}"; fi
  echo ${@:-"[  OK  ]"}
  if [ "$TERM_DISPLAY_MODE" = "color" ]; then echo -en "${TERM_COLOR_NORMAL}"; fi
  return 0
}

# Environment:
# - TERM_DISPLAY_MODE
# - TERM_DISPLAY_RED
# - TERM_DISPLAY_NORMAL
echo_failure() {
  if [ "$TERM_DISPLAY_MODE" = "color" ]; then echo -en "${TERM_COLOR_RED}"; fi
  echo ${@:-"[FAILED]"}
  if [ "$TERM_DISPLAY_MODE" = "color" ]; then echo -en "${TERM_COLOR_NORMAL}"; fi
  return 1
}

# Environment:
# - TERM_DISPLAY_MODE
# - TERM_DISPLAY_YELLOW
# - TERM_DISPLAY_NORMAL
echo_warning() {
  if [ "$TERM_DISPLAY_MODE" = "color" ]; then echo -en "${TERM_COLOR_YELLOW}"; fi
  echo ${@:-"[WARNING]"}
  if [ "$TERM_DISPLAY_MODE" = "color" ]; then echo -en "${TERM_COLOR_NORMAL}"; fi
  return 1
}


echo_info() {
  if [ "$TERM_DISPLAY_MODE" = "color" ]; then echo -en "${TERM_COLOR_CYAN}"; fi
  echo ${@:-"[INFO]"}
  if [ "$TERM_DISPLAY_MODE" = "color" ]; then echo -en "${TERM_COLOR_NORMAL}"; fi
  return 1
}

############################################################################################################

# Product installation is only allowed if we are running as root
if [ $EUID -ne 0 ];  then
  echo_failure "Workload service installation has to run as root. Exiting"
  exit 1
fi

# Make sure that we are running in the same directory as the install script
cd "$( dirname "$0" )"

# load installer environment file, if present
if [ -f ~/workload-service.env ]; then
  echo_info "Loading environment variables from $(cd ~ && pwd)/workload-service.env"
  . ~/workload-service.env
  env_file_exports=$(cat ~/workload-service.env | grep -E '^[A-Z0-9_]+\s*=' | cut -d = -f 1)
  if [ -n "$env_file_exports" ]; then eval export $env_file_exports; fi
else
  echo_info "workload-service.env not found. Using existing exported variables or default ones"
fi

export LOG_LEVEL=${LOG_LEVEL:-"info"}

# Load local configurations
directory_layout() {
export WORKLOAD_SERVICE_CONFIGURATION=/etc/workload-service
export WORKLOAD_SERVICE_LOGS=/var/log/workload-service
export WORKLOAD_SERVICE_HOME=/opt/workload-service
export WORKLOAD_SERVICE_BIN=$WORKLOAD_SERVICE_HOME/bin
export INSTALL_LOG_FILE=$WORKLOAD_SERVICE_LOGS/install.log
}
directory_layout

mkdir -p $(dirname $INSTALL_LOG_FILE)
if [ $? -ne 0 ]; then
  echo_failure "Cannot create directory: $(dirname $INSTALL_LOG_FILE)"
  exit 1
fi
logfile=$INSTALL_LOG_FILE
date >> $logfile

echo_info "Installing workload service..." >> $logfile

echo_info "Creating Workload Service User ..."
id -u wls 2> /dev/null || useradd wls

# Create application directories (chown will be repeated near end of this script, after setup)
for directory in $WORKLOAD_SERVICE_CONFIGURATION $WORKLOAD_SERVICE_BIN $WORKLOAD_SERVICE_LOGS; do
  # mkdir -p will return 0 if directory exists or is a symlink to an existing directory or directory and parents can be created
  mkdir -p $directory 
  if [ $? -ne 0 ]; then
    echo_failure "Cannot create directory: $directory" | tee -a $logfile
    exit 1
  fi
  chown -R wls:wls $directory
  chmod 700 $directory
done

mkdir -p /etc/workload-service/cacerts
chown wls:wls /etc/workload-service/cacerts

mkdir -p /etc/workload-service/jwt
chown wls:wls /etc/workload-service/jwt

# Create PID file directory in /var/run
mkdir -p /var/run/workload-service
chown wls:wls /var/run/workload-service

# Copy workload service installer to workload-service bin directory and create a symlink
cp -f workload-service $WORKLOAD_SERVICE_BIN
ln -sfT $WORKLOAD_SERVICE_BIN/workload-service /usr/local/bin/workload-service
chown wls:wls /usr/local/bin/workload-service

cp -f workload-service.service $WORKLOAD_SERVICE_HOME
systemctl enable $WORKLOAD_SERVICE_HOME/workload-service.service | tee -a $logfile

# exit workload-service setup if WORKLOAD_SERVICE_NOSETUP is set
if [ -n "$WLS_NOSETUP" ]; then
  echo_info "WLS_NOSETUP is set. So, skipping the workload-service setup task." | tee -a $logfile
  exit 0
fi

# run setup tasks
echo_info "Running setup tasks ..."
workload-service setup
SETUP_RESULT=$?

# start wls server
if [ ${SETUP_RESULT} == 0 ]; then
    systemctl start workload-service
    if [ $? == 0 ]; then
      echo_success "Installation completed Successfully" | tee -a $logfile
    else
      echo_failure "Installation failed to complete successfully" | tee -a $logfile
    fi
fi
