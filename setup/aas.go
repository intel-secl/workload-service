/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"flag"
	"fmt"
	"intel/isecl/lib/clients"
	"intel/isecl/lib/common/crypt"
	commLog "intel/isecl/lib/common/log"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"
	consts "intel/isecl/workload-service/constants"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
)

var log = commLog.GetDefaultLogger()
var seclog = commLog.GetSecurityLogger()

// AASConnection is a setup task for setting roles in AAS
type AASConnection struct {
	Flags []string
}

// Run will run the AAS Connection setup task, but will skip if Validate() returns no errors
func (aas AASConnection) Run(c csetup.Context) error {
	log.Trace("setup/aas:Run() Entering")
	defer log.Trace("setup/aas:Run() Leaving")

	var err error

	fmt.Println("Running setup task: aasconnection")

	fs := flag.NewFlagSet("aasconnection", flag.ExitOnError)
	force := fs.Bool("force", false, "force rerun of AAS config setup")

	err = fs.Parse(aas.Flags)
	if err != nil {
		fmt.Println("setup aasconnection: Unable to parse flags")
		return fmt.Errorf("setup aasconnection: Unable to parse flags")
	}

	if aas.Validate(c) == nil && !*force {
		fmt.Println("setup aasconnection: setup task already complete. Skipping...")
		log.Info("setup/aas:Run() AAS configuration config already setup, skipping ...")
		return nil
	}

	var aasURL string
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

	log.Info("setup/aas:Run() aasconnection setup task successful")
	return nil
}

// Validate checks whether or not the AAS Connection setup task was completed successfully
func (aas AASConnection) Validate(c csetup.Context) error {
	log.Trace("setup/aas:Validate() Entering")
	defer log.Trace("setup/aas:Validate() Leaving")

	_, err := os.Stat(consts.TrustedJWTSigningCertsDir)
	if os.IsNotExist(err) {
		return errors.Wrap(err, "setup/aas:Validate() JWT certificate directory does not exist")
	}

	isJWTCertExist := isPathContainPemFile(consts.TrustedJWTSigningCertsDir)

	if !isJWTCertExist {
		return errors.New("setup/aas:Validate() AAS JWT certs not found")
	}

	return nil
}

func isPathContainPemFile(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	// read in ONLY one file
	fname, err := f.Readdir(1)

	// if EOF detected path is empty
	if err != io.EOF && len(fname) > 0 && strings.HasSuffix(fname[0].Name(), ".pem") {
		log.Trace("setup/aas:isPathContainPemFile() fname is ", fname[0].Name())
		_, errs := crypt.GetCertFromPemFile(name + "/" + fname[0].Name())
		if errs == nil {
			log.Trace("setup/aas:isPathContainPemFile() full path valid PEM ", name+"/"+fname[0].Name())
			return true
		}
	}
	return false
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
