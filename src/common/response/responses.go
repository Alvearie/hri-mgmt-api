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

func Success(statusCode int, body map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"statusCode": statusCode,
		"body":       body,
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
