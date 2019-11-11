/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package config

import (
	"fmt"
	commLog "intel/isecl/lib/common/log"
	commLogInt "intel/isecl/lib/common/log/setup"
	"intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/constants"
	"io"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	CmsTlsCertDigest string
	Postgres struct {
		DBName   string
		User     string
		Password string
		Hostname string
		Port     int
		SSLMode  bool
	}
	HVS_API_URL  string
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
	LogLevel          logrus.Level
	KEY_CACHE_SECONDS int
}

var log = commLog.GetDefaultLogger()
var secLog = commLog.GetSecurityLogger()

var (
	LogWriter io.Writer
)

func SaveConfiguration(c setup.Context) error {
	log.Trace("config/config:SaveConfiguration() Entering")
	defer log.Trace("config/config:SaveConfiguration() Leaving")
	var err error = nil

	tlsCertDigest, err := c.GetenvString(constants.CmsTlsCertDigestEnv, "CMS TLS certificate digest")
	if err == nil &&  tlsCertDigest != "" {
		Configuration.CmsTlsCertDigest = tlsCertDigest
	} else if Configuration.CmsTlsCertDigest == "" {
		return errors.Wrap(err, "config/config:SaveConfiguration() CMS_TLS_CERT_SHA384 is not defined in environment or configuration file")
	}

	cmsBaseUrl, err := c.GetenvString(constants.CmsBaseUrlEnv, "CMS Base URL")
	if err == nil && cmsBaseUrl != "" {
		Configuration.CMS_BASE_URL = cmsBaseUrl
	} else if Configuration.CMS_BASE_URL == "" {
		return errors.Wrap(err, "config/config:SaveConfiguration() CMS_BASE_URL is not defined in environment or configuration file")
	}

	aasAPIUrl, err := c.GetenvString(AAS_API_URL, "AAS API URL")
	if err == nil && aasAPIUrl != "" {
		Configuration.AAS_API_URL = aasAPIUrl
	} else if Configuration.AAS_API_URL == "" {
		return errors.Wrap(err, "config/config:SaveConfiguration() AAS_API_URL is not defined in environment or configuration file")
	}

	hvsAPIURL, err := c.GetenvString(HVS_URL, "Verification Service URL")
	if err == nil && hvsAPIURL != "" {
		Configuration.HVS_API_URL = hvsAPIURL
	} else if Configuration.HVS_API_URL == "" {
		return errors.Wrap(err, "config/config:SaveConfiguration() HVS_URL is not defined in environment or configuration file")
	}

	wlsAASUser, err := c.GetenvString(WLS_USER, "WLS AAS User")
	if err == nil && wlsAASUser != "" {
		Configuration.WLS.User = wlsAASUser
	} else if Configuration.WLS.User == "" {
		return errors.Wrap(err, "config/config:SaveConfiguration() WLS_USER is not defined in environment or configuration file")
	}

	wlsAASPassword, err := c.GetenvString(WLS_PASSWORD, "WLS AAS Password")
	if err == nil && wlsAASPassword != "" {
		Configuration.WLS.User = wlsAASPassword
	} else if Configuration.WLS.Password == "" {
		return errors.Wrap(err, "config/config:SaveConfiguration() WLS_PASSWORD is not defined in environment or configuration file")
	}

	tlsCertCN, err := c.GetenvString(constants.WlsTLsCertCnEnv, "WLS TLS Certificate Common Name")
	if err == nil && tlsCertCN != "" {
		Configuration.Subject.TLSCertCommonName = tlsCertCN
	} else if Configuration.Subject.TLSCertCommonName == "" {
		log.Info("config/config:SaveConfiguration() WLS TLS Certificate Common Name not defined, using default value")
		Configuration.Subject.TLSCertCommonName = constants.DefaultWlsTlsCn
	}

	certOrg, err := c.GetenvString(constants.WlsCertOrgEnv, "WLS Certificate Organization")
	if err == nil && certOrg != "" {
		Configuration.Subject.Organization = certOrg
	} else if Configuration.Subject.Organization == "" {
		log.Info("config/config:SaveConfiguration() WLS Certificate Organization not defined, using default value")
		Configuration.Subject.Organization = constants.DefaultWlsCertOrganization
	}

	certCountry, err := c.GetenvString(constants.WlsCertCountryEnv, "WLS Certificate Country")
	if err == nil && certCountry != "" {
		Configuration.Subject.Country = certCountry
	} else if Configuration.Subject.Country == "" {
		log.Info("config/config:SaveConfiguration() WLS Certificate Country not defined, using default value")
		Configuration.Subject.Country = constants.DefaultWlsCertCountry
	}

	certProvince, err := c.GetenvString(constants.WlsCertProvinceEnv, "WLS Certificate Province")
	if err == nil && certProvince != "" {
		Configuration.Subject.Province = certProvince
	} else if Configuration.Subject.Province == "" {
		log.Info("config/config:SaveConfiguration() WLS Certificate Province not defined, using default value")
		Configuration.Subject.Province = constants.DefaultWlsCertProvince
	}

	certLocality, err := c.GetenvString(constants.WlsCertLocalityEnv, "WLS Certificate Locality")
	if err == nil && certLocality != "" {
		Configuration.Subject.Locality = certLocality
	} else if Configuration.Subject.Locality == "" {
		log.Info("config/config:SaveConfiguration() WLS Certificate Locality not defined, using default value")
		Configuration.Subject.Locality = constants.DefaultWlsCertLocality
	}

	keyCacheSeconds, err := c.GetenvString(constants.KeyCacheSeconds, "Key Cache Seconds")
	if err == nil && keyCacheSeconds != "" {
		Configuration.KEY_CACHE_SECONDS, _ = strconv.Atoi(keyCacheSeconds)
	} else if Configuration.KEY_CACHE_SECONDS <= 0 {
		log.Info("config/config:SaveConfiguration() Key Cache Seconds not defined, using default value")
		Configuration.KEY_CACHE_SECONDS = constants.DefaultKeyCacheSeconds
	}

	ll, err := c.GetenvString(WLS_LOGLEVEL, "Logging Level")
	if err != nil {
		fmt.Fprintln(os.Stderr, "No logging level specified, using default logging level: Error")
		Configuration.LogLevel = logrus.ErrorLevel
	}
	Configuration.LogLevel, err = logrus.ParseLevel(ll)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid logging level specified, using default logging level: Error")
		Configuration.LogLevel = logrus.ErrorLevel
	}

	fmt.Println("Configuration Loaded")
	log.Info("config/config:SaveConfiguration() Saving Environment variables inside the configuration file")
	return Save()
}

// Save the configuration struct into /etc/workload-service/config.ynml
func Save() error {
	log.Trace("config/config:Save() Entering")
	defer log.Trace("config/config:Save() Leaving")

	file, err := os.OpenFile(constants.ConfigFile, os.O_RDWR, 0)
	if err != nil {
		// we have an error
		if os.IsNotExist(err) {
			// error is that the config doesnt yet exist, create it
			log.Debug("config/config:Save() File does not exist, creating a file... ")
			file, err = os.Create(constants.ConfigFile)
			if err != nil {
				return errors.Wrap(err, "config/config:Save() Error in file creation")
			}
		} else {
			// someother I/O related error
			return errors.Wrap(err, "config/config:Save() I/O related error")
		}
	}
	defer file.Close()
	return yaml.NewEncoder(file).Encode(Configuration)
}

func init() {
	log.Trace("config/config:init() Entering")
	defer log.Trace("config/config:init() Leaving")

	// load from config
	file, err := os.Open(constants.ConfigFile)
	if err == nil {
		defer file.Close()
		yaml.NewDecoder(file).Decode(&Configuration)
	}
	LogWriter = os.Stdout
}

func LogConfiguration(stdOut, logFile bool) {
	log.Trace("config/config:LogConfiguration() Entering")
	defer log.Trace("config/config:LogConfiguration() Leaving")

	// creating the log file if not preset
	var ioWriterDefault io.Writer
	secLogFile, _ := os.OpenFile(constants.SecurityLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	defaultLogFile, _ := os.OpenFile(constants.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)

	ioWriterDefault = defaultLogFile
	if stdOut && logFile {
		ioWriterDefault = io.MultiWriter(os.Stdout, defaultLogFile)
	}
	if stdOut && !logFile {
		ioWriterDefault = os.Stdout
	}
	ioWriterSecurity := io.MultiWriter(ioWriterDefault, secLogFile)

	commLogInt.SetLogger(commLog.DefaultLoggerName, Configuration.LogLevel, nil, ioWriterDefault, false)
	commLogInt.SetLogger(commLog.SecurityLoggerName, Configuration.LogLevel, nil, ioWriterSecurity, false)
	secLog.Trace("config/config:LogConfiguration() Security log initiated")
	log.Trace("config/config:LogConfiguration() Loggers setup finished")
}
