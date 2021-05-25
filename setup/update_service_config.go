/*
 * Copyright (C) 2021 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	commLog "intel/isecl/lib/common/v4/log"
	csetup "intel/isecl/lib/common/v4/setup"
	"intel/isecl/workload-service/v4/config"
	"intel/isecl/workload-service/v4/constants"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Update_Service_Config struct {
	Flags []string
}

var log = commLog.GetDefaultLogger()
var seclog = commLog.GetSecurityLogger()

// Run will configure the parameters for the WLS web service layer. This will be skipped if Validate() returns no errors
func (uc Update_Service_Config) Run(c csetup.Context) error {
	log.Trace("setup/update_service_config:Run() Entering")
	defer log.Trace("setup/update_service_config:Run() Leaving")

	fmt.Println("Running setup task: update_service_config")

	fs := flag.NewFlagSet("server", flag.ExitOnError)
	force := fs.Bool("force", false, "force re-run of update_service_config setup task")

	err := fs.Parse(uc.Flags)
	if err != nil {
		fmt.Println("setup/update_service_config:Run() Unable to parse flags")
		return fmt.Errorf("setup/update_service_config:Run() Unable to parse flags")
	}

	if !*force && uc.Validate(c) == nil {
		fmt.Println("WLS Update_Service_Config config variables already set, so skipping server setup task...")
		log.Info("setup/update_service_config:Run() WLS Update_Service_Config setup already complete, skipping ...")
		return nil
	}

	config.Configuration.Port, err = c.GetenvInt(constants.WlsPortEnv, "Webserver Port")
	if err != nil {
		log.Info("setup/update_service_config:Run() Listen port not specified.Using default webserver port: 5000")
		config.Configuration.Port = 5000
	}
	if config.Configuration.WLS.User, err = c.GetenvString(constants.WlsUserEnv, "Workload Service User"); err != nil {
		return err
	}
	if config.Configuration.WLS.Password, err = c.GetenvSecret(constants.WlsPasswordEnv, "Workload Service Password"); err != nil {
		return err
	}

	keyCacheSeconds, err := c.GetenvString(constants.KeyCacheSecondsEnv, "Key Cache Seconds")
	if err == nil && keyCacheSeconds != "" {
		config.Configuration.KeyCacheSeconds, _ = strconv.Atoi(keyCacheSeconds)
	} else if config.Configuration.KeyCacheSeconds <= 0 {
		log.Infof("setup/update_service_config:Run() %s not defined, using default value", constants.KeyCacheSecondsEnv)
		config.Configuration.KeyCacheSeconds = constants.DefaultKeyCacheSeconds
	}

	ll, err := c.GetenvString(constants.WlsLoglevelEnv, "Logging Level")
	if err != nil {
		if config.Configuration.LogLevel == "" {
			log.Infof("setup/update_service_config:Run() %s not defined, using default log level: %s", constants.WlsLoglevelEnv, logrus.InfoLevel.String())
			config.Configuration.LogLevel = logrus.InfoLevel.String()
		}
	} else {
		llp, err := logrus.ParseLevel(ll)
		if err != nil {
			log.Infof("setup/update_service_config:Run() Invalid log level specified in env, using default log level: %s", logrus.InfoLevel.String())
			config.Configuration.LogLevel = logrus.InfoLevel.String()
		} else {
			config.Configuration.LogLevel = llp.String()
			log.Infof("setup/update_service_config:Run() Log level set %s\n", ll)
		}
	}

	logEntryMaxLength, err := c.GetenvInt(constants.LogEntryMaxlengthEnv, "Maximum length of each entry in a log")
	if err == nil && logEntryMaxLength >= 300 {
		config.Configuration.LogEntryMaxLength = logEntryMaxLength
	} else {
		log.Infof("setup/update_service_config:Run() Invalid Log Entry Max Length defined (should be >=  %d ), using default value: %d", constants.DefaultLogEntryMaxlength, constants.DefaultLogEntryMaxlength)
		config.Configuration.LogEntryMaxLength = constants.DefaultLogEntryMaxlength
	}

	readTimeout, err := c.GetenvInt(constants.WlsServerReadTimeoutEnv, "Workload Service Read Timeout")
	if err != nil {
		config.Configuration.ReadTimeout = constants.DefaultReadTimeout
	} else {
		config.Configuration.ReadTimeout = time.Duration(readTimeout) * time.Second
	}

	readHeaderTimeout, err := c.GetenvInt(constants.WlsServerReadHeaderTImeoutEnv, "Workload Service Read Header Timeout")
	if err != nil {
		config.Configuration.ReadHeaderTimeout = constants.DefaultReadHeaderTimeout
	} else {
		config.Configuration.ReadHeaderTimeout = time.Duration(readHeaderTimeout) * time.Second
	}

	writeTimeout, err := c.GetenvInt(constants.WlsServerWriteTimeoutEnv, "Workload Service Write Timeout")
	if err != nil {
		config.Configuration.WriteTimeout = constants.DefaultWriteTimeout
	} else {
		config.Configuration.WriteTimeout = time.Duration(writeTimeout) * time.Second
	}

	idleTimeout, err := c.GetenvInt(constants.WlsServerIdleTimeoutEnv, "Workload Service Idle Timeout")
	if err != nil {
		config.Configuration.IdleTimeout = constants.DefaultIdleTimeout
	} else {
		config.Configuration.IdleTimeout = time.Duration(idleTimeout) * time.Second
	}

	maxHeaderBytes, err := c.GetenvInt(constants.WlsServerMaxHeaderBytesEnv, "Workload Service Max Header Bytes Timeout")
	if err != nil {
		config.Configuration.MaxHeaderBytes = constants.DefaultMaxHeaderBytes
	} else {
		config.Configuration.MaxHeaderBytes = maxHeaderBytes
	}

	logEnableStdout, err := c.GetenvString(constants.WLSConsoleEnableEnv, "Workload Service enable standard output")
	if err == nil && logEnableStdout != "" {
		config.Configuration.LogEnableStdout, err = strconv.ParseBool(logEnableStdout)
		if err != nil {
			log.Infof("setup/update_service_config:Run() Error while parsing the variable %s setting to default value false", constants.WLSConsoleEnableEnv)
		}
	}

	return config.Save()
}

// Validate checks whether or not the Update_Service_Config task configured successfully or not
func (uc Update_Service_Config) Validate(c csetup.Context) error {
	log.Trace("setup/update_service_config:Validate() Entering")
	defer log.Trace("setup/update_service_config:Validate() Leaving")
	// validate that the port variable is not the zero value of its type
	if config.Configuration.Port == 0 {
		return errors.Errorf("setup/update_service_config:Validate() Update_Service_Config: %s is not set", constants.WlsPortEnv)
	}
	wls := &config.Configuration.WLS
	if wls.User == "" {
		return errors.Errorf("setup/update_service_config:Validate() %s is not set", constants.WlsUserEnv)
	}
	if wls.Password == "" {
		return errors.Errorf("setup/update_service_config:Validate() %s is not set ", constants.WlsPasswordEnv)
	}
	return nil
}
