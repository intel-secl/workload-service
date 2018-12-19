package config

import (
	"io"
	"os"
	"strconv"
)

// Config is a global structure of config values
var Port int
var UseTLS bool
var Postgres struct {
	DBName   string
	User     string
	Password string
	Hostname string
	Port     int
	SSLMode  bool
}
var LogWriter io.Writer

func init() {
	Port, _ = strconv.Atoi(os.Getenv("WORKLOAD_SERVICE_PORTNUM"))
	UseTLS = false
	Postgres.DBName = os.Getenv("DATABASE_SCHEMA")
	Postgres.User = os.Getenv("DATABASE_USERNAME")
	Postgres.Password = os.Getenv("DATABASE_PASSWORD")
	Postgres.Hostname = os.Getenv("DATABASE_HOSTNAME")
	Postgres.Port, _ = strconv.Atoi(os.Getenv("DATABASE_PORTNUM"))
	Postgres.SSLMode = false
	LogWriter = os.Stdout
}
