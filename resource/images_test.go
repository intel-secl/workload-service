package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"intel/isecl/lib/flavor"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestImagesResource(t *testing.T) {
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

	// First Create a Flavor, and store it in DB
	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)

	// setup Flavor Endpoints
	r := mux.NewRouter()
	SetFlavorsEndpoints(r.PathPrefix("/wls/flavors").Subrouter(), db)

	// Post a new Flavor
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer(fJSON))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Setup Images Endpoints
	SetImagesEndpoints(r.PathPrefix("/wls/images").Subrouter(), db)

	// Post a new Image association
	recorder = httptest.NewRecorder()
	id, _ := uuid.NewV4()
	newImage := CreateImage{ImageID: id.String(), FlavorID: f.Image.Meta.ID}
	newImageJSON, _ := json.Marshal(newImage)
	req = httptest.NewRequest("POST", "/wls/images", bytes.NewBuffer(newImageJSON))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Check and see if the Image has been created in the db
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/images/"+newImage.ImageID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	// Create another Image Association
	uuid2, _ := uuid.NewV4()
	newImage2 := CreateImage{ImageID: uuid2.String(), FlavorID: f.Image.Meta.ID}
	newImage2JSON, _ := json.Marshal(newImage2)
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/images", bytes.NewBuffer(newImage2JSON))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// query all by flavorID and see if we can find boths
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/images?flavor_id="+f.Image.Meta.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var response struct {
		ImageIDs []string `json:"image_ids"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	checkErr(err)
	assert.NotEmpty(response.ImageIDs)
	assert.ElementsMatch([]string{newImage.ImageID, newImage2.ImageID}, response.ImageIDs)

	// Delete  the first one we created
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/images/"+newImage.ImageID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	// Assert that it doesn't exist anymore
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/images/"+newImage.ImageID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNotFound, recorder.Code)

	// Clean up Flavor, and see if images associated are gone too.
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f.Image.Meta.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	// Check to see if the second image we created was implicitly deleted
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/images/"+newImage2.ImageID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNotFound, recorder.Code)
}
