/*
 * Copyright (C) 2020 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package resource

import (
	"encoding/xml"
	"github.com/google/uuid"
	"github.com/intel-secl/intel-secl/v3/pkg/clients/hvsclient"
	"github.com/intel-secl/intel-secl/v3/pkg/clients/kbs"
	"github.com/intel-secl/intel-secl/v3/pkg/lib/common/crypt"
	samlVerifier "github.com/intel-secl/intel-secl/v3/pkg/lib/saml"
	"github.com/intel-secl/intel-secl/v3/pkg/model/hvs"
	"intel/isecl/lib/common/v3/log/message"
	"intel/isecl/lib/common/v3/validation"
	"intel/isecl/workload-service/v3/config"
	"intel/isecl/workload-service/v3/constants"
	consts "intel/isecl/workload-service/v3/constants"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Verifies host and retrieves key from KMS
// getFlavor is true for the images API and false for the keys API
// id is only required when using the images API
func transfer_key(getFlavor bool, hwid string, kUrl string, id string) ([]byte, error) {
	var endpoint, funcName, retrievalErr string
	if getFlavor {
		endpoint = "resource/images"
		funcName = "retrieveFlavorandKeyForImageID()"
		retrievalErr = "Failed to retrieve Flavor/Key for Image"
	} else {
		endpoint = "resource/keys"
		funcName = "retrieveKey()"
		retrievalErr = "Failed to retrieve Key for Image"
	}
	// we have key URL
	// http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer"
	// post HVS with hardwareUUID
	// extract key_id from KeyUrl
	cLog := log.WithField("hardwareUUID", hwid).WithField("keyUrl", kUrl)
	if getFlavor {
		cLog = cLog.WithField("id", id)
	}
	cLog.Debugf("%s:%s KeyUrl is present", endpoint, funcName)
	keyUrl, err := url.Parse(kUrl)
	if err != nil {
		cLog.WithError(err).Errorf("%s:%s %s : KeyUrl is malformed", endpoint, funcName, message.InvalidInputProtocolViolation)
		log.Tracef("%+v", err)
		return nil, &endpointError{
			Message:    retrievalErr + " - KeyUrl is malformed",
			StatusCode: http.StatusBadRequest,
		}
	}
	re := regexp.MustCompile("(?i)([0-9A-F]{8}-[0-9A-F]{4}-4[0-9A-F]{3}-[89AB][0-9A-F]{3}-[0-9A-F]{12})")
	keyID := re.FindString(keyUrl.Path)

	// retrieve host SAML report from HVS
	vsClientFactory, err := hvsclient.NewVSClientFactoryWithUserCredentials(config.Configuration.HVS_API_URL, config.Configuration.AAS_API_URL, config.Configuration.WLS.User, config.Configuration.WLS.Password, constants.TrustedCaCertsDir)
	if err != nil {
		cLog.WithError(err).Error("Error while instantiating VSClientFactory")
		return nil, &endpointError{
			Message:    "Error while instantiating VSClientFactory",
			StatusCode: http.StatusInternalServerError,
		}
	}

	reportsClient, err := vsClientFactory.ReportsClient()
	if err != nil {
		cLog.WithError(err).Error("Error while instantiating ReportsClient")
		return nil, &endpointError{
			Message:    "Error while instantiating ReportsClient",
			StatusCode: http.StatusInternalServerError,
		}
	}
	reportCreateRequest := hvs.ReportCreateRequest{
		HardwareUUID: uuid.MustParse(hwid),
	}
	saml, err := reportsClient.CreateSAMLReport(reportCreateRequest)
	if err != nil {
		cLog.WithError(err).Errorf("%s:%s %s : Failed to read HVS response body", endpoint, funcName, message.BadConnection)
		log.Tracef("%+v", err)
		return nil, &endpointError{
			Message:    retrievalErr + " - Failed to read HVS response",
			StatusCode: http.StatusInternalServerError,
		}
	}

	// validate the response from HVS
	if err = validation.ValidateXMLString(string(saml)); err != nil {
		cLog.WithError(err).Errorf("%s:%s %s : HVS response validation failed", endpoint, funcName, message.AppRuntimeErr)
		return nil, &endpointError{
			Message:    retrievalErr + " - Invalid SAML report format received from HVS",
			StatusCode: http.StatusInternalServerError,
		}
	}

	var samlStruct Saml
	cLog.WithField("saml", string(saml)).Debugf("%s:%s Successfully got SAML report from HVS", endpoint, funcName)
	err = xml.Unmarshal(saml, &samlStruct)
	if err != nil {
		cLog.WithError(err).Errorf("%s:%s %s : Failed to unmarshal host SAML report", endpoint, funcName, message.AppRuntimeErr)
		log.Tracef("%+v", err)
		return nil, &endpointError{
			Message:    retrievalErr + " - Failed to unmarshal host SAML report",
			StatusCode: http.StatusInternalServerError,
		}
	}

	// verify saml cert chain
	verified := samlVerifier.VerifySamlSignature(string(saml), constants.SamlCaCertFilePath, constants.TrustedCaCertsDir)
	if !verified {
		cLog.WithError(err).Errorf("%s:%s SAML certificate chain verification failed", endpoint, funcName)
		return nil, &endpointError{
			Message:    retrievalErr + " - SAML signature or certificate chain verification failed",
			StatusCode: http.StatusInternalServerError,
		}
	}

	var key []byte
	for i := 0; i < len(samlStruct.Attribute); i++ {
		if samlStruct.Attribute[i].Name == "TRUST_OVERALL" {
                        if samlStruct.Attribute[i].AttributeValue == "false"{
				return nil, &endpointError{
                        		Message:    retrievalErr + " - Host is untrusted",
		                        StatusCode: http.StatusInternalServerError,
                		}
			}
			// check if the key is cached and retrieve it
			// try to obtain the key from the cache. If the key is not found in the cache,
			// then it will return and error. In this case, we ignore it and pro

			var cachedKeyID string
			cachedKey, err := getKeyFromCache(hwid)
			if err == nil {
				cachedKeyID = cachedKey.ID
				cLog.Infof("%s:%s %s : Retrieved Key from in-memory cache. key ID: %s", endpoint, funcName, message.EncKeyUsed, cachedKey.ID)
			}
			// check if the key cached is same as the one in the flavor
			if cachedKeyID != "" && cachedKeyID == keyID {
				key = cachedKey.Bytes
			} else {
				//Load trusted CA certificates
				caCerts, err := crypt.GetCertsFromDir(consts.TrustedCaCertsDir)
				if err != nil {
					cLog.WithError(err).Errorf("%s:%s %s : Failed to load CA certificates", endpoint, funcName, message.AppRuntimeErr)
					return nil, &endpointError{
						Message:    retrievalErr + " - Unable to load CA certificates",
						StatusCode: http.StatusInternalServerError,
					}
				}

				baseUrl := strings.TrimSuffix(re.Split(kUrl, 2)[0], "/keys/")
				kbsUrl, _ := url.Parse(baseUrl)
				//Initialize the KBS client
				kc := kbs.NewKBSClient(nil, kbsUrl, "", "", caCerts)

				// post to KBS client with saml
				cLog.Infof("%s:%s baseURL: %s, keyID: %s : start to retrieve key from KMS", endpoint, funcName, baseUrl, keyID)
				key, err = kc.TransferKeyWithSaml(keyID, string(saml))
				if err != nil {
					cLog.WithError(err).Errorf("%s:%s %s : Failed to retrieve key from KMS", endpoint, funcName, message.AppRuntimeErr)
					return nil, &endpointError{
						Message:    "Failed to retrieve key ",
						StatusCode: http.StatusInternalServerError,
					}
				}
				cLog.Infof("%s:%s Successfully got key from KMS", endpoint, funcName)
				cacheKeyInMemory(hwid, keyID, key)
			}
		}
	}
	return key, nil
}
