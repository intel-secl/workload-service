#!/bin/bash

# Preconditions:
# * http_proxy and https_proxy are already set, if required
# * the mtwilson linux util functions already sourced
#   (for add_package_repository, echo_success, echo_failure)
# * WORKLOAD_SERVICE_HOME is set, for example /opt/workloadservice
# * WORKLOAD_SERVICE_INSTALL_LOG_FILE is set, for example /opt/workloadservice/logs/install.log

# Postconditions:
# * All messages logged to stdout/stderr; caller redirect to logfile as needed

# NOTE:  \cp escapes alias, needed because some systems alias cp to always prompt before override

# Outline:

# source functions file

# TODO : no setup-prereqs here... so just an empty file for now
echo "No setup pre-req tasks currently... so exiting this script for now"