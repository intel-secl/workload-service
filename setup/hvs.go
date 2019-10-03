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
	"strings"
)

// HVSConnection is a setup task for setting up the connection to the Host Verification Service (HVS)
type HVSConnection struct{}

// Run will run the HVS Connection setup task, but will skip if Validate() returns no errors
func (hvs HVSConnection) Run(c csetup.Context) error {
	if hvs.Validate(c) == nil {
		fmt.Println("HVS connection already setup, skipping ...")
		return nil
	}
	fmt.Println("Setting up HVS configuration ...")
	var err error
	var hvsURL string
	if hvsURL, err = c.GetenvString(config.HVS_URL, "Host Verification Service URL"); err != nil {
		return err
	}
	if strings.HasSuffix(hvsURL, "/") {
		config.Configuration.HVS_API_URL = hvsURL
	} else {
		config.Configuration.HVS_API_URL = hvsURL + "/"
	}
	return config.Save()
}

// Validate checks whether or not the HVS Connection setup task was completed successfully
func (hvs HVSConnection) Validate(c csetup.Context) error {
	if config.Configuration.HVS_API_URL == "" {
		return errors.New("HVS Connection: URL is not set")
	}
	return nil
}
