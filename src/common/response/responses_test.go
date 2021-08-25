/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package response

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetErrorDetail(t *testing.T) {
	requestId := "requestId"
	description := "Could not perform elasticsearch health check: elasticsearch client error: client error"

	result := NewErrorDetail(requestId, description)
	expectedErrorDetail := &ErrorDetail{ErrorEventId: requestId, ErrorDescription: description}

	if !reflect.DeepEqual(result, expectedErrorDetail) {
		t.Errorf("expected [%v] but have [%v]", expectedErrorDetail, result)
	}
}

func TestErrorDetailToJson(t *testing.T) {
	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	requestId := "requestId"
	description := "Could not perform elasticsearch health check: elasticsearch client error: client error"
	result := NewErrorDetail(requestId, description)

	context.JSON(http.StatusOK, result)
	assert.Equal(t, "{\"errorEventId\":\""+requestId+"\",\"errorDescription\":\""+description+"\"}\n", recorder.Body.String())
}

func TestNewErrorDetailResponse(t *testing.T) {
	code := http.StatusBadRequest
	requestId := "requestId"
	description := "Could not perform elasticsearch health check: elasticsearch client error: client error"

	result := NewErrorDetailResponse(code, requestId, description)
	expectedErrorDetail := &ErrorDetailResponse{
		Code: code,
		Body: &ErrorDetail{
			ErrorEventId:     requestId,
			ErrorDescription: description,
		},
	}

	if !reflect.DeepEqual(result, expectedErrorDetail) {
		t.Errorf("expected [%v] but have [%v]", expectedErrorDetail, result)
	}
}
