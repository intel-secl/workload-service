#!/bin/bash

DESC="Workload Service"
NAME=workloadservice.bin

if [[ ${container} == "docker" ]]; then
    DOCKER=true
else
    DOCKER=false
fi

###################################################################################################
#Set environment specific variables here 
###################################################################################################

# the home directory must be defined before we load any environment or
# configuration files; it is explicitly passed through the sudo command
export WORKLOAD_SERVICE_HOME=${WORKLOAD_SERVICE_HOME:-/opt/workloadservice}

# the env directory is not configurable; it is defined as WORKLOADSERVICE_HOME/env.d and the
# administrator may use a symlink if necessary to place it anywhere else
export WORKLOAD_SERVICE_ENV=$WORKLOAD_SERVICE_HOME/env.d

workloadservice_load_env() {
  local env_files="$@"
  local env_file_exports
  for env_file in $env_files; do
    if [ -n "$env_file" ] && [ -f "$env_file" ]; then
      . $env_file
      env_file_exports=$(cat $env_file | grep -E '^[A-Z0-9_]+\s*=' | cut -d = -f 1)
      if [ -n "$env_file_exports" ]; then eval export $env_file_exports; fi
    fi
  done
}

load_workloadservice_prov_env() {
  local env_file="$@"
  if [ -z $env_file ]; then
    echo "No environment file provided"
    return
  fi

  # load installer environment file, if present
  if [ -r $env_file ]; then
    echo "Loading environment variables from $env_file"
    . $env_file
    env_file_exports=$(cat $env_file | grep -E '^[A-Z0-9_]+\s*=' | cut -d = -f 1)
    #echo $env_file_exports
    if [ -n "$env_file_exports" ]; then eval export $env_file_exports; fi
  else
    echo "Workload Service does not have permission to read environment file"
  fi
}

if [ -a "$2" ]; then
  load_workloadservice_prov_env $2 2>&1 >/dev/null
fi


# load environment variables; these override any existing environment variables.
# the idea is that if someone wants to override these, they must have write
# access to the environment files that we load here. 
if [ -d $WORKLOAD_SERVICE_ENV ]; then
  workloadservice_load_env $(ls -1 $WORKLOAD_SERVICE_ENV/*)
fi

###################################################################################################

# if non-root execution is specified, and we are currently root, start over; the WORKLOADSERVICE_SUDO variable limits this to one attempt
# we make an exception for the following commands:
# - 'uninstall' may require root access to delete users and certain directories
# - 'update-system-info' requires root access to use dmidecode and virsh commands
# - 'restart' requires root access as it calls workloadservice_update_system_info to update system information
if [ -n "$WORKLOADSERVICE_USERNAME" ] && [ "$WORKLOADSERVICE_USERNAME" != "root" ] && [ $(whoami) == "root" ] && [ -z "$WORKLOADSERVICE_SUDO" ] && [ "$1" != "uninstall" ] && [ "$1" != "restart" ] && [[ "$1" != "replace-"* ]]; then
  export WORKLOADSERVICE_SUDO=true
  sudo -u $WORKLOADSERVICE_USERNAME -H -E $WORKLOADSERVICE_BIN/workloadservice $*
  exit $?
fi

###################################################################################################


# default directory layout follows the 'home' style
WORKLOAD_SERVICE_CONFIGURATION=${WORKLOAD_SERVICE_CONFIGURATION:-${WORKLOAD_SERVICE_CONF:-$WORKLOAD_SERVICE_HOME/configuration}}
WORKLOAD_SERVICE_BIN=${WORKLOAD_SERVICE_BIN:-$WORKLOAD_SERVICE_HOME/bin}
WORKLOAD_SERVICE_ENV=${WORKLOAD_SERVICE_ENV:-$WORKLOAD_SERVICE_HOME/env.d}
WORKLOAD_SERVICE_VAR=${WORKLOAD_SERVICE_VAR:-$WORKLOAD_SERVICE_HOME/var}
WORKLOAD_SERVICE_LOGS=${WORKLOAD_SERVICE_LOGS:-$WORKLOAD_SERVICE_HOME/logs}
WORKLOAD_SERVICE_SHARE=${WORKLOAD_SERVICE_SHARE:-$WORKLOAD_SERVICE_HOME/share}

###################################################################################################

# load linux utility
if [ -f "$WORKLOADSERVICE_HOME/share/scripts/functions.sh" ]; then
  . $WORKLOADSERVICE_HOME/share/scripts/functions.sh
fi

# stored master password
if [ -z "$WORKLOADSERVICE_PASSWORD" ] && [ -f $WORKLOADSERVICE_CONFIGURATION/.workloadservice_password ]; then
  export WORKLOADSERVICE_PASSWORD=$(cat $WORKLOADSERVICE_CONFIGURATION/.workloadservice_password)
fi

###################################################################################################

# all other variables with defaults
WORKLOADSERVICE_PID_FILE=$WORKLOADSERVICE_HOME/workloadservice.pid
WORKLOADSERVICE_HTTP_LOG_FILE=$WORKLOADSERVICE_LOGS/http.log
WORKLOADSERVICE_SETUP_TASKS="create-keystore-password create-tls-keypair create-admin-user initialize-database and create-key-vault"
WORKLOADSERVICE_BINARY_NAME=$NAME
###################################################################################################

# ensure that our commands can be found
export PATH=$WORKLOAD_SERVICE_BIN/bin:$PATH

# run a workloadservice command
workloadservice_run() {
  local args="$*"
  $NAME $args
  return $?
}

# arguments are optional, if provided they are the names of the tasks to run, in order
workloadservice_setup() {
  local tasklist="$*"
  if [ -z "$tasklist" ]; then
    tasklist=$WORKLOAD_SERVICE_SETUP_TASKS
  fi
  $NAME setup $tasklist
  return $?
}

workloadservice_status() {
    # check if we're already running - don't start a second instance
    if workloadservice_is_running; then
        echo "Workload Service is running"
    else
        echo "Workload Service is not running"
    fi
    return 0
}

workloadservice_start() {
    # check if we're already running - don't start a second instance
    if workloadservice_is_running; then
        echo "Workload Service is running"
        return 0
    fi

    # the subshell allows the java process to have a reasonable current working
    # directory without affecting the user's working directory. 
    # the last background process pid $! must be stored from the subshell.
    (
      cd /opt/workloadservice
      $WORKLOAD_SERVICE_BIN/workloadservice.bin start >>$WORKLOADSERVICE_HTTP_LOG_FILE 2>&1 &
      echo $! > $WORKLOADSERVICE_PID_FILE
    )

    if workloadservice_is_running; then
      echo_success "Started workload service"
    else
      echo_failure "Failed to start workload service"
    fi
}

# returns 0 if workload service is running, 1 if not running
# side effects: sets WORKLOADSERVICE_PID if workload service is running, or to empty otherwise
workloadservice_is_running() {
  WORKLOADSERVICE_PID=
  if [ -f $WORKLOADSERVICE_PID_FILE ]; then
    WORKLOADSERVICE_PID=$(cat $WORKLOADSERVICE_PID_FILE)
    local is_running=`ps -eo pid | grep "^\s*${WORKLOADSERVICE_PID}$"`
    if [ -z "$is_running" ]; then
      # stale PID file
      WORKLOADSERVICE_PID=
    fi
  fi
  if [ -z "$WORKLOADSERVICE_PID" ]; then
    # check the process list just in case the pid file is stale
    WORKLOADSERVICE_PID=$(ps ww | grep -v grep | grep "workloadservice.bin" | awk '{ print $1 }')
  fi
  if [ -z "$WORKLOADSERVICE_PID" ]; then
    #echo "workload service is not running"
    return 1
  fi
  # workload service is running and WORKLOADSERVICE_PID is set
  return 0
}


workloadservice_stop() {
  if workloadservice_is_running; then
    kill -9 $WORKLOADSERVICE_PID
    if [ $? ]; then
      echo "Stopped workload service"
      rm -f $WORKLOADSERVICE_PID_FILE
    else
      echo "Failed to stop workload service"
    fi
  fi
}

# backs up the configuration directory and removes all workloadservice files,
# except for configuration files which are saved and restored
## TODO: complete the uninstall.. current code does not work if not installed
## to default directories

workloadservice_uninstall() {
    datestr=`date +%Y-%m-%d.%H%M`
    mkdir -p /tmp/workloadservice.configuration.$datestr
    chmod 500 /tmp/workloadservice.configuration.$datestr
	rm -f /usr/local/bin/workloadservice
    if [ -n "$WORKLOAD_SERVICE_HOME" ] && [ -d "$WORKLOAD_SERVICE_HOME" ]; then
      rm -rf $WORKLOAD_SERVICE_HOME
    fi
    remove_startup_script workloadservice
}


## TODO fix help section
print_help() {
    echo "Usage: $0 status|uninstall|zeroize|version"
    echo "Usage: $0 setup [--force|--noexec] [task1 task2 ...]"
    echo "Usage: $0 export-config [outfile|--in=infile|--out=outfile|--stdout] [--env-password=PASSWORD_VAR]"
    echo "Usage: $0 config [key] [--delete|newValue]"
    echo "Available setup tasks:"
    echo $WORKLOAD_SERVICE_SETUP_TASKS | tr ' ' '\n'
}

###################################################################################################

# here we look for specific commands first that we will handle in the
# script, and anything else we send to the java application

case "$1" in
  help)
    print_help
    ;;
  setup)
    shift
    workloadservice_setup $*
    ;;
  start)
    workloadservice_start
    ;;
  status)
    workloadservice_status
    ;;
  stop)
    workloadservice_stop
    ;;
  restart)
      workloadservice_stop
      workloadservice_start
    ;;
  uninstall)
    workloadservice_uninstall
    groupdel workloadservice > /dev/null 2>&1
    userdel workloadservice > /dev/null 2>&1
    ;;
  *)
    echo $NAME
    
    if [ -z "$*" ]; then
      print_help
    fi
    ;;
esac

exit $?
