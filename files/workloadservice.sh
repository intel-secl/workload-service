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

# the env directory is not configurable; it is defined as WORKLOAD_SERVICE_HOME/env.d and the
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

# load environment variables; these override any existing environment variables.
# the idea is that if someone wants to override these, they must have write
# access to the environment files that we load here. 
if [ -d $WORKLOAD_SERVICE_ENV ]; then
  workloadservice_load_env $(ls -1 $WORKLOAD_SERVICE_ENV/*)
fi

###################################################################################################

# if non-root execution is specified, and we are currently root, start over; the WORKLOAD_SERVICE_SUDO variable limits this to one attempt
# we make an exception for the following commands:
# - 'uninstall' may require root access to delete users and certain directories
if [ -n "$WORKLOAD_SERVICE_USERNAME" ] && [ "$WORKLOAD_SERVICE_USERNAME" != "root" ] && [ $(whoami) == "root" ] && [ -z "$WORKLOAD_SERVICE_SUDO" ] && [ "$1" != "uninstall" ] && [ "$1" != "restart" ] && [[ "$1" != "replace-"* ]]; then
  export WORKLOAD_SERVICE_SUDO=true
  sudo -u $WORKLOAD_SERVICE_USERNAME -H -E $WORKLOAD_SERVICE_BIN/workloadservice $*
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
if [ -f "$WORKLOAD_SERVICE_HOME/share/scripts/functions.sh" ]; then
  . $WORKLOAD_SERVICE_HOME/share/scripts/functions.sh
fi

# stored master password
if [ -z "$WORKLOAD_SERVICE_PASSWORD" ] && [ -f $WORKLOAD_SERVICE_CONFIGURATION/.workloadservice_password ]; then
  export WORKLOAD_SERVICE_PASSWORD=$(cat $WORKLOAD_SERVICE_CONFIGURATION/.workloadservice_password)
fi

###################################################################################################
# all other variables with defaults
WORKLOAD_SERVICE_PID_FILE=$WORKLOAD_SERVICE_HOME/workloadservice.pid
WORKLOAD_SERVICE_HTTP_LOG_FILE=$WORKLOAD_SERVICE_LOGS/http.log
WORKLOAD_SERVICE_SETUP_TASKS="SampleSetupTask" # initialize-database and create-key-vault"
###################################################################################################
# ensure that our commands can be found
export PATH=$WORKLOAD_SERVICE_BIN/bin:$PATH

# arguments are optional, if provided they are the names of the tasks to run, in order
workloadservice_setup() {
  local tasklist="$*"
  if [ -z "$tasklist" ]; then
    tasklist=$WORKLOAD_SERVICE_SETUP_TASKS
  fi
  $WORKLOAD_SERVICE_BIN/workloadservice.bin setup $tasklist
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
      $WORKLOAD_SERVICE_BIN/workloadservice.bin start >>$WORKLOAD_SERVICE_HTTP_LOG_FILE 2>&1 &
      echo $! > $WORKLOAD_SERVICE_PID_FILE
    )

    if workloadservice_is_running; then
      echo_success "Started workload service"
    else
      echo_failure "Failed to start workload service"
    fi
}

# returns 0 if workload service is running, 1 if not running
# side effects: sets WORKLOAD_SERVICE_PID if workload service is running, or to empty otherwise
workloadservice_is_running() {
  WORKLOAD_SERVICE_PID=
  if [ -f $WORKLOAD_SERVICE_PID_FILE ]; then
    WORKLOAD_SERVICE_PID=$(cat $WORKLOAD_SERVICE_PID_FILE)
    local is_running=`ps -eo pid | grep "^\s*${WORKLOAD_SERVICE_PID}$"`
    if [ -z "$is_running" ]; then
      # stale PID file
      WORKLOAD_SERVICE_PID=
    fi
  fi
  if [ -z "$WORKLOAD_SERVICE_PID" ]; then
    # check the process list just in case the pid file is stale
    WORKLOAD_SERVICE_PID=$(ps ww | grep -v grep | grep "workloadservice\.bin" | awk '{ print $1 }')
  fi
  if [ -z "$WORKLOAD_SERVICE_PID" ]; then
    #echo "workload service is not running"
    return 1
  fi
  # workload service is running and WORKLOAD_SERVICE_PID is set
  return 0
}

workloadservice_stop() {
  if workloadservice_is_running; then
    kill -9 $WORKLOAD_SERVICE_PID
    if [ $? ]; then
      echo "Stopped workload service"
      rm -f $WORKLOAD_SERVICE_PID_FILE
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
