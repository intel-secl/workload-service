/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"flag"
	"fmt"
	"github.com/intel-secl/intel-secl/v3/pkg/clients/hvsclient"
	"github.com/pkg/errors"
	csetup "intel/isecl/lib/common/v2/setup"
	"intel/isecl/workload-service/v2/config"
	"intel/isecl/workload-service/v2/constants"
	"io/ioutil"
	"os"
)

type Download_Saml_Ca_Cert struct {
	Flags []string
}

func (dc Download_Saml_Ca_Cert) Run(c csetup.Context) error {
	log.Trace("setup/download_saml_ca_cert:Run() Entering")
	defer log.Trace("setup/download_saml_ca_cert:Run() Leaving")

	fs := flag.NewFlagSet("Download_Saml_Ca_Cert", flag.ExitOnError)
	force := fs.Bool("force", false, "force rerun of WLS setup to download SAML CA certificates from HVS")

	err := fs.Parse(dc.Flags)
	if err != nil {
		fmt.Println("setup/download_saml_ca_cert: Unable to parse flags")
		return fmt.Errorf("setup/download_saml_ca_cert: Unable to parse flags")
	}

	if dc.Validate(c) == nil && !*force {
		fmt.Println("setup/download_saml_ca_cert:Run() SAML CA certificates are already downloaded by WLS. Skipping setup task execution...")
		log.Info("setup/download_saml_ca_cert:Run() SAML CA certificates are already downloaded by WLS. Skipping setup task execution...")
		return nil
	}
	log.Info("setup/download_saml_ca_cert:Run() Downloading SAML CA certificates.")
	jwtToken, err := c.GetenvSecret(constants.BearerToken, "BEARER_TOKEN")
	if jwtToken == "" || err != nil {
		fmt.Fprintln(os.Stderr, "BEARER_TOKEN is not defined in environment")
		return errors.Wrap(err, "BEARER_TOKEN is not defined in environment")
	}

	vsClientFactory, err := hvsclient.NewVSClientFactory(config.Configuration.HVS_API_URL, jwtToken, constants.TrustedCaCertsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr,"setup/download_saml_ca_cert:Run() Error while instantiating VSClientFactory")
		return errors.Wrap(err, "setup/download_saml_ca_cert:Run() Error while instantiating VSClientFactory")
	}

	caCertsClient, err := vsClientFactory.CACertificatesClient()
	if err != nil {
		fmt.Fprintln(os.Stderr,"setup/download_saml_ca_cert:Run() Error while getting CACertificatesClient")
		return errors.Wrap(err, "setup/download_saml_ca_cert:Run() Error while getting CACertificatesClient")
	}

	cacerts, err := caCertsClient.GetCaCertsInPem("saml")
	if err != nil {
		log.Error("setup/download_saml_ca_cert:Run() Failed to read HVS response body for GET SAML ca-certificates API")
		return errors.Wrap(err, "setup/download_saml_ca_cert:Run() Error while getting SAML CA certificates")
	}

	//write the output to a file
	err = ioutil.WriteFile(constants.SamlCaCertFilePath, cacerts, 0644)
	if err != nil {
		return errors.Wrapf(err, "setup/download_saml_ca_cert:Run() Error while writing file:%s", constants.SamlCaCertFilePath)
	}
	return nil
}

func (dc Download_Saml_Ca_Cert) Validate(c csetup.Context) error {
	log.Trace("setup/download_saml_ca_cert:Validate() Entering")
	defer log.Trace("setup/download_saml_ca_cert:Validate() Leaving")

	log.Info("setup/download_saml_ca_cert:Validate() Validation for downloading SAML CA certificates from HVS.")

	if _, err := os.Stat(constants.SamlCaCertFilePath); os.IsNotExist(err) {
		return errors.Wrap(err, "setup/download_saml_ca_cert:Validate() Error while validating download_saml_ca_cert setup task")
	}

	return nil
}
