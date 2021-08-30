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
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

const topic = "ingest.08.claims.phil.collins"
const startDate = "2020-10-30T12:34:00Z"

func TestNewHandler(t *testing.T) {
	config := config.Config{}
	config.ElasticUrl = "https://fake-elastic.com"
	config.AuthDisabled = false

	handler := NewHandler(config).(*theHandler)
	assert.Equal(t, config, handler.config)
	assert.NotNil(t, handler.jwtValidator)
	// This asserts that they are the same function by memory address;
	assert.Equal(t, reflect.ValueOf(Create), reflect.ValueOf(handler.create))
	assert.Equal(t, reflect.ValueOf(GetById), reflect.ValueOf(handler.getById))
	assert.Equal(t, reflect.ValueOf(Get), reflect.ValueOf(handler.get))
	assert.Equal(t, reflect.ValueOf(SendComplete), reflect.ValueOf(handler.sendComplete))
	assert.Equal(t, reflect.ValueOf(Terminate), reflect.ValueOf(handler.terminate))
	assert.Equal(t, reflect.ValueOf(ProcessingComplete), reflect.ValueOf(handler.processingComplete))
	assert.Equal(t, reflect.ValueOf(Fail), reflect.ValueOf(handler.fail))
}

func TestNewHandlerNoAuthFunctions(t *testing.T) {
	config := config.Config{}
	config.ElasticUrl = "https://fake-elastic.com"
	config.AuthDisabled = true

	handler := NewHandler(config).(*theHandler)
	assert.Equal(t, config, handler.config)
	assert.Nil(t, handler.jwtValidator)

	assert.Equal(t, reflect.ValueOf(CreateNoAuth), reflect.ValueOf(handler.create))
	assert.Equal(t, reflect.ValueOf(GetNoAuth), reflect.ValueOf(handler.get))
	assert.Equal(t, reflect.ValueOf(GetByIdNoAuth), reflect.ValueOf(handler.getById))

	assert.Equal(t, reflect.ValueOf(SendCompleteNoAuth), reflect.ValueOf(handler.sendComplete))
	assert.Equal(t, reflect.ValueOf(TerminateNoAuth), reflect.ValueOf(handler.terminate))
	assert.Equal(t, reflect.ValueOf(ProcessingCompleteNoAuth), reflect.ValueOf(handler.processingComplete))
	assert.Equal(t, reflect.ValueOf(FailNoAuth), reflect.ValueOf(handler.fail))
}

// Fake for the auth.Validator interface; just returns the desired values
type fakeAuthValidator struct {
	claims  auth.HriClaims
	errResp *response.ErrorDetailResponse
}

func (f fakeAuthValidator) GetValidatedClaims(_ string, _ string, _ string) (auth.HriClaims, *response.ErrorDetailResponse) {
	return f.claims, f.errResp
}

func Test_theHandler_Create(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	var testConfig = createDefaultTestConfig()
	validTenantId := "tenant_33-z"
	invalidTenantIdUpper := "BAD-93TENant-1"
	invalidTenantIdSpecChar := "invalid$$??##^^%**!!"
	var unauthorizedTenantId = "unnauthorized_tenant"
	emptyTenantId := ""
	status := status.Started.String()

	validReqBody := fmt.Sprintf(`{"name": "%s",
		"integratorId": "%s",
		"topic": "%s",
		"dataType": "%s",
		"invalidThreshold":10,
		"status": "%s",
		"startDate": "%s",
		"metadata":{"batchContact":"Samuel L. Jackson","finalRecordCount":20}}`,
		batchName, integratorId, topic, batchDataType, status, startDate)

	validUtf8ReqBody := fmt.Sprintf(`{"name": "%s中文",
		"integratorId": "%s中文",
		"topic": "%s中文",
		"dataType": "%s中文",
		"invalidThreshold":10,
		"status": "%s",
		"startDate": "%s",
		"metadata":{"batchContact":"中文","finalRecordCount":20}}`,
		batchName, integratorId, topic, batchDataType, status, startDate)

	invalidReqBody := fmt.Sprintf(`{"name": "%s",
		"integratorId": "%s",
		"invalidThreshold":10,
		"status": "%s",
		"startDate": "%s"}`, batchName, integratorId, status, startDate)

	noThresholdReqBody := fmt.Sprintf(`{"name": "%s",
		"integratorId": "%s",
		"topic": "%s",
		"dataType": "%s",
		"status": "%s",
		"startDate": "%s"}`,
		batchName, integratorId, topic, batchDataType, status, startDate)

	var invalidBodyParam = "invalidBodyParam{}][][??##^^%**!!"
	specialCharInDatatypeReqBody := fmt.Sprintf(`{"name": "%s",
		"integratorId": "%s",
		"topic": "%s",
		"dataType": "%s",
		"status": "%s",
		"startDate": "%s"}`,
		batchName, integratorId, topic, invalidBodyParam, status, startDate)

	specialCharInTopicReqBody := fmt.Sprintf(`{"name": "%s",
		"integratorId": "%s",
		"topic": "%s",
		"dataType": "%s",
		"status": "%s",
		"startDate": "%s"}`,
		batchName, integratorId, invalidBodyParam, batchDataType, status, startDate)

	tests := []struct {
		name         string
		handler      theHandler
		tenant       string
		expectedCode int
		requestBody  string
		expectedBody string
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusCreated, map[string]interface{}{"batchId": "1234-unique-id"}
				},
			},
			tenant:       validTenantId,
			expectedCode: http.StatusCreated,
			requestBody:  validReqBody,
			expectedBody: "{\"batchId\":\"1234-unique-id\"}\n",
		},
		{
			name: "should handle UTF8 chars",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{Subject: integratorId + "中文"},
					errResp: nil,
				},
				create: func(_ string, create model.CreateBatch, claims auth.HriClaims, _ *elasticsearch.Client, _ kafka.Writer) (int, interface{}) {
					assert.Equal(t, batchName+"中文", create.Name)
					assert.Equal(t, integratorId+"中文", claims.Subject)
					assert.Equal(t, topic+"中文", create.Topic)
					assert.Equal(t, batchDataType+"中文", create.DataType)
					assert.Equal(t, "中文", create.Metadata["batchContact"])
					return http.StatusCreated, map[string]interface{}{"batchId": "1234-unique-id"}
				},
			},
			tenant:       validTenantId,
			expectedCode: http.StatusCreated,
			requestBody:  validUtf8ReqBody,
			expectedBody: "{\"batchId\":\"1234-unique-id\"}\n",
		},
		{
			name: "Bad TenantId Param contains UpperCase chars",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusBadRequest,
			tenant:       invalidTenantIdUpper,
			requestBody:  validReqBody,
			expectedBody: fmt.Sprintf(`{"errorEventId":"","errorDescription":"invalid request arguments:\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'"}`) + "\n",
		},
		{
			name: "Bad TenantId Param contains Special Chars",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusBadRequest,
			tenant:       invalidTenantIdSpecChar,
			requestBody:  validReqBody,
			expectedBody: fmt.Sprintf(`{"errorEventId":"","errorDescription":"invalid request arguments:\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'"}`) + "\n",
		},
		{
			name: "Empty TenantId Param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusBadRequest,
			tenant:       emptyTenantId,
			requestBody:  validReqBody,
			expectedBody: fmt.Sprintf(`{"errorEventId":"","errorDescription":"invalid request arguments:\n- tenantId (url path parameter) is a required field"}`) + "\n",
		},
		{
			name: "Failure Binding Invalid Body Param datatype for name param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusBadRequest,
			tenant:       validTenantId,
			requestBody:  "{\"id\" :\"\",\"name\":9858,\"integratorId\":\"abcd-12345-bfc8151f5522\",\"topic\":false,\"dataType\":\"claims\",\"invalidThreshold\":10,\"status\":\"started\",\"startDate\":\"2019-10-30T12:34:00Z\"}",
			expectedBody: fmt.Sprintf(`{"errorEventId":"","errorDescription":"invalid request param \"name\": expected type string, but received type number"}`) + "\n",
		},
		{
			name: "Invalid Body Params: missing topic/datatype",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusBadRequest,
			tenant:       validTenantId,
			requestBody:  invalidReqBody,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"invalid request arguments:\\n- dataType (json field in request body) is a required field\\n- topic (json field in request body) is a required field\"}\n",
		},
		{
			name: "Missing Invalid Threshold Body Param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusCreated, map[string]interface{}{"batchId": "1234-unique-id"}
				},
			},
			expectedCode: http.StatusCreated,
			tenant:       validTenantId,
			requestBody:  noThresholdReqBody,
			expectedBody: "{\"batchId\":\"1234-unique-id\"}\n",
		},
		{
			name: "Invalid JWT Claim Unauthorized Tenant",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, "Unauthorized tenant access. Tenant '"+unauthorizedTenantId+"' is not included in the authorized scopes."),
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusUnauthorized,
			tenant:       unauthorizedTenantId,
			requestBody:  validReqBody,
			expectedBody: "{\"errorEventId\":\"" + requestId + "\",\"errorDescription\":\"Unauthorized tenant access. Tenant '" + unauthorizedTenantId + "' is not included in the authorized scopes.\"}\n",
		},
		{
			name: "Invalid characters in 'Datatype' body param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusBadRequest,
			tenant:       validTenantId,
			requestBody:  specialCharInDatatypeReqBody,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"invalid request arguments:\\n- dataType (json field in request body) must not contain the following characters: \\\"=\\u003c\\u003e[]{}\"}\n",
		},
		{
			name: "Invalid characters in 'Topic' body param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusBadRequest,
			tenant:       validTenantId,
			requestBody:  specialCharInTopicReqBody,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"invalid request arguments:\\n- topic (json field in request body) must not contain the following characters: \\\"=\\u003c\\u003e[]{}\"}\n",
		},
		{
			name: "Elastic client error",
			handler: theHandler{
				config: badEsConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: nil,
			},
			expectedCode: http.StatusInternalServerError,
			tenant:       validTenantId,
			requestBody:  validReqBody,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"error getting Elastic client: cannot create client: cannot parse url: parse \\\"https:// a bad url.com\\\": invalid character \\\" \\\" in host name\"}\n",
		},
		{
			name: "Kafka writer error",
			handler: theHandler{
				config: badKafkaConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				create: nil,
			},
			expectedCode: http.StatusInternalServerError,
			tenant:       validTenantId,
			requestBody:  validReqBody,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"error constructing Kafka producer: Invalid value for configuration property \\\"message.max.bytes\\\"\"}\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenant/:" + param.TenantId + "/batches")
			context.SetParamNames(param.TenantId)
			context.SetParamValues(tt.tenant)

			if assert.NoError(t, tt.handler.Create(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func Test_theHandler_CreateNoAuth(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	testConfig.AuthDisabled = true

	testName := "create no auth happy path"
	validTenantId := "tenant-3428"
	status := status.Started.String()
	requestBody := fmt.Sprintf(`{"name": "%s",
		"integratorId": "%s",
		"topic": "%s",
		"dataType": "%s",
		"invalidThreshold":10,
		"status": "%s",
		"startDate": "%s",
		"metadata":{"batchContact":"Porcypine Jones","finalRecordCount":45}}`,
		batchName, integratorId, topic, batchDataType, status, startDate)

	handler := theHandler{
		config: testConfig,
		create: func(string, model.CreateBatch, auth.HriClaims, *elasticsearch.Client, kafka.Writer) (int, interface{}) {
			return http.StatusCreated, map[string]interface{}{"batchId": "batch-675-unique-id"}
		},
	}
	expectedReturnCode := http.StatusCreated
	responseBody := "{\"batchId\":\"batch-675-unique-id\"}\n"

	e := test.GetTestServer()

	t.Run(testName, func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(requestBody))
		context, recorder := test.PrepareHeadersContextRecorder(request, e)
		context.SetPath("/hri/tenant/:" + param.TenantId + "/batches")
		context.SetParamNames(param.TenantId)
		context.SetParamValues(validTenantId)

		if assert.NoError(t, handler.Create(context)) {
			assert.Equal(t, expectedReturnCode, recorder.Code)
			assert.Equal(t, responseBody, recorder.Body.String())
		}
	})
}

func Test_myHandler_GetById(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	validTenantId := "tenant_33-z"
	validBatchId := "batch7j3"

	tests := []struct {
		name         string
		handler      theHandler
		tenant       string
		batchId      string
		expectedCode int
		expectedBody string
	}{
		{
			name: "success case",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				getById: func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"id": "batch7j3", "name": "monkeyBatch", "status": "started", "startDate": "2019-12-13", "dataType": "claims", "topic": "ingest-test", "recordCount": float64(1), "expectedRecordCount": float64(1)}
				},
			},
			tenant:       validTenantId,
			batchId:      validBatchId,
			expectedCode: http.StatusOK,
			expectedBody: "{\"dataType\":\"claims\",\"expectedRecordCount\":1,\"id\":\"batch7j3\",\"name\":\"monkeyBatch\",\"recordCount\":1,\"startDate\":\"2019-12-13\",\"status\":\"started\",\"topic\":\"ingest-test\"}\n",
		},
		{
			name: "Bad TenantId Param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				getById: func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusNotFound, response.NewErrorDetail("", "The document for tenantId: BAD-93TENant-1 with document (batch) ID: batch7j3 was not found")
				},
			},
			expectedCode: http.StatusNotFound,
			tenant:       "BAD-93TENant-1",
			batchId:      validBatchId,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"The document for tenantId: BAD-93TENant-1 with document (batch) ID: batch7j3 was not found\"}\n",
		},
		{
			name: "Empty TenantId Param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				getById: func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusBadRequest,
			batchId:      validBatchId,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) is a required field\"}\n",
		},
		{
			name: "Empty BatchId Param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				getById: func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusBadRequest,
			tenant:       validTenantId,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"invalid request arguments:\\n- id (url path parameter) is a required field\"}\n",
		},
		{
			name: "Invalid JWT Claim Unauthorized Tenant",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: response.NewErrorDetailResponse(http.StatusUnauthorized, "requestId", "Unauthorized tenant access. Tenant 'unauthorized_tenant' is not included in the authorized scopes."),
				},
				getById: func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			expectedCode: http.StatusUnauthorized,
			tenant:       "unauthorized_tenant",
			batchId:      validBatchId,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"Unauthorized tenant access. Tenant 'unauthorized_tenant' is not included in the authorized scopes.\"}\n",
		},
		{
			name: "Elastic client error",
			handler: theHandler{
				config: badEsConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				getById: nil,
			},
			tenant:       validTenantId,
			batchId:      validBatchId,
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"\",\"errorDescription\":\"error getting Elastic client: cannot create client: cannot parse url: parse \\\"https:// a bad url.com\\\": invalid character \\\" \\\" in host name\"}\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenants/:" + param.TenantId + "/batches/:" + param.BatchId)
			context.SetParamNames(param.TenantId, param.BatchId)
			context.SetParamValues(tt.tenant, tt.batchId)

			if assert.NoError(t, tt.handler.GetById(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func Test_theHandler_GetByIdNoAuth(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	testConfig.AuthDisabled = true

	testName := "getById-no-auth happy path"
	myTenantId := "tenant-3428"
	myBatchId := "batch92dz3"
	topic := "ingest.03.phil.collins"
	batchName := "porcypineBatch"
	recCount := 30
	expectedReturnCode := http.StatusOK

	handler := theHandler{
		config: testConfig,
		getById: func(string, model.GetByIdBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
			return http.StatusOK, map[string]interface{}{"id": myBatchId, "name": batchName, "status": "started",
				"startDate": "2019-12-13", "dataType": "claims", "topic": topic,
				"recordCount": float64(recCount), "expectedRecordCount": float64(recCount)}
		},
	}
	responseBody := "{\"dataType\":\"claims\",\"expectedRecordCount\":30,\"id\":\"batch92dz3\",\"name\":\"porcypineBatch\",\"recordCount\":30,\"startDate\":\"2019-12-13\",\"status\":\"started\",\"topic\":\"ingest.03.phil.collins\"}\n"

	e := test.GetTestServer()
	t.Run(testName, func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		context, recorder := test.PrepareHeadersContextRecorder(request, e)
		context.SetPath("/hri/tenants/:" + param.TenantId + "/batches/:" + param.BatchId)
		context.SetParamNames(param.TenantId, param.BatchId)
		context.SetParamValues(myTenantId, myBatchId)

		if assert.NoError(t, handler.GetById(context)) {
			assert.Equal(t, expectedReturnCode, recorder.Code)
			assert.Equal(t, responseBody, recorder.Body.String())
		}
	})
}

func Test_myHandler_Get(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	validTenantId := "tenant_33-z"
	invalidTenantId := "BAD-93TENant-1"
	unauthorizedTenantId := "unauthorized_tenant"

	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		nameParam    string
		statusParam  string
		gteDateParam string
		lteDateParam string
		sizeParam    string
		fromParam    string
		responseCode int
		responseBody string
	}{
		{
			name: "success case minimal",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				get: func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			responseCode: http.StatusOK,
			responseBody: "{\"results\":[{\"dataType\":\"rspec-batch\",\"id\":\"uuid\",\"integratorId\":\"modified-integrator-id\",\"invalidThreshold\":-1,\"metadata\":{\"rspec1\":\"test1\"},\"name\":\"mybatch\",\"startDate\":\"01/02/2019\",\"status\":\"started\",\"topic\":\"ingest.test.claims.in\"}],\"total\":1}\n",
		},
		{
			name: "success case maximal",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				get: func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			nameParam:    "niceName",
			statusParam:  "niceStatus",
			gteDateParam: "01/01/2020",
			lteDateParam: "01/01/2021",
			sizeParam:    "1",
			fromParam:    "2",
			responseCode: http.StatusOK,
			responseBody: "{\"results\":[{\"dataType\":\"rspec-batch\",\"id\":\"uuid\",\"integratorId\":\"modified-integrator-id\",\"invalidThreshold\":-1,\"metadata\":{\"rspec1\":\"test1\"},\"name\":\"mybatch\",\"startDate\":\"01/02/2019\",\"status\":\"started\",\"topic\":\"ingest.test.claims.in\"}],\"total\":1}\n",
		},
		{
			name: "missing tenantId param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				get: func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			responseCode: http.StatusBadRequest,
			responseBody: "{\"errorEventId\":\"\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) is a required field\"}\n",
		},
		{
			name: "bad tenantId param",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				get: func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusNotFound, response.NewErrorDetail("", "The document for tenantId: "+invalidTenantId+" was not found")
				},
			},
			tenantId:     invalidTenantId,
			responseCode: http.StatusNotFound,
			responseBody: "{\"errorEventId\":\"\",\"errorDescription\":\"The document for tenantId: " + invalidTenantId + " was not found\"}\n",
		},
		{
			name: "bad name param prohibited char",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				get: func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			tenantId:     validTenantId,
			nameParam:    "{[]//sqlInjection[]}",
			responseCode: http.StatusBadRequest,
			responseBody: "{\"errorEventId\":\"\",\"errorDescription\":\"invalid request arguments:\\n- name (request query parameter) must not contain the following characters: \\\"=\\u003c\\u003e[]{}\"}\n",
		},
		{
			name: "bad status param prohibited char",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				get: func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			tenantId:     validTenantId,
			statusParam:  "{[]//sqlInjection[]}",
			responseCode: http.StatusBadRequest,
			responseBody: "{\"errorEventId\":\"\",\"errorDescription\":\"invalid request arguments:\\n- status (request query parameter) must not contain the following characters: \\\"=\\u003c\\u003e[]{}\"}\n",
		},
		{
			name: "bad date param prohibited char",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				get: func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			tenantId:     validTenantId,
			gteDateParam: "01/01/2019",
			lteDateParam: "{}[][sqlInjection]\"}",
			responseCode: http.StatusBadRequest,
			responseBody: "{\"errorEventId\":\"\",\"errorDescription\":\"invalid request arguments:\\n- lteDate (request query parameter) must not contain the following characters: \\\"=\\u003c\\u003e[]{}\"}\n",
		},
		{
			name: "invalid JWT claim unauthorized tenant",
			handler: theHandler{
				config: testConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: response.NewErrorDetailResponse(http.StatusUnauthorized, "requestId", "Unauthorized tenant access. Tenant '"+unauthorizedTenantId+"' is not included in the authorized scopes."),
				},
				get: func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
					return http.StatusForbidden, map[string]interface{}{"NO_CALL": "This Function Should Never Get Called"}
				},
			},
			tenantId:     unauthorizedTenantId,
			responseCode: http.StatusUnauthorized,
			responseBody: "{\"errorEventId\":\"\",\"errorDescription\":\"Unauthorized tenant access. Tenant '" + unauthorizedTenantId + "' is not included in the authorized scopes.\"}\n",
		},
		{
			name: "Elastic client error",
			handler: theHandler{
				config: badEsConfig,
				jwtValidator: fakeAuthValidator{
					claims:  auth.HriClaims{},
					errResp: nil,
				},
				get: nil,
			},
			tenantId:     validTenantId,
			responseCode: http.StatusInternalServerError,
			responseBody: "{\"errorEventId\":\"\",\"errorDescription\":\"error getting Elastic client: cannot create client: cannot parse url: parse \\\"https:// a bad url.com\\\": invalid character \\\" \\\" in host name\"}\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := make(url.Values)
			queryParams := [][]string{
				{param.Name, tt.nameParam},
				{param.Status, tt.statusParam},
				{param.GteDate, tt.gteDateParam},
				{param.LteDate, tt.lteDateParam},
				{param.Size, tt.sizeParam},
				{param.From, tt.fromParam},
			}
			for _, paramPair := range queryParams {
				if len(paramPair[1]) > 0 {
					q.Set(paramPair[0], paramPair[1])
				}
			}

			request := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenants/:" + param.TenantId + "/batches")
			context.SetParamNames(param.TenantId)
			context.SetParamValues(tt.tenantId)
			if assert.NoError(t, tt.handler.Get(context)) {
				assert.Equal(t, tt.responseCode, recorder.Code)
				assert.Equal(t, tt.responseBody, recorder.Body.String())
			}
		})
	}
}

func Test_theHandler_GetNoAuth(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	testConfig.AuthDisabled = true

	testName := "get-no-auth happy path-minimal"
	myTenantId := "tenant-55"
	myBatchId := "batch0881gd5"
	topic := "ingest.03.phil.collins"
	batchName := "porcypineBatch"
	expectedRespCode := http.StatusOK
	startDate := "06/08/2019"
	metadata := map[string]interface{}{"contact": "Sam L. Jackson"}

	handler := theHandler{
		config: testConfig,
		get: func(string, model.GetBatch, auth.HriClaims, *elasticsearch.Client) (int, interface{}) {
			return http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": myBatchId, "dataType": "claims",
				"invalidThreshold": float64(-1), "name": batchName,
				"startDate": startDate, "status": "started",
				"topic": topic, "integratorId": "modified-integrator-id",
				"metadata": metadata}}}
		},
	}

	expectedResponseBody := "{\"results\":[{\"dataType\":\"claims\",\"id\":\"" + myBatchId +
		"\",\"integratorId\":\"modified-integrator-id\",\"invalidThreshold\":-1,\"metadata\":{\"contact\":\"Sam L. Jackson\"},\"name\":\"" + batchName +
		"\",\"startDate\":\"" + startDate + "\",\"status\":\"started\",\"topic\":\"" + topic + "\"}],\"total\":1}\n"

	e := test.GetTestServer()

	t.Run(testName, func(t *testing.T) {
		q := make(url.Values)
		request := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
		context, recorder := test.PrepareHeadersContextRecorder(request, e)
		context.SetPath("/hri/tenants/:" + param.TenantId + "/batches")
		context.SetParamNames(param.TenantId)
		context.SetParamValues(myTenantId)
		if assert.NoError(t, handler.Get(context)) {
			assert.Equal(t, expectedRespCode, recorder.Code)
			assert.Equal(t, expectedResponseBody, recorder.Body.String())
		}
	})
}

func createDefaultTestConfig() config.Config {
	config := config.Config{}
	config.ConfigPath = "/some/fake/path"
	config.OidcIssuer = "https://us-south.appid.blorg.forg"
	config.JwtAudienceId = "1234-990-g-catnip-e9ec09a5b2f3"
	config.Validation = false
	config.ElasticUrl = "https://elastic-JanetJackson.com"
	config.ElasticServiceCrn = "elasticUsername"
	config.KafkaBrokers = []string{"broker-1", "broker-2", "broker-DefLeppard", "broker-CoreyHart"}
	return config
}
