package config

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Do not use this casing for GoLang constants unless you are making it match environment variable syntax in bash

// WLS_NOSETUP is a boolean environment variable for skipping WLS Setup tasks
const WLS_NOSETUP = "WLS_NOSETUP"

// WLS_PORT is an integer environment variable for specifying the port WLS should listen on
const WLS_PORT = "WLS_PORT"

// WLS_DB is a string environment variable for specifying the db name to use in the database
const WLS_DB = "WLS_DB"

// WLS_DB_USERNAME is a string environment variable for specifying the username to use for the database connection
const WLS_DB_USERNAME = "WLS_DB_USERNAME"

// WLS_DB_PASSWORD is a string environment variable for specifying the password to use for the database connection
const WLS_DB_PASSWORD = "WLS_DB_PASSWORD"

// WLS_DB_PORT is an integer environment variable for specifying the port to use for the database connection
const WLS_DB_PORT = "WLS_DB_PORT"

// WLS_DB_HOSTNAME is a string environment variable for specifying the database hostname to connect to
const WLS_DB_HOSTNAME = "WLS_DB_HOSTNAME"

// KMS_URL is a string environment variable for specifying the url pointing to the kms, such as https://kms:443/v1/...
const KMS_URL = "KMS_URL"

// KMS_USER is a string environment variable for specifying the username to connect to the KMS
const KMS_USER = "KMS_USER"

// KMS_PASSWORD is a string environment variable for specifying the password to connect to the KMS
const KMS_PASSWORD = "KMS_PASSWORD"

// HVS_URL is a string environment variable for specifying the url pointing to the hvs, such as https://host-verification:8443/mtwilson/v2
const HVS_URL = "HVS_URL"

// HVS_USER is a string environment variable for specifying the username to connect to the HVS
const HVS_USER = "HVS_USER"

// HVS_PASSWORD is a string environment variable for specifying the password to connect to the HVS
const HVS_PASSWORD = "HVS_PASSWORD"

const WLS_LOGLEVEL = "WLS_LOGLEVEL"

const AAS_URL = "AAS_URL"
const AAS_BEARER_TOKEN = "AAS_BEARER_TOKEN"

// Configuration is the global configuration struct that is marshalled/unmarshaled to a persisted yaml file
var Configuration struct {
	Port     int
	TLS      bool
	Postgres struct {
		DBName   string
		User     string
		Password string
		Hostname string
		Port     int
		SSLMode  bool
	}
	HVS struct {
		URL      string
		User     string
		Password string
	}
	KMS struct {
		URL      string
		User     string
		Password string
	}
	AAS_URL string
	
	LogLevel log.Level
}

// Save the configuration struct into /etc/workload-service/config.ynml
func Save() error {
	file, err := os.OpenFile("/etc/workload-service/config.yml", os.O_RDWR, 0)
	if err != nil {
		// we have an error
		if os.IsNotExist(err) {
			// error is that the config doesnt yet exist, create it
			file, err = os.Create("/etc/workload-service/config.yml")
			if err != nil {
				return err
			}
		} else {
			// someother I/O related error
			return err
		}
	}
	defer file.Close()
	return yaml.NewEncoder(file).Encode(Configuration)
}

func init() {
	// load from config
	file, err := os.Open("/etc/workload-service/config.yml")
	if err == nil {
		defer file.Close()
		yaml.NewDecoder(file).Decode(&Configuration)
	}
}
