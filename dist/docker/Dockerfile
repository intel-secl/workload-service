FROM centos:8

RUN useradd -u 1001 --comment "Workload Service" --home /opt/workload-service --shell /bin/false wls

RUN mkdir -p /opt/workload-service/bin
COPY workload-service  /usr/bin/workload-service

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT [ "/usr/local/bin/entrypoint.sh" ]
