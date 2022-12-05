/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	config := config.Config{
		ConfigPath: "",
		Validation: false,
	}

	handler := NewHandler(config).(*theHandler)
	assert.True(t, reflect.DeepEqual(config, handler.config))

	// Can't check partitionReaderFromConfig, because it's an anonymous function
	// This asserts that they are the same function by memory address
	assert.Equal(t, reflect.ValueOf(GetCheck), reflect.ValueOf(handler.hriHealthcheck))
}

func TestHealthcheckHandler(t *testing.T) {
	validConfig := config.Config{
		ConfigPath:        "",
		AzKafkaBrokers:    []string{"broker1", "broker2"},
		AzKafkaProperties: config.StringMap{"message.max.bytes": "10000"},
		Validation:        false,
	}

	tests := []struct {
		name         string
		handler      *theHandler
		expectedCode int
		expectedBody string
	}{
		{
			name: "Good healthcheck",
			handler: &theHandler{
				config: validConfig,
				hriHealthcheck: func(requestId string, healthChecker kafka.HealthChecker) (int, *response.ErrorDetail) {
					return http.StatusOK, nil
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name: "Bad healthcheck",
			handler: &theHandler{
				config: config.Config{},
				hriHealthcheck: func(requestId string, healthChecker kafka.HealthChecker) (int, *response.ErrorDetail) {
					return http.StatusServiceUnavailable, response.NewErrorDetail(requestId, "Cosmos not available")
				},
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"" + requestId + "\",\"errorDescription\":\"Cosmos not available\"}\n",
		},

		{
			name: "Kafka client error",
			handler: &theHandler{
				config:         config.Config{AzKafkaProperties: config.StringMap{"message.max.bytes": "bad_value"}},
				hriHealthcheck: nil,
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"" + requestId + "\",\"errorDescription\":\"error constructing Kafka admin client: Invalid value for configuration property \\\"message.max.bytes\\\"\"}\n",
		},
	}

	logwrapper.Initialize("error", os.Stdout)
	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.Response().Header().Add(echo.HeaderXRequestID, requestId)

			if assert.NoError(t, tt.handler.HriHealthcheck(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}
