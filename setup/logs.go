/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"fmt"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"

	log "github.com/sirupsen/logrus"
)

type Logs struct{}

// Run will configure the parameters for the WLS web service layer. This will be skipped if Validate() returns no errors
func (l Logs) Run(c csetup.Context) error {
	fmt.Println("Setting up webserver ...")
	var err error
	ll, err := c.GetenvString(config.WLS_LOGLEVEL, "Logging Level")
	if err != nil {
		fmt.Println("No logging level specified, using default logging level: Error")
		config.Configuration.LogLevel = log.ErrorLevel
	}
	config.Configuration.LogLevel, err = log.ParseLevel(ll)
	if err != nil {
		fmt.Println("Invalid logging level specified, using default logging level: Error")
		config.Configuration.LogLevel = log.ErrorLevel
	}
	return config.Save()
}

// Validate checks whether or not the Server task configured successfully or not
func (l Logs) Validate(c csetup.Context) error {
	return nil
}
