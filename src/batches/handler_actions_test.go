/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package batches

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

type fakeAction struct {
	t               *testing.T
	expectedRequest interface{}
	expectedStatus  status.BatchStatus
	code            int
	body            interface{}
}

func (fake fakeAction) sendComplete(_ string, request *model.SendCompleteRequest, _ auth.HriClaims, _ *elasticsearch.Client, _ kafka.Writer, currentStatus status.BatchStatus) (int, interface{}) {
	if !reflect.DeepEqual(fake.expectedRequest, request) {
		fake.t.Errorf("Request is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedRequest, request)
	}
	if fake.expectedStatus != currentStatus {
		fake.t.Errorf("Current Batch Status is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedStatus, currentStatus)
	}
	return fake.code, fake.body
}

func (fake fakeAction) terminate(_ string, request *model.TerminateRequest, _ auth.HriClaims, _ *elasticsearch.Client, _ kafka.Writer, currentStatus status.BatchStatus) (int, interface{}) {
	if !reflect.DeepEqual(fake.expectedRequest, request) {
		fake.t.Errorf("Request is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedRequest, request)
	}
	if fake.expectedStatus != currentStatus {
		fake.t.Errorf("Current Batch Status is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedStatus, currentStatus)
	}
	return fake.code, fake.body
}

func (fake fakeAction) processingComplete(_ string, request *model.ProcessingCompleteRequest, _ auth.HriClaims, _ *elasticsearch.Client, _ kafka.Writer, currentStatus status.BatchStatus) (int, interface{}) {
	if !reflect.DeepEqual(fake.expectedRequest, request) {
		fake.t.Errorf("Request is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedRequest, request)
	}
	if fake.expectedStatus != currentStatus {
		fake.t.Errorf("Current Batch Status is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedStatus, currentStatus)
	}
	return fake.code, fake.body
}

func (fake fakeAction) fail(_ string, request *model.FailRequest, _ auth.HriClaims, _ *elasticsearch.Client, _ kafka.Writer, currentStatus status.BatchStatus) (int, interface{}) {
	if !reflect.DeepEqual(fake.expectedRequest, request) {
		fake.t.Errorf("Request is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedRequest, request)
	}
	if fake.expectedStatus != currentStatus {
		fake.t.Errorf("Current Batch Status is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedStatus, currentStatus)
	}
	return fake.code, fake.body
}

const topicBase = "awesomeTopic"
const defaultTenantId = test.ValidTenantId
const defaultBatchId = test.ValidBatchId
const defaultBatchName = "monkeeBatch2"
const defaultBatchDataType = "claims"
const defaultTopicBase = "awesomeTopic"
const defaultInputTopic = topicBase + inputSuffix
const defaultBatchStatus = status.Started //Started

func Test_theHandler_SendComplete(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	var configNoValidation = createDefaultTestConfig()
	var configWithValidation = createDefaultTestConfig()
	configWithValidation.Validation = true

	tenantId := test.ValidTenantId
	batchId := test.ValidBatchId
	batchName := defaultBatchName
	batchDataType := defaultBatchDataType
	topicBase := defaultTopicBase
	inputTopic := topicBase + inputSuffix
	defaultStatus := status.Started //Started

	returnBatch := map[string]interface{}{
		"id":                  batchId,
		"name":                batchName,
		"status":              defaultStatus.String(),
		"startDate":           "2019-12-13",
		"dataType":            batchDataType,
		"topic":               inputTopic,
		"recordCount":         float64(1),
		"expectedRecordCount": float64(1)}

	tests := []struct {
		name                   string
		tenantId               string
		batchId                string
		handler                theHandler
		requestBody            string
		expectedCode           int
		expectedBody           string
		expectedGetByIdRequest model.GetByIdBatch
		expectedGetByIdCode    int
		expectedGetByIdBody    interface{} //expected Return for getById() call to retrieve batch.status
	}{
		{
			name:     "success with expectedRecordCount and no Validation",
			tenantId: tenantId,
			batchId:  batchId,
			handler: theHandler{
				config: configNoValidation,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				sendComplete: fakeAction{
					t:               t,
					expectedRequest: getTestSendCompleteRequest(intPtr(100), nil, nil, false),
					expectedStatus:  defaultStatus,
					code:            http.StatusOK,
					body:            nil,
				}.sendComplete,
			},
			requestBody:            `{"expectedRecordCount": 100}`,
			expectedCode:           http.StatusOK,
			expectedBody:           "",
			expectedGetByIdRequest: getTestGetByIdBatch(tenantId, batchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
		{
			name:     "success with recordCount, metadata, and Validation",
			tenantId: tenantId,
			batchId:  batchId,
			handler: theHandler{
				config: configWithValidation,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				sendComplete: fakeAction{
					t: t,
					expectedRequest: getTestSendCompleteRequest(
						nil,
						intPtr(100),
						map[string]interface{}{"field1": "field1", "field2": float64(10)},
						true,
					),
					expectedStatus: defaultStatus,
					code:           http.StatusOK,
					body:           nil,
				}.sendComplete,
			},
			requestBody:            `{"recordCount": 100, "metadata":{"field1":"field1","field2":10}}`,
			expectedCode:           http.StatusOK,
			expectedBody:           "",
			expectedGetByIdRequest: getTestGetByIdBatch(tenantId, batchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
		{
			name:         "400 request binding failure",
			tenantId:     tenantId,
			batchId:      batchId,
			handler:      theHandler{},
			requestBody:  `{"expectedRecordCount": "100"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request param \"expectedRecordCount\": expected type int, but received type string"}`, requestId) + "\n",
		},
		{
			name:         "400 empty tenant and batch params",
			tenantId:     "",
			batchId:      "",
			handler:      theHandler{},
			requestBody:  `{"expectedRecordCount": 100}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request arguments:\n- id (url path parameter) is a required field\n- tenantId (url path parameter) is a required field"}`, requestId) + "\n",
		},
		{
			name:         "400 missing both record count fields",
			tenantId:     tenantId,
			batchId:      batchId,
			handler:      theHandler{},
			requestBody:  `{}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request arguments:\n- expectedRecordCount (json field in request body) must be present if recordCount (json field in request body) is not present\n- recordCount (json field in request body) must be present if expectedRecordCount (json field in request body) is not present"}`, requestId) + "\n",
		},
		{
			name:         "400 request validation failure",
			tenantId:     test.ValidTenantId,
			batchId:      test.ValidBatchId,
			handler:      theHandler{},
			requestBody:  `{"expectedRecordCount": -100}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request arguments:\n- expectedRecordCount (json field in request body) must be 0 or greater"}`, requestId) + "\n",
		},
		{
			name:     "401 unauthorized failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "missing tenant scope"),
				},
			},
			requestBody:  `{"expectedRecordCount": 100}`,
			expectedCode: http.StatusUnauthorized,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"missing tenant scope"}`, requestId) + "\n",
		},
		{
			name:     "500 elastic client failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: badEsConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
			},
			requestBody:  `{"expectedRecordCount": 100}`,
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"error getting Elastic client: cannot create client: cannot parse url: parse \"https:// a bad url.com\": invalid character \" \" in host name"}`, requestId) + "\n",
		},
		{
			name:     "500 sendCompete failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: configNoValidation,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				sendComplete: fakeAction{
					t:               t,
					expectedRequest: getTestSendCompleteRequest(intPtr(100), nil, nil, false),
					expectedStatus:  defaultStatus,
					code:            http.StatusInternalServerError,
					body:            response.NewErrorDetail(requestId, "something bad happened"),
				}.sendComplete,
			},
			requestBody:            `{"expectedRecordCount": 100}`,
			expectedCode:           http.StatusInternalServerError,
			expectedBody:           fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"something bad happened"}`, requestId) + "\n",
			expectedGetByIdRequest: getTestGetByIdBatch(tenantId, batchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.requestBody))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenant/:tenantId/batches/:batchId/action/sendComplete")
			context.SetParamNames(param.TenantId, param.BatchId)
			context.SetParamValues(tt.tenantId, tt.batchId)
			context.Response().Header().Add(echo.HeaderXRequestID, requestId)

			//Check for call to getById to retrieve Batch info (status)
			var expectGetByIdCall = tt.expectedGetByIdCode != 0
			getByIdCalled := false
			if expectGetByIdCall {
				tt.handler.getByIdNoAuth = func(_ string, requestBatch model.GetByIdBatch, _ auth.HriClaims, esClient *elasticsearch.Client) (int, interface{}) {
					getByIdCalled = true
					if !reflect.DeepEqual(requestBatch, tt.expectedGetByIdRequest) {
						t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tt.expectedGetByIdRequest, requestBatch))
					}

					return tt.expectedGetByIdCode, tt.expectedGetByIdBody
				}
			}

			if assert.NoError(t, tt.handler.SendComplete(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
				if expectGetByIdCall && !getByIdCalled {
					t.Error(fmt.Printf("Expected function call that did NOT happen: getById()"))
				}
			}
		})
	}
}

func Test_theHandler_SendCompleteNoAuth(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	testConfig.AuthDisabled = true

	returnBatch := map[string]interface{}{
		"id":                  defaultBatchId,
		"name":                batchName,
		"status":              defaultBatchStatus.String(),
		"startDate":           "2019-12-13",
		"dataType":            batchDataType,
		"topic":               defaultInputTopic,
		"recordCount":         float64(1),
		"expectedRecordCount": float64(1)}

	requestBody := `{"expectedRecordCount": 100}`
	expectedRespCode := http.StatusOK
	expectedRespBody := ""
	expectedGetByIdRequest := getTestGetByIdBatch(defaultTenantId, defaultBatchId)
	expectedGetByIdCode := http.StatusOK
	expectedGetByIdBody := returnBatch

	testName := "sendComplete-no-auth happy path with expectedRecordCount and no Validation"
	handler := theHandler{
		config: testConfig,
		sendComplete: fakeAction{
			t:               t,
			expectedRequest: getTestSendCompleteRequest(intPtr(100), nil, nil, false),
			expectedStatus:  defaultBatchStatus,
			code:            http.StatusOK,
			body:            nil,
		}.sendComplete,
	}

	e := test.GetTestServer()
	t.Run(testName, func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(requestBody))
		context, recorder := test.PrepareHeadersContextRecorder(request, e)
		context.SetPath("/hri/tenant/:tenantId/batches/:batchId/action/sendComplete")
		context.SetParamNames(param.TenantId, param.BatchId)
		context.SetParamValues(test.ValidTenantId, test.ValidBatchId)
		context.Response().Header().Add(echo.HeaderXRequestID, requestId)

		//Check for call to getById to retrieve Batch info (status)
		var expectGetByIdCall = expectedGetByIdCode != 0
		getByIdCalled := false
		if expectGetByIdCall {
			handler.getByIdNoAuth = func(_ string, requestBatch model.GetByIdBatch, _ auth.HriClaims, esClient *elasticsearch.Client) (int, interface{}) {
				getByIdCalled = true
				if !reflect.DeepEqual(requestBatch, expectedGetByIdRequest) {
					t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", expectedGetByIdRequest, requestBatch))
				}

				return expectedGetByIdCode, expectedGetByIdBody
			}
		}

		if assert.NoError(t, handler.SendComplete(context)) {
			assert.Equal(t, expectedRespCode, recorder.Code)
			assert.Equal(t, expectedRespBody, recorder.Body.String())
			if expectGetByIdCall && !getByIdCalled {
				t.Error(fmt.Printf("Expected function call that did NOT happen: getById()"))
			}
		}
	})
}

func Test_theHandler_Terminate(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	var defaultConfig = createDefaultTestConfig()
	const batchStatus = status.SendCompleted

	returnBatch := map[string]interface{}{
		"id":                  defaultBatchId,
		"name":                batchName,
		"status":              batchStatus.String(),
		"startDate":           "2019-12-13",
		"dataType":            batchDataType,
		"topic":               defaultInputTopic,
		"recordCount":         float64(1),
		"expectedRecordCount": float64(1)}

	tests := []struct {
		name                   string
		tenantId               string
		batchId                string
		handler                theHandler
		requestBody            string
		expectedCode           int
		expectedBody           string
		expectedGetByIdRequest model.GetByIdBatch
		expectedGetByIdCode    int
		expectedGetByIdBody    interface{} //expected Return for getById() call to retrieve batch.status
	}{
		{
			name:     "success without metadata",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: defaultConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				terminate: fakeAction{
					t:               t,
					expectedRequest: getTestTerminateRequest(nil),
					expectedStatus:  batchStatus,
					code:            http.StatusOK,
					body:            nil,
				}.terminate,
			},
			requestBody:            ``,
			expectedCode:           http.StatusOK,
			expectedBody:           "",
			expectedGetByIdRequest: getTestGetByIdBatch(defaultTenantId, defaultBatchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
		{
			name:     "success with metadata",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: defaultConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				terminate: fakeAction{
					t:               t,
					expectedRequest: getTestTerminateRequest(map[string]interface{}{"field1": "field1", "field2": float64(10)}),
					expectedStatus:  batchStatus,
					code:            http.StatusOK,
					body:            nil,
				}.terminate,
			},
			requestBody:            `{"metadata":{"field1":"field1","field2":10}}`,
			expectedCode:           http.StatusOK,
			expectedBody:           "",
			expectedGetByIdRequest: getTestGetByIdBatch(defaultTenantId, defaultBatchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
		{
			name:         "400 request binding failure",
			tenantId:     test.ValidTenantId,
			batchId:      test.ValidBatchId,
			handler:      theHandler{},
			requestBody:  `{"metadata": 100}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request param \"metadata\": expected type map[string]interface {}, but received type number"}`, requestId) + "\n",
		},
		{
			name:         "400 request binding failure",
			tenantId:     "",
			batchId:      "",
			handler:      theHandler{},
			requestBody:  ``,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request arguments:\n- id (url path parameter) is a required field\n- tenantId (url path parameter) is a required field"}`, requestId) + "\n",
		},
		{
			name:     "401 unauthorized failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "missing tenant scope"),
				},
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"missing tenant scope"}`, requestId) + "\n",
		},
		{
			name:     "500 elastic client failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: badEsConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"error getting Elastic client: cannot create client: cannot parse url: parse \"https:// a bad url.com\": invalid character \" \" in host name"}`, requestId) + "\n",
		},
		{
			name:     "500 terminate failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: defaultConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				terminate: fakeAction{
					t:               t,
					expectedRequest: getTestTerminateRequest(nil),
					expectedStatus:  batchStatus,
					code:            http.StatusInternalServerError,
					body:            response.NewErrorDetail(requestId, "something bad happened"),
				}.terminate,
			},
			expectedCode:           http.StatusInternalServerError,
			expectedBody:           fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"something bad happened"}`, requestId) + "\n",
			expectedGetByIdRequest: getTestGetByIdBatch(defaultTenantId, defaultBatchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.requestBody))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenant/:tenantId/batches/:batchId/action/terminate")
			context.SetParamNames(param.TenantId, param.BatchId)
			context.SetParamValues(tt.tenantId, tt.batchId)
			context.Response().Header().Add(echo.HeaderXRequestID, requestId)

			//Check for call to getById to retrieve Batch info (status)
			var expectGetByIdCall = tt.expectedGetByIdCode != 0
			getByIdCalled := false
			if expectGetByIdCall {
				tt.handler.getByIdNoAuth = func(_ string, requestBatch model.GetByIdBatch, _ auth.HriClaims, esClient *elasticsearch.Client) (int, interface{}) {
					getByIdCalled = true
					if !reflect.DeepEqual(requestBatch, tt.expectedGetByIdRequest) {
						t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tt.expectedGetByIdRequest, requestBatch))
					}

					return tt.expectedGetByIdCode, tt.expectedGetByIdBody
				}
			}

			if assert.NoError(t, tt.handler.Terminate(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
				if expectGetByIdCall && !getByIdCalled {
					t.Error(fmt.Printf("Expected function call that did NOT happen: getById()"))
				}
			}
		})
	}
}

func Test_theHandler_TerminateNoAuth(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	testConfig.AuthDisabled = true
	var batchStatus = status.SendCompleted

	testName := "Terminate-no-auth happy path no Metadata"

	handler := theHandler{
		config: testConfig,
		terminate: fakeAction{
			t:               t,
			expectedRequest: getTestTerminateRequest(nil),
			expectedStatus:  batchStatus,
			code:            http.StatusOK,
			body:            nil,
		}.terminate,
	}

	returnBatch := map[string]interface{}{
		"id":                  defaultBatchId,
		"name":                batchName,
		"status":              batchStatus.String(),
		"startDate":           "2019-12-13",
		"dataType":            batchDataType,
		"topic":               defaultInputTopic,
		"recordCount":         float64(1),
		"expectedRecordCount": float64(1)}

	requestBody := ""
	returnCode := http.StatusOK
	responseBody := ""
	expectedGetByIdRequest := getTestGetByIdBatch(defaultTenantId, defaultBatchId)
	expectedGetByIdCode := http.StatusOK
	expectedGetByIdBody := returnBatch

	e := test.GetTestServer()
	t.Run(testName, func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(requestBody))
		context, recorder := test.PrepareHeadersContextRecorder(request, e)
		context.SetPath("/hri/tenant/:tenantId/batches/:batchId/action/terminate")
		context.SetParamNames(param.TenantId, param.BatchId)
		context.SetParamValues(test.ValidTenantId, test.ValidBatchId)
		context.Response().Header().Add(echo.HeaderXRequestID, requestId)

		//Check for call to getById to retrieve Batch info (status)
		var expectGetByIdCall = expectedGetByIdCode != 0
		getByIdCalled := false
		if expectGetByIdCall {
			handler.getByIdNoAuth = func(_ string, requestBatch model.GetByIdBatch, _ auth.HriClaims, esClient *elasticsearch.Client) (int, interface{}) {
				getByIdCalled = true
				if !reflect.DeepEqual(requestBatch, expectedGetByIdRequest) {
					t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", expectedGetByIdRequest, requestBatch))
				}
				return expectedGetByIdCode, expectedGetByIdBody
			}
		}

		if assert.NoError(t, handler.Terminate(context)) {
			assert.Equal(t, returnCode, recorder.Code)
			assert.Equal(t, responseBody, recorder.Body.String())
			if expectGetByIdCall && !getByIdCalled {
				t.Error(fmt.Printf("Expected function call that did NOT happen: getById()"))
			}
		}
	})
}

func Test_theHandler_ProcessingComplete(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	var defaultConfig = createDefaultTestConfig()
	const currentStatus = status.SendCompleted

	returnBatch := map[string]interface{}{
		"id":                  defaultBatchId,
		"name":                batchName,
		"status":              currentStatus.String(),
		"startDate":           "2019-12-13",
		"dataType":            batchDataType,
		"topic":               defaultInputTopic,
		"recordCount":         float64(1),
		"expectedRecordCount": float64(1)}

	tests := []struct {
		name                   string
		tenantId               string
		batchId                string
		handler                theHandler
		requestBody            string
		expectedCode           int
		expectedBody           string
		expectedGetByIdRequest model.GetByIdBatch
		expectedGetByIdCode    int
		expectedGetByIdBody    interface{} //expected Return for getById() call to retrieve batch.status

	}{
		{
			name:     "success",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: defaultConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				processingComplete: fakeAction{
					t:               t,
					expectedRequest: getTestProcessingCompleteRequest(100, 10),
					expectedStatus:  currentStatus,
					code:            http.StatusOK,
					body:            nil,
				}.processingComplete,
			},
			requestBody:            `{"actualRecordCount":100,"invalidRecordCount":10}`,
			expectedCode:           http.StatusOK,
			expectedBody:           "",
			expectedGetByIdRequest: getTestGetByIdBatch(defaultTenantId, defaultBatchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
		{
			name:     "success with 0 records",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: defaultConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				processingComplete: fakeAction{
					t:               t,
					expectedRequest: getTestProcessingCompleteRequest(0, 0),
					expectedStatus:  currentStatus,
					code:            http.StatusOK,
					body:            nil,
				}.processingComplete,
			},
			requestBody:            `{"actualRecordCount":0,"invalidRecordCount":0}`,
			expectedCode:           http.StatusOK,
			expectedBody:           "",
			expectedGetByIdRequest: getTestGetByIdBatch(defaultTenantId, defaultBatchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
		{
			name:         "400 request binding failure",
			tenantId:     test.ValidTenantId,
			batchId:      test.ValidBatchId,
			handler:      theHandler{},
			requestBody:  `{"actualRecordCount": "100"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request param \"actualRecordCount\": expected type int, but received type string"}`, requestId) + "\n",
		},
		{
			name:         "400 empty tenant and batch params",
			tenantId:     "",
			batchId:      "",
			handler:      theHandler{},
			requestBody:  `{"actualRecordCount": 100,"invalidRecordCount":10}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request arguments:\n- id (url path parameter) is a required field\n- tenantId (url path parameter) is a required field"}`, requestId) + "\n",
		},
		{
			name:         "400 request validation failure",
			tenantId:     test.ValidTenantId,
			batchId:      test.ValidBatchId,
			handler:      theHandler{},
			requestBody:  `{"actualRecordCount": -100,"invalidRecordCount":10}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request arguments:\n- actualRecordCount (json field in request body) must be 0 or greater"}`, requestId) + "\n",
		},
		{
			name:     "400 request missing invalidRecordCount failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: defaultConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				processingComplete: fakeAction{
					t:               t,
					expectedRequest: getTestProcessingCompleteRequest(15, 0),
					expectedStatus:  currentStatus,
					code:            http.StatusOK,
					body:            nil,
				}.processingComplete,
			},
			requestBody:  `{"actualRecordCount":15}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request arguments:\n- invalidRecordCount (json field in request body) is a required field"}`, requestId) + "\n",
		},
		{
			name:     "401 unauthorized failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "missing tenant scope"),
				},
			},
			requestBody:  `{"actualRecordCount":100,"invalidRecordCount":10}`,
			expectedCode: http.StatusUnauthorized,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"missing tenant scope"}`, requestId) + "\n",
		},
		{
			name:     "500 elastic client failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: badEsConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
			},
			requestBody:  `{"actualRecordCount":100,"invalidRecordCount":10}`,
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"error getting Elastic client: cannot create client: cannot parse url: parse \"https:// a bad url.com\": invalid character \" \" in host name"}`, requestId) + "\n",
		},
		{
			name:     "500 processingCompete failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: defaultConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				processingComplete: fakeAction{
					t:               t,
					expectedRequest: getTestProcessingCompleteRequest(100, 10),
					expectedStatus:  currentStatus,
					code:            http.StatusInternalServerError,
					body:            response.NewErrorDetail(requestId, "something bad happened"),
				}.processingComplete,
			},
			requestBody:            `{"actualRecordCount":100,"invalidRecordCount":10}`,
			expectedCode:           http.StatusInternalServerError,
			expectedBody:           fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"something bad happened"}`, requestId) + "\n",
			expectedGetByIdRequest: getTestGetByIdBatch(defaultTenantId, defaultBatchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.requestBody))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenant/:tenantId/batches/:batchId/action/processingComplete")
			context.SetParamNames(param.TenantId, param.BatchId)
			context.SetParamValues(tt.tenantId, tt.batchId)
			context.Response().Header().Add(echo.HeaderXRequestID, requestId)

			//Check for call to getById to retrieve Batch info (status)
			var expectGetByIdCall = tt.expectedGetByIdCode != 0
			getByIdCalled := false
			if expectGetByIdCall {
				tt.handler.getByIdNoAuth = func(_ string, requestBatch model.GetByIdBatch, _ auth.HriClaims, esClient *elasticsearch.Client) (int, interface{}) {
					getByIdCalled = true
					if !reflect.DeepEqual(requestBatch, tt.expectedGetByIdRequest) {
						t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tt.expectedGetByIdRequest, requestBatch))
					}
					return tt.expectedGetByIdCode, tt.expectedGetByIdBody
				}
			}

			if assert.NoError(t, tt.handler.ProcessingComplete(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
				if expectGetByIdCall && !getByIdCalled {
					t.Error(fmt.Printf("Expected function call that did NOT happen: getById()"))
				}
			}
		})
	}
}

func Test_theHandler_ProcessingCompleteNoAuth(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	testConfig.AuthDisabled = true
	const currentStatus = status.SendCompleted

	returnBatch := map[string]interface{}{
		"id":                  defaultBatchId,
		"name":                batchName,
		"status":              currentStatus.String(),
		"startDate":           "2019-12-13",
		"dataType":            batchDataType,
		"topic":               defaultInputTopic,
		"recordCount":         float64(1),
		"expectedRecordCount": float64(1)}

	requestBody := `{"actualRecordCount":125,"invalidRecordCount":7}`
	responseCode := http.StatusOK
	responseBody := ""
	expectedGetByIdRequest := getTestGetByIdBatch(defaultTenantId, defaultBatchId)
	expectedGetByIdCode := http.StatusOK
	expectedGetByIdBody := returnBatch

	testName := "processingComplete no auth success"
	handler := theHandler{
		config: testConfig,
		processingComplete: fakeAction{
			t:               t,
			expectedRequest: getTestProcessingCompleteRequest(125, 7),
			expectedStatus:  currentStatus,
			code:            http.StatusOK,
			body:            nil,
		}.processingComplete,
	}

	e := test.GetTestServer()
	t.Run(testName, func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(requestBody))
		context, recorder := test.PrepareHeadersContextRecorder(request, e)
		context.SetPath("/hri/tenant/:tenantId/batches/:batchId/action/processingComplete")
		context.SetParamNames(param.TenantId, param.BatchId)
		context.SetParamValues(test.ValidTenantId, test.ValidBatchId)
		context.Response().Header().Add(echo.HeaderXRequestID, requestId)

		//Check for call to getById to retrieve Batch info (status)
		var expectGetByIdCall = expectedGetByIdCode != 0
		getByIdCalled := false
		if expectGetByIdCall {
			handler.getByIdNoAuth = func(_ string, requestBatch model.GetByIdBatch, _ auth.HriClaims, esClient *elasticsearch.Client) (int, interface{}) {
				getByIdCalled = true
				if !reflect.DeepEqual(requestBatch, expectedGetByIdRequest) {
					t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", expectedGetByIdRequest, requestBatch))
				}
				return expectedGetByIdCode, expectedGetByIdBody
			}
		}

		if assert.NoError(t, handler.ProcessingComplete(context)) {
			assert.Equal(t, responseCode, recorder.Code)
			assert.Equal(t, responseBody, recorder.Body.String())

			if expectGetByIdCall && !getByIdCalled {
				t.Error(fmt.Printf("Expected function call that did NOT happen: getById()"))
			}
		}
	})
}

func Test_theHandler_Fail(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	var defaultConfig = createDefaultTestConfig()
	const currentStatus = status.Started

	returnBatch := map[string]interface{}{
		"id":                  defaultBatchId,
		"name":                batchName,
		"status":              currentStatus.String(),
		"startDate":           "2019-12-13",
		"dataType":            batchDataType,
		"topic":               defaultInputTopic,
		"recordCount":         float64(1),
		"expectedRecordCount": float64(1)}

	tests := []struct {
		name                   string
		tenantId               string
		batchId                string
		handler                theHandler
		requestBody            string
		expectedCode           int
		expectedBody           string
		expectedGetByIdRequest model.GetByIdBatch
		expectedGetByIdCode    int
		expectedGetByIdBody    interface{} //expected Return for getById() call to retrieve batch.status
	}{
		{
			name:     "success",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: defaultConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				fail: fakeAction{
					t:               t,
					expectedRequest: getTestFailRequest(100, 10, "a bad batch"),
					expectedStatus:  currentStatus,
					code:            http.StatusOK,
					body:            nil,
				}.fail,
			},
			requestBody:            `{"actualRecordCount":100,"invalidRecordCount":10,"failureMessage":"a bad batch"}`,
			expectedCode:           http.StatusOK,
			expectedBody:           "",
			expectedGetByIdRequest: getTestGetByIdBatch(defaultTenantId, defaultBatchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
		{
			name:         "400 request binding failure",
			tenantId:     test.ValidTenantId,
			batchId:      test.ValidBatchId,
			handler:      theHandler{},
			requestBody:  `{"actualRecordCount":"100","invalidRecordCount":10,"failureMessage":"a bad batch"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request param \"actualRecordCount\": expected type int, but received type string"}`, requestId) + "\n",
		},
		{
			name:         "400 missing tenant and batch id",
			tenantId:     "",
			batchId:      "",
			handler:      theHandler{},
			requestBody:  `{"actualRecordCount":100,"invalidRecordCount":10,"failureMessage":"a bad batch"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request arguments:\n- id (url path parameter) is a required field\n- tenantId (url path parameter) is a required field"}`, requestId) + "\n",
		},
		{
			name:         "400 request validation failure",
			tenantId:     test.ValidTenantId,
			batchId:      test.ValidBatchId,
			handler:      theHandler{},
			requestBody:  `{"actualRecordCount":100,"invalidRecordCount":10,"failureMessage":""}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"invalid request arguments:\n- failureMessage (json field in request body) is a required field"}`, requestId) + "\n",
		},
		{
			name:     "401 unauthorized failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "missing tenant scope"),
				},
			},
			requestBody:  `{"actualRecordCount":100,"invalidRecordCount":10,"failureMessage":"a bad batch"}`,
			expectedCode: http.StatusUnauthorized,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"missing tenant scope"}`, requestId) + "\n",
		},
		{
			name:     "500 elastic client failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: badEsConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
			},
			requestBody:  `{"actualRecordCount":100,"invalidRecordCount":10,"failureMessage":"a bad batch"}`,
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"error getting Elastic client: cannot create client: cannot parse url: parse \"https:// a bad url.com\": invalid character \" \" in host name"}`, requestId) + "\n",
		},
		{
			name:     "500 fail-action failure",
			tenantId: test.ValidTenantId,
			batchId:  test.ValidBatchId,
			handler: theHandler{
				config: defaultConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				fail: fakeAction{
					t:               t,
					expectedRequest: getTestFailRequest(100, 10, "a bad batch"),
					expectedStatus:  currentStatus,
					code:            http.StatusInternalServerError,
					body:            response.NewErrorDetail(requestId, "something bad happened"),
				}.fail,
			},
			requestBody:            `{"actualRecordCount":100,"invalidRecordCount":10,"failureMessage":"a bad batch"}`,
			expectedCode:           http.StatusInternalServerError,
			expectedBody:           fmt.Sprintf(`{"errorEventId":"%s","errorDescription":"something bad happened"}`, requestId) + "\n",
			expectedGetByIdRequest: getTestGetByIdBatch(defaultTenantId, defaultBatchId),
			expectedGetByIdCode:    http.StatusOK,
			expectedGetByIdBody:    returnBatch,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.requestBody))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenant/:tenantId/batches/:batchId/action/fail")
			context.SetParamNames(param.TenantId, param.BatchId)
			context.SetParamValues(tt.tenantId, tt.batchId)
			context.Response().Header().Add(echo.HeaderXRequestID, requestId)

			//Check for call to getById to retrieve Batch info (status)
			var expectGetByIdCall = tt.expectedGetByIdCode != 0
			getByIdCalled := false
			if expectGetByIdCall {
				tt.handler.getByIdNoAuth = func(_ string, requestBatch model.GetByIdBatch, _ auth.HriClaims, esClient *elasticsearch.Client) (int, interface{}) {
					getByIdCalled = true
					if !reflect.DeepEqual(requestBatch, tt.expectedGetByIdRequest) {
						t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tt.expectedGetByIdRequest, requestBatch))
					}
					return tt.expectedGetByIdCode, tt.expectedGetByIdBody
				}
			}

			if assert.NoError(t, tt.handler.Fail(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
				if expectGetByIdCall && !getByIdCalled {
					t.Error(fmt.Printf("Expected function call that did NOT happen: getById()"))
				}
			}
		})
	}
}

func Test_theHandler_FailNoAuth(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	testConfig.AuthDisabled = true
	const currentStatus = status.Started

	returnBatch := map[string]interface{}{
		"id":                  defaultBatchId,
		"name":                batchName,
		"status":              currentStatus.String(),
		"startDate":           "2019-12-13",
		"dataType":            batchDataType,
		"topic":               defaultInputTopic,
		"recordCount":         float64(1),
		"expectedRecordCount": float64(1)}

	requestBody := `{"actualRecordCount":200,"invalidRecordCount":10,"failureMessage":"a bad-azz batch"}`
	code := http.StatusOK
	responseBody := ""
	expectedGetByIdRequest := getTestGetByIdBatch(defaultTenantId, defaultBatchId)
	expectedGetByIdCode := http.StatusOK
	expectedGetByIdBody := returnBatch

	testName := "fail no-auth success"
	handler := theHandler{
		config: testConfig,
		fail: fakeAction{
			t:               t,
			expectedRequest: getTestFailRequest(200, 10, "a bad-azz batch"),
			expectedStatus:  currentStatus,
			code:            http.StatusOK,
			body:            nil,
		}.fail,
	}

	e := test.GetTestServer()
	t.Run(testName, func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(requestBody))
		context, recorder := test.PrepareHeadersContextRecorder(request, e)
		context.SetPath("/hri/tenant/:tenantId/batches/:batchId/action/fail")
		context.SetParamNames(param.TenantId, param.BatchId)
		context.SetParamValues(test.ValidTenantId, test.ValidBatchId)
		context.Response().Header().Add(echo.HeaderXRequestID, requestId)

		//Check for call to getById to retrieve Batch info (status)
		var expectGetByIdCall = expectedGetByIdCode != 0
		getByIdCalled := false
		if expectGetByIdCall {
			handler.getByIdNoAuth = func(_ string, requestBatch model.GetByIdBatch, _ auth.HriClaims, esClient *elasticsearch.Client) (int, interface{}) {
				getByIdCalled = true
				if !reflect.DeepEqual(requestBatch, expectedGetByIdRequest) {
					t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", expectedGetByIdRequest, requestBatch))
				}
				return expectedGetByIdCode, expectedGetByIdBody
			}
		}

		if assert.NoError(t, handler.Fail(context)) {
			assert.Equal(t, code, recorder.Code)
			assert.Equal(t, responseBody, recorder.Body.String())
			if expectGetByIdCall && !getByIdCalled {
				t.Error(fmt.Printf("Expected function call that did NOT happen: getById()"))
			}
		}
	})
}

func Test_theHandler_getCurrentBatchStatus(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	testConfig.AuthDisabled = true
	var prefix = "handler_test/test_getCurrentBatchStatus"
	var logger = logwrapper.GetMyLogger("", prefix)

	var requestId = "fakeGetByIdRequest"
	var tenantId = "testTenant"
	var batchId = "funbatch1"
	var topic = "ingest.15.phil.collins"
	var batchName = "monkeyBatch"
	var recCount = 20
	var defaultBatchStatus = status.Started
	var defaultGetByIdResult = map[string]interface{}{
		"id":                  batchId,
		"name":                batchName,
		"status":              defaultBatchStatus.String(),
		"startDate":           "2019-12-13",
		"dataType":            "claims",
		"topic":               topic,
		"recordCount":         float64(recCount),
		"expectedRecordCount": float64(recCount)}
	var badGetByIdResult = map[string]interface{}{
		"id":     batchId,
		"name":   batchName,
		"status": status.Unknown.String(),
	}
	var getBatchRequest = model.GetByIdBatch{
		TenantId: tenantId,
		BatchId:  batchId,
	}

	tests := []struct {
		name                string
		handler             theHandler
		expectedBatchStatus status.BatchStatus
		expectedErrResponse *response.ErrorDetailResponse
		expectedErrRtnCode  int
		expectedErrDscr     string
	}{
		{
			name: "getBatchStatus success case",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				getByIdNoAuth: func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusOK, defaultGetByIdResult
				},
			},
			expectedErrResponse: nil,
			expectedBatchStatus: status.Started,
		},
		{
			name: "getBatchStatus returns 404 Not Found for getById batch Not found",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				getByIdNoAuth: func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusNotFound, response.NewErrorDetail("", "The document for tenantId: testTenant with document (batch) ID: funbatch1 was not found")
				},
			},
			expectedErrRtnCode:  http.StatusNotFound,
			expectedErrDscr:     "error getting current Batch Status: The document for tenantId: testTenant with document (batch) ID: funbatch1 was not found",
			expectedBatchStatus: status.Unknown,
		}, {
			name: "getBatchStatus returns 500 InternalServerError for error extracting batch status",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				getByIdNoAuth: func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusOK, badGetByIdResult
				},
			},
			expectedErrRtnCode:  http.StatusInternalServerError,
			expectedErrDscr:     "error getting current Batch Status: Error extracting Batch Status: Invalid 'status' value: unknown",
			expectedBatchStatus: status.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esClient, err := elastic.ClientFromTransport(test.NewFakeTransport(t))
			if err != nil {
				t.Error(err)
			}
			batchStatus, errDetail := getCurrentBatchStatus(&tt.handler, requestId, getBatchRequest, esClient, logger)

			assert.Equal(t, tt.expectedBatchStatus, batchStatus)
			if batchStatus == status.Unknown || tt.expectedErrResponse != nil {
				assert.Equal(t, tt.expectedErrRtnCode, errDetail.Code)
				assert.Equal(t, tt.expectedErrDscr, errDetail.Body.ErrorDescription)
			}

		})

	}

}

var badEsConfig = config.Config{
	ConfigPath:        "/some/fake/path",
	OidcIssuer:        "https://us-south.appid.blorg.forg",
	JwtAudienceId:     "1234-990-g-catnip-e9ec09a5b2f3",
	Validation:        false,
	ElasticUrl:        "https:// a bad url.com",
	ElasticServiceCrn: "elasticUsername",
	KafkaUsername:     "duranDuran",
	KafkaPassword:     "Toto",
	KafkaBrokers:      []string{"broker-1", "broker-2", "broker-DefLeppard", "broker-CoreyHart"},
}
