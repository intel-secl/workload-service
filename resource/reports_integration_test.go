// +build integration

/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package resource

import (
	"bytes"
	"crypto"
	"encoding/json"
	"fmt"
	"intel/isecl/lib/common/v3/crypt"
	"intel/isecl/lib/common/v3/middleware"
	"intel/isecl/lib/common/v3/pkg/instance"
	"intel/isecl/lib/flavor/v3"
	flavorUtil "intel/isecl/lib/flavor/v3/util"
	"intel/isecl/lib/verifier/v3"
	"intel/isecl/workload-service/v3/model"
	"intel/isecl/workload-service/v3/repository/postgres"
	"github.com/intel-secl/intel-secl/v3/pkg/model/wls"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"io/ioutil"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	// Import Postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func TestReportResource(t *testing.T) {
	log.Trace("resource/reports_integration_test:TestReportResource() Entering")
	defer log.Trace("resource/reports_integration_test:TestReportResource() Leaving")
	var signedFlavor wls.SignedImageFlavor
	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	_, ci := os.LookupEnv("CI")
	var host string
	if ci {
		host = "postgres"
	} else {
		host = "localhost"
	}

	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=5432 user=runner dbname=wls password=test sslmode=disable", host))
	checkErr(err)
	wlsDB := postgres.PostgresDatabase{DB: db.Debug()}
	currDir, _ := os.Getwd()
	wlsDB.Migrate()
	flavor, err := flavor.GetImageFlavor("Cirros-enc", true,
		"https://kbs.server.com:9443/v1/keys/4377ae27-5b48-4301-9684-3a5f39123511/transfer", "Flt8imDt7sqtRAwcBAFzFxz8j0/E1xMulfEOYlu7FesUt/oief0AB4G0gbYaRM6x")
	flavorJSON, err := json.Marshal(flavor)

	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(flavorJSON), "../repository/mock/flavor-signing-key.pem")
	manifest := instance.Manifest{InstanceInfo: instance.Info{InstanceID: "7b280921-83f7-4f44-9f8d-2dcf36e7af33", HostHardwareUUID: "59EED8F0-28C5-4070-91FC-F5E2E5443F6B", ImageID: "670f263e-b34e-4e07-a520-40ac9a89f62d"}, ImageEncrypted: true}
	json.Unmarshal([]byte(signedFlavorString), &signedFlavor)

	report, err := verifier.Verify(&manifest, &signedFlavor, "../repository/mock/flavor-signing-cert.pem", currDir, false)
	instanceReport, _ := report.(*verifier.InstanceTrustReport)

	fJSON, err := json.Marshal(instanceReport)
	checkErr(err)

	//create an rsa keypair, and test certificate
	rsaPriv, cert, err := crypt.CreateSelfSignedCertAndRSAPrivKeys()
	checkErr(err)

	signature, err := crypt.HashAndSignPKCS1v15([]byte(fJSON), rsaPriv, crypto.SHA256)
	checkErr(err)

	signedReport := crypt.SignedData{fJSON, crypt.GetHashingAlgorithmName(crypto.SHA256), cert, signature}

	signedJSON, err := json.Marshal(signedReport)
	checkErr(err)

	r := mux.NewRouter()
	r.Use(middleware.NewTokenAuth("../mockJWTDir", "../mockJWTDir", mockRetrieveJWTSigningCerts, cacheTime))
	SetReportsEndpoints(r.PathPrefix("/wls/reports").Subrouter(), wlsDB)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/wls/reports", bytes.NewBuffer(signedJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	rbody, _ := ioutil.ReadAll(recorder.Result().Body)
	log.Infof("%v", string(rbody))
	assert.Equal(http.StatusCreated, recorder.Code)

	// ISECL-3639: a GET without parameters to /wls/reports should return 400 and an error message
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)

	// reports with filter=false should return all reports
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?filter=false", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?instance_id=7b280921-83f7-4f44-9f8d-2dcf36e7af33&&from_date=2017-08-26T11:45:42", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?instance_id=7b280921-83f7-4f44-9f8d-2dcf36e7af33&&from_date=2017-08-26T11:45:42&&latest_per_vm=false", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var rResponse []model.Report
	checkErr(json.Unmarshal(recorder.Body.Bytes(), &rResponse))

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?hardware_uuid=59EED8F0-28C5-4070-91FC-F5E2E5443F6B&&to_date=2019-08-26T11:45:42", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?to_date=2019-08-26T11:45:42", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?from_date=2017-08-26T11:45:42", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?num_of_days=3&&latest_per_vm=false", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?num_of_days=3", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	// TODO: Fix failing tests - after Sprint 29

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?report_id="+rResponse[0].ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/reports/"+rResponse[0].ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?report_id="+rResponse[0].ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	var rResponse1 []model.Report
	checkErr(json.Unmarshal(recorder.Body.Bytes(), &rResponse1))
	assert.Equal(0, len(rResponse1))

}
