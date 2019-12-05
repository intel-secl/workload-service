/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package config

import (
	"fmt"
	commLog "intel/isecl/lib/common/log"
	commLogInt "intel/isecl/lib/common/log/setup"
	cos "intel/isecl/lib/common/os"
	"intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/constants"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

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

// WLS_SERVICE_USERNAME is a string environment variable for specifying  user to get token from AAS
const WLS_USER = "WLS_SERVICE_USERNAME"

// WLS_SERVICE_PASSWORD is a string environment variable for specifying the password get token from AAS
const WLS_PASSWORD = "WLS_SERVICE_PASSWORD"

const WLS_LOGLEVEL = "WLS_LOGLEVEL"

const AAS_API_URL = "AAS_API_URL"

const KEY_CACHE_SECONDS = "KEY_CACHE_SECONDS"

// Configuration is the global configuration struct that is marshalled/unmarshaled to a persisted yaml file
var Configuration struct {
	Port             int
	CmsTlsCertDigest string
	Postgres         struct {
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
	LogLevel          string
	LogEntryMaxLength int
	KEY_CACHE_SECONDS int
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
	CertSANList       string
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

	wlsPort, err := c.GetenvInt(constants.CmsTlsCertDigestEnv, "WLS Listener Port")
	if err == nil && wlsPort > 0 {
		Configuration.Port = wlsPort
	} else if Configuration.Port <= 0 {
		Configuration.Port = constants.DefaultWLSListenerPort
		log.Info("config/config:SaveConfiguration() WLS_PORT not defined, using default value: ", constants.DefaultWLSListenerPort)
	}

	tlsCertDigest, err := c.GetenvString(constants.CmsTlsCertDigestEnv, "CMS TLS certificate digest")
	if err == nil && tlsCertDigest != "" {
		Configuration.CmsTlsCertDigest = tlsCertDigest
	} else if strings.TrimSpace(Configuration.CmsTlsCertDigest) == "" {
		return errors.Wrap(err, "CMS_TLS_CERT_SHA384 is not defined in environment or configuration file")
	}

	cmsBaseUrl, err := c.GetenvString(constants.CmsBaseUrlEnv, "CMS Base URL")
	if err == nil && cmsBaseUrl != "" {
		Configuration.CMS_BASE_URL = cmsBaseUrl
	} else if strings.TrimSpace(Configuration.CMS_BASE_URL) == "" {
		return errors.Wrap(err, "CMS_BASE_URL is not defined in environment or configuration file")
	}

	aasAPIUrl, err := c.GetenvString(AAS_API_URL, "AAS API URL")
	if err == nil && aasAPIUrl != "" {
		Configuration.AAS_API_URL = aasAPIUrl
	} else if strings.TrimSpace(Configuration.AAS_API_URL) == "" {
		return errors.Wrap(err, "AAS_API_URL is not defined in environment or configuration file")
	}

	hvsAPIURL, err := c.GetenvString(HVS_URL, "Verification Service URL")
	if err == nil && hvsAPIURL != "" {
		Configuration.HVS_API_URL = hvsAPIURL
	} else if strings.TrimSpace(Configuration.HVS_API_URL) == "" {
		return errors.Wrap(err, "HVS_URL is not defined in environment or configuration file")
	}

	wlsAASUser, err := c.GetenvString(WLS_USER, "WLS Service Username")
	if err == nil && wlsAASUser != "" {
		Configuration.WLS.User = wlsAASUser
	} else if Configuration.WLS.User == "" {
		return errors.Wrap(err, "WLS_SERVICE_USERNAME is not defined in environment or configuration file")
	}

	wlsAASPassword, err := c.GetenvString(WLS_PASSWORD, "WLS Service Password")
	if err == nil && wlsAASPassword != "" {
		Configuration.WLS.Password = wlsAASPassword
	} else if strings.TrimSpace(Configuration.WLS.Password) == "" {
		return errors.Wrap(err, "WLS_SERVICE_PASSWORD is not defined in environment or configuration file")
	}

	// Postgres DB configuration
	wlsDBHostname, err := c.GetenvString(WLS_DB_HOSTNAME, "WLS DB Hostname")
	if err == nil && wlsDBHostname != "" {
		Configuration.Postgres.Hostname = wlsDBHostname
	} else if strings.TrimSpace(Configuration.Postgres.Hostname) == "" {
		return errors.Wrap(err, "WLS_DB_HOSTNAME is not defined in environment or configuration file")
	}

	wlsDBPort, err := c.GetenvInt(WLS_DB_PORT, "WLS DB Port")
	if err == nil && wlsDBPort > 0 {
		Configuration.Postgres.Port = wlsDBPort
	} else if Configuration.Postgres.Port == 0 {
		return errors.Wrap(err, "WLS_DB_PORT is not defined in environment or configuration file")
	}

	wlsDBUsername, err := c.GetenvString(WLS_DB_USERNAME, "WLS DB Username")
	if err == nil && wlsDBUsername != "" {
		Configuration.Postgres.User = wlsDBUsername
	} else if strings.TrimSpace(Configuration.Postgres.User) == "" {
		return errors.Wrap(err, "WLS_DB_USERNAME is not defined in environment or configuration file")
	}

	wlsDBPassword, err := c.GetenvString(WLS_DB_PASSWORD, "WLS DB Password")
	if err == nil && wlsDBPassword != "" {
		Configuration.Postgres.Password = wlsDBPassword
	} else if strings.TrimSpace(Configuration.Postgres.Password) == "" {
		return errors.Wrap(err, "WLS_DB_PASSWORD is not defined in environment or configuration file")
	}

	wlsDBName, err := c.GetenvString(WLS_DB, "WLS DB Name")
	if err == nil && wlsDBName != "" {
		Configuration.Postgres.DBName = wlsDBName
	} else if strings.TrimSpace(Configuration.Postgres.DBName) == "" {
		return errors.Wrap(err, "WLS_DB is not defined in environment or configuration file")
	}

	tlsCertCN, err := c.GetenvString(constants.WlsTLsCertCnEnv, "WLS TLS Certificate Common Name")
	if err == nil && tlsCertCN != "" {
		Configuration.Subject.TLSCertCommonName = tlsCertCN
	} else if strings.TrimSpace(Configuration.Subject.TLSCertCommonName) == "" {
		log.Info("config/config:SaveConfiguration() WLS_TLS_CERT_CN not defined, using default value")
		Configuration.Subject.TLSCertCommonName = constants.DefaultWlsTlsCn
	}

	certOrg, err := c.GetenvString(constants.WlsCertOrgEnv, "WLS Certificate Organization")
	if err == nil && certOrg != "" {
		Configuration.Subject.Organization = certOrg
	} else if strings.TrimSpace(Configuration.Subject.Organization) == "" {
		log.Info("config/config:SaveConfiguration() WLS_CERT_ORG not defined, using default value")
		Configuration.Subject.Organization = constants.DefaultWlsCertOrganization
	}

	certCountry, err := c.GetenvString(constants.WlsCertCountryEnv, "WLS Certificate Country")
	if err == nil && certCountry != "" {
		Configuration.Subject.Country = certCountry
	} else if strings.TrimSpace(Configuration.Subject.Country) == "" {
		log.Info("config/config:SaveConfiguration() WLS_CERT_COUNTRY not defined, using default value")
		Configuration.Subject.Country = constants.DefaultWlsCertCountry
	}

	certProvince, err := c.GetenvString(constants.WlsCertProvinceEnv, "WLS Certificate Province")
	if err == nil && certProvince != "" {
		Configuration.Subject.Province = certProvince
	} else if strings.TrimSpace(Configuration.Subject.Province) == "" {
		log.Info("config/config:SaveConfiguration() WLS_CERT_PROVINCE not defined, using default value")
		Configuration.Subject.Province = constants.DefaultWlsCertProvince
	}

	certLocality, err := c.GetenvString(constants.WlsCertLocalityEnv, "WLS Certificate Locality")
	if err == nil && certLocality != "" {
		Configuration.Subject.Locality = certLocality
	} else if strings.TrimSpace(Configuration.Subject.Locality) == "" {
		log.Info("config/config:SaveConfiguration() WLS_CERT_LOCALITY not defined, using default value")
		Configuration.Subject.Locality = constants.DefaultWlsCertLocality
	}

	certSANList, err := c.GetenvString(constants.WlsCertSANList, "WLS Certificate SAN List")
	if err == nil && certSANList != "" {
		Configuration.CertSANList = certSANList
	} else if strings.TrimSpace(Configuration.CertSANList) == "" {
		log.Info("config/config:SaveConfiguration() WLS_CERT_SAN List not defined, using default value")
		Configuration.CertSANList = constants.DefaultWlsTlsSan
	}

	keyCacheSeconds, err := c.GetenvString(constants.KeyCacheSeconds, "Key Cache Seconds")
	if err == nil && keyCacheSeconds != "" {
		Configuration.KEY_CACHE_SECONDS, _ = strconv.Atoi(keyCacheSeconds)
	} else if Configuration.KEY_CACHE_SECONDS <= 0 {
		log.Info("config/config:SaveConfiguration() KEY_CACHE_SECONDS not defined, using default value")
		Configuration.KEY_CACHE_SECONDS = constants.DefaultKeyCacheSeconds
	}

	ll, err := c.GetenvString(WLS_LOGLEVEL, "Logging Level")
	if err != nil {
		if Configuration.LogLevel == "" {
			log.Infof("config/config:SaveConfiguration() %s not defined, using default log level: Info", WLS_LOGLEVEL)
			Configuration.LogLevel = logrus.InfoLevel.String()
		}
	} else {
		llp, err := logrus.ParseLevel(ll)
		if err != nil {
			log.Info("config/config:SaveConfiguration() Invalid log level specified in env, using default log level: Info")
			Configuration.LogLevel = logrus.InfoLevel.String()
		} else {
			Configuration.LogLevel = llp.String()
			log.Infof("config/config:SaveConfiguration() Log level set %s\n", ll)
		}
	}

	logEntryMaxLength, err := c.GetenvInt(constants.LogEntryMaxlengthEnv, "Maximum length of each entry in a log")
	if err == nil && logEntryMaxLength >= 100 {
		Configuration.LogEntryMaxLength = logEntryMaxLength
	} else {
		log.Info("config/config:SaveConfiguration() Invalid Log Entry Max Length defined (should be > 100), " +
			"using default value")
		Configuration.LogEntryMaxLength = constants.DefaultLogEntryMaxlength
	}

	log.Info("config/config:SaveConfiguration() Saving Environment variables inside the configuration file")

	return Save()
}

// Save the configuration struct into /etc/workload-service/config.ynml
func Save() error {
	log.Trace("config/config:Save() Entering")
	defer log.Trace("config/config:Save() Leaving")

	file, err := os.OpenFile(constants.ConfigFile, os.O_WRONLY, 0)
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

func TakeOwnershipFileWLS(filename string) error {
	// when successful, we update the ownership of the config/certs updated by the setup tasks
	// all of them are likely to be found in /etc/workload-service/ path
	wlsUser, err := user.Lookup(constants.WLSRuntimeUser)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Unable to find service user - wls. Failed to set permissions on WLS configuration files")
		return errors.Wrapf(err, "Failed to set permissions on %s", filename)
	}
	wlsGroup, err := user.LookupGroup(constants.WLSRuntimeGroup)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Unable to find service group - wls. Failed to set permissions on WLS configuration files")
		return errors.Wrapf(err, "Failed to set permissions on %s", filename)
	}
	wlsuid, err := strconv.Atoi(wlsUser.Uid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Unable to get service user id - wls")
		return errors.Wrapf(err, "Failed to set permissions on %s", filename)
	}
	wlsgid, err := strconv.Atoi(wlsGroup.Gid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Unable to get service group id - wls")
		return errors.Wrapf(err, "Failed to set permissions on %s", filename)
	}
	err = cos.ChownR(filename, wlsuid, wlsgid)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Failed to set permissions on "+filename)
		return errors.Wrapf(err, "Failed to set permissions on %s", filename)
	}

	if err != nil {
		fmt.Println("Error updating permissions on WLS config ", constants.ConfigDir)
		log.Errorf("config/config:TakeOwnershipFileWLS() Error updating permissions on config path %s: %s", constants.ConfigDir, err)
	}
	return nil
}

func LogConfiguration(stdOut, logFile bool) {
	log.Trace("config/config:LogConfiguration() Entering")
	defer log.Trace("config/config:LogConfiguration() Leaving")

	// creating the log file if not preset
	var ioWriterDefault io.Writer
	secLogFile, _ := os.OpenFile(constants.SecurityLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	defaultLogFile, _ := os.OpenFile(constants.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)

	err := TakeOwnershipFileWLS(constants.LogDir)
	if err != nil {
		fmt.Println("Error taking log path ownership ", constants.LogDir)
		os.Exit(-1)
	}

	ioWriterDefault = defaultLogFile
	if stdOut && logFile {
		ioWriterDefault = io.MultiWriter(os.Stdout, defaultLogFile)
	}
	if stdOut && !logFile {
		ioWriterDefault = os.Stdout
	}
	ioWriterSecurity := io.MultiWriter(ioWriterDefault, secLogFile)

	if Configuration.LogLevel == "" {
		log.Infof("config/config:SaveConfiguration() %s not defined, using default log level: Info\n", WLS_LOGLEVEL)
		Configuration.LogLevel = logrus.InfoLevel.String()
	}

	llp, _ := logrus.ParseLevel(Configuration.LogLevel)
	commLogInt.SetLogger(commLog.DefaultLoggerName, llp, &commLog.LogFormatter{MaxLength: Configuration.LogEntryMaxLength}, ioWriterDefault, false)
	commLogInt.SetLogger(commLog.SecurityLoggerName, llp, &commLog.LogFormatter{MaxLength: Configuration.LogEntryMaxLength}, ioWriterSecurity, false)

	secLog.Trace("config/config:LogConfiguration() Security log initiated")
	log.Trace("config/config:LogConfiguration() Loggers setup finished")
}
