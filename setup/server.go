/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"errors"
	"fmt"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"
)

type Server struct{}

// Run will configure the parameters for the WLS web service layer. This will be skipped if Validate() returns no errors
func (ss Server) Run(c csetup.Context) error {
	if ss.Validate(c) == nil {
		fmt.Println("Webserver already setup, skipping ...")
		return nil
	}
	fmt.Println("Setting up webserver ...")
	var err error
	config.Configuration.Port, err = c.GetenvInt(config.WLS_PORT, "Webserver Port")
	if err != nil {
		fmt.Println("Using default webserver port: 5000")
		config.Configuration.Port = 5000
	}
	if config.Configuration.WLS.User, err = c.GetenvString(config.WLS_USER, "Workload Service User"); err != nil {
		return err
	}
	if config.Configuration.WLS.Password, err = c.GetenvSecret(config.WLS_PASSWORD, "Workload Service Password"); err != nil {
		return err
	}
	return config.Save()
}

// Validate checks whether or not the Server task configured successfully or not
func (ss Server) Validate(c csetup.Context) error {
	// validate that the port variable is not the zero value of its type
	if config.Configuration.Port == 0 {
		return errors.New("Server: Port is not set")
	}
	wls := &config.Configuration.WLS
	if wls.User == "" {
		return errors.New("WLS User is not set")
	}
	if wls.Password == "" {
		return errors.New("WLS Password is not set ")
	}
	return nil
}
