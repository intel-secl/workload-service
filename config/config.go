/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package config

import (
	"intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/constants"
	"os"
    "strconv"
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

// HVS_URL is a string environment variable for specifying the url pointing to the hvs, such as https://host-verification:8443/mtwilson/v2
const HVS_URL = "HVS_URL"

// WLS_USER is a string environment variable for specifying  user to get token from AAS
const WLS_USER = "WLS_USER"

// WLS_PASSWORD is a string environment variable for specifying the password get token from AAS
const WLS_PASSWORD = "WLS_PASSWORD"

const WLS_LOGLEVEL = "WLS_LOGLEVEL"

const AAS_API_URL = "AAS_API_URL"

const KEY_CACHE_SECONDS = "KEY_CACHE_SECONDS"
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
	HVS_API_URL string
	CMS_BASE_URL string
	AAS_API_URL  string
	Subject      struct {
		TLSCertCommonName string
		Organization      string
		Country           string
		Province          string
		Locality          string
	}
	WLS struct {
		User     string
		Password string
	}
	LogLevel log.Level
    KEY_CACHE_SECONDS int

}

func SaveConfiguration(c setup.Context) error {
	var err error = nil

	cmsBaseUrl, err := c.GetenvString(constants.CmsBaseUrlEnv, "CMS Base URL")
	if err == nil && cmsBaseUrl != "" {
		Configuration.CMS_BASE_URL = cmsBaseUrl
	} else if Configuration.CMS_BASE_URL == "" {
		log.Error("CMS_BASE_URL is not defined in environment or configuration file")
	}

	tlsCertCN, err := c.GetenvString(constants.WlsTLsCertCnEnv, "WLS TLS Certificate Common Name")
	if err == nil && tlsCertCN != "" {
		Configuration.Subject.TLSCertCommonName = tlsCertCN
	} else if Configuration.Subject.TLSCertCommonName == "" {
		Configuration.Subject.TLSCertCommonName = constants.DefaultWlsTlsCn
	}

	certOrg, err := c.GetenvString(constants.WlsCertOrgEnv, "WLS Certificate Organization")
	if err == nil && certOrg != "" {
		Configuration.Subject.Organization = certOrg
	} else if Configuration.Subject.Organization == "" {
		Configuration.Subject.Organization = constants.DefaultWlsCertOrganization
	}

	certCountry, err := c.GetenvString(constants.WlsCertCountryEnv, "WLS Certificate Country")
	if err == nil && certCountry != "" {
		Configuration.Subject.Country = certCountry
	} else if Configuration.Subject.Country == "" {
		Configuration.Subject.Country = constants.DefaultWlsCertCountry
	}

	certProvince, err := c.GetenvString(constants.WlsCertProvinceEnv, "WLS Certificate Province")
	if err == nil && certProvince != "" {
		Configuration.Subject.Province = certProvince
	} else if Configuration.Subject.Province == "" {
		Configuration.Subject.Province = constants.DefaultWlsCertProvince
	}

	certLocality, err := c.GetenvString(constants.WlsCertLocalityEnv, "WLS Certificate Locality")
	if err == nil && certLocality != "" {
		Configuration.Subject.Locality = certLocality
	} else if Configuration.Subject.Locality == "" {
		Configuration.Subject.Locality = constants.DefaultWlsCertLocality
	}

	keyCacheSeconds, err := c.GetenvString(constants.KeyCacheSeconds, "Key Cache Seconds")
	if err == nil && keyCacheSeconds != "" {
		Configuration.KEY_CACHE_SECONDS, _ = strconv.Atoi(keyCacheSeconds)
	} else if Configuration.KEY_CACHE_SECONDS <= 0 {
		Configuration.KEY_CACHE_SECONDS = constants.DefaultKeyCacheSeconds
	}

	return Save()

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
