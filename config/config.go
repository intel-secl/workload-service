package config

import (
	"io"
	"os"
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
	Port = 8444
	UseTLS = false
	Postgres.DBName = "wls"
	Postgres.User = "wlsadmin"
	Postgres.Password = "password"
	Postgres.Hostname = "localhost"
	Postgres.Port = 5432
	Postgres.SSLMode = false
	LogWriter = os.Stdout
}
