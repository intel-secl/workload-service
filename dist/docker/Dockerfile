FROM centos:8

RUN mkdir -p /opt/workload-service/bin

COPY workload-service  /usr/bin/workload-service

COPY entrypoint.sh /entrypoint.sh

RUN touch /.container-env && chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
