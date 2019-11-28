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

Installer Bin will be available in out/wls-*.bin Exportable docker image will be available in out/ as well

### Deploy

```console
> ./wls-*.bin
```

OR

```console
> docker-compose -f dist/docker/docker-compose.yml up
```

### Deployment Config

The table below provides some details on the deployment configuration required in the /root/workload-service.env. A sample is also provided in the dist/linux path.

Variable            | Data Type      | Required? | Default Value       | Description                                                                      | Example
------------------- | -------------- | --------- | ------------------- | -------------------------------------------------------------------------------- | -----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
WLS_LOGLEVEL        | String         | No        | INFO                | Logging level of the Workload Service                                            | Info/Error/Debug
WLS_PORT            | Integer        | No        | 5000                | Listener Port of the Workload Service                                            | 5000
WLS_DB_HOSTNAME     | String         | Yes       | -                   | Hostname for Postgres DB instance                                                | localhost
WLS_DB              | String         | Yes       | -                   | Database schema name for WLS                                                     | wls_pgdb
WLS_DB_PORT         | String         | Yes       | -                   | Postgres DB connection Port                                                      | 5432
WLS_DB_USERNAME     | String         | Yes       | -                   | Postgres DB username                                                             | wlsDbUser
WLS_DB_PASSWORD     | String         | Yes       | -                   | Password for Postgres DB                                                         | wlsDbPassword
HVS_URL             | URL            | Yes       | -                   | Host Verification Service Endpoint                                               | <https://hvs.example.com:8443:/mtwilson/v2/>
CMS_BASE_URL        | URL            | Yes       | -                   | Cert Management Service Endpoint                                                 | <https://certservice.example.com:8445:/cms/v1/>
CMS_TLS_CERT_SHA384 | String         | Yes       | -                   | Sha384 Hash value of the CMS TLS Certificate - required to validate CMS TLS cert
AAS_API_URL         | URL            | Yes       | -                   | AAS Endpoint                                                                     | <https://authservice.example.com:8444/aas>
BEARER_TOKEN        | JWT Token      | Yes       | -                   | JWT token from AAS containing roles required by WLS for setup tasks              | eyJhbGciOiJSUzM4NCIsImtpZCI6IjI1YmY0Mjc2MzNiNTk1YTNmNWU2NjcwYzkxMzc2NjJlMzZiZDJhNDkiLCJ0eXAiOiJKV1QifQ.eyJyb2xlcyI6W3sic2VydmljZSI6IkNNUyIsIm5hbWUiOiJDZXJ0QXBwcm92ZXIiLCJjb250ZXh0IjoiQ049V1BNIEZsYXZvciBTaWduaW5nIENlcnRpZmljYXRlO2NlcnRUeXBlPVNpZ25pbmcifSx7InNlcnZpY2UiOiJDTVMiLCJuYW1lIjoiQ2VydEFwcHJvdmVyIiwiY29udGV4dCI6IkNOPW10d2lsc29uLXNhbWw7Y2VydFR5cGU9U2lnbmluZyJ9LHsic2VydmljZSI6IkNNUyIsIm5hbWUiOiJDZXJ0QXBwcm92ZXIiLCJjb250ZXh0IjoiQ049TXQgV2lsc29uIFRMUyBDZXJ0aWZpY2F0ZTtTQU49MTAuMTA1LjE2OC4xNjAsZGV2MTYwLmZtLmludGVsLmNvbTtjZXJ0VHlwZT1UTFMifSx7InNlcnZpY2UiOiJDTVMiLCJuYW1lIjoiQ2VydEFwcHJvdmVyIiwiY29udGV4dCI6IkNOPUF0dGVzdGF0aW9uIEh1YiBUTFMgQ2VydGlmaWNhdGU7U0FOPTEwLjEwNS4xNjguMTYwLGRldjE2MC5mbS5pbnRlbC5jb207Y2VydFR5cGU9VExTIn0seyJzZXJ2aWNlIjoiQ01TIiwibmFtZSI6IkNlcnRBcHByb3ZlciIsImNvbnRleHQiOiJDTj1XTFMgVExTIENlcnRpZmljYXRlO1NBTj0xMC4xMDUuMTY4LjE2MCxkZXYxNjAuZm0uaW50ZWwuY29tO2NlcnRUeXBlPVRMUyJ9LHsic2VydmljZSI6IkNNUyIsIm5hbWUiOiJDZXJ0QXBwcm92ZXIiLCJjb250ZXh0IjoiQ049S01TIFRMUyBDZXJ0aWZpY2F0ZTtTQU49MTAuMTA1LjE2OC4xNzcsZGV2MTc3LmZtLmludGVsLmNvbTtjZXJ0VHlwZT1UTFMifSx7InNlcnZpY2UiOiJDTVMiLCJuYW1lIjoiQ2VydEFwcHJvdmVyIiwiY29udGV4dCI6IkNOPVZTIEZsYXZvciBTaWduaW5nIENlcnRpZmljYXRlO2NlcnRUeXBlPVNpZ25pbmcifSx7InNlcnZpY2UiOiJWUyIsIm5hbWUiOiJBZG1pbmlzdHJhdG9yIn1dLCJwZXJtaXNzaW9ucyI6W3sic2VydmljZSI6IlZTIiwicnVsZXMiOlsiKjoqOioiXX1dLCJleHAiOjE1NzQ5NTYxNzUsImlhdCI6MTU3NDc4MzM3NSwiaXNzIjoiQUFTIEpXVCBJc3N1ZXIiLCJzdWIiOiJzdXBlcmFkbWluIn0.S6ORVEC8Jz2jrXaMpr0SCnNm4OxandAQkDpEJxthorDaMd8kPCllMri6Fo1Sid7xra28AvWVPAn5WQY1qBCUhYC9ruOLQ3_DgLqmdwg1t-hcgVezSzFJXYhP00Y0IiBZ3KRk5R7FKtoz3JSMJ9b0ecBBCoHHJJvqOx7kWbLbPnEUK3nIJN-hCEgUxrGAkk4WaXzJo4dcEakX37Na5P5-NGawquZb2ryOUy5g8YVDA6239mcWC-RN51VowRtpoG9cJBC2vDDxXkOMnGen7lLpoFgW8LqSOOO6MqxQgoqA04m1kGIe0JHwDkevK4o0U7pa6yRMT9SMJiymP7cnEAoDitxPnAKwx_ipNDFA0Nh-7MRfR6hXzsTb9xJ8BwCIf9wNx9et6mFFlXYsIdRl2c8llNNBfQxIIJp16hJ973Hkl3OzqMBzJGSwr81TnRjcrbRafJYqKQfYTX8XKnF7pYN3RdebstrCKufwrK1xLyoXkW3M4iw153Q3svV9wxZOs3pu
WLS_SERVICE_USERNAME | String | Yes | - | Username in AAS which has the relevant roles assigned for WLS | admin@wls
WLS_SERVICE_PASSWORD        | String         | Yes       | -                   | Password for AAS user account assigned to WLS                                    | wlsAdminPassword
WLS_NOSETUP         | boolean        | No        | WLS No Setup Flag   | If set to "true" the setup tasks are skipped, else the setup tasks are skipped   | true/false
WLS_TLS_CERT_CN     | String         | No        | WLS TLS Certificate | Entry in Common Name field in WLS TLS Certificate                                | My WLS Cert
WLS_CERT_ORG        | String         | No        | INTEL               | Entry in WLS TLS Certificate Organization field                                  | Acmecorp
WLS_CERT_COUNTRY    | String         | No        | US                  | Entry in WLS TLS Certificate Country field                                       | US
WLS_CERT_PROVINCE   | String         | No        | SF                  | Entry in WLS TLS Certificate Province field                                      | CA
WLS_CERT_LOCALITY   | String         | No        | SC                  | Entry in WLS TLS Certificate Locality field                                      | Los Angeles
SAN_LIST            | CSV of strings | No        | 127.0.0.1,localhost | List of FQDNs to be added on Cert Request to CMS                                 | wls.example.com,workloadserivce.example.com,10.12.34.56

## Manage service

- Start service

  - workload-service start

- Stop service

  - workload-service stop

- Status of service

  - workload-service status
