/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"fmt"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"
	"strings"

	"github.com/pkg/errors"
)

// HVSConnection is a setup task for setting up the connection to the Host Verification Service (HVS)
type HVSConnection struct{}

// Run will run the HVS Connection setup task, but will skip if Validate() returns no errors
func (hvs HVSConnection) Run(c csetup.Context) error {
	log.Trace("setup/hvs:Run() Entering")
	defer log.Trace("setup/hvs:Run() Leaving")
	if hvs.Validate(c) == nil {
		log.Info("setup/hvs:Run() HVS connection already setup, skipping ...")
		return nil
	}
	fmt.Println("Setting up HVS configuration ...")
	var err error
	var hvsURL string
	if hvsURL, err = c.GetenvString(config.HVS_URL, "Host Verification Service URL"); err != nil {
		return errors.Wrap(err, "setup/hvs:Run() Missing HVS Endpoint URL in environment")
	}
	if strings.HasSuffix(hvsURL, "/") {
		config.Configuration.HVS_API_URL = hvsURL
	} else {
		config.Configuration.HVS_API_URL = hvsURL + "/"
	}
	log.Info("setup/hvs:Run() Updated HVS endpoint in configuration")
	return config.Save()
}

// Validate checks whether or not the HVS Connection setup task was completed successfully
func (hvs HVSConnection) Validate(c csetup.Context) error {
	log.Trace("setup/hvs:Validate() Entering")
	defer log.Trace("setup/hvs:Validate() Leaving")
	if config.Configuration.HVS_API_URL == "" {
		return errors.New("setup/hvs:Validate() HVS Connection: URL is not set")
	}
	return nil
}
