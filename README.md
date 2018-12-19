# Go Workload Service

### Build
1. cd build
2. ./buildinstallerlocal.sh

Build will be available in build/target/workloadservice-*.bin

### Deploy
./workloadservice-*.bin

### Config
Add /Update following configuration in workloadservice.env

WORKLOAD_SERVICE_SETUP_PREREQS=yes <br />
WORKLOAD_SERVICE_NOSETUP=yes <br />
WORKLOAD_SERVICE_PORTNUM=8444 <br />
DATABASE_SCHEMA=mw_ws <br />
DATABASE_USERNAME=root <br />
DATABASE_PASSWORD=dbpassword <br />
DATABASE_HOSTNAME=localhost <br />
DATABASE_PORTNUM=5432 <br />

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
