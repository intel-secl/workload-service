/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package vsclient

import (
	"bytes"
	"intel/isecl/lib/clients"
	"intel/isecl/lib/clients/aas"
	commLog "intel/isecl/lib/common/log"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/lib/common/validation"
	config "intel/isecl/workload-service/config"
	consts "intel/isecl/workload-service/constants"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"sync"
	"fmt"
	"net/http"
	"net/url"
)

var log = commLog.GetDefaultLogger()
var aasClient = aas.NewJWTClient(config.Configuration.AAS_API_URL)
var aasRWLock = sync.RWMutex{}

func init() {
	aasRWLock.Lock()
	if aasClient.HTTPClient == nil {
		c, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
		if err != nil {
			return
		}
		aasClient.HTTPClient = c
	}
	aasRWLock.Unlock()
}

func addJWTToken(req *http.Request) error {
	log.Trace("vsclient/client.go:addJWTToken() Entering")
	defer log.Trace("vsclient/client.go:addJWTToken() Leaving")
	if aasClient.BaseURL == "" {
		aasClient = aas.NewJWTClient(config.Configuration.AAS_API_URL)
		if aasClient.HTTPClient == nil {
			c, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
			if err != nil {
				return errors.Wrap(err, "vsclient/client.go:addJWTToken() Error initializing http client")
			}
			aasClient.HTTPClient = c
		}
	}
	aasRWLock.RLock()
	jwtToken, err := aasClient.GetUserToken(config.Configuration.WLS.User)
	aasRWLock.RUnlock()
	// something wrong
	if err != nil {
		// lock aas with w lock
		aasRWLock.Lock()
		// check if other thread fix it already
		jwtToken, err = aasClient.GetUserToken(config.Configuration.WLS.User)
		// it is not fixed
		if err != nil {
			// these operation cannot be done in init() because it is not sure
			// if config.Configuration is loaded at that time
			aasClient.AddUser(config.Configuration.WLS.User, config.Configuration.WLS.Password)
			err = aasClient.FetchAllTokens()
			if err != nil {
				return errors.Wrap(err, "vsclient/client.go:addJWTToken() Could not fetch token")
			}
		}
		aasRWLock.Unlock()
	}
	log.Debug("vsclient/client.go:addJWTToken() successfully added jwt bearer token")
	req.Header.Set("Authorization", "Bearer "+string(jwtToken))
	return nil
}

//SendRequest method is used to create an http client object and send the request to the server
func sendRequest(req *http.Request) ([]byte, error) {
	log.Trace("vsclient/client.go:sendRequest() Entering")
	defer log.Trace("vsclient/client.go:sendRequest() Leaving")

	client, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client.go:sendRequest() Failed to create http client")
	}
	err = addJWTToken(req)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client.go:sendRequest() Failed to add JWT token")
	}

	response, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client.go:sendRequest() Error from response")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized {
		// fetch token and try again
		aasRWLock.Lock()
		aasClient.FetchAllTokens()
		aasRWLock.Unlock()
		err = addJWTToken(req)
		if err != nil {
			return nil, errors.Wrap(err, "vsclient/client.go:sendRequest() Failed to add JWT token")
		}
		response, err = client.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "vsclient/client.go:sendRequest() Error from response")
		}
	}
	if response.StatusCode == http.StatusNotFound {
		return nil, errors.Wrap(err, "vsclient/client.go:sendRequest() HTTP Response: 404 Not found")
	}

	//create byte array of HTTP response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client.go:sendRequest() sendRequest() Error while reading the response body")
	}
	log.Info("vsclient/client.go:sendRequest() Recieved the response successfully")
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

	rsp, err := sendRequest(req)
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

// GetCaCerts method is used to get all the CA certs of HVS
func GetCaCerts(domain string) ([]byte, error) {
	log.Trace("vsclient/client:GetCaCerts() Entering")
	defer log.Trace("vsclient/client:GetCaCerts() Leaving")

	requestURL, err := url.Parse(config.Configuration.HVS_API_URL + "ca-certificates?domain=" + domain)
	if err != nil {
		return nil, errors.Wrap(err,"vsclient/client:GetCaCerts() error forming GET ca-certificates API URL for HVS")
	}

	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err,"vsclient/client:GetCaCerts() Error while forming a new http request from client to server")
	}

	req.Header.Set("Accept", "application/x-pem-file")
	req.Header.Set("Content-Type", "application/json")
	var c csetup.Context
	jwtToken, err := c.GetenvString(consts.BearerToken, "BEARER_TOKEN")
	if jwtToken == "" || err != nil {
		fmt.Fprintln(os.Stderr, "BEARER_TOKEN is not defined in environment")
		return nil, errors.Wrap(err, "BEARER_TOKEN is not defined in environment")
	}
	req.Header.Set("Authorization", "Bearer "+ jwtToken)
	client, err := clients.HTTPClientWithCADir(consts.TrustedCaCertsDir)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client:GetCaCerts() Failed to create http client")
	}
	rsp, err := client.Do(req)

	if err != nil {
		log.Error("vsclient/client:GetCaCerts() Error while sending request from client to server")
		log.Tracef("%+v", err)
		return nil, errors.Wrap(err,"vsclient/client:GetCaCerts() Error while sending request from client to server")
	}

	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "vsclient/client:GetCaCerts() Error while reading response body")
	}

	return body, nil
}
