// Xbuild integration

package resource

import (
	"bytes"
	"crypto"
	"encoding/json"
	"fmt"
	"intel/isecl/workload-service/repository/postgres"
	"intel/isecl/lib/common/crypt"
	"intel/isecl/lib/common/pkg/instance"
	flavorUtil "intel/isecl/lib/flavor/util"
	"intel/isecl/lib/flavor"
	"intel/isecl/lib/verifier"
	"intel/isecl/workload-service/model"
	"intel/isecl/lib/common/middleware"
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

func TestReportResource(t *testing.T) {
	var signedFlavor flavor.SignedImageFlavor
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
	wlsDB.Migrate()

	flavor, err := flavor.GetImageFlavor("Cirros-enc", true,
		"http://10.1.68.21:20080/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "261209df1789073192285e4e408addadb35068421ef4890a5d4d434")
	flavorJSON, err := json.Marshal(flavor)
	signedFlavorString, err := flavorUtil.GetSignedFlavor(string(flavorJSON), "../repository/mock/flavor-signing-key.pem")
	manifest := instance.Manifest{InstanceInfo: instance.Info{InstanceID: "7B280921-83F7-4F44-9F8D-2DCF36E7AF33", HostHardwareUUID: "59EED8F0-28C5-4070-91FC-F5E2E5443F6B", ImageID: "670F263E-B34E-4E07-A520-40AC9A89F62D"}, ImageEncrypted: true}
	json.Unmarshal([]byte(signedFlavorString), &signedFlavor)
	report, err := verifier.Verify(&manifest, &signedFlavor, "../repository/mock/flavor-signing-cert.pem", false)
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
	r.Use(middleware.NewTokenAuth("../mockJWTDir", "../mockJWTDir", mockRetrieveJWTSigningCerts))
	SetReportsEndpoints(r.PathPrefix("/wls/reports").Subrouter(), wlsDB)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/wls/reports", bytes.NewBuffer(signedJSON))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// ISECL-3639: a GET without parameters to /wls/reports should return 400 and an error message
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?vm_id=7b280921-83f7-4f44-9f8d-2dcf36e7af33&&from_date=2017-08-26T11:45:42", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)

	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/wls/reports?vm_id=7b280921-83f7-4f44-9f8d-2dcf36e7af33&&from_date=2017-08-26T11:45:42&&latest_per_vm=false", nil)
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
