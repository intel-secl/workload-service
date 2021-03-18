// +build integration

/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */

package resource

import (
	"bytes"
	"encoding/json"
	flavorUtil "intel/isecl/lib/flavor/v3/util"
	"intel/isecl/workload-service/v3/model"
	"net/http"
	"net/http/httptest"
	"testing"

	uuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestImagesResourceIntegration(t *testing.T) {
	log.Trace("resource/images_integration_test:TestImagesResourceIntegration() Entering")
	defer log.Trace("resource/images_integration_test:TestImagesResourceIntegration() Leaving")

	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	r := setupServer(t)
	// First Create a Flavor, and store it in DB
	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)

	// Free standing flavor that wont be associated with any images
	f2, err := flavor.GetImageFlavor("Bad-guy", true, "http://kbs.server.com:20080/v1/keys/83755fdb-c910-46be-821f-e8ddeab189e8/transfer", "2260f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	f2JSON, err := json.Marshal(f2)
	checkErr(err)

	// Post a new Flavor
	recorder := httptest.NewRecorder()
	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(fJSON), "../repository/mock/flavor-signing-key.pem")
	req := httptest.NewRequest("POST", "/wls/v1/flavors", bytes.NewBuffer([]byte(signedFlavorString)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post second Flavor
	recorder = httptest.NewRecorder()
	signedFlavorString2, err := flavorUtil.GetSignedFlavor(string(f2JSON), "../repository/mock/flavor-signing-key.pem")
	req = httptest.NewRequest("POST", "/wls/v1/flavors", bytes.NewBuffer([]byte(signedFlavorString2)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post a new Image
	recorder = httptest.NewRecorder()
	id, _ := uuid.NewV4()
	newImage := model.Image{ID: id.String(), FlavorIDs: []string{f.Image.Meta.ID}}
	newImageJSON, _ := json.Marshal(newImage)
	req = httptest.NewRequest("POST", "/wls/v1/images", bytes.NewBuffer(newImageJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Check and see if the Image has been created in the db
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images/"+newImage.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var getResp model.Image
	_ = json.Unmarshal(recorder.Body.Bytes(), &getResp)
	assert.Equal(newImage, getResp)

	// Check and see if the Image flavor has been associated in the db correctly
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images/"+newImage.ID+"/flavors?flavor_part="+f.Image.Meta.Description.FlavorPart, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	// Create another Image Association
	uuid2, _ := uuid.NewV4()
	newImage2 := model.Image{ID: uuid2.String(), FlavorIDs: []string{f.Image.Meta.ID}}
	newImage2JSON, _ := json.Marshal(newImage2)
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/v1/images", bytes.NewBuffer(newImage2JSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// query all by flavorID and see if we can find both
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images?flavor_id="+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var response []model.Image
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	checkErr(err)
	assert.NotEmpty(response)
	i1 := model.Image{
		ID:        newImage.ID,
		FlavorIDs: newImage.FlavorIDs,
	}
	i2 := model.Image{
		ID:        newImage2.ID,
		FlavorIDs: newImage2.FlavorIDs,
	}
	assert.ElementsMatch([]model.Image{i1, i2}, response)

	// make sure we only get a single image when filtering by image_id (ISECL-3557)
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images?image_id="+newImage.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	checkErr(err)
	assert.NotEmpty(response)
	assert.ElementsMatch([]model.Image{i1}, response)

	// test using both image_id and flavor_id parameters
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images?image_id="+newImage.ID+"&flavor_id="+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	checkErr(err)
	assert.NotEmpty(response)
	assert.ElementsMatch([]model.Image{i1}, response)

	// test 'wls/v1/image?filter=false (should return all new images)
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images?filter=false", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	checkErr(err)
	assert.NotEmpty(response)
	assert.ElementsMatch([]model.Image{i1, i2}, response)

	// Delete  the first one we created
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/images/"+newImage.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	// Assert that it doesn't exist anymore
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images/"+newImage.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNotFound, recorder.Code)

	// Clean up Flavor
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/flavors/"+f2.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	// Clean up Image
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/images/"+newImage2.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}

func TestImageDuplicate(t *testing.T) {
	log.Trace("resource/images_integration_test:TestImageDuplicate() Entering")
	defer log.Trace("resource/images_integration_test:TestImageDuplicate() Leaving")
	assert := assert.New(t)
	r := setupServer(t)

	// Post a new Image association
	recorder := httptest.NewRecorder()
	id, _ := uuid.NewV4()
	newImage := model.Image{ID: id.String(), FlavorIDs: nil}
	newImageJSON, _ := json.Marshal(newImage)
	req := httptest.NewRequest("POST", "/wls/v1/images", bytes.NewBuffer(newImageJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post Duplicate, expect conflict
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/v1/images", bytes.NewBuffer(newImageJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Delete it
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/images/"+newImage.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}

func TestImageAssociatedFlavors(t *testing.T) {
	log.Trace("resource/images_integration_test:TestImageAssociatedFlavors() Entering")
	defer log.Trace("resource/images_integration_test:TestImageAssociatedFlavors() Leaving")
	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	r := setupServer(t)
	// First Create a Flavor, and store it in DB
	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)

	// Free standing falvor that wont be associated with any images
	f2, err := flavor.GetImageFlavor("PretendSoftwareFlavor", true, "http://kbs.server.com:20080/v1/keys/83755fdb-c910-46be-821f-e8ddeab189e8/transfer", "2260f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	f2.Image.Meta.Description.FlavorPart = "IMAGE"
	checkErr(err)
	f2JSON, err := json.Marshal(f2)
	checkErr(err)

	// Post first Flavor
	recorder := httptest.NewRecorder()
	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(fJSON), "../repository/mock/flavor-signing-key.pem")
	req := httptest.NewRequest("POST", "/wls/v1/flavors", bytes.NewBuffer([]byte(signedFlavorString)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post Second Flavor
	recorder = httptest.NewRecorder()
	signedFlavorString2, err := flavorUtil.GetSignedFlavor(string(f2JSON), "../repository/mock/flavor-signing-key.pem")
	req = httptest.NewRequest("POST", "/wls/v1/flavors", bytes.NewBuffer([]byte(signedFlavorString2)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post a new Image association with only the First Flavor associated
	recorder = httptest.NewRecorder()
	id, _ := uuid.NewV4()
	newImage := model.Image{ID: id.String(), FlavorIDs: []string{f.Image.Meta.ID}}
	newImageJSON, _ := json.Marshal(newImage)
	req = httptest.NewRequest("POST", "/wls/v1/images", bytes.NewBuffer(newImageJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Check to see if /imageId/flavors/flavorId works
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images/"+newImage.ID+"/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var getResp model.Flavor
	_ = json.Unmarshal(recorder.Body.Bytes(), &getResp)
	assert.Equal((model.Flavor)(*f), getResp)

	// Add the second Flavor Association to the Image
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/wls/v1/images/"+newImage.ID+"/flavors/"+f2.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Verify that the Image now has another Flavors associated with it
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images/"+newImage.ID+"/flavors/"+f2.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var getResp2 model.Flavor
	_ = json.Unmarshal(recorder.Body.Bytes(), &getResp2)
	assert.Equal((model.Flavor)(*f2), getResp2)

	// Verify it again by querying all associated flavors
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images/"+newImage.ID+"/flavors", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var getAllResp []model.Flavor
	_ = json.Unmarshal(recorder.Body.Bytes(), &getAllResp)
	assert.Len(getAllResp, 2)

	// Delete the second Flavor Association from the Image
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/images/"+newImage.ID+"/flavors/"+f2.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	// Verify that the Image now only has one Flavor associated with it
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/v1/images/"+newImage.ID+"/flavors/"+f2.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNotFound, recorder.Code)

	// Delete Flavors
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/flavors/"+f2.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	// Delete Image

	// Delete it
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/images/"+newImage.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}

func TestImageBadFlavorID(t *testing.T) {
	log.Trace("resource/images_integration_test:TestImageBadFlavorID() Entering")
	defer log.Trace("resource/images_integration_test:TestImageBadFlavorID() Leaving")
	assert := assert.New(t)
	r := setupServer(t)
	// Post a new Image association with Invalid ID
	recorder := httptest.NewRecorder()
	id, _ := uuid.NewV4()
	newImage := model.Image{ID: id.String(), FlavorIDs: []string{"DSFLKDJSFKLJDKSLFJDKD"}}
	newImageJSON, _ := json.Marshal(newImage)
	req := httptest.NewRequest("POST", "/wls/v1/images", bytes.NewBuffer(newImageJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
}

func TestImageDuplicateFlavorIDs(t *testing.T) {
	log.Trace("resource/images_integration_test:TestImageDuplicateFlavorIDs() Entering")
	defer log.Trace("resource/images_integration_test:TestImageDuplicateFlavorIDs() Leaving")
	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	r := setupServer(t)
	// First Create a Flavor, and store it in DB
	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://kbs.server.com:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(fJSON), "../repository/mock/flavor-signing-key.pem")
	checkErr(err)

	// Post first Flavor
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/wls/v1/flavors", bytes.NewBuffer([]byte(signedFlavorString)))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post a new Image association with dupe of the First Flavor associated
	recorder = httptest.NewRecorder()
	id, _ := uuid.NewV4()
	newImage := model.Image{ID: id.String(), FlavorIDs: []string{f.Image.Meta.ID, f.Image.Meta.ID}}
	newImageJSON, _ := json.Marshal(newImage)
	req = httptest.NewRequest("POST", "/wls/v1/images", bytes.NewBuffer(newImageJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusConflict, recorder.Code)

	// Create it normally, but then try and add a dupe via /id/flavors/flavorID
	recorder = httptest.NewRecorder()
	validImage := model.Image{ID: id.String(), FlavorIDs: []string{f.Image.Meta.ID}}
	validImageJSON, _ := json.Marshal(validImage)
	req = httptest.NewRequest("POST", "/wls/v1/images", bytes.NewBuffer(validImageJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// PUT is idempotent, so PUT to this ID should result in no error
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/wls/v1/images/"+newImage.ID+"/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Delete Flavors
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/flavors/"+f.Image.Meta.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	// Delete it
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/images/"+newImage.ID, nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}

func TestCreateImageEmptyFlavorsIntegration(t *testing.T) {
	log.Trace("resource/images_integration_test:TestCreateImageEmptyFlavorsIntegration() Entering")
	defer log.Trace("resource/images_integration_test:TestCreateImageEmptyFlavorsIntegration() Leaving")
	assert := assert.New(t)
	r := setupServer(t)

	recorder := httptest.NewRecorder()
	iJSON := `{"id": "fffe021e-9669-4e53-9224-8880fb4e4080", "flavor_ids":[]}`
	req := httptest.NewRequest("POST", "/wls/v1/images", bytes.NewBufferString(iJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Delete it
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/v1/images/fffe021e-9669-4e53-9224-8880fb4e4080", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}
