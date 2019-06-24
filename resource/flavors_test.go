package resource

import (
	"intel/isecl/workload-service/repository/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"bytes"

	"github.com/jinzhu/gorm"

	"github.com/stretchr/testify/assert"
)

func TestDeleteNonExistentFlavorID(t *testing.T) {
	assert := assert.New(t)
	db := new(mock.Database)
	db.MockFlavor.DeleteByUUIDFn = func(uuid string) error {
		return gorm.ErrRecordNotFound
	}
	r := setupMockServer(db)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/wls/flavors/dddd021e-9669-4e53-9224-8880fb4e4080", nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNotFound, recorder.Code)
}

func TestInvalidFlavorID(t *testing.T) {
	assert := assert.New(t)
	db := new(mock.Database)
	db.MockFlavor.DeleteByUUIDFn = func(uuid string) error {
		return gorm.ErrRecordNotFound
	}
	r := setupMockServer(db)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/wls/flavors/yaddablahblahblbahlbah", nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Contains(recorder.Body.String(), "is not uuidv4 compliant")
}

func TestFlavorPartValidation(t *testing.T) {
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)

	// Invalid flavor part (from ISECL-3459) should fail
	badFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"flavor_part":"image123","label":"label_image-test-3"}},"encryption":	{"encryption_required":true,"key_url":"https://10.105.168.234:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}}}"`
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(badFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)

	// "IMAGE" flavor part should be created
	imageFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"flavor_part":"IMAGE","label":"label_image-test-3"}},"encryption":	{"encryption_required":true,"key_url":"https://10.105.168.234:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}}}"`
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(imageFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// "CONTAINER_IMAGE" flavor part should be created
	containerImageFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"flavor_part":"CONTAINER_IMAGE","label":"label_image-test-3"}},"encryption":	{"encryption_required":true,"key_url":"https://10.105.168.234:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}}}"`
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(containerImageFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Empty flavor part should fail
	emptyImageFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"flavor_part":"","label":"label_image-test-3"}},"encryption":	{"encryption_required":true,"key_url":"https://10.105.168.234:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}}}"`
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(emptyImageFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)

	// Omitted flavor part should fail
	omittedImageFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"label":"label_image-test-3"}},"encryption":	{"encryption_required":true,"key_url":"https://10.105.168.234:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}}}"`
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(omittedImageFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)

}