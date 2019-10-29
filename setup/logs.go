/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"fmt"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"

	"github.com/sirupsen/logrus"
)

type Logs struct{}

// Run will configure the parameters for the WLS web service layer. This will be skipped if Validate() returns no errors
func (l Logs) Run(c csetup.Context) error {
	log.Trace("setup/logs:Run() Entering")
	defer log.Trace("setup/logs:Run() Leaving")
	fmt.Println("Setting up webserver ...")
	var err error
	ll, err := c.GetenvString(config.WLS_LOGLEVEL, "Logging Level")
	if err != nil {
		fmt.Println("No logging level specified, using default logging level: Error")
		config.Configuration.LogLevel = logrus.ErrorLevel
	}
	config.Configuration.LogLevel, err = logrus.ParseLevel(ll)
	if err != nil {
		fmt.Println("Invalid logging level specified, using default logging level: Error")
		config.Configuration.LogLevel = logrus.ErrorLevel
	}
	return config.Save()
}

// Validate checks whether or not the Server task configured successfully or not
func (l Logs) Validate(c csetup.Context) error {
	log.Trace("setup/logs:Validate() Entering")
	defer log.Trace("setup/logs:Validate() Leaving")
	return nil
}
