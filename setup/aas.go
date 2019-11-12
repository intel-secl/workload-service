/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"intel/isecl/lib/clients"
	"intel/isecl/lib/common/crypt"
	commLog "intel/isecl/lib/common/log"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"
	consts "intel/isecl/workload-service/constants"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

var log = commLog.GetDefaultLogger()
var seclog = commLog.GetSecurityLogger()

// AASConnection is a setup task for setting roles in AAS
type AASConnection struct{}

// Run will run the HVS Connection setup task, but will skip if Validate() returns no errors
func (aas AASConnection) Run(c csetup.Context) error {
	log.Trace("setup/aas:Run() Entering")
	defer log.Trace("setup/aas:Run() Leaving")

	var aasURL string
	var err error
	if aasURL, err = c.GetenvString(config.AAS_API_URL, "AAS Server URL"); err != nil {
		return errors.Wrap(err, "setup/aas:Run AAS endpoint not set in environment")
	}

	if strings.HasSuffix(aasURL, "/") {
		config.Configuration.AAS_API_URL = aasURL
	} else {
		config.Configuration.AAS_API_URL = aasURL + "/"
	}

	config.Save()
	log.Info("setup/aas:Run() AAS endpoint updated")

	//Fetch JWT Certificate from AAS
	err = fnGetJwtCerts()
	if err != nil {
		log.Tracef("%+v", err)
		return errors.Wrap(err, "Failed to fetch JWT Auth Certs")
	}

	return nil
}

// Validate checks whether or not the AAS Connection setup task was completed successfully
func (aas AASConnection) Validate(c csetup.Context) error {
	log.Trace("setup/aas:Validate() Entering")
	defer log.Trace("setup/aas:Validate() Leaving")
	return nil
}

func fnGetJwtCerts() error {
	log.Trace("setup/aas:fnGetJwtCerts() Entering")
	defer log.Trace("setup/aas:fnGetJwtCerts() Leaving")
	url := config.Configuration.AAS_API_URL + "noauth/jwt-certificates"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/x-pem-file")
	seclog.Debugf("setup/aas:fnGetJwtCerts() Connecting to AAS Endpoint %s", url)

	hc, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
	if err != nil {
		return errors.Wrapf(err, "setup/aas:fnGetJwtCerts() Error setting up HTTP client: %s", err.Error())
	}

	res, err := hc.Do(req)
	if err != nil {
		return errors.Wrap(err, "setup/aas:fnGetJwtCerts() Could not retrieve jwt certificate")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.Wrap(err, "setup/aas:fnGetJwtCerts() Error while reading response body")
	}

	err = crypt.SavePemCertWithShortSha1FileName(body, consts.TrustedJWTSigningCertsDir)
	if err != nil {
		return errors.Wrap(err, "setup/aas:fnGetJwtCerts() Error in certificate setup")
	}

	return nil
}
