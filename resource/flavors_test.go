/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"bytes"
	"intel/isecl/workload-service/v3/repository/mock"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jinzhu/gorm"

	"github.com/stretchr/testify/assert"
)

func TestDeleteNonExistentFlavorID(t *testing.T) {
	log.Trace("resource/flavors_test:TestDeleteNonExistentFlavorID() Entering")
	defer log.Trace("resource/flavors_test:TestDeleteNonExistentFlavorID() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	db.MockFlavor.DeleteByUUIDFn = func(uuid string) error {
		return gorm.ErrRecordNotFound
	}
	r := setupMockServer(db)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/wls/flavors/dddd021e-9669-4e53-9224-8880fb4e4080", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusNotFound, recorder.Code)
}

func TestInvalidFlavorID(t *testing.T) {
	log.Trace("resource/flavors_test:TestInvalidFlavorID() Entering")
	defer log.Trace("resource/flavors_test:TestInvalidFlavorID() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	db.MockFlavor.DeleteByUUIDFn = func(uuid string) error {
		return gorm.ErrRecordNotFound
	}
	r := setupMockServer(db)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/wls/flavors/yaddablahblahblbahlbah", nil)
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
	assert.Contains(recorder.Body.String(), "not uuidv4 compliant")
}

func TestFlavorPartValidation(t *testing.T) {
	log.Trace("resource/flavors_test:TestFlavorPartValidation() Entering")
	defer log.Trace("resource/flavors_test:TestFlavorPartValidation() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)

	// Invalid flavor part (from ISECL-3459) should fail
	badFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"flavor_part":"image123","label":"label_image-test-3"}},"encryption_required":true,"encryption":{"key_url":"https://kbs.server.com:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}},"signature": "CStRpWgj0De7+xoX1uFSOacLAZeEcodUuvH62B4hVoiIEriVaHxrLJhBjnIuSPmIoZewCdTShw7GxmMQiMikCrVhaUilYk066TckOcLW/E3K+7NAiZ5kuS96J6dVxgJ+9k7iKf7Z+6lnWUJz92VWLP4U35WK4MtV+MPTYn2Zj1p+/tTUuSqlk8KCmpywzI1J1/XXjvqee3M9cGInnbOUGEFoLBAO1+w30yptoNxKEaB/9t3qEYywk8buT5GEMYUjJEj9PGGaW+lR37x0zcXggwMg/RsijMV6rNKsjjC0fN1vGswzoaIJPD1RJkQ8X9l3AaM0qhLBQDrurWxKK4KSQSpI0BziGPkKi5vAeeRkVfU5JXNdPxdOkyXVebeMQR9bYntXtZl41qjOZ0zIOKAHNJiBLyMYausbTZHVCwDuA/HBAT8i7JAIesxexX89bL+khPebHWkHaifS4NejymbGzM+n62EHuoeIo33qDMQ/U0FA3i6gRy0s/sFQVXR0xk8l"}`
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(badFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)

	// "IMAGE" flavor part should be created
	imageFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"flavor_part":"IMAGE","label":"label_image-test-3"}},"encryption_required":true, "encryption":{"key_url":"https://kbs.server.com:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}}, "signature": "CStRpWgj0De7+xoX1uFSOacLAZeEcodUuvH62B4hVoiIEriVaHxrLJhBjnIuSPmIoZewCdTShw7GxmMQiMikCrVhaUilYk066TckOcLW/E3K+7NAiZ5kuS96J6dVxgJ+9k7iKf7Z+6lnWUJz92VWLP4U35WK4MtV+MPTYn2Zj1p+/tTUuSqlk8KCmpywzI1J1/XXjvqee3M9cGInnbOUGEFoLBAO1+w30yptoNxKEaB/9t3qEYywk8buT5GEMYUjJEj9PGGaW+lR37x0zcXggwMg/RsijMV6rNKsjjC0fN1vGswzoaIJPD1RJkQ8X9l3AaM0qhLBQDrurWxKK4KSQSpI0BziGPkKi5vAeeRkVfU5JXNdPxdOkyXVebeMQR9bYntXtZl41qjOZ0zIOKAHNJiBLyMYausbTZHVCwDuA/HBAT8i7JAIesxexX89bL+khPebHWkHaifS4NejymbGzM+n62EHuoeIo33qDMQ/U0FA3i6gRy0s/sFQVXR0xk8l"}`
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(imageFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// "CONTAINER_IMAGE" flavor part should be created
	containerImageFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"flavor_part":"CONTAINER_IMAGE","label":"label_image-test-3"}},"encryption_required":true,"encryption":	{"key_url":"https://kbs.server.com:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}},"signature": "CStRpWgj0De7+xoX1uFSOacLAZeEcodUuvH62B4hVoiIEriVaHxrLJhBjnIuSPmIoZewCdTShw7GxmMQiMikCrVhaUilYk066TckOcLW/E3K+7NAiZ5kuS96J6dVxgJ+9k7iKf7Z+6lnWUJz92VWLP4U35WK4MtV+MPTYn2Zj1p+/tTUuSqlk8KCmpywzI1J1/XXjvqee3M9cGInnbOUGEFoLBAO1+w30yptoNxKEaB/9t3qEYywk8buT5GEMYUjJEj9PGGaW+lR37x0zcXggwMg/RsijMV6rNKsjjC0fN1vGswzoaIJPD1RJkQ8X9l3AaM0qhLBQDrurWxKK4KSQSpI0BziGPkKi5vAeeRkVfU5JXNdPxdOkyXVebeMQR9bYntXtZl41qjOZ0zIOKAHNJiBLyMYausbTZHVCwDuA/HBAT8i7JAIesxexX89bL+khPebHWkHaifS4NejymbGzM+n62EHuoeIo33qDMQ/U0FA3i6gRy0s/sFQVXR0xk8l"}`
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(containerImageFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusCreated, recorder.Code)

	// Empty flavor part should fail
	emptyImageFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"flavor_part":"","label":"label_image-test-3"}},"encryption_required":true,"encryption":{"key_url":"https://kbs.server.com:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}},"signature": "CStRpWgj0De7+xoX1uFSOacLAZeEcodUuvH62B4hVoiIEriVaHxrLJhBjnIuSPmIoZewCdTShw7GxmMQiMikCrVhaUilYk066TckOcLW/E3K+7NAiZ5kuS96J6dVxgJ+9k7iKf7Z+6lnWUJz92VWLP4U35WK4MtV+MPTYn2Zj1p+/tTUuSqlk8KCmpywzI1J1/XXjvqee3M9cGInnbOUGEFoLBAO1+w30yptoNxKEaB/9t3qEYywk8buT5GEMYUjJEj9PGGaW+lR37x0zcXggwMg/RsijMV6rNKsjjC0fN1vGswzoaIJPD1RJkQ8X9l3AaM0qhLBQDrurWxKK4KSQSpI0BziGPkKi5vAeeRkVfU5JXNdPxdOkyXVebeMQR9bYntXtZl41qjOZ0zIOKAHNJiBLyMYausbTZHVCwDuA/HBAT8i7JAIesxexX89bL+khPebHWkHaifS4NejymbGzM+n62EHuoeIo33qDMQ/U0FA3i6gRy0s/sFQVXR0xk8l"}`
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(emptyImageFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)

	// Omitted flavor part should fail
	omittedImageFlavorPartJson := `{"flavor":{"meta":{"id":"d6129610-4c8f-4ac4-8823-df4e925688c3","description":{"label":"label_image-test-3"}},"encryption_required":true,"encryption":{"key_url":"https://kbs.server.com:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer","digest":"3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="}},"signature": "CStRpWgj0De7+xoX1uFSOacLAZeEcodUuvH62B4hVoiIEriVaHxrLJhBjnIuSPmIoZewCdTShw7GxmMQiMikCrVhaUilYk066TckOcLW/E3K+7NAiZ5kuS96J6dVxgJ+9k7iKf7Z+6lnWUJz92VWLP4U35WK4MtV+MPTYn2Zj1p+/tTUuSqlk8KCmpywzI1J1/XXjvqee3M9cGInnbOUGEFoLBAO1+w30yptoNxKEaB/9t3qEYywk8buT5GEMYUjJEj9PGGaW+lR37x0zcXggwMg/RsijMV6rNKsjjC0fN1vGswzoaIJPD1RJkQ8X9l3AaM0qhLBQDrurWxKK4KSQSpI0BziGPkKi5vAeeRkVfU5JXNdPxdOkyXVebeMQR9bYntXtZl41qjOZ0zIOKAHNJiBLyMYausbTZHVCwDuA/HBAT8i7JAIesxexX89bL+khPebHWkHaifS4NejymbGzM+n62EHuoeIo33qDMQ/U0FA3i6gRy0s/sFQVXR0xk8l"}`
	recorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/wls/flavors", bytes.NewBufferString(omittedImageFlavorPartJson))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	assert.Equal(http.StatusBadRequest, recorder.Code)
}

//TestGetAllFlavors checks if all flavors are returned without filter
func TestGetFlavorNoFilter(t *testing.T) {
	log.Trace("resource/flavors_test:TestGetFlavorNoFilter() Entering")
	defer log.Trace("resource/flavors_test:TestGetFlavorNoFilter() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/flavors", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	// check we got a good response
	assert.Equal(http.StatusOK, recorder.Code)
}

//TestGetAllFlavorsFilter checks if all flavors are returned without filter
func TestGetFlavorsFilterByLabel(t *testing.T) {
	log.Trace("resource/flavors_test:TestGetFlavorsFilterByLabel() Entering")
	defer log.Trace("resource/flavors_test:TestGetFlavorsFilterByLabel() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/flavors?label=label_image-test-3", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	// check we got a good response
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestGetFlavorsFilterByUUID(t *testing.T) {
	log.Trace("resource/flavors_test:TestGetFlavorsFilterByUUID() Entering")
	defer log.Trace("resource/flavors_test:TestGetFlavorsFilterByUUID() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/flavors?id=d6129610-4c8f-4ac4-8823-df4e925688c3", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	// check we got a good response
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestGetFlavorsFilterWithTrue(t *testing.T) {
	log.Trace("resource/flavors_test:TestGetFlavorsFilterWithTrue() Entering")
	defer log.Trace("resource/flavors_test:TestGetFlavorsFilterWithTrue() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/flavors?id=d6129610-4c8f-4ac4-8823-df4e925688c3&filter=true", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	// check we got a good response
	assert.Equal(http.StatusOK, recorder.Code)
}

func TestGetFlavorsNegFilterTrue(t *testing.T) {
	log.Trace("resource/flavors_test:TestGetFlavorsNegFilterTrue() Entering")
	defer log.Trace("resource/flavors_test:TestGetFlavorsNegFilterTrue() Leaving")
	assert := assert.New(t)
	db := new(mock.Database)
	r := setupMockServer(db)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/wls/flavors?filter=true", nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+BearerToken)
	r.ServeHTTP(recorder, req)
	// check we got a good response
	assert.Equal(http.StatusBadRequest, recorder.Code)
}
