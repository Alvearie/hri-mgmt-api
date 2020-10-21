/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package response

import (
	"fmt"
	"net/http"
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

func TestSuccess(t *testing.T) {
	_ = os.Setenv(EnvOwActivationId, activationId)
	inputBody := map[string]interface{}{"batchId": "batch123"}
	response := Success(http.StatusCreated, inputBody)

	if response["statusCode"] != http.StatusCreated {
		t.Errorf("Unexpected StatusCode. Expected: [%v], Actual: [%v]", http.StatusCreated, response["statusCode"])
	}
	if !reflect.DeepEqual(response["body"], inputBody) {
		t.Errorf("Unexpected body. Expected: [%v], Actual: [%v]", inputBody, response["body"])
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
