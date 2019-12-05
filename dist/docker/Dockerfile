FROM centos:latest

COPY wls-*.bin /tmp/wls.bin
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN env WLS_NOSETUP=true /tmp/wls.bin && rm /tmp/wls.bin
USER wls
ENTRYPOINT [ "/usr/local/bin/entrypoint.sh" ]