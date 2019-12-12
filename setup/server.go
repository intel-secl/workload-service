/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"flag"
	"fmt"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"

	"github.com/pkg/errors"
)

type Server struct {
	Flags []string
}

// Run will configure the parameters for the WLS web service layer. This will be skipped if Validate() returns no errors
func (ss Server) Run(c csetup.Context) error {
	log.Trace("setup/server:Run() Entering")
	defer log.Trace("setup/server:Run() Leaving")

	var err error

	fmt.Println("Running setup task: server")

	fs := flag.NewFlagSet("server", flag.ExitOnError)
	force := fs.Bool("force", false, "force re-run of server setup task")

	err = fs.Parse(ss.Flags)
	if err != nil {
		fmt.Println("WLS Server setup: Unable to parse flags")
		return fmt.Errorf("WLS Server setup: Unable to parse flags")
	}

	if !*force && ss.Validate(c) == nil {
		fmt.Println("setup server: setup task already complete. Skipping...")
		log.Info("setup/server:Run() WLS Server setup already complete, skipping ...")
		return nil
	}

	log.Info("setup/server:Run() Setting up webserver ...")

	config.Configuration.Port, err = c.GetenvInt(config.WLS_PORT, "Webserver Port")
	if err != nil {
		log.Info("setup/server:Run() Listen port not specified.Using default webserver port: 5000")
		config.Configuration.Port = 5000
	}
	if config.Configuration.WLS.User, err = c.GetenvString(config.WLS_USER, "Workload Service User"); err != nil {
		return err
	}
	if config.Configuration.WLS.Password, err = c.GetenvSecret(config.WLS_PASSWORD, "Workload Service Password"); err != nil {
		return err
	}

	log.Info("setup/server:Run() Updated WLS user credentials and server port in configuration")
	return nil
}

// Validate checks whether or not the Server task configured successfully or not
func (ss Server) Validate(c csetup.Context) error {
	log.Trace("setup/server:Validate() Entering")
	defer log.Trace("setup/server:Validate() Leaving")
	// validate that the port variable is not the zero value of its type
	if config.Configuration.Port == 0 {
		return errors.New("setup/server:Validate() Server: Port is not set")
	}
	wls := &config.Configuration.WLS
	if wls.User == "" {
		return errors.New("setup/server:Validate() WLS User is not set")
	}
	if wls.Password == "" {
		return errors.New("setup/server:Validate() WLS Password is not set ")
	}
	return nil
}
