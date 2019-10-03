/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package vsclient

import (
	"bytes"
	"intel/isecl/lib/clients"
	"intel/isecl/lib/clients/aas"
	config "intel/isecl/workload-service/config"
	consts "intel/isecl/workload-service/constants"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"fmt"
	"net/http"
	"net/url"
)

var aasClient = aas.NewJWTClient(config.Configuration.AAS_API_URL)

func init() {
	if aasClient.HTTPClient == nil {
		c, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
		if err != nil {
			return
		}
		aasClient.HTTPClient = c
	}
}

func addJWTToken(req *http.Request) error {
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

	client, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
	if err != nil {
		log.Error("Error while verifying root ca certificate")
		return nil, err
	}
	err = addJWTToken(req)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	
	if response.StatusCode == http.StatusNotFound {
		return nil, err
	}
	//create byte array of HTTP response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func CreateSAMLReport(hwid string) ([]byte, error) {

	criteriaJSON := []byte(fmt.Sprintf(`{"hardware_uuid":"%s"}`, hwid))
	url, err := url.Parse(config.Configuration.HVS_API_URL)
	if err != nil {
		log.Error("Configured HVS URL is malformed: ", err)
		return nil, err
	}
	reports, _ := url.Parse("reports")
	endpoint := url.ResolveReference(reports)
	req, err := http.NewRequest("POST", endpoint.String(), bytes.NewBuffer(criteriaJSON))

	if err != nil {
		log.Error("Failed to instantiate http request to HVS")
		return nil, err
	}

	req.Header.Set("Accept", "application/samlassertion+xml")
	req.Header.Set("Content-Type", "application/json")

	rsp, err := sendRequest(req, true)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}
