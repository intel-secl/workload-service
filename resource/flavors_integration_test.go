// +build integration

/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"intel/isecl/lib/common/v3/middleware"
	"intel/isecl/lib/flavor/v3"
	flavorUtil "intel/isecl/lib/flavor/v3/util"
	"intel/isecl/workload-service/v3/model"
	"intel/isecl/workload-service/v3/repository/postgres"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	// Import Postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// 	"flavor": {
// 	  "id": "string",
// 	  "meta": {
// 		"description": {
// 		  "flavor_part": "IMAGE",
// 		  "label": "Cirros-enc"
// 		},
// 		"encryption": {
// 		  "encryption_required": true,
// 		  "key_URL": "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer"
// 		}
// 	  }
// 	}
//   }

func setupFlavorServer(t *testing.T) *mux.Router {
	log.Trace("resource/flavors_integration_test:setupFlavorServer() Entering")
	defer log.Trace("resource/flavors_integration_test:setupFlavorServer() Leaving")
	checkErr := func(e error) {
		assert.NoError(t, e)
		if e != nil {
			assert.FailNow(t, "fatal error, cannot continue test")
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

	r := mux.NewRouter()
	r.Use(middleware.NewTokenAuth("../mockJWTDir", "../mockJWTDir", mockRetrieveJWTSigningCerts, cacheTime))
	wlsDB := postgres.PostgresDatabase{DB: db.Debug()}
	wlsDB.Migrate()
	SetFlavorsEndpoints(r.PathPrefix("/wls/flavors").Subrouter(), wlsDB)
	return r
}

func TestFlavorResource(t *testing.T) {
	log.Trace("resource/flavors_integration_test:TestFlavorResource() Entering")
	defer log.Trace("resource/flavors_integration_test:TestFlavorResource() Leaving")
	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	r := setupFlavorServer(t)

	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)

	recorder := httptest.NewRecorder()
	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(fJSON), "../repository/mock/flavor-signing-key.pem")
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer([]byte(signedFlavorString)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)
	var postResponse model.Flavor
	json.Unmarshal(recorder.Body.Bytes(), &postResponse)
	assert.Equal((model.Flavor)(*f), postResponse)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var fResponse model.Flavor
	checkErr(json.Unmarshal(recorder.Body.Bytes(), &fResponse))
	assert.Equal((model.Flavor)(*f), fResponse)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}

func TestDuplicate(t *testing.T) {
	log.Trace("resource/flavors_integration_test:TestDuplicate() Entering")
	defer log.Trace("resource/flavors_integration_test:TestDuplicate() Leaving")
	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	r := setupFlavorServer(t)

	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)

	// Post it once
	recorder := httptest.NewRecorder()
	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(fJSON), "../repository/mock/flavor-signing-key.pem")
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer([]byte(signedFlavorString)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post it again
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer([]byte(signedFlavorString)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusConflict, recorder.Code)

	// Delete
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

}

func TestFlavorDuplicateLabel(t *testing.T) {
	log.Trace("resource/flavors_integration_test:TestFlavorDuplicateLabel() Entering")
	defer log.Trace("resource/flavors_integration_test:TestFlavorDuplicateLabel() Leaving")
	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	r := setupFlavorServer(t)

	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)
	f2, err := flavor.GetImageFlavor("Cirros-enc", true, "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	f2JSON, err := json.Marshal(f2)
	checkErr(err)

	// Post it once
	recorder := httptest.NewRecorder()
	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(fJSON), "../repository/mock/flavor-signing-key.pem")
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer([]byte(signedFlavorString)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post it again
	recorder = httptest.NewRecorder()
	signedFlavorString2, err := flavorUtil.GetSignedFlavor(string(f2JSON), "../repository/mock/flavor-signing-key.pem")
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer([]byte(signedFlavorString2)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusConflict, recorder.Code)

	// Delete
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

}

func TestFlavorInvalidJson(t *testing.T) {
	log.Trace("resource/flavors_integration_test:TestFlavorInvalidJson() Entering")
	defer log.Trace("resource/flavors_integration_test:TestFlavorInvalidJson() Leaving")
	assert := assert.New(t)
	r := setupFlavorServer(t)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString("asdlkfjaksdlfjklsjfd"))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
}

func TestFlavorDeleteByLabel(t *testing.T) {
	log.Trace("resource/flavors_integration_test:TestFlavorDeleteByLabel() Entering")
	defer log.Trace("resource/flavors_integration_test:TestFlavorDeleteByLabel() Leaving")
	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	r := setupFlavorServer(t)

	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)

	// Post it once
	recorder := httptest.NewRecorder()
	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(fJSON), "../repository/mock/flavor-signing-key.pem")
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer([]byte(signedFlavorString)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Delete
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f.Image.Meta.Description.Label, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Contains(recorder.Body.String(), "compliant")

	// Actually Delete it
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}

func TestGetByLabel(t *testing.T) {
	log.Trace("resource/flavors_integration_test:TestGetByLabel() Entering")
	defer log.Trace("resource/flavors_integration_test:TestGetByLabel() Leaving")
	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	r := setupFlavorServer(t)

	f, err := flavor.GetImageFlavor("TestGetByLabel", true, "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)

	// Post it once
	recorder := httptest.NewRecorder()
	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(fJSON), "../repository/mock/flavor-signing-key.pem")
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer([]byte(signedFlavorString)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Get
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/flavors/"+f.Image.Meta.Description.Label, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	// Actually Delete it
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}
