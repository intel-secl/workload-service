# Go Workload Service

## Software requirements
- git
- makeself
- `go` version >= `go1.11.4` & <= `go1.12.12`

### Install `go` version >= `go1.11.4` & <= `go1.12.12`
The `Workload Service` requires Go version 1.11.4 that has support for `go modules`. The build was validated with the latest version 1.12.12 of `go`. It is recommended that you use 1.12.12 version of `go`. More recent versions may introduce compatibility issues. You can use the following to install `go`.
```shell
wget https://dl.google.com/go/go1.12.12.linux-amd64.tar.gz
tar -xzf go1.12.12.linux-amd64.tar.gz
sudo mv go /usr/local
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

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
Add / Update following configuration in workload-service.env

    WLS_LOGLEVEL=INFO
    WLS_PORT=5000
    WLS_DB_HOSTNAME=wls-pg-db
    WLS_DB=wls
    WLS_DB_PORT=5432
    WLS_DB_USERNAME=runner
    WLS_DB_PASSWORD=test
    KMS_URL=kmsurl
    HVS_URL=hvsurl
    HVS_USER=hvsuser
    HVS_PASSWORD=hvspass
    CMS_BASE_URL=https://certservice.example.com:8445:/cms/v1/
    AAS_API_URL=https://authservice.example.com:8444/aas
    SAN_LIST=127.0.0.1,localhost
    # BEARER_TOKEN JWT token from AAS for setup tasks
    BEARER_TOKEN=


### Manage service
* Start service
    * workloadservice start
* Stop service
    * workloadservice stop
* Restart service
    * workloadservice restart
* Status of service
    * workloadservice status

