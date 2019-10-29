/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"bytes"
	"encoding/json"
	"intel/isecl/workload-service/config"
	"intel/isecl/workload-service/model"
	"intel/isecl/workload-service/repository"
	"intel/isecl/workload-service/repository/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlavorKey(t *testing.T) {
	log.Trace("resource/images_test:TestFlavorKey() Entering")
	defer log.Trace("resource/images_test:TestFlavorKey() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)
	config.Configuration.HVS_API_URL = "http://localhost:1338/mtwilson/v2/"

	k := mockKMS(":1337")
	defer k.Close()
	h := mockHVS(":1338")
	defer h.Close()
	time.Sleep(1 * time.Second)

	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestFlavorKeyMissingHWUUID(t *testing.T) {
	log.Trace("resource/images_test:TestFlavorKeyMissingHWUUID() Entering")
	defer log.Trace("resource/images_test:TestFlavorKeyMissingHWUUID() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)
	config.Configuration.HVS_API_URL = "http://localhost:2338/mtwilson/v2/"

	k := mockKMS(":2337")
	defer k.Close()
	h := mockHVS(":2338")
	defer h.Close()
	time.Sleep(1 * time.Second)
	// Test Flavor-Key with no hardware_uuid
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Equal("Missing query parameters: [hardware_uuid]\n", recorder.Body.String())
}

func TestFlavorKeyEmptyHWUUID(t *testing.T) {
	log.Trace("resource/images_test:TestFlavorKeyEmptyHWUUID() Entering")
	defer log.Trace("resource/images_test:TestFlavorKeyEmptyHWUUID() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)
	config.Configuration.HVS_API_URL = "http://localhost:3338/mtwilson/v2/"

	k := mockKMS(":3337")
	defer k.Close()
	h := mockHVS(":3338")
	defer h.Close()
	time.Sleep(1 * time.Second)
	// Test Flavor-Key with no hardware_uuid
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Contains(recorder.Body.String(), "Invalid hardware uuid")
}

func TestFlavorKeyHVSDown(t *testing.T) {
	log.Trace("resource/images_test:TestFlavorKeyHVSDown() Entering")
	defer log.Trace("resource/images_test:TestFlavorKeyHVSDown() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)
	config.Configuration.HVS_API_URL = "http://localhost:4338/mtwilson/v2/"

	k := mockKMS(":4337")
	defer k.Close()
	time.Sleep(1 * time.Second)
	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	t.Log(recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
}

func TestFlavorKyHVSBadRequest(t *testing.T) {
	log.Trace("resource/images_test:TestFlavorKyHVSBadRequest() Entering")
	defer log.Trace("resource/images_test:TestFlavorKyHVSBadRequest() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)
	config.Configuration.HVS_API_URL = "http://localhost:5338/mtwilson/v2/"

	k := mockKMS(":5337")
	defer k.Close()
	h := badHVS(":5338")
	defer h.Close()
	time.Sleep(1 * time.Second)
	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	t.Log(recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
}

func TestFlavorKeyKMSDown(t *testing.T) {
	log.Trace("resource/images_test:TestFlavorKeyKMSDown() Entering")
	defer log.Trace("resource/images_test:TestFlavorKeyKMSDown() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)
	config.Configuration.HVS_API_URL = "http://localhost:6338/mtwilson/v2/"

	h := mockHVS(":6338")
	defer h.Close()
	time.Sleep(1 * time.Second)
	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	t.Log(recorder.Body.String())
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestFlavorKeyKMSBadRequest(t *testing.T) {
	log.Trace("resource/images_test:TestFlavorKeyKMSBadRequest() Entering")
	defer log.Trace("resource/images_test:TestFlavorKeyKMSBadRequest() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)
	config.Configuration.HVS_API_URL = "http://localhost:7338/mtwilson/v2/"

	h := mockHVS(":7338")
	defer h.Close()
	k := badKMS(":7337")
	defer k.Close()
	time.Sleep(1 * time.Second)
	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	t.Log(recorder.Body.String())
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestQueryEmptyImagesResource(t *testing.T) {
	log.Trace("resource/images_test:TestQueryEmptyImagesResource() Entering")
	defer log.Trace("resource/images_test:TestQueryEmptyImagesResource() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	db.MockImage.RetrieveByFilterCriteriaFn = func(repository.ImageFilter) ([]model.Image, error) {
		return nil, nil
	}
	r := setupMockServer(db)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
}

func TestQueryImagesResource(t *testing.T) {
	log.Trace("resource/images_test:TestQueryImagesResource() Entering")
	defer log.Trace("resource/images_test:TestQueryImagesResource() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	db.MockImage.RetrieveByFilterCriteriaFn = func(repository.ImageFilter) ([]model.Image, error) {
		return []model.Image{
			model.Image{ID: "ffff021e-9669-4e53-9224-8880fb4e4080"},
			model.Image{ID: "ffff021e-9669-4e53-9224-8880fb4e4081"},
		}, nil
	}
	r := setupMockServer(db)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images?filter=false", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	t.Log(recorder.Body.String())
	assert.Equal(http.StatusOK, recorder.Code)
	var models []model.Image
	json.Unmarshal(recorder.Body.Bytes(), &models)
	assert.Len(models, 2)
	assert.Equal("ffff021e-9669-4e53-9224-8880fb4e4080", models[0].ID)
	assert.Equal("ffff021e-9669-4e53-9224-8880fb4e4081", models[1].ID)
}

func TestInvalidImageID(t *testing.T) {
	log.Trace("resource/images_test:TestInvalidImageID() Entering")
	defer log.Trace("resource/images_test:TestInvalidImageID() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/wls/images/yaddablahblahblbahlbah", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Contains(recorder.Body.String(), "is not uuidv4 compliant")
}

func TestCreateImageEmptyFlavors(t *testing.T) {
	log.Trace("resource/images_test:TestCreateImageEmptyFlavors() Entering")
	defer log.Trace("resource/images_test:TestCreateImageEmptyFlavors() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)

	recorder := httptest.NewRecorder()
	iJSON := `{"id": "ffff021e-9669-4e53-9224-8880fb4e4080", "flavor_ids":[]}`
	req := httptest.NewRequest("POST", "/wls/images", bytes.NewBufferString(iJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)
}
