#!/bin/bash

# READ .env file 
echo PWD IS $(pwd)
if [ -f ~/workload-service.env ]; then 
    echo Reading installation options from `realpath ~/workload-service.env`
    source ~/workload-service.env
elif [ -f ../workload-service.env ]; then
    echo Reading installation options from `realpath ../workload-service.env`
    source ../workload-service.env
else
    echo No .env file found
fi

# Export all known variables
export WLS_DB_HOSTNAME
export WLS_DB_PORT
export WLS_DB_USERNAME
export WLS_DB_PASSWORD
export WLS_DB

# Check required variables
if [ -z $WLS_DB_HOSTNAME ] ; then
    echo "DB hostname is not given"
    exit 1
fi
if [ -z $WLS_DB_PORT ] ; then
    echo "DB port is not given"
    exit 1
fi
if [ -z $WLS_DB ] ; then
    echo "DB name is not given"
    exit 1
fi
if [ -z $WLS_DB_USERNAME ] ; then
    echo "DB username is not given"
    exit 1
fi
if [ -z $WLS_DB_PASSWORD ] ; then
    echo "DB password is not given"
    exit 1
fi

echo "Installing postgres database version 11 and its rpm repo for RHEL 7 x86_64 ..."

cd /tmp
log_file=/dev/null
if [ -z $SAVE_DB_INSTALL_LOG ] ; then
	log_file=~/db_install_log
fi

# download postgres repo
yum install https://download.postgresql.org/pub/repos/yum/11/redhat/rhel-7-x86_64/pgdg-redhat-repo-latest.noarch.rpm -y &>> $log_file
yum install postgresql11 postgresql11-server postgresql11-contrib postgresql11-libs -y &>> $log_file

if [ $? -ne 0 ] ; then
	echo "yum installation fail"
	exit 1
fi

echo "Initializing postgres database ..."

# required env variables
export PGDATA=/usr/local/pgsql/data
export PGHOST=$WLS_DB_HOSTNAME
export PGPORT=$WLS_DB_PORT

# if there is no preset database folder, set it up
if [ ! -f $PGDATA/pg_hba.conf ] ; then
	# cleanup and create folders for db
	rm -Rf /usr/local/pgsql
	mkdir -p /usr/local/pgsql/data
	chown -R postgres:postgres /usr/local/pgsql

	sudo -u postgres /usr/pgsql-11/bin/pg_ctl initdb -D $PGDATA &>> $log_file
	
	mv $PGDATA/pg_hba.conf $PGDATA/pg_hba-template.conf
	echo "local all postgres peer" >> $PGDATA/pg_hba.conf
	echo "local all all md5" >> $PGDATA/pg_hba.conf
	echo "host all all 127.0.0.1/32 md5" >> $PGDATA/pg_hba.conf
fi

echo "Setting up systemctl for postgres database ..."

# setup systemd startup for postgresql
pg_systemd=/usr/lib/systemd/system/postgresql-11.service
rm -rf $pg_systemd
echo "[Unit]" >> $pg_systemd
echo "Description=PostgreSQL database server" >> $pg_systemd
echo "Documentation=https://www.postgresql.org/docs/11/static/" >> $pg_systemd
echo "After=network.target" >> $pg_systemd
echo "After=syslog.target" >> $pg_systemd
echo "" >> $pg_systemd
echo "[Service]" >> $pg_systemd
echo "Type=forking" >> $pg_systemd
echo "User=postgres" >> $pg_systemd
echo "Group=postgres" >> $pg_systemd
echo "Environment=PGDATA=${PGDATA}" >> $pg_systemd
echo "OOMScoreAdjust=-1000" >> $pg_systemd
echo "Environment=PG_OOM_ADJUST_FILE=/proc/self/oom_score_adj" >> $pg_systemd
echo "Environment=PG_OOM_ADJUST_VALUE=0" >> $pg_systemd
echo "ExecStart=/usr/pgsql-11/bin/pg_ctl start -D ${PGDATA} -l ${PGDATA}/pg_log" >> $pg_systemd
echo "ExecStop=/usr/pgsql-11/bin/pg_ctl stop -D ${PGDATA}" >> $pg_systemd
echo "ExecReload=/usr/pgsql-11/bin/pg_ctl reload -D ${PGDATA}" >> $pg_systemd
echo "" >> $pg_systemd
echo "TimeoutSec=300" >> $pg_systemd
echo "" >> $pg_systemd
echo "[Install]" >> $pg_systemd
echo "WantedBy=multi-user.target" >> $pg_systemd
echo "" >> $pg_systemd

systemctl daemon-reload &>> $log_file
systemctl enable postgresql-11.service &>> $log_file

echo "Setting up postgres database ..."

# start the database service
service postgresql-11 start &>> $log_file

sudo -u postgres psql postgres -c "alter system set log_connections = 'on';" &>> $log_file
sudo -u postgres psql postgres -c "alter system set log_disconnections = 'on';" &>> $log_file
sudo -u postgres psql postgres -c "CREATE EXTENSION \"uuid-ossp\";" &>> $log_file
sudo -u postgres psql postgres -c "CREATE USER ${WLS_DB_USERNAME} WITH PASSWORD '${WLS_DB_PASSWORD}';" &>> $log_file
sudo -u postgres psql postgres -c "CREATE DATABASE ${WLS_DB}" &>> $log_file
sudo -u postgres psql postgres -c "GRANT ALL PRIVILEGES ON DATABASE ${WLS_DB} TO ${WLS_DB_USERNAME};" &>> $log_file
sudo -u postgres psql postgres -c "ALTER ROLE ${WLS_DB_USERNAME} NOSUPERUSER;" &>> $log_file
sudo -u postgres psql postgres -c "ALTER ROLE ${WLS_DB_USERNAME} NOCREATEROLE;" &>> $log_file
sudo -u postgres psql postgres -c "ALTER ROLE ${WLS_DB_USERNAME} NOCREATEDB;" &>> $log_file
sudo -u postgres psql postgres -c "ALTER ROLE ${WLS_DB_USERNAME} NOREPLICATION;" &>> $log_file
sudo -u postgres psql postgres -c "ALTER ROLE ${WLS_DB_USERNAME} NOBYPASSRLS;" &>> $log_file
sudo -u postgres psql postgres -c "ALTER ROLE ${WLS_DB_USERNAME} NOINHERIT;" &>> $log_file
