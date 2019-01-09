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
	go mockKMS()
	go mockHVS()

	// Test Flavor-Key
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/images/dddd021e-9669-4e53-9224-8880fb4e4080/flavor-key?hardware_uuid=ecee021e-9669-4e53-9224-8880fb4e4080", nil)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusOK, recorder.Code)
}
