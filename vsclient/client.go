/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package vsclient

import (
	"bytes"
	"github.com/pkg/errors"
	"intel/isecl/lib/clients"
	"intel/isecl/lib/clients/aas"
	commLog "intel/isecl/lib/common/log"
	"intel/isecl/lib/common/validation"
	config "intel/isecl/workload-service/config"
	consts "intel/isecl/workload-service/constants"
	"io/ioutil"

	"fmt"
	"net/http"
	"net/url"
)

var log = commLog.GetDefaultLogger()
var aasClient = aas.NewJWTClient(config.Configuration.AAS_API_URL)

func init() {
	log.Trace("vsclient/client:init() Entering")
	defer log.Trace("vsclient/client:init() Leaving")

	if aasClient.HTTPClient == nil {
		c, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
		if err != nil {
			return
		}
		aasClient.HTTPClient = c
	}
}

func addJWTToken(req *http.Request) error {
	log.Trace("vsclient/client:addJWTToken() Entering")
	defer log.Trace("vsclient/client:addJWTToken() Leaving")

	jwtToken, err := aasClient.GetUserToken(config.Configuration.WLS.User)
	if err != nil {
		aasClient.AddUser(config.Configuration.WLS.User, config.Configuration.WLS.Password)
		aasClient.FetchTokenForUser(config.Configuration.WLS.User)
		jwtToken, err = aasClient.GetUserToken(config.Configuration.WLS.User)
	}
	req.Header.Set("Authorization", "Bearer "+string(jwtToken))
	return nil
}

//SendRequest method is used to create an http client object and send the request to the server
func sendRequest(req *http.Request, insecureConnection bool) ([]byte, error) {
	log.Trace("vsclient/client:sendRequest() Entering")
	defer log.Trace("vsclient/client:sendRequest() Leaving")

	client, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client:sendRequest() Error while verifying root ca certificate")
	}
	err = addJWTToken(req)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client:sendRequest() Failed to add JWT token to request header")
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client:sendRequest() Client failed to connect to server")
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, errors.Wrap(err, "vsclient/client:sendRequest() HTTP Response: 404 Not found")
	}
	//create byte array of HTTP response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client:sendRequest() Error while reading the response body")
	}
	return body, nil
}

func CreateSAMLReport(hwid string) ([]byte, error) {
	log.Trace("vsclient/client:CreateSAMLReport() Entering")
	defer log.Trace("vsclient/client:CreateSAMLReport() Leaving")

	criteriaJSON := []byte(fmt.Sprintf(`{"hardware_uuid":"%s"}`, hwid))
	url, err := url.Parse(config.Configuration.HVS_API_URL)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client:CreateSAMLReport() Configured HVS URL is malformed")
	}
	reports, _ := url.Parse("reports")
	endpoint := url.ResolveReference(reports)
	req, err := http.NewRequest("POST", endpoint.String(), bytes.NewBuffer(criteriaJSON))

	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client:CreateSAMLReport() Failed to instantiate http request to HVS")
	}

	req.Header.Set("Accept", "application/samlassertion+xml")
	req.Header.Set("Content-Type", "application/json")

	rsp, err := sendRequest(req, true)
	if err != nil {
		log.Error("vsclient/client:CreateSAMLReport() Error while sending request from client to server")
		log.Tracef("%+v", err)
		return nil, err
	}

	// now validate SAML
	err = validation.ValidateXMLString(string(rsp))
	if err != nil {
		return nil, err
	}

	return rsp, nil
}
