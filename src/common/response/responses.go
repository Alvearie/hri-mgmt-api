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
)

const EnvOwActivationId string = "__OW_ACTIVATION_ID"
const missingParamsMsg string = "Missing required parameter(s): %v"
const invalidParamsMsg string = "Invalid parameter type(s): %v"

func Error(statusCode int, description string) map[string]interface{} {
	activationId := os.Getenv(EnvOwActivationId)

	return map[string]interface{}{
		"error": map[string]interface{}{
			"statusCode": statusCode,
			"body": map[string]interface{}{
				"errorEventId":     activationId,
				"errorDescription": description,
			},
		},
	}
}

func MissingParams(params ...string) map[string]interface{} {
	desc := fmt.Sprintf(missingParamsMsg, params)
	return Error(http.StatusBadRequest, desc)
}

func InvalidParams(params ...string) map[string]interface{} {
	desc := fmt.Sprintf(invalidParamsMsg, params)
	return Error(http.StatusBadRequest, desc)
}

type ErrorDetail struct {
	ErrorEventId     string `json:"errorEventId"`
	ErrorDescription string `json:"errorDescription"`
}

func NewErrorDetail(requestId string, description string) *ErrorDetail {
	errorDetail := ErrorDetail{
		ErrorEventId:     requestId,
		ErrorDescription: description,
	}
	return &errorDetail
}

type ErrorDetailResponse struct {
	Code int
	Body *ErrorDetail
}

func NewErrorDetailResponse(code int, requestId string, description string) *ErrorDetailResponse {
	return &ErrorDetailResponse{
		Code: code,
		Body: &ErrorDetail{
			ErrorEventId:     requestId,
			ErrorDescription: description,
		},
	}
}
