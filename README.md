# Go Workload Service

### Build
```console
> make all
```

Installer Bin will be available in out/wls-*.bin
Exportable docker image will be available in out/ as well


### Deploy
```console
> ./wls-*.bin
```

OR

```console
> docker-compose -f dist/docker/docker-compose.yml up
```

### Config
Add / Update following configuration in wls.env

    WLS_LOGLEVEL=DEBUG
    WLS_PORT=5000
    WLS_DB_HOSTNAME=wls-pg-db
    WLS_DB=wls
    WLS_DB_PORT=5432
    WLS_DB_USERNAME=runner
    WLS_DB_PASSWORD=test
    KMS_URL=kmsurl
    KMS_USER=kmsuser
    KMS_PASSWORD=kmspass
    HVS_URL=hvsurl
    HVS_USER=hvsuser
    HVS_PASSWORD=hvspass

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
