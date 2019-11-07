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
Add / Update following configuration in /root/workload-service.env

    WLS_LOGLEVEL=DEBUG
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
    BEARER_TOKEN=eyJhbGciOiJSUzM4NCIsImtpZCI6ImM3MWM3MjU4Njg3M2MyYmI5Y2MwMjExMWEwYzM5M2JhNjlhYTJhZWQiLCJ0eXAiOiJKV1QifQ.eyJyb2xlcyI6W3sic2VydmljZSI6IkNNUyIsIm5hbWUiOiJDZXJ0QXBwcm92ZXIiLCJjb250ZXh0IjoiQ049V0xTIFRMUyBDZXJ0aWZpY2F0ZTtTQU49MTI3LjAuMC4xLGxvY2FsaG9zdCwxMC4xMDUuMTY4LjE2MDtjZXJ0VHlwZT1UTFMifSx7InNlcnZpY2UiOiJBQVMiLCJuYW1lIjoiUm9sZU1hbmFnZXIiLCJjb250ZXh0IjoiV0xTIn0seyJzZXJ2aWNlIjoiV0xTIiwibmFtZSI6IkFkbWluaXN0cmF0b3IifV0sImV4cCI6MTU3MjcwMDMzNCwiaWF0IjoxNTcyNTI3NTM0LCJpc3MiOiJBQVMgSldUIElzc3VlciIsInN1YiI6ImFkbWluQHdscyJ9.wxqkIA-Kf85N49O0tcNu-ThK6lmRn3rtXgzAFALcpFve6vpzkyUVxsRsd5IgIfC19H5Ds3e5C8QDVZl5Fv0N5DnHVF4DzqimM6EWDF8zXsw8WP3kq1OEuwv0oO5RwfpABZ-Lh2d1y-E2MZi5DReADD_sgDX2uD1jXuiRLiaopPATe6un-5y59ADWryqjqPgDpdQgJ9olcz9Qk7hAz8y-OvTpyak9r8suemgovGSqjad9XCDDVnvFuGSU5pOCkzEufiQ4rGNDXDnxIa8QRC0ehb-cV363fDp9R0jgneegG6qZBGuXh576huq4jvnQ_0BJ4IVbnMjlzn6UfBbIm9z0bv6egOLeg0RR33qA5OC4CEueIWPTdW-gdO5IwLpNmjff1qZd3b0KXnpKrucH71ecgjzFMVkn2jpKFcmtvduluw34TYHgi69lEcVvYFtQPzXvHFFlzM__EMnRNNVijqw-kK2gXi_bO9mCfYj589yMHG1Thop-8gKBW8-j5htK0CF1


### Manage service
* Start service
    * workloadservice start
* Stop service
    * workloadservice stop
* Restart service
    * workloadservice restart
* Status of service
    * workloadservice status

