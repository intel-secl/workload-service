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
// 		  "key_URL": "http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer"
// 		}
// 	  }
// 	}
//   }

func TestFlavorResource(t *testing.T) {
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
	f, err := flavor.GetImageFlavor("Cirros-enc", true, "http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
	checkErr(err)
	fJSON, err := json.Marshal(f)
	checkErr(err)

	r := mux.NewRouter()
	SetFlavorsEndpoints(r.PathPrefix("/wls/flavors").Subrouter(), db)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBuffer(fJSON))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/flavors/"+f.Image.Meta.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
	var fResponse flavor.ImageFlavor
	checkErr(json.Unmarshal(recorder.Body.Bytes(), &fResponse))
	assert.Equal(*f, fResponse)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("DELETE", "/wls/flavors/"+f.Image.Meta.ID, nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNoContent, recorder.Code)
}
