# Go Workload Service

### Build
1. make all

Build will be available in out/wls-*.bin

### Deploy
./wls-*.bin

### Config
Add /Update following configuration in wls.env

WLS_NOSETUP=yes
WLS_PORT=8444
WLS_DB=wls
WLS_DB_USERNAME=root
WLS_DB_PASSWORD=dbpassword
WLS_DB_HOSTNAME=localhost
WLS_DB_PORT=5432

### Manage service
* Start service
    * workloadservice start
* Stop service
    * workloadservice stop
* Restart service
    * workloadservice restart
* Status of service
    * workloadservice status

### v1.0/develop CI Status
[![v1.0/develop pipeline status](https://gitlab.devtools.intel.com/sst/isecl/workload-service/badges/v1.0/develop/pipeline.svg)](https://gitlab.devtools.intel.com/sst/isecl/workload-service/commits/v1.0/develop)
[![v1.0/develop coverage report](https://gitlab.devtools.intel.com/sst/isecl/workload-service/badges/v1.0/develop/coverage.svg)](https://gitlab.devtools.intel.com/sst/isecl/workload-service/commits/v1.0/develop)
