package resource

import (
	"intel/isecl/workload-service/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlavorKey(t *testing.T) {
	assert := assert.New(t)
	r := setupMockServer(t)
	config.Configuration.KMS.URL = "http://localhost:1337/v1/"
	config.Configuration.KMS.User = "user"
	config.Configuration.KMS.Password = "pass"
	config.Configuration.HVS.URL = "http://localhost:1338/mtwilson/v2/"
	config.Configuration.HVS.User = "user"
	config.Configuration.HVS.Password = "pass"
	k := mockKMS(":1337")
	defer k.Close()
	h := mockHVS(":1338")
	defer h.Close()

	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestFlavorKeyMissingHWUUID(t *testing.T) {
	assert := assert.New(t)
	r := setupMockServer(t)
	config.Configuration.KMS.URL = "http://localhost:2337/v1/"
	config.Configuration.KMS.User = "user"
	config.Configuration.KMS.Password = "pass"
	config.Configuration.HVS.URL = "http://localhost:2338/mtwilson/v2/"
	config.Configuration.HVS.User = "user"
	config.Configuration.HVS.Password = "pass"
	k := mockKMS(":2337")
	defer k.Close()
	h := mockHVS(":2338")
	defer h.Close()

	// Test Flavor-Key with no hardware_uuid
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key", nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Equal("Missing query parameters: [hardware_uuid]\n", recorder.Body.String())
}

func TestFlavorKeyEmptyHWUUID(t *testing.T) {
	assert := assert.New(t)
	r := setupMockServer(t)
	config.Configuration.KMS.URL = "http://localhost:3337/v1/"
	config.Configuration.KMS.User = "user"
	config.Configuration.KMS.Password = "pass"
	config.Configuration.HVS.URL = "http://localhost:3338/mtwilson/v2/"
	config.Configuration.HVS.User = "user"
	config.Configuration.HVS.Password = "pass"
	k := mockKMS(":3337")
	defer k.Close()
	h := mockHVS(":3338")
	defer h.Close()

	// Test Flavor-Key with no hardware_uuid
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid", nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Contains(recorder.Body.String(), "cannot be nil")
}

func TestFlavorKeyHVSDown(t *testing.T) {
	assert := assert.New(t)
	r := setupMockServer(t)
	config.Configuration.KMS.URL = "http://localhost:4337/v1/"
	config.Configuration.KMS.User = "user"
	config.Configuration.KMS.Password = "pass"
	config.Configuration.HVS.URL = "http://localhost:4338/mtwilson/v2/"
	config.Configuration.HVS.User = "user"
	config.Configuration.HVS.Password = "pass"
	k := mockKMS(":4337")
	defer k.Close()

	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	r.ServeHTTP(recorder, req)
	t.Log(recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
}

func TestFlavorKyHVSBadRequest(t *testing.T) {
	assert := assert.New(t)
	r := setupMockServer(t)
	config.Configuration.KMS.URL = "http://localhost:5337/v1/"
	config.Configuration.KMS.User = "user"
	config.Configuration.KMS.Password = "pass"
	config.Configuration.HVS.URL = "http://localhost:5338/mtwilson/v2/"
	config.Configuration.HVS.User = "user"
	config.Configuration.HVS.Password = "pass"
	k := mockKMS(":5337")
	defer k.Close()
	h := badHVS(":5338")
	defer h.Close()
	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	r.ServeHTTP(recorder, req)
	t.Log(recorder.Body.String())
	assert.Equal(http.StatusBadRequest, recorder.Code)
}

func TestFlavorKeyKMSDown(t *testing.T) {
	assert := assert.New(t)
	r := setupMockServer(t)
	config.Configuration.KMS.URL = "http://localhost:6337/v1/"
	config.Configuration.KMS.User = "user"
	config.Configuration.KMS.Password = "pass"
	config.Configuration.HVS.URL = "http://localhost:6338/mtwilson/v2/"
	config.Configuration.HVS.User = "user"
	config.Configuration.HVS.Password = "pass"
	h := mockHVS(":6338")
	defer h.Close()
	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	r.ServeHTTP(recorder, req)
	t.Log(recorder.Body.String())
	assert.Equal(http.StatusInternalServerError, recorder.Code)
}

func TestFlavorKeyKMSBadRequest(t *testing.T) {
	assert := assert.New(t)
	r := setupMockServer(t)
	config.Configuration.KMS.URL = "http://localhost:7337/v1/"
	config.Configuration.KMS.User = "user"
	config.Configuration.KMS.Password = "pass"
	config.Configuration.HVS.URL = "http://localhost:7338/mtwilson/v2/"
	config.Configuration.HVS.User = "user"
	config.Configuration.HVS.Password = "pass"
	h := mockHVS(":7337")
	defer h.Close()
	k := badKMS(":7338")
	defer k.Close()

	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	r.ServeHTTP(recorder, req)
	t.Log(recorder.Body.String())
	assert.Equal(http.StatusBadRequest, recorder.Code)
}
