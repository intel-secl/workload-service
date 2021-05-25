/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"flag"
	"fmt"
	csetup "intel/isecl/lib/common/v4/setup"
	"intel/isecl/workload-service/v4/config"
	"intel/isecl/workload-service/v4/constants"
	"strings"

	"github.com/pkg/errors"
)

// HVSConnection is a setup task for setting up the connection to the Host Verification Service (HVS)
type HVSConnection struct {
	Flags []string
}

// Run will run the HVS Connection setup task, but will skip if Validate() returns no errors
func (hvs HVSConnection) Run(c csetup.Context) error {
	log.Trace("setup/hvs:Run() Entering")
	defer log.Trace("setup/hvs:Run() Leaving")

	var err error

	fmt.Println("Running setup task: hvsconnection")

	fs := flag.NewFlagSet("hvsconnection", flag.ExitOnError)
	force := fs.Bool("force", false, "force rerun of HVS config setup")

	err = fs.Parse(hvs.Flags)
	if err != nil {
		fmt.Println("HVS Connection setup: Unable to parse flags")
		return fmt.Errorf("HVS Connection setup: Unable to parse flags")
	}

	if !*force && hvs.Validate(c) == nil {
		fmt.Println("setup hvsconnection: HVS config variables already set, so skipping hvs setup task...")
		log.Info("setup/hvs:Run() HVS config already setup, skipping ...")
		return nil
	}

	fmt.Println("Setting up HVS configuration ...")
	var hvsURL string
	if hvsURL, err = c.GetenvString(constants.HvsUrlEnv, "Host Verification Service URL"); err != nil {
		return errors.Wrap(err, "setup/hvs:Run() Missing HVS Endpoint URL in environment")
	}
	if strings.HasSuffix(hvsURL, "/") {
		config.Configuration.HvsApiUrl = hvsURL
	} else {
		config.Configuration.HvsApiUrl = hvsURL + "/"
	}

	log.Info("setup/hvs:Run() Updated HVS endpoint in configuration")
	return config.Save()
}

// Validate checks whether or not the HVS Connection setup task was completed successfully
func (hvs HVSConnection) Validate(c csetup.Context) error {
	log.Trace("setup/hvs:Validate() Entering")
	defer log.Trace("setup/hvs:Validate() Leaving")
	if config.Configuration.HvsApiUrl == "" {
		return errors.New("setup/hvs:Validate() HVS Connection: URL is not set")
	}
	return nil
}
