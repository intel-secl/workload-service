package resource

import (
	"intel/isecl/workload-service/repository/mock"
	"net/http"
	"net/http/httptest"
	"testing"

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
