/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package healthcheck

/*
func TestNewHandler(t *testing.T) {
	config := config.Config{
		ConfigPath:      "",
		OidcIssuer:      "",
		JwtAudienceId:   "",
		Validation:      false,
		ElasticUrl:      "",
		ElasticUsername: "",
		ElasticPassword: "",
		ElasticCert:     "",
	}

	handler := NewHandler(config).(*theHandler)
	assert.True(t, reflect.DeepEqual(config, handler.config))

	// Can't check partitionReaderFromConfig, because it's an anonymous function
	// This asserts that they are the same function by memory address
	assert.Equal(t, reflect.ValueOf(Get), reflect.ValueOf(handler.healthcheck))
}

func TestHealthcheckHandler(t *testing.T) {
	tests := []struct {
		name         string
		handler      *theHandler
		expectedCode int
		expectedBody string
	}{
		{
			name: "Good healthcheck",
			handler: &theHandler{
				config: config.Config{},
				healthcheck: func(requestId string, client *elasticsearch.Client, healthChecker kafka.HealthChecker) (int, *response.ErrorDetail) {
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
				healthcheck: func(requestId string, client *elasticsearch.Client, healthChecker kafka.HealthChecker) (int, *response.ErrorDetail) {
					return http.StatusServiceUnavailable, response.NewErrorDetail(requestId, "Elastic not available")
				},
			},
			expectedCode: http.StatusServiceUnavailable,
			expectedBody: "{\"errorEventId\":\"" + requestId + "\",\"errorDescription\":\"Elastic not available\"}\n",
		},
		{
			name: "Elastic client error",
			handler: &theHandler{
				config:      config.Config{ElasticUrl: "https://an.invalid url.com/", ElasticCert: "Invalid Cert"},
				healthcheck: nil,
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"" + requestId + "\",\"errorDescription\":\"cannot create client: cannot parse url: parse \\\"https://an.invalid url.com\\\": invalid character \\\" \\\" in host name\"}\n",
		},
		{
			name: "Kafka client error",
			handler: &theHandler{
				config:      config.Config{KafkaProperties: config.StringMap{"message.max.bytes": "bad_value"}},
				healthcheck: nil,
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

			if assert.NoError(t, tt.handler.Healthcheck(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}*/
