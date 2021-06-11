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
if [ $EUID -ne 0 ]; then
  echo_failure "Workload service installation has to run as root. Exiting"
  exit 1
fi

# Make sure that we are running in the same directory as the install script
cd "$(dirname "$0")"

COMPONENT_NAME=workload-service
# Upgrade if component is already installed
if command -v $COMPONENT_NAME &>/dev/null; then
  n=0
  until [ "$n" -ge 3 ]
  do
  echo "$COMPONENT_NAME is already installed, Do you want to proceed with the upgrade? [y/n]"
  read UPGRADE_NEEDED
  if [ $UPGRADE_NEEDED == "y" ] || [ $UPGRADE_NEEDED == "Y" ] ; then
    echo "Proceeding with the upgrade.."
    ./${COMPONENT_NAME}_upgrade.sh
    exit $?
  elif [ $UPGRADE_NEEDED == "n" ] || [ $UPGRADE_NEEDED == "N" ] ; then
    echo "Exiting the installation.."
    exit 0
  fi
  n=$((n+1))
  done
  echo "Exiting the installation.."
  exit 0
fi

# load installer environment file, if present
if [ -f ~/workload-service.env ]; then
  echo_info "Loading environment variables from $(cd ~ && pwd)/workload-service.env"
  . ~/workload-service.env
  env_file_exports=$(cat ~/workload-service.env | grep -E '^[A-Z0-9_]+\s*=' | cut -d = -f 1)
  if [ -n "$env_file_exports" ]; then eval export $env_file_exports; fi
else
  echo_info "workload-service.env not found. Using existing exported variables or default ones"
fi

# Load local configurations
directory_layout() {
  export WORKLOAD_SERVICE_CONFIGURATION=/etc/workload-service
  export WORKLOAD_SERVICE_LOGS=/var/log/workload-service
  export WORKLOAD_SERVICE_HOME=/opt/workload-service
  export WORKLOAD_SERVICE_BIN=$WORKLOAD_SERVICE_HOME/bin
}
directory_layout

echo_info "Installing workload service..."

echo_info "Creating Workload Service User ..."
id -u wls 2>/dev/null || useradd --comment "Workload Service" --home $WORKLOAD_SERVICE_HOME --system --shell /bin/false wls

# Create application directories (chown will be repeated near end of this script, after setup)
for directory in $WORKLOAD_SERVICE_CONFIGURATION $WORKLOAD_SERVICE_BIN $WORKLOAD_SERVICE_LOGS; do
  # mkdir -p will return 0 if directory exists or is a symlink to an existing directory or directory and parents can be created
  mkdir -p $directory
  if [ $? -ne 0 ]; then
    echo_failure "Cannot create directory: $directory"
    exit 1
  fi
  chown -R wls:wls $directory
  chmod 700 $directory
done

# change the ownership of the WLS_HOME path
chown wls:wls $WORKLOAD_SERVICE_HOME

# log file permission change
chmod 740 $WORKLOAD_SERVICE_LOGS 

mkdir -p /etc/workload-service/certs/trustedca
chown wls:wls /etc/workload-service/certs/trustedca

mkdir -p /etc/workload-service/certs/trustedjwt
chown wls:wls /etc/workload-service/certs/trustedjwt

# Create PID file directory in /var/run
mkdir -p /var/run/workload-service
chown wls:wls /var/run/workload-service

# Copy workload service installer to workload-service bin directory and create a symlink
cp -f workload-service $WORKLOAD_SERVICE_BIN
ln -sfT $WORKLOAD_SERVICE_BIN/workload-service /usr/local/bin/workload-service
chown wls:wls /usr/local/bin/workload-service

cp -f workload-service.service $WORKLOAD_SERVICE_HOME
chown wls:wls $WORKLOAD_SERVICE_HOME/workload-service.service
systemctl enable $WORKLOAD_SERVICE_HOME/workload-service.service

#Install log rotation
auto_install() {
  local component=${1}
  local cprefix=${2}
  local yum_packages=$(eval "echo \$${cprefix}_YUM_PACKAGES")
  # detect available package management tools. start with the less likely ones to differentiate.
  yum -y install $yum_packages
}

# SCRIPT EXECUTION
logRotate_clear() {
  logrotate=""
}

logRotate_detect() {
  local logrotaterc=$(ls -1 /etc/logrotate.conf 2>/dev/null | tail -n 1)
  logrotate=$(which logrotate 2>/dev/null)
  if [ -z "$logrotate" ] && [ -f "/usr/sbin/logrotate" ]; then
    logrotate="/usr/sbin/logrotate"
  fi
}

logRotate_install() {
  LOGROTATE_YUM_PACKAGES="logrotate"
  if [ "$(whoami)" = "root" ]; then
    auto_install "Log Rotate" "LOGROTATE"
    if [ $? -ne 0 ]; then
      echo_failure "Failed to install logrotate"
      exit -1
    fi
  fi
  logRotate_clear
  logRotate_detect
  if [ -z "$logrotate" ]; then
    echo_failure "logrotate is not installed"
  else
    echo "logrotate installed in $logrotate"
  fi
}

logRotate_install
export WLS_LOGLEVEL=${WLS_LOGLEVEL:-info}
export LOG_ROTATION_PERIOD=${LOG_ROTATION_PERIOD:-weekly}
export LOG_COMPRESS=${LOG_COMPRESS:-compress}
export LOG_DELAYCOMPRESS=${LOG_DELAYCOMPRESS:-delaycompress}
export LOG_COPYTRUNCATE=${LOG_COPYTRUNCATE:-copytruncate}
export LOG_SIZE=${LOG_SIZE:-100M}
export LOG_OLD=${LOG_OLD:-12}

mkdir -p /etc/logrotate.d

if [ ! -a /etc/logrotate.d/wls ]; then
  echo "/var/log/workload-service/*.log {
    missingok
    notifempty
    rotate $LOG_OLD
    maxsize $LOG_SIZE
    nodateext
    $LOG_ROTATION_PERIOD
    $LOG_COMPRESS
    $LOG_DELAYCOMPRESS
    $LOG_COPYTRUNCATE
}" >/etc/logrotate.d/wls
fi

# exit workload-service setup if WLS_NOSETUP is set
if [ "${WLS_NOSETUP,,}" == "true" ]; then
  echo_info "WLS_NOSETUP is set. So, skipping the workload-service setup task."
  echo_info "Execute 'workload-service setup' to run all the setup tasks that are required for workload-service to perform it's functions"
  exit 0
fi

# a global value to indicate if all the needed environment variables are present
# this is initially set to true. The check_env_var_present function would set this
# to false if and of the conditions are not met. This will be used to later decide
# whether to proceed with the setup
all_env_vars_present=1

# check_env_var_present is used to check if an environment variable that we expect
# is present. It prints a warning to the console if it does not exist
# Also, sets the the all_env_vars_present to false
# Arguments
#      $1 - var_name - the environment variable name that we are checking
#      $2 - empty_okay - (Optional) - empty_okay implies that environment variable needs
#           to be present - but it is acceptable for it to be empty
#           For most variables that we use, we won't pass it meaning that empty
#           strings are not acceptable
# Return
#      0 - function succeeds
#      1 - function fauls
check_env_var_present() {
  # check if we were passed in an empty string
  if [[ -z $1 ]]; then return 1; fi

  if [[ $# > 1 ]] && [[ $2 == "true" || $2 == "1" ]]; then
    if [ "${!1:-}" ]; then
      return 0
    else
      echo_warning "$1 must be set and exported (empty value is okay)"
      all_env_vars_present=0
      return 1
    fi
  fi

  if [ "${!1:-}" ]; then
    return 0
  else
    echo_warning "$1 must be set and exported"
    all_env_vars_present=0
    return 1
  fi
}

# Validate the required environment variables for the setup. We are validating this in the
# binary. However, for someone to figure out what are the ones that need to be set, they can
# check here

# start with all_env_vars_present=1 and let the the check_env_vars_present() method override
# to false if any of the required vars are not set

all_env_vars_present=1

required_vars="CMS_BASE_URL HVS_URL AAS_API_URL WLS_SERVICE_USERNAME WLS_SERVICE_PASSWORD CMS_TLS_CERT_SHA384"
for env_var in $required_vars; do
  check_env_var_present $env_var
done

chmod 700 /opt/workload-service/bin/workload-service

# Call workload-service setup if all the required env variables are set
if [[ $all_env_vars_present -eq 1 ]]; then
  # run setup tasks
  echo_info "Running setup tasks ..."
  workload-service setup all
  SETUP_RESULT=${PIPESTATUS[0]}
else
  echo_failure "One or more environment variables are not present. Setup cannot proceed. Aborting..."
  echo_failure "Please export the missing environment variables and run setup again"
  exit 1
fi

# start wls server
if [ ${SETUP_RESULT} -eq 0 ]; then
  echo_success "Installation completed Successfully"
  echo_info "Starting Workload Service"
  systemctl start workload-service
  if [ $? -eq 0 ]; then
    echo_success "Started Workload Service"
  else
    echo_failure "Workload service failed to start"
  fi
else
  echo_failure "Installation failed to complete successfully"
fi
