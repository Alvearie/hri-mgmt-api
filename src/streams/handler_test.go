/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package streams

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

const (
	validToken           = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNjUyMTA4MTQ0LCJleHAiOjI1NTIxMTE3NDR9.XxTTNBtgjX48iCM4FaV_hhhGenzhzrUaTWn6ooepK14" // expires in 2050
	expiredBearerToken   = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNDUyMTA4MTQ0LCJleHAiOjE1NTIxMTE3NDR9.JCYxVQmkSoHtmcpl_AjIH_SD2fDDQvldwYyCU0xQcYw"
	malformedBearerToken = "WQiOiJJQk1pZC0yNzAwMDdEMEhXIiwiaWQiOiJJQk1pZC0yNzAwMDdEMEhXIiwicmVhbG1pZCI6IklCTWlkIiwic2Vzc2lvbl9pZCI6IkMtYmUxODY1MjUtZWU0Yy00YWU1LWI3NGYtZjMyMTZlYjIxNWRhIiwic2Vzc2lvbl9leHBfbWF4IjoxNjUxODU5NTM2LCJzZXNzaW9uX2V4cF9uZXh0IjoxNjUxNzg0NDQyLCJqdGkiOiJiY2QI6IkJBWFRFUiIsIm5hbWUiOiJEQU4gQkFYVEVSIiwiZW1haWwiOiJkamJheHRlckB1cy5pYm0uY29tIiwic3ViIjoiZGpiYXh0ZXJAdXMuaWJtLmNvbSIsImF1dGhuIjp7InN1YiI6ImRqYmF4dGVyQHVzLmlibS5jb20iLCJpYW1faWQiOiJJQk1pZC0yNzAwMDdEMEhXIiwibmFtZSI6IkRBTiBCQVhURVIiLCJnaXZlbl9uYW1lIjoiREFOIiwiZmFtaWx5X25hbWUiOiJCQVhURVIiLCJlbhbGlkIjp0cnVlLCJic3MiOiI1MjM2NmM5YWIyMTQ0MDJmOWU5NjkxN2IxYjI4NTBlOSIsImltc191c2VyX2lkIjoiOTA1MzM5MiIsImZyb3plbiI6dHJ1ZSwiaW1zIjoiMjI5MzE0MiJ9LCJtZmEiOnsiaW1zIjp0cnVlfSwiaWF0IjoxNjUxNZSI6ImlibSBvcGVuaWQiLCJjbGllbnRfaWQiOiJieCIsImFjciI6MYeLCxvzTzjd9aacHnm6TNjMeMX5U3OVOdC_enTW7WXVUNWcTVRb8"

	expiredTokenErrMsg   = "Must supply an unexpired token:"
	malformedTokenErrMsg = "unexpected error parsing bearer token, could not parse jwt token"
)

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
	assert.Equal(t, reflect.ValueOf(Create), reflect.ValueOf(handler.create))
	assert.Equal(t, reflect.ValueOf(Delete), reflect.ValueOf(handler.delete))
	assert.Equal(t, reflect.ValueOf(Get), reflect.ValueOf(handler.get))
}

func TestHandlerCreate(t *testing.T) {
	validRequest := `{
		  "numPartitions": 1,
	      "retentionMs": 3600000
		}`

	logwrapper.Initialize("error", os.Stdout)

	tests := []struct {
		name                 string
		handler              theHandler
		request              string
		tenantId             string
		streamId             string
		bearerTokens         []string
		expectedCode         int
		expectedBody         string
		expectedCreateTopics []string
		createErrMessage     string
		createReturnCode     int
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
				},
			},
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{validToken},
			expectedCode: http.StatusCreated,
			expectedBody: `{"id":"stream_id"}`,
		},
		{
			name: "failed create with no auth token",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
				},
			},
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"missing header 'Authorization'"}`,
		},
		{
			name: "failed create with expired auth token",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
				},
			},
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{expiredBearerToken},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"` + expiredTokenErrMsg + `.*`,
		},
		{
			name: "failed create with malformed auth token",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
				},
			},
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{malformedBearerToken},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"` + malformedTokenErrMsg + `"}`,
		},
		{
			name: "failed with bad tenant id",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
				},
			},
			tenantId:     "INVALID",
			streamId:     "stream_id",
			bearerTokens: []string{validToken},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- tenantId \(url path parameter\) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'"}`,
		},
		{
			name: "failed with bad stream id",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
				},
			},
			tenantId:     "tenant_id",
			streamId:     "INVALID",
			bearerTokens: []string{validToken},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- id \(url path parameter\) may only contain lower-case alpha-numeric characters, no more than one '.', and the following 2 special chars: '-', '_'"}`,
		},
		{
			name: "failed with invalid request fields",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
				},
			},
			request:      `{"cleanupPolicy": "bogus"}`,
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{validToken},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- cleanupPolicy \(json field in request body\) must be one of \[delete compact\]\\n- numPartitions \(json field in request body\) is a required field\\n- retentionMs \(json field in request body\) is a required field"}`,
		},
		{
			name: "failed with invalid json",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					return []string{"in", "out", "invalid", "notification"}, http.StatusCreated, nil
				},
			},
			request:      `{`,
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{validToken},
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"unable to parse request body due to unexpected EOF"}`,
		},
		{
			name: "create fails and topic deletion succeeds",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					message := "create failure message"
					return []string{"in", "out"}, http.StatusInternalServerError, fmt.Errorf(message)
				},
			},
			expectedCreateTopics: []string{"in", "out"},
			createReturnCode:     http.StatusOK,
			tenantId:             "tenant_id",
			streamId:             "stream_id",
			bearerTokens:         []string{validToken},
			expectedCode:         http.StatusInternalServerError,
			expectedBody:         `{"errorEventId":"test-request-id","errorDescription":"create failure message"}`,
		},
		{
			name: "create fails and topic deletion fails",
			handler: theHandler{
				config: config.Config{},
				create: func(model.CreateStreamsRequest, string, string, bool, string, kafka.KafkaAdmin) ([]string, int, error) {
					message := "create failure message"
					return []string{"in", "out"}, http.StatusInternalServerError, fmt.Errorf(message)
				},
			},
			expectedCreateTopics: []string{"in", "out"},
			createReturnCode:     http.StatusInternalServerError,
			createErrMessage:     "delete failure message",
			tenantId:             "tenant_id",
			streamId:             "stream_id",
			bearerTokens:         []string{validToken},
			expectedCode:         http.StatusInternalServerError,
			expectedBody:         `{"errorEventId":"test-request-id","errorDescription":"create failure message\\ndelete failure message"}`,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestBody := validRequest
			if tt.request != "" {
				requestBody = tt.request
			}
			request := httptest.NewRequest(http.MethodPost, "/hri/tenant/test/streams/streamId", strings.NewReader(requestBody))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			for _, token := range tt.bearerTokens {
				request.Header.Add(echo.HeaderAuthorization, token)
			}
			context.SetPath("/tenants/:tenantId/streams/:id")
			context.SetParamNames(param.TenantId, param.StreamId)
			context.SetParamValues(tt.tenantId, tt.streamId)
			context.Response().Header().Add(echo.HeaderXRequestID, "test-request-id")

			if tt.createReturnCode != 0 {
				tt.handler.delete = func(requestId string, topicsToCreate []string, service kafka.KafkaAdmin) (int, error) {
					if !reflect.DeepEqual(topicsToCreate, tt.expectedCreateTopics) {
						t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tt.expectedCreateTopics, topicsToCreate))
					}

					if tt.createErrMessage == "" {
						return tt.createReturnCode, nil
					}

					return tt.createReturnCode, fmt.Errorf(tt.createErrMessage)
				}
			}

			if assert.NoError(t, tt.handler.Create(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				actualBody := strings.Trim(recorder.Body.String(), "\n")
				matched, _ := regexp.MatchString(tt.expectedBody, actualBody)
				if !matched {
					t.Errorf("Returned body did not match expected.\nExpected: %s, Actual: %s", tt.expectedBody, actualBody)
				}
			}
		})
	}
}

func TestHandlerDelete(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	tests := []struct {
		name                string
		handler             theHandler
		tenantId            string
		streamId            string
		expectedStreamNames []string
		bearerTokens        []string
		expectedCode        int
		expectedBody        string
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: config.Config{
					Validation: true,
				},
			},
			tenantId: "tenant_id",
			streamId: "stream_id",
			expectedStreamNames: []string{
				"ingest.tenant_id.stream_id.in",
				"ingest.tenant_id.stream_id.notification",
				"ingest.tenant_id.stream_id.out",
				"ingest.tenant_id.stream_id.invalid",
			},
			bearerTokens: []string{validToken},
			expectedCode: http.StatusOK,
		},
		{
			name: "happy path without validation",
			handler: theHandler{
				config: config.Config{
					Validation: false,
				},
			},
			tenantId: "tenant_id",
			streamId: "stream_id",
			expectedStreamNames: []string{
				"ingest.tenant_id.stream_id.in",
				"ingest.tenant_id.stream_id.notification",
			},
			bearerTokens: []string{validToken},
			expectedCode: http.StatusOK,
		},
		{
			name:         "failed delete with no auth token",
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"missing header 'Authorization'"}`,
		},
		{
			name:         "failed delete with expired auth token",
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{expiredBearerToken},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"` + expiredTokenErrMsg + `.*`,
		},
		{
			name:         "failed delete with malformed auth token",
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{malformedBearerToken},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"` + malformedTokenErrMsg + `"}`,
		},
		{
			name: "failed with empty tenant id",
			handler: theHandler{
				config: config.Config{
					Validation: true,
				},
			},
			tenantId:            "",
			streamId:            "stream_id",
			expectedStreamNames: []string{},
			bearerTokens:        []string{validToken},
			expectedCode:        http.StatusBadRequest,
			expectedBody:        `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- tenantId \(url path parameter\) is a required field"}`,
		},
		{
			name: "failed with empty stream id",
			handler: theHandler{
				config: config.Config{
					Validation: true,
				},
			},
			tenantId:            "tenant_id",
			streamId:            "",
			expectedStreamNames: []string{},
			bearerTokens:        []string{validToken},
			expectedCode:        http.StatusBadRequest,
			expectedBody:        `{"errorEventId":"test-request-id","errorDescription":"invalid request arguments:\\n- id \(url path parameter\) is a required field"}`,
		},
		{
			name: "delete failed",
			handler: theHandler{
				config: config.Config{
					Validation: true,
				},
				delete: func(string, []string, kafka.KafkaAdmin) (int, error) {
					message := "delete failure message"
					return http.StatusInternalServerError, fmt.Errorf(message)
				},
			},
			tenantId:     "tenant_id",
			streamId:     "stream_id",
			bearerTokens: []string{validToken},
			expectedCode: http.StatusInternalServerError,
			expectedBody: `{"errorEventId":"test-request-id","errorDescription":"delete failure message"}`,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/hri/tenant/test/streams/streamId", nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			for _, token := range tt.bearerTokens {
				request.Header.Add(echo.HeaderAuthorization, token)
			}
			context.SetPath("/tenants/:tenantId/streams/:id")
			context.SetParamNames(param.TenantId, param.StreamId)
			context.SetParamValues(tt.tenantId, tt.streamId)
			context.Response().Header().Add(echo.HeaderXRequestID, "test-request-id")

			if tt.handler.delete == nil {
				// The handler wasn't mocked in the test case. Assert that the proper arguments were sent to
				// the delete handler and return a 200.
				tt.handler.delete = func(requestId string, actualStreamNames []string, service kafka.KafkaAdmin) (int, error) {
					assert.NotNil(t, service)

					if !reflect.DeepEqual(actualStreamNames, tt.expectedStreamNames) {
						t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tt.expectedStreamNames, actualStreamNames))
					}

					return http.StatusOK, nil
				}
			}

			if assert.NoError(t, tt.handler.Delete(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				actualBody := strings.Trim(recorder.Body.String(), "\n")
				matched, _ := regexp.MatchString(tt.expectedBody, actualBody)
				if !matched {
					t.Errorf("Returned body did not match expected.\nExpected: %s, Actual: %s", tt.expectedBody, actualBody)
				}
			}
		})
	}
}

func TestHandlerGet(t *testing.T) {
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
				config: config.Config{},
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
				config: config.Config{},
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
			name:         "failed delete with expired auth token",
			tenantId:     "tenant_id",
			bearerTokens: []string{expiredBearerToken},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errorEventId":"req42","errorDescription":"` + expiredTokenErrMsg + `.*`,
		},
		{
			name:         "failed delete with malformed auth token",
			tenantId:     "tenant_id",
			bearerTokens: []string{malformedBearerToken},
			expectedCode: http.StatusUnauthorized,
			expectedBody: `{"errorEventId":"req42","errorDescription":"` + malformedTokenErrMsg + `"}`,
		},
		{
			name: "Return Bad Request for missing TenantId Param ",
			handler: theHandler{
				config: config.Config{},
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
				config: config.Config{},
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
}
