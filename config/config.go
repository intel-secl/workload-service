/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package config

import (
	"fmt"
	"github.com/intel-secl/intel-secl/v3/pkg/lib/common/utils"
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
	HvsApiUrl  string `yaml:"hvs_api_url"`
	CmsBaseUrl string `yaml:"cms_base_url"`
	AasApiUrl  string `yaml:"aas_api_url"`
	Subject    struct {
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
	KeyCacheSeconds   int `yaml:"key_cache_seconds"`
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

	wlsPort, err := c.GetenvInt(constants.WlsPortEnv, "WLS Listener Port")
	if err == nil && wlsPort > 0 {
		Configuration.Port = wlsPort
	} else if Configuration.Port <= 0 {
		Configuration.Port = constants.DefaultWLSListenerPort
		log.Infof("config/config:SaveConfiguration() %s not defined, using default value: %d", constants.WlsPortEnv, constants.DefaultWLSListenerPort)
	}

	tlsCertDigest, err := c.GetenvString(constants.CmsTlsCertDigestEnv, "CMS TLS certificate digest")
	if err == nil && tlsCertDigest != "" {
		Configuration.CmsTlsCertDigest = tlsCertDigest
	} else if strings.TrimSpace(Configuration.CmsTlsCertDigest) == "" {
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrapf(err, "%s is not defined in environment or configuration file", constants.CmsTlsCertDigestEnv)
	}

	cmsBaseUrl, err := c.GetenvString(constants.CmsBaseUrlEnv, "CMS Base URL")
	if err == nil && cmsBaseUrl != "" {
		Configuration.CmsBaseUrl = cmsBaseUrl
	} else if strings.TrimSpace(Configuration.CmsBaseUrl) == "" {
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrapf(err, "%s is not defined in environment or configuration file", constants.CmsBaseUrlEnv)
	}

	aasAPIUrl, err := c.GetenvString(constants.AasApiUrlEnv, "AAS API URL")
	if err == nil && aasAPIUrl != "" {
		Configuration.AasApiUrl = aasAPIUrl
	} else if strings.TrimSpace(Configuration.AasApiUrl) == "" {
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrapf(err, "%s is not defined in environment or configuration file", constants.AasApiUrlEnv)
	}

	hvsAPIURL, err := c.GetenvString(constants.HvsUrlEnv, "Verification Service URL")
	if err == nil && hvsAPIURL != "" {
		Configuration.HvsApiUrl = hvsAPIURL
	} else if strings.TrimSpace(Configuration.HvsApiUrl) == "" {
		log.Error(message.InvalidInputProtocolViolation)
		return errors.Wrapf(err, "%s is not defined in environment or configuration file", constants.HvsUrlEnv)
	}

	tlsKeyPath, err := c.GetenvString(constants.TLSKeyPathEnv, "Path of file where TLS key needs to be stored")
	if err == nil && tlsKeyPath != "" {
		Configuration.TLSKeyFile = tlsKeyPath
	} else if Configuration.TLSKeyFile == "" {
		Configuration.TLSKeyFile = constants.DefaultTLSKeyPath
	}

	tlsCertPath, err := c.GetenvString(constants.TLSCertPathEnv, "Path of file/directory where TLS certificate needs to be stored")
	if err == nil && tlsCertPath != "" {
		Configuration.TLSCertFile = tlsCertPath
	} else if Configuration.TLSCertFile == "" {
		Configuration.TLSCertFile = constants.DefaultTLSCertPath
	}

	tlsCertCN, err := c.GetenvString(constants.WlsTLsCertCnEnv, "WLS TLS Certificate Common Name")
	if err == nil && tlsCertCN != "" {
		Configuration.Subject.TLSCertCommonName = tlsCertCN
	} else if strings.TrimSpace(Configuration.Subject.TLSCertCommonName) == "" {
		log.Infof("config/config:SaveConfiguration() %s not defined, using default value", constants.WlsTLsCertCnEnv)
		Configuration.Subject.TLSCertCommonName = constants.DefaultWlsTlsCn
	}

	certSANList, err := c.GetenvString(constants.WlsCertSANListEnv, "WLS Certificate SAN List")
	if err == nil && certSANList != "" {
		Configuration.CertSANList = certSANList
	} else if strings.TrimSpace(Configuration.CertSANList) == "" {
		log.Infof("config/config:SaveConfiguration() %s List not defined, using default value", constants.WlsCertSANListEnv)
		Configuration.CertSANList = constants.DefaultWlsTlsSan
	}

	ll, err := c.GetenvString(constants.WlsLoglevelEnv, "Logging Level")
	if err != nil {
		if Configuration.LogLevel == "" {
			log.Infof("config/config:SaveConfiguration() %s not defined, using default log level: %s", constants.WlsLoglevelEnv, logrus.InfoLevel.String())
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

	Configuration.LogEnableStdout = false
	logEnableStdout, err := c.GetenvString(constants.WLSConsoleEnableEnv, "Workload Service enable standard output")
	if err == nil && logEnableStdout != "" {
		Configuration.LogEnableStdout, err = strconv.ParseBool(logEnableStdout)
		if err != nil {
			log.Infof("Error while parsing the variable %s setting to default value false", constants.WLSConsoleEnableEnv)
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
	if utils.IsContainerEnv() {
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
	var err error = nil
	// creating the log file if not preset
	var ioWriterDefault io.Writer
	secLogFile, _ := os.OpenFile(constants.SecurityLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)

	err = os.Chmod(constants.SecurityLogFile, 0640)
	if err != nil {
		log.Errorf("config/config:LogConfiguration() error in setting file permission for file : %s", secLogFile)
	}

	defaultLogFile, _ := os.OpenFile(constants.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0640)
	err = os.Chmod(constants.LogFile, 0640)
	if err != nil {
		log.Errorf("config/config:LogConfiguration() error in setting file permission for file : %s", defaultLogFile)
	}

	err = TakeOwnershipFileWLS(constants.LogDir)
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
