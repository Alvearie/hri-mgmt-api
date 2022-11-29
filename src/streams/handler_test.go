/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package streams

// import (
// 	"github.com/Alvearie/hri-mgmt-api/common/response"
// )

// const (
// 	validToken   = "BEaRer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNjUyMTA4MTQ0LCJleHAiOjI1NTIxMTE3NDR9.XxTTNBtgjX48iCM4FaV_hhhGenzhzrUaTWn6ooepK14" // expires in 2050
// 	validAztoken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IjJaUXBKM1VwYmpBWVhZR2FYRUpsOGxWMFRPSSIsImtpZCI6IjJaUXBKM1VwYmpBWVhZR2FYRUpsOGxWMFRPSSJ9.eyJhdWQiOiJjMzNhYzRkYS0yMWM2LTQyNmItYWJjYy0yN2UyNGZmMWNjZjkiLCJpc3MiOiJodHRwczovL3N0cy53aW5kb3dzLm5ldC9jZWFhNjNhYS01ZDVjLTRjN2QtOTRiMC0wMmY5YTNhYjZhOGMvIiwiaWF0IjoxNjYzNzQyMTM0LCJuYmYiOjE2NjM3NDIxMzQsImV4cCI6MTY2Mzc0NjAzNCwiYWlvIjoiRTJaZ1lGaHdablhvSG84elJvOHpQUlpNMU9VNENBQT0iLCJhcHBpZCI6ImMzM2FjNGRhLTIxYzYtNDI2Yi1hYmNjLTI3ZTI0ZmYxY2NmOSIsImFwcGlkYWNyIjoiMSIsImlkcCI6Imh0dHBzOi8vc3RzLndpbmRvd3MubmV0L2NlYWE2M2FhLTVkNWMtNGM3ZC05NGIwLTAyZjlhM2FiNmE4Yy8iLCJvaWQiOiI4YjFlN2E4MS03ZjRhLTQxYjAtYTE3MC1hZTE5Zjg0M2YyN2MiLCJyaCI6IjAuQVZBQXFtT3F6bHhkZlV5VXNBTDVvNnRxak5yRU9zUEdJV3RDcTh3bjRrX3h6UGxfQUFBLiIsInJvbGVzIjpbImhyaS5ocmlfaW50ZXJuYWwiLCJ0ZW5hbnRfcHJvdmlkZXIxMjM0IiwidGVuYW50X3BlbnRlc3QiLCJ0ZXN0X3JvbGUiLCJ0ZXN0IiwiaHJpX2NvbnN1bWVyIiwiaHJpX2RhdGFfaW50ZWdyYXRvciIsInByb3ZpZGVyMTIzNCJdLCJzdWIiOiI4YjFlN2E4MS03ZjRhLTQxYjAtYTE3MC1hZTE5Zjg0M2YyN2MiLCJ0aWQiOiJjZWFhNjNhYS01ZDVjLTRjN2QtOTRiMC0wMmY5YTNhYjZhOGMiLCJ1dGkiOiJnaUJlZUliWk9rS0ZYbGFIaHNfZ0FBIiwidmVyIjoiMS4wIn0.LdwhQpf5M1LSprQ9gk9abisbucKhNQtDnYEN1GLw_SqJ23DIFlfevlLikw075rVYvwf-4p_MJN3-7QZ2gMzTsqQ-G2x9IH4BO-oULlXeoHBQllDtmnYQFEesGogM0OjtXvoIAzUXCTPyxbjzTX3sPvghXuCSWPfu9ehVn8mRVtXuH0LWaU47XjTYzDE-RIFM2S80UCv7ZQErLrshC91OI0rNyc8ARPEc-TlnIK-KQ8HgehjFaapO6VL15s3YLO0zGA1v4RLnxbd36SdFfGxE_Vlv7WSLR5nB_n403FbiUUpwdIORaFRdMBEtNDbuI2RwHesUIEL6lrBrDxXPuaLIsA"
// )

// // Fake for the auth.Validator interface; just returns the desired values
// type fakeAuthValidator struct {
// 	errResp *response.ErrorDetailResponse
// }

// func (f fakeAuthValidator) GetValidatedClaimsForTenant(_ string, _ string) *response.ErrorDetailResponse {
// 	return f.errResp
// }

/*func TestNewHandler(t *testing.T) {
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
	assert.Equal(t, reflect.ValueOf(Create), reflect.ValueOf(handler.create))
	assert.Equal(t, reflect.ValueOf(Delete), reflect.ValueOf(handler.delete))
	assert.Equal(t, reflect.ValueOf(Get), reflect.ValueOf(handler.get))
}*/
// func TestHandlerCreateStream(t *testing.T) {
// 	validConfig := config.Config{
// 		/*KafkaProperties: map[string]string{
// 			 "security.protocol": "sasl_ssl",
// 		 },*/
// 	}

// 	validRequest := `{
// 		   "numPartitions": 1,
// 		   "retentionMs": 3600000
// 		 }`

// 	logwrapper.Initialize("error", os.Stdout)

// 	tests := []struct {
// 		name                 string
// 		handler              theHandler
// 		request              string
// 		tenantId             string
// 		streamId             string
// 		bearerTokens         []string
// 		expectedCode         int
// 		expectedBody         string
// 		expectedCreateTopics []string
// 		createErrMessage     string
// 		createReturnCode     int
// 	}{
// 		{
// 			name: "happy path",
// 			handler: theHandler{
// 				config: validConfig,
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 				createStream: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
// 					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
// 				},
// 			},
// 			tenantId:     "tenant_id",
// 			streamId:     "stream_id",
// 			bearerTokens: []string{validToken},
// 			expectedCode: http.StatusCreated,
// 			expectedBody: `{"id":"stream_id"}`,
// 		},
// 		{
// 			name: "failed create with no auth token",
// 			handler: theHandler{
// 				config: validConfig,
// 				createStream: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
// 					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
// 				},
// 			},
// 			tenantId:     "tenant_id",
// 			streamId:     "stream_id",
// 			bearerTokens: []string{},
// 			expectedCode: http.StatusUnauthorized,
// 			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"missing header 'Authorization'"}`,
// 		},
// 		{
// 			name: "failed with bad tenant id",
// 			handler: theHandler{
// 				config: validConfig,
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 				createStream: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
// 					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
// 				},
// 			},
// 			tenantId:     "INVALID",
// 			streamId:     "stream_id",
// 			bearerTokens: []string{validToken},
// 			expectedCode: http.StatusBadRequest,
// 			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- tenantId \(url path parameter\) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'"}`,
// 		},
// 		{
// 			name: "failed with bad stream id",
// 			handler: theHandler{
// 				config: validConfig,
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 				createStream: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
// 					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
// 				},
// 			},
// 			tenantId:     "tenant_id",
// 			streamId:     "INVALID",
// 			bearerTokens: []string{validToken},
// 			expectedCode: http.StatusBadRequest,
// 			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- id \(url path parameter\) may only contain lower-case alpha-numeric characters, no more than one '.', and the following 2 special chars: '-', '_'"}`,
// 		},
// 		{
// 			name: "failed with invalid request fields",
// 			handler: theHandler{
// 				config: validConfig,
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 				createStream: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
// 					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
// 				},
// 			},
// 			request:      `{"cleanupPolicy": "bogus"}`,
// 			tenantId:     "tenant_id",
// 			streamId:     "stream_id",
// 			bearerTokens: []string{validToken},
// 			expectedCode: http.StatusBadRequest,
// 			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- cleanupPolicy \(json field in request body\) must be one of \[delete compact\]\\n- numPartitions \(json field in request body\) is a required field\\n- retentionMs \(json field in request body\) is a required field"}`,
// 		},
// 		{
// 			name: "failed with invalid json",
// 			handler: theHandler{
// 				config: validConfig,
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 				createStream: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
// 					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
// 				},
// 			},
// 			request:      `{`,
// 			tenantId:     "tenant_id",
// 			streamId:     "stream_id",
// 			bearerTokens: []string{validToken},
// 			expectedCode: http.StatusBadRequest,
// 			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"unable to parse request body due to unexpected EOF"}`,
// 		},
// 		{
// 			name: "create fails and topic deletion succeeds",
// 			handler: theHandler{
// 				config: validConfig,
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 				createStream: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
// 					message := "create failure message"
// 					return []string{"in", "out"}, http.StatusInternalServerError, fmt.Errorf(message)
// 				},
// 			},
// 			expectedCreateTopics: []string{"in", "out"},
// 			createReturnCode:     http.StatusOK,
// 			tenantId:             "tenant_id",
// 			streamId:             "stream_id",
// 			bearerTokens:         []string{validToken},
// 			expectedCode:         http.StatusInternalServerError,
// 			expectedBody:         `{"errorEventId":"test-request-id","errorDescription":"create failure message"}`,
// 		},
// 		{
// 			name: "create fails and topic deletion fails",
// 			handler: theHandler{
// 				config: validConfig,
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 				createStream: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
// 					message := "create failure message"
// 					return []string{"in", "out"}, http.StatusInternalServerError, fmt.Errorf(message)
// 				},
// 			},
// 			expectedCreateTopics: []string{"in", "out"},
// 			createReturnCode:     http.StatusInternalServerError,
// 			createErrMessage:     "delete failure message",
// 			tenantId:             "tenant_id",
// 			streamId:             "stream_id",
// 			bearerTokens:         []string{validToken},
// 			expectedCode:         http.StatusInternalServerError,
// 			expectedBody:         `{"errorEventId":"test-request-id","errorDescription":"create failure message\\ndelete failure message"}`,
// 		},
// 	}

// 	e := test.GetTestServer()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			requestBody := validRequest
// 			if tt.request != "" {
// 				requestBody = tt.request
// 			}
// 			request := httptest.NewRequest(http.MethodPost, "/hri/tenant/test/streams/streamId", strings.NewReader(requestBody))
// 			context, recorder := test.PrepareHeadersContextRecorder(request, e)
// 			for _, token := range tt.bearerTokens {
// 				request.Header.Add(echo.HeaderAuthorization, token)
// 			}
// 			context.SetPath("/tenants/:tenantId/streams/:id")
// 			context.SetParamNames(param.TenantId, param.StreamId)
// 			context.SetParamValues(tt.tenantId, tt.streamId)
// 			context.Response().Header().Add(echo.HeaderXRequestID, "test-request-id")

// 			if tt.createReturnCode != 0 {
// 				tt.handler.deleteStream = func(requestId string, topicsToCreate []string, service kafka.KafkaAdmin) (int, error) {
// 					if !reflect.DeepEqual(topicsToCreate, tt.expectedCreateTopics) {
// 						t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tt.expectedCreateTopics, topicsToCreate))
// 					}

// 					if tt.createErrMessage == "" {
// 						return tt.createReturnCode, nil
// 					}

// 					return tt.createReturnCode, fmt.Errorf(tt.createErrMessage)
// 				}
// 			}

// 			if assert.NoError(t, tt.handler.CreateStream(context)) {
// 				assert.Equal(t, tt.expectedCode, recorder.Code)
// 				actualBody := strings.Trim(recorder.Body.String(), "\n")
// 				matched, _ := regexp.MatchString(tt.expectedBody, actualBody)
// 				if !matched {
// 					t.Errorf("Returned body did not match expected.\nExpected: %s, Actual: %s", tt.expectedBody, actualBody)
// 				}
// 			}
// 		})
// 	}
// }
// func TestHandlerDeleteStream(t *testing.T) {

// 	logwrapper.Initialize("error", os.Stdout)

// 	tests := []struct {
// 		name                string
// 		handler             theHandler
// 		tenantId            string
// 		streamId            string
// 		expectedStreamNames []string
// 		bearerTokens        []string
// 		expectedCode        int
// 		expectedBody        string
// 	}{
// 		{
// 			name: "happy path",
// 			handler: theHandler{
// 				config: config.Config{
// 					Validation: true,
// 					//KafkaProperties: kafkaProperties,
// 				},
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 			},
// 			tenantId: "tenant_id",
// 			streamId: "stream_id",
// 			expectedStreamNames: []string{
// 				"ingest.tenant_id.stream_id.in",
// 				"ingest.tenant_id.stream_id.notification",
// 				"ingest.tenant_id.stream_id.out",
// 				"ingest.tenant_id.stream_id.invalid",
// 			},
// 			bearerTokens: []string{validAztoken},
// 			expectedCode: http.StatusOK,
// 		},
// 		{
// 			name: "Without bearer token",
// 			handler: theHandler{
// 				config: config.Config{
// 					Validation: true,
// 					//KafkaProperties: kafkaProperties,
// 				},
// 			},
// 			tenantId: "tenant_id",
// 			streamId: "stream_id",
// 			expectedStreamNames: []string{
// 				"ingest.tenant_id.stream_id.in",
// 				"ingest.tenant_id.stream_id.notification",
// 				"ingest.tenant_id.stream_id.out",
// 				"ingest.tenant_id.stream_id.invalid",
// 			},
// 			bearerTokens: []string{},
// 			expectedCode: http.StatusUnauthorized,
// 		},

// 		{
// 			name: "failed with empty tenant id",
// 			handler: theHandler{
// 				config: config.Config{
// 					Validation: true,
// 					//KafkaProperties: kafkaProperties,
// 				},
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 			},
// 			tenantId:            "",
// 			streamId:            "stream_id",
// 			expectedStreamNames: []string{},
// 			bearerTokens:        []string{validToken},
// 			expectedCode:        http.StatusBadRequest,
// 			expectedBody:        `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- tenantId \(url path parameter\) is a required field"}`,
// 		},
// 		{
// 			name: "failed with empty stream id",
// 			handler: theHandler{
// 				config: config.Config{
// 					Validation: true,
// 					//KafkaProperties: kafkaProperties,
// 				},
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 			},
// 			tenantId:            "tenant_id",
// 			streamId:            "",
// 			expectedStreamNames: []string{},
// 			bearerTokens:        []string{validToken},
// 			expectedCode:        http.StatusBadRequest,
// 			expectedBody:        `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- id \(url path parameter\) is a required field"}`,
// 		},
// 		{
// 			name: "delete failed",
// 			handler: theHandler{
// 				config: config.Config{
// 					Validation: true,
// 					//KafkaProperties: kafkaProperties,
// 				},
// 				jwtValidator: fakeAuthValidator{
// 					errResp: nil,
// 				},
// 				deleteStream: func(string, []string, kafka.KafkaAdmin) (int, error) {
// 					message := "delete failure message"
// 					return http.StatusInternalServerError, fmt.Errorf(message)
// 				},
// 			},
// 			tenantId:     "tenant_id",
// 			streamId:     "stream_id",
// 			bearerTokens: []string{validToken},
// 			expectedCode: http.StatusInternalServerError,
// 			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"delete failure message"}`,
// 		},
// 	}

// 	e := test.GetTestServer()
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			request := httptest.NewRequest(http.MethodDelete, "/hri/tenant/test/streams/streamId", nil)
// 			context, recorder := test.PrepareHeadersContextRecorder(request, e)
// 			for _, token := range tt.bearerTokens {
// 				request.Header.Add(echo.HeaderAuthorization, token)
// 			}
// 			context.SetPath("/tenants/:tenantId/streams/:id")
// 			context.SetParamNames(param.TenantId, param.StreamId)
// 			context.SetParamValues(tt.tenantId, tt.streamId)
// 			context.Response().Header().Add(echo.HeaderXRequestID, "test-request-id")

// 			if tt.handler.deleteStream == nil {
// 				// The handler wasn't mocked in the test case. Assert that the proper arguments were sent to
// 				// the delete handler and return a 200.
// 				tt.handler.deleteStream = func(requestId string, actualStreamNames []string, service kafka.KafkaAdmin) (int, error) {
// 					assert.NotNil(t, service)

// 					if !reflect.DeepEqual(actualStreamNames, tt.expectedStreamNames) {
// 						t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tt.expectedStreamNames, actualStreamNames))
// 					}

// 					return http.StatusOK, nil
// 				}
// 			}

// 			if assert.NoError(t, tt.handler.DeleteStream(context)) {
// 				assert.Equal(t, tt.expectedCode, recorder.Code)
// 				actualBody := strings.Trim(recorder.Body.String(), "\n")
// 				matched, _ := regexp.MatchString(tt.expectedBody, actualBody)
// 				if !matched {
// 					t.Errorf("Returned body did not match expected.\nExpected: %s, Actual: %s", tt.expectedBody, actualBody)
// 				}
// 			}
// 		})
// 	}
// }

/*func TestHandlerGet(t *testing.T) {
	var validTenantId = "tenantA3"
	var requestId = "req42"
	var streamId1 = "CountChocula.qualifier330"
	var streamId2NoQualifier = "PorcupinePaul"
	var streamId3 = "dataIntegrator887"
	goodRequestStreams := []map[string]interface{}{
		{param.StreamId: streamId1},
		{param.StreamId: streamId2NoQualifier},
		{param.StreamId: streamId3},
	}
	emptyStreamsResults := []map[string]interface{}{}
	validConfig := config.Config{
		KafkaProperties: map[string]string{
			"security.protocol": "sasl_ssl",
		},
	}
	logwrapper.Initialize("error", os.Stdout)

	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		bearerTokens []string
		expectedCode int
		expectedBody string
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: validConfig,
				get: func(string, string, kafka.KafkaAdmin) (int, interface{}) {
					return http.StatusOK, goodRequestStreams
				},
			},
			tenantId:     validTenantId,
			bearerTokens: []string{validToken},
			expectedCode: http.StatusOK,
			expectedBody: `\[{"id":"CountChocula.qualifier330"},{"id":"PorcupinePaul"},{"id":"dataIntegrator887"}\]`,
		},
		{
			name: "list streams returns no results",
			handler: theHandler{
				config: validConfig,
				get: func(string, string, kafka.KafkaAdmin) (int, interface{}) {
					return http.StatusOK, emptyStreamsResults
				},
			},
			tenantId:     validTenantId,
			bearerTokens: []string{validToken},
			expectedCode: http.StatusOK,
			expectedBody: `\[\]`,
		},
		{
			name:         "failed delete with no auth token",
			tenantId:     "tenant_id",
			bearerTokens: []string{},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errorEventId":"req42","errorDescription":"missing header 'Authorization'"}`,
		},
		{
			name: "Return Bad Request for missing TenantId Param ",
			handler: theHandler{
				config: validConfig,
				get: func(string, string, kafka.KafkaAdmin) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			bearerTokens: []string{validToken},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errorEventId":"req42","errorDescription":"invalid request arguments:\\n- tenantId \(url path parameter\) is a required field"}`,
		},
		{
			name: "List streams function call fails",
			handler: theHandler{
				config: validConfig,
				get: func(string, string, kafka.KafkaAdmin) (int, interface{}) {
					return http.StatusInternalServerError,
						response.NewErrorDetail(requestId, "Error List Streams: Unable to connect to Kafka")
				},
			},
			tenantId:     validTenantId,
			bearerTokens: []string{validToken},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"errorEventId":"req42","errorDescription":"Error List Streams: Unable to connect to Kafka"}`,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			for _, token := range tt.bearerTokens {
				request.Header.Add(echo.HeaderAuthorization, token)
			}
			context.SetPath("/tenants/:tenantId/streams")
			context.SetParamNames(param.TenantId)
			context.SetParamValues(tt.tenantId)
			context.Response().Header().Add(echo.HeaderXRequestID, requestId)

			if assert.NoError(t, tt.handler.Get(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				actualBody := strings.Trim(recorder.Body.String(), "\n")
				matched, _ := regexp.MatchString(tt.expectedBody, actualBody)
				if !matched {
					t.Errorf("Returned body did not match expected.\nExpected: %s, Actual: %s", tt.expectedBody, actualBody)
				}
			}
		})
	}
}*/
