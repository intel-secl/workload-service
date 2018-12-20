package resource

import (
	"intel/isecl/workload-service/repository/postgres"
	"bytes"
	"encoding/json"
	"fmt"
	"intel/isecl/lib/flavor"
	"intel/isecl/workload-service/model"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func setupServer(t *testing.T) *mux.Router {
	_, ci := os.LookupEnv("CI")
	var host string
	if ci {
		host = "postgres"
	} else {
		host = "localhost"
	}
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=5432 user=runner dbname=wls password=test sslmode=disable", host))
	if err != nil {
		t.Fatal("could not open DB")
	}
	r := mux.NewRouter()
	wlsDB := postgres.PostgresDatabase{DB: db.Debug()}
	wlsDB.Migrate()
	SetFlavorsEndpoints(r.PathPrefix("/wls/flavors").Subrouter(), wlsDB)
	// Setup Images Endpoints
	SetImagesEndpoints(r.PathPrefix("/wls/images").Subrouter(), wlsDB)

	return r
}

func TestImagesResource(t *testing.T) {
	assert := assert.New(t)
	checkErr := func(e error) {
		assert.NoError(e)
		if e != nil {
			assert.FailNow("fatal error, cannot continue test")
		}
	}
	r := setupServer(t)
	// First Create a Flavor, and store it in DB
	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)

	// Free standinf falvor that wont be associated with any images
	f2, err := flavor.GetImageFlavor("Bad-guy", true, "http://10.1.68.21:20080/v1/keys/83755fdb-c910-46be-821f-e8ddeab189e8/transfer", "2260f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd");
	checkErr(err)
	f2JSON, err := json.Marshal(f2)
	checkErr(err)

	// Post a new Flavor
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer(fJSON))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post second Flavor
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer(f2JSON))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Post a new Image association
	recorder = httptest.NewRecorder()
	id, _ := uuid.NewV4()
	newImage := model.Image{ID: id.String(), FlavorIDs: []string{f.Image.Meta.ID}}
	newImageJSON, _ := json.Marshal(newImage)
	req = httptest.NewRequest("POST", "/wls/images", bytes.NewBuffer(newImageJSON))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Check and see if the Image has been created in the db
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/images/"+newImage.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var getResp model.Image
	_ = json.Unmarshal(recorder.Body.Bytes(), &getResp)
	assert.Equal(newImage, getResp)

	// Create another Image Association
	uuid2, _ := uuid.NewV4()
	newImage2 := model.Image{ID: uuid2.String(), FlavorIDs: []string{f.Image.Meta.ID}}
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
	var response []model.Image
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	checkErr(err)
	assert.NotEmpty(response)
	i1 := model.Image{
		ID:       newImage.ID,
		FlavorIDs: newImage.FlavorIDs,
	}
	i2 := model.Image{
		ID:       newImage2.ID,
		FlavorIDs: newImage2.FlavorIDs,
	}
	assert.ElementsMatch([]model.Image{i1, i2}, response)

	// Delete  the first one we created
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/images/"+newImage.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	// Assert that it doesn't exist anymore
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/images/"+newImage.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNotFound, recorder.Code)

	// Clean up Flavor
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f.Image.Meta.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f2.Image.Meta.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)

	// Clean up Image
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/images/"+newImage2.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}

