#!/bin/bash

# This is a script file for the libvirt hook. In order for the wsservice to function, it needs the
# home path of the binary. From the binary path, we can find the environment variables and other 
# configuration information.

# For now, we are going to export the home directory. This exact path will be filled out during
# installation time. For now, any variable that is set with VARIABLE_NAME = <AUTOFILL_AT_INSTALL>
# will be replaced with the right value from the variable value. 

export WORKLOAD_SERVICE_HOME = <AUTOFILL_AT_INSTALL>
