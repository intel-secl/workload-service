/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"fmt"
	"intel/isecl/lib/clients"
	aasClient "intel/isecl/lib/clients/aas"
	"intel/isecl/lib/common/crypt"
	commLog "intel/isecl/lib/common/log"
	csetup "intel/isecl/lib/common/setup"
	aasTypes "intel/isecl/lib/common/types/aas"
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

	fmt.Println("Setting up roles in AAS ...")
	var aasURL string
	var aasBearerToken string
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
	if aasBearerToken, err = c.GetenvString(consts.BearerToken, "AAS Bearer Token"); err != nil {
		return errors.Wrap(err, "setup/aas:Run AAS bearer token not set in environment")
	}

	hc, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
	if err != nil {
		return errors.Wrapf(err, "setup/aas:Run() Error setting up HTTP client: %s", err.Error())
	}

	ac := &aasClient.Client{
		BaseURL:    aasURL,
		JWTToken:   []byte(aasBearerToken),
		HTTPClient: hc,
	}

	roles := [3]string{consts.FlavorImageRetrievalGroupName, consts.ReportCreationGroupName, consts.AdministratorGroupName}

	var role_ids []string
	for _, role := range roles {
		roleCreate := aasTypes.RoleCreate{
			RoleInfo: aasTypes.RoleInfo{
				Name:    role,
				Service: consts.ServiceName,
			},
		}
		roleCreateResponse, err := ac.CreateRole(roleCreate)
		if err != nil {
			if strings.Contains(err.Error(), "same role exists") {
				seclog.Debugf("setup/aas:Run() Role %s already exists in AAS. Role creation skipped: %v", role, err)
				log.Tracef("%+v", err)
				continue
			}
			return errors.Wrapf(err, "setup/aas:Run() Error in role %s creation", role)
		}
		log.Debugf("setup/aas:Run() Role %s created in AAS with ID %s ", role, roleCreateResponse.ID)
		log.Infof("setup/aas:Run() Role %s created in AAS ", role)

		role_ids = append(role_ids, roleCreateResponse.ID)
	}

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
