/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	commLog "intel/isecl/lib/common/v3/log"
	"intel/isecl/lib/common/v3/log/message"
	commLogInt "intel/isecl/lib/common/v3/log/setup"
	cos "intel/isecl/lib/common/v3/os"
	"intel/isecl/lib/common/v3/setup"
	"intel/isecl/workload-service/v3/constants"
	"io"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Do not use this casing for GoLang constants unless you are making it match environment variable syntax in bash

// WLS_NOSETUP is a boolean environment variable for skipping WLS Setup tasks
const WLS_NOSETUP = "WLS_NOSETUP"

// WLS_PORT is an integer environment variable for specifying the port WLS should listen on
const WLS_PORT = "WLS_PORT"

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
		UserName string
		Password string
		Hostname string
		Port     int
		SSLMode  string
		SSLCert  string
	}
	HVS_API_URL  string
	CMS_BASE_URL string
	AAS_API_URL  string
	Subject      struct {
		TLSCertCommonName string
	}
	WLS struct {
		User     string
		Password string
	}
	TLSKeyFile        string
	TLSCertFile       string
	LogLevel          string
	LogEnableStdout   bool
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

	wlsPort, err := c.GetenvInt(WLS_PORT, "WLS Listener Port")
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
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrap(err, "CMS_TLS_CERT_SHA384 is not defined in environment or configuration file")
	}

	cmsBaseUrl, err := c.GetenvString(constants.CmsBaseUrlEnv, "CMS Base URL")
	if err == nil && cmsBaseUrl != "" {
		Configuration.CMS_BASE_URL = cmsBaseUrl
	} else if strings.TrimSpace(Configuration.CMS_BASE_URL) == "" {
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrap(err, "CMS_BASE_URL is not defined in environment or configuration file")
	}

	aasAPIUrl, err := c.GetenvString(AAS_API_URL, "AAS API URL")
	if err == nil && aasAPIUrl != "" {
		Configuration.AAS_API_URL = aasAPIUrl
	} else if strings.TrimSpace(Configuration.AAS_API_URL) == "" {
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrap(err, "AAS_API_URL is not defined in environment or configuration file")
	}

	hvsAPIURL, err := c.GetenvString(HVS_URL, "Verification Service URL")
	if err == nil && hvsAPIURL != "" {
		Configuration.HVS_API_URL = hvsAPIURL
	} else if strings.TrimSpace(Configuration.HVS_API_URL) == "" {
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrap(err, "HVS_URL is not defined in environment or configuration file")
	}

	wlsAASUser, err := c.GetenvString(WLS_USER, "WLS Service Username")
	if err == nil && wlsAASUser != "" {
		Configuration.WLS.User = wlsAASUser
	} else if Configuration.WLS.User == "" {
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrap(err, "WLS_SERVICE_USERNAME is not defined in environment or configuration file")
	}

	wlsAASPassword, err := c.GetenvSecret(WLS_PASSWORD, "WLS Service Password")
	if err == nil && wlsAASPassword != "" {
		Configuration.WLS.Password = wlsAASPassword
	} else if strings.TrimSpace(Configuration.WLS.Password) == "" {
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrap(err, "WLS_SERVICE_PASSWORD is not defined in environment or configuration file")
	}

	tlsKeyPath, err := c.GetenvString("KEY_PATH", "Path of file where TLS key needs to be stored")
	if err == nil && tlsKeyPath != "" {
		Configuration.TLSKeyFile = tlsKeyPath
	} else if Configuration.TLSKeyFile == "" {
		Configuration.TLSKeyFile = constants.DefaultTLSKeyPath
	}

	tlsCertPath, err := c.GetenvString("CERT_PATH", "Path of file/directory where TLS certificate needs to be stored")
	if err == nil && tlsCertPath != "" {
		Configuration.TLSCertFile = tlsCertPath
	} else if Configuration.TLSCertFile == "" {
		Configuration.TLSCertFile = constants.DefaultTLSCertPath
	}

	tlsCertCN, err := c.GetenvString(constants.WlsTLsCertCnEnv, "WLS TLS Certificate Common Name")
	if err == nil && tlsCertCN != "" {
		Configuration.Subject.TLSCertCommonName = tlsCertCN
	} else if strings.TrimSpace(Configuration.Subject.TLSCertCommonName) == "" {
		log.Info("config/config:SaveConfiguration() WLS_TLS_CERT_CN not defined, using default value")
		Configuration.Subject.TLSCertCommonName = constants.DefaultWlsTlsCn
	}

	certSANList, err := c.GetenvString(constants.WlsCertSANList, "WLS Certificate SAN List")
	if err == nil && certSANList != "" {
		Configuration.CertSANList = certSANList
	} else if strings.TrimSpace(Configuration.CertSANList) == "" {
		log.Info("config/config:SaveConfiguration() SAN_LIST List not defined, using default value")
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
	if err == nil && logEntryMaxLength >= 300 {
		Configuration.LogEntryMaxLength = logEntryMaxLength
	} else {
		log.Info("config/config:SaveConfiguration() Invalid Log Entry Max Length defined (should be >= ", constants.DefaultLogEntryMaxlength, "), using default value:", constants.DefaultLogEntryMaxlength)
		Configuration.LogEntryMaxLength = constants.DefaultLogEntryMaxlength
	}

	readTimeout, err := c.GetenvInt("WLS_SERVER_READ_TIMEOUT", "Workload Service Read Timeout")
	if err != nil {
		Configuration.ReadTimeout = constants.DefaultReadTimeout
	} else {
		Configuration.ReadTimeout = time.Duration(readTimeout) * time.Second
	}

	readHeaderTimeout, err := c.GetenvInt("WLS_SERVER_READ_HEADER_TIMEOUT", "Workload Service Read Header Timeout")
	if err != nil {
		Configuration.ReadHeaderTimeout = constants.DefaultReadHeaderTimeout
	} else {
		Configuration.ReadHeaderTimeout = time.Duration(readHeaderTimeout) * time.Second
	}

	writeTimeout, err := c.GetenvInt("WLS_SERVER_WRITE_TIMEOUT", "Workload Service Write Timeout")
	if err != nil {
		Configuration.WriteTimeout = constants.DefaultWriteTimeout
	} else {
		Configuration.WriteTimeout = time.Duration(writeTimeout) * time.Second
	}

	idleTimeout, err := c.GetenvInt("WLS_SERVER_IDLE_TIMEOUT", "Workload Service Idle Timeout")
	if err != nil {
		Configuration.IdleTimeout = constants.DefaultIdleTimeout
	} else {
		Configuration.IdleTimeout = time.Duration(idleTimeout) * time.Second
	}

	maxHeaderBytes, err := c.GetenvInt("WLS_SERVER_MAX_HEADER_BYTES", "Workload Service Max Header Bytes Timeout")
	if err != nil {
		Configuration.MaxHeaderBytes = constants.DefaultMaxHeaderBytes
	} else {
		Configuration.MaxHeaderBytes = maxHeaderBytes
	}

	Configuration.LogEnableStdout = false
	logEnableStdout, err := c.GetenvString(constants.WLSConsoleEnableEnv, "Workload Service enable standard output")
	if err == nil && logEnableStdout != "" {
		Configuration.LogEnableStdout, err = strconv.ParseBool(logEnableStdout)
		if err != nil {
			log.Info("Error while parsing the variable ", constants.WLSConsoleEnableEnv, " setting to default value false")
		}
	}

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
			file, err = os.OpenFile(constants.ConfigFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
			if err != nil {
				return errors.Wrap(err, "config/config:Save() Error in file creation")
			}
		} else {
			// someother I/O related error
			return errors.Wrap(err, "config/config:Save() I/O related error")
		}
	}

	log.Info(message.ConfigChanged)

	defer func() {
		perr := file.Close()
		if perr != nil {
			fmt.Fprintln(os.Stderr, "Error while closing file : "+perr.Error())
		}
	}()
	return yaml.NewEncoder(file).Encode(Configuration)
}

func init() {
	log.Trace("config/config:init() Entering")
	defer log.Trace("config/config:init() Leaving")

	// load from config
	file, err := os.Open(constants.ConfigFile)
	if err == nil {
		defer func() {
			perr := file.Close()
			if perr != nil {
				fmt.Fprintln(os.Stderr, "Error while closing file : "+perr.Error())
			}
		}()
		err = yaml.NewDecoder(file).Decode(&Configuration)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Unable to decode configuration")
		}
	}
	LogWriter = os.Stdout
}

func TakeOwnershipFileWLS(filename string) error {
	// Containers are always run as non root users, does not require changing ownership of config directories
	if _, err := os.Stat("/.container-env"); err == nil {
		return nil
	}

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
		Configuration.LogLevel = logrus.InfoLevel.String()
	}

	llp, err := logrus.ParseLevel(Configuration.LogLevel)
	if err != nil {
		Configuration.LogLevel = logrus.InfoLevel.String()
		llp, _ = logrus.ParseLevel(Configuration.LogLevel)
	}
	commLogInt.SetLogger(commLog.DefaultLoggerName, llp, &commLog.LogFormatter{MaxLength: Configuration.LogEntryMaxLength}, ioWriterDefault, false)
	commLogInt.SetLogger(commLog.SecurityLoggerName, llp, &commLog.LogFormatter{MaxLength: Configuration.LogEntryMaxLength}, ioWriterSecurity, false)

	secLog.Infof("config/config:LogConfiguration() %s", message.LogInit)
	log.Infof("config/config:LogConfiguration() %s", message.LogInit)
}
