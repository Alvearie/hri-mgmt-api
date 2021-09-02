/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package response

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

const activationId string = "activation123"

func TestError(t *testing.T) {
	_ = os.Setenv(EnvOwActivationId, activationId)
	inputDesc := "Caller is not authorized"
	response := Error(http.StatusUnauthorized, inputDesc)

	err, ok := response["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing error entry, or it's not an object")
	}
	if err["statusCode"] != http.StatusUnauthorized {
		t.Errorf("Unexpected StatusCode. Expected: [%v], Actual: [%v]", http.StatusUnauthorized, err["statusCode"])
	}

	body, ok := err["body"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing body entry, or it's not an object")
	}
	if body["errorEventId"] != activationId {
		t.Errorf("Unexpected errorEventId. Expected: [%v], Actual: [%v]", activationId, body["errorEventId"])
	}
	if body["errorDescription"] != inputDesc {
		t.Errorf("Unexpected errorDescription. Expected: [%v], Actual: [%v]", inputDesc, body["errorDescription"])
	}
}

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

func TestMissingParams(t *testing.T) {
	_ = os.Setenv(EnvOwActivationId, activationId)
	p1 := "param1"
	p2 := "param2"
	p3 := "param3"
	expectedMsg := fmt.Sprintf(missingParamsMsg, []string{p1, p2, p3})
	response := MissingParams(p1, p2, p3)

	err, ok := response["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing error entry, or it's not an object")
	}
	if err["statusCode"] != http.StatusBadRequest {
		t.Errorf("Unexpected StatusCode. Expected: [%v], Actual: [%v]", http.StatusBadRequest, err["statusCode"])
	}

	body, ok := err["body"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing body entry, or it's not an object")
	}
	if body["errorEventId"] != activationId {
		t.Errorf("Unexpected errorEventId. Expected: [%v], Actual: [%v]", activationId, body["errorEventId"])
	}
	if body["errorDescription"] != expectedMsg {
		t.Errorf("Unexpected errorDescription. Expected: [%v], Actual: [%v]", expectedMsg, body["errorDescription"])
	}
}

func TestInvalidParams(t *testing.T) {
	_ = os.Setenv(EnvOwActivationId, activationId)
	p1 := "param1"
	p2 := "param2"
	p3 := "param3"
	expectedMsg := fmt.Sprintf(invalidParamsMsg, []string{p1, p2, p3})
	response := InvalidParams(p1, p2, p3)

	err, ok := response["error"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing error entry, or it's not an object")
	}
	if err["statusCode"] != http.StatusBadRequest {
		t.Errorf("Unexpected StatusCode. Expected: [%v], Actual: [%v]", http.StatusBadRequest, err["statusCode"])
	}

	body, ok := err["body"].(map[string]interface{})
	if !ok {
		t.Fatalf("Missing body entry, or it's not an object")
	}
	if body["errorEventId"] != activationId {
		t.Errorf("Unexpected errorEventId. Expected: [%v], Actual: [%v]", activationId, body["errorEventId"])
	}
	if body["errorDescription"] != expectedMsg {
		t.Errorf("Unexpected errorDescription. Expected: [%v], Actual: [%v]", expectedMsg, body["errorDescription"])
	}
}
