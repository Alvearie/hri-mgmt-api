package batches

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

const topic = "ingest.08.claims.phil.collins"
const startDate = "2020-10-30T12:34:00Z"

// Fake for the auth.Validator interface; just returns the desired values
type fakeAuthValidator struct {
	claims  auth.HriAzClaims
	errResp *response.ErrorDetailResponse
}

func (f fakeAuthValidator) GetValidatedRoles(_ string, _ string, _ string) (auth.HriAzClaims, *response.ErrorDetailResponse) {
	return f.claims, f.errResp
}

func createDefaultTestConfig() config.Config {
	config := config.Config{}
	config.ConfigPath = "/some/fake/path"
	config.AzOidcIssuer = "https://us-south.appid.blorg.forg"
	config.AzJwtAudienceId = "1234-990-g-catnip-e9ec09a5b2f3"
	config.Validation = false
	config.AzKafkaBrokers = []string{"broker-1", "broker-2", "broker-DefLeppard", "broker-CoreyHart"}
	return config
}

func createDefaultTestConfigAuthDisabled() config.Config {
	config := config.Config{}
	config.ConfigPath = "/some/fake/path"
	config.AzOidcIssuer = "https://us-south.appid.blorg.forg"
	config.AzJwtAudienceId = "1234-990-g-catnip-e9ec09a5b2f3"
	config.Validation = false
	config.AzKafkaBrokers = []string{"broker-1", "broker-2", "broker-DefLeppard", "broker-CoreyHart"}
	config.AuthDisabled = true
	return config
}
func TestBatchHandlerAuthEnabled(t *testing.T) {
	config := config.Config{}

	handler := NewHandler(config).(*theHandler)
	assert.Equal(t, config, handler.config)
	// This asserts that they are the same function by memory address
	assert.Equal(t, reflect.ValueOf(CreateBatch), reflect.ValueOf(handler.createBatch))
	assert.Equal(t, reflect.ValueOf(GetByBatchId), reflect.ValueOf(handler.getByBatchId))
	assert.Equal(t, reflect.ValueOf(GetBatch), reflect.ValueOf(handler.getBatch))
	assert.Equal(t, reflect.ValueOf(SendStatusComplete), reflect.ValueOf(handler.sendStatusComplete))
	assert.Equal(t, reflect.ValueOf(SendFail), reflect.ValueOf(handler.sendFail))
	assert.Equal(t, reflect.ValueOf(TerminateBatch), reflect.ValueOf(handler.terminateBatch))
	assert.Equal(t, reflect.ValueOf(ProcessingCompleteBatch), reflect.ValueOf(handler.processingCompleteBatch))

}

func TestBatchHandlerAuthDisabled(t *testing.T) {
	config := config.Config{AuthDisabled: true}

	handler := NewHandler(config).(*theHandler)
	assert.Equal(t, config, handler.config)

}

func Test_myHandler_Get(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	var AuthDisabledconfig = createDefaultTestConfigAuthDisabled()
	validTenantId := "tenant_33-z"

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
				jwtBatchValidator: fakeAuthValidator{
					claims:  auth.HriAzClaims{},
					errResp: nil,
				},
				getBatch: func(string, model.GetBatch, auth.HriAzClaims) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			responseCode: http.StatusOK,
			responseBody: "{\"results\":[{\"dataType\":\"rspec-batch\",\"id\":\"uuid\",\"integratorId\":\"modified-integrator-id\",\"invalidThreshold\":-1,\"metadata\":{\"rspec1\":\"test1\"},\"name\":\"mybatch\",\"startDate\":\"01/02/2019\",\"status\":\"started\",\"topic\":\"ingest.test.claims.in\"}],\"total\":1}\n",
		},
		{
			name: "success case",
			handler: theHandler{
				config: testConfig,
				jwtBatchValidator: fakeAuthValidator{

					errResp: response.NewErrorDetailResponse(http.StatusBadRequest, "test-request-id", "jwtValidator error"),
				},
				getBatch: func(string, model.GetBatch, auth.HriAzClaims) (int, interface{}) {
					return http.StatusBadRequest, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			responseCode: http.StatusBadRequest,
			responseBody: "{\"errorEventId\":\"\",\"errorDescription\":\"jwtValidator error\"}\n",
		},
		{
			name: "AuthDisabledconfig",
			handler: theHandler{
				config: AuthDisabledconfig,
				jwtBatchValidator: fakeAuthValidator{
					claims:  auth.HriAzClaims{},
					errResp: nil,
				},
				getBatch: func(string, model.GetBatch, auth.HriAzClaims) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			responseCode: http.StatusOK,
			responseBody: "{\"results\":[{\"dataType\":\"rspec-batch\",\"id\":\"uuid\",\"integratorId\":\"modified-integrator-id\",\"invalidThreshold\":-1,\"metadata\":{\"rspec1\":\"test1\"},\"name\":\"mybatch\",\"startDate\":\"01/02/2019\",\"status\":\"started\",\"topic\":\"ingest.test.claims.in\"}],\"total\":1}\n",
		},
		{
			name: "NoConfig",
			handler: theHandler{
				jwtBatchValidator: fakeAuthValidator{
					claims:  auth.HriAzClaims{},
					errResp: nil,
				},
				getBatch: func(string, model.GetBatch, auth.HriAzClaims) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			responseCode: http.StatusOK,
			responseBody: "{\"results\":[{\"dataType\":\"rspec-batch\",\"id\":\"uuid\",\"integratorId\":\"modified-integrator-id\",\"invalidThreshold\":-1,\"metadata\":{\"rspec1\":\"test1\"},\"name\":\"mybatch\",\"startDate\":\"01/02/2019\",\"status\":\"started\",\"topic\":\"ingest.test.claims.in\"}],\"total\":1}\n",
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
			if assert.NoError(t, tt.handler.GetBatch(context)) {
				assert.Equal(t, tt.responseCode, recorder.Code)
				assert.Equal(t, tt.responseBody, recorder.Body.String())
			}
		})
	}
}
func Test_myHandler_Create(t *testing.T) {
	var testConfig = createDefaultTestConfig()
	var AuthDisabledconfig = createDefaultTestConfigAuthDisabled()
	validTenantId := "tenant_33-z"

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
				jwtBatchValidator: fakeAuthValidator{
					claims:  auth.HriAzClaims{},
					errResp: nil,
				},
				getBatch: func(string, model.GetBatch, auth.HriAzClaims) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			responseCode: http.StatusOK,
			responseBody: "{\"results\":[{\"dataType\":\"rspec-batch\",\"id\":\"uuid\",\"integratorId\":\"modified-integrator-id\",\"invalidThreshold\":-1,\"metadata\":{\"rspec1\":\"test1\"},\"name\":\"mybatch\",\"startDate\":\"01/02/2019\",\"status\":\"started\",\"topic\":\"ingest.test.claims.in\"}],\"total\":1}\n",
		},
		{
			name: "success case",
			handler: theHandler{
				config: testConfig,
				jwtBatchValidator: fakeAuthValidator{

					errResp: response.NewErrorDetailResponse(http.StatusBadRequest, "test-request-id", "jwtValidator error"),
				},
				getBatch: func(string, model.GetBatch, auth.HriAzClaims) (int, interface{}) {
					return http.StatusBadRequest, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			responseCode: http.StatusBadRequest,
			responseBody: "{\"errorEventId\":\"\",\"errorDescription\":\"jwtValidator error\"}\n",
		},
		{
			name: "AuthDisabledconfig",
			handler: theHandler{
				config: AuthDisabledconfig,
				jwtBatchValidator: fakeAuthValidator{
					claims:  auth.HriAzClaims{},
					errResp: nil,
				},
				getBatch: func(string, model.GetBatch, auth.HriAzClaims) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			responseCode: http.StatusOK,
			responseBody: "{\"results\":[{\"dataType\":\"rspec-batch\",\"id\":\"uuid\",\"integratorId\":\"modified-integrator-id\",\"invalidThreshold\":-1,\"metadata\":{\"rspec1\":\"test1\"},\"name\":\"mybatch\",\"startDate\":\"01/02/2019\",\"status\":\"started\",\"topic\":\"ingest.test.claims.in\"}],\"total\":1}\n",
		},
		{
			name: "NoConfig",
			handler: theHandler{
				jwtBatchValidator: fakeAuthValidator{
					claims:  auth.HriAzClaims{},
					errResp: nil,
				},
				getBatch: func(string, model.GetBatch, auth.HriAzClaims) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "01/02/2019", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}}
				},
			},
			tenantId:     validTenantId,
			responseCode: http.StatusOK,
			responseBody: "{\"results\":[{\"dataType\":\"rspec-batch\",\"id\":\"uuid\",\"integratorId\":\"modified-integrator-id\",\"invalidThreshold\":-1,\"metadata\":{\"rspec1\":\"test1\"},\"name\":\"mybatch\",\"startDate\":\"01/02/2019\",\"status\":\"started\",\"topic\":\"ingest.test.claims.in\"}],\"total\":1}\n",
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
			if assert.NoError(t, tt.handler.GetBatch(context)) {
				assert.Equal(t, tt.responseCode, recorder.Code)
				assert.Equal(t, tt.responseBody, recorder.Body.String())
			}
		})
	}
}
func Test_myHandler_Fail(t *testing.T) {
	validConfig := config.Config{}
	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		batchId      string
		requestBody  string
		expectedCode int
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: validConfig,
				jwtBatchValidator: fakeAuthValidator{
					errResp: nil,
				},
				sendFail: func(string, *model.FailRequest, auth.HriAzClaims, kafka.Writer) (int, interface{}) {
					return http.StatusOK, nil
				},
			},
			tenantId:     "1_a-tenant-id",
			batchId:      "test-batch-id",
			requestBody:  `{"actualRecordCount":100,"invalidRecordCount":10,"failureMessage":"a bad batch"}`,
			expectedCode: 200,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(tt.requestBody))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenant/:tenantId/batches/:batchId/action/sendFail")
			context.SetParamNames(param.TenantId, param.BatchId)
			context.SetParamValues(tt.tenantId, tt.batchId)
			context.Response().Header().Add(echo.HeaderXRequestID, requestId)
			if assert.NoError(t, tt.handler.SendFail(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)

			}
		})
	}
}

func Test_myHandler_CreateBatch(t *testing.T) {
	validConfig := config.Config{}

	logwrapper.Initialize("error", os.Stdout)

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

	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		requestBody  string
		expectedCode int
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: validConfig,
				jwtBatchValidator: fakeAuthValidator{
					errResp: nil,
				},
				createBatch: func(string, model.CreateBatch, auth.HriAzClaims, kafka.Writer) (int, interface{}) {
					return http.StatusCreated, map[string]interface{}{param.BatchId: "test-batch-id"}
				},
			},
			tenantId:     "1_a-tenant-id",
			requestBody:  validReqBody,
			expectedCode: 201,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.requestBody))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenant/:" + param.TenantId + "/batches")
			context.SetParamNames(param.TenantId)
			context.SetParamValues(tt.tenantId)

			if assert.NoError(t, tt.handler.CreateBatch(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)

			}
		})
	}
}

func Test_myHandler_GetByBatchId(t *testing.T) {
	validConfig := config.Config{}

	logwrapper.Initialize("error", os.Stdout)

	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		batchId      string
		expectedCode int
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: validConfig,
				jwtBatchValidator: fakeAuthValidator{
					errResp: nil,
				},
				getByBatchId: func(string, model.GetByIdBatch, auth.HriAzClaims) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"id": "batch7j3", "name": "monkeyBatch", "status": "started", "startDate": "2019-12-13", "dataType": "claims", "topic": "ingest-test", "recordCount": float64(1), "expectedRecordCount": float64(1)}
				},
			},
			tenantId:     "1_a-tenant-id",
			batchId:      "test-batch-id",
			expectedCode: 200,
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			context.SetPath("/hri/tenants/:" + param.TenantId + "/batches/:" + param.BatchId)
			context.SetParamNames(param.TenantId, param.BatchId)
			context.SetParamValues(tt.tenantId, tt.batchId)

			if assert.NoError(t, tt.handler.GetByBatchId(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)

			}
		})
	}
}

func Test_myHandler_Send_status_complete(t *testing.T) {
	validConfig := config.Config{}

	logwrapper.Initialize("error", os.Stdout)

	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		batchId      string
		requestBody  string
		expectedCode int
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: validConfig,
				jwtBatchValidator: fakeAuthValidator{
					errResp: nil,
				},
				sendStatusComplete: func(string, *model.SendCompleteRequest, auth.HriAzClaims, kafka.Writer) (int, interface{}) {
					return http.StatusOK, nil
				},
			},
			tenantId:     "1_a-tenant-id",
			batchId:      "test-batch-id",
			requestBody:  `{"expectedRecordCount": 100}`,
			expectedCode: 200,
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

			if assert.NoError(t, tt.handler.SendStatusComplete(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)

			}
		})
	}
}

func Test_myHandler_Process_Complete(t *testing.T) {
	validConfig := config.Config{}

	logwrapper.Initialize("error", os.Stdout)

	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		batchId      string
		requestBody  string
		expectedCode int
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: validConfig,
				jwtBatchValidator: fakeAuthValidator{
					errResp: nil,
				},
				processingCompleteBatch: func(string, *model.ProcessingCompleteRequest, auth.HriAzClaims, kafka.Writer) (int, interface{}) {
					return http.StatusOK, nil
				},
			},
			tenantId:     "1_a-tenant-id",
			batchId:      "test-batch-id",
			requestBody:  `{"actualRecordCount":100,"invalidRecordCount":10}`,
			expectedCode: 200,
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

			if assert.NoError(t, tt.handler.ProcessingCompleteBatch(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)

			}
		})
	}
}

func Test_myHandler_terminate(t *testing.T) {
	validConfig := config.Config{}

	logwrapper.Initialize("error", os.Stdout)

	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		batchId      string
		requestBody  string
		expectedCode int
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: validConfig,
				jwtBatchValidator: fakeAuthValidator{
					errResp: nil,
				},
				terminateBatch: func(string, *model.TerminateRequest, auth.HriAzClaims, kafka.Writer) (int, interface{}) {
					return http.StatusOK, nil
				},
			},
			tenantId:     "1_a-tenant-id",
			batchId:      "test-batch-id",
			requestBody:  `{"metadata":{"field1":"field1","field2":10}}`,
			expectedCode: 200,
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

			if assert.NoError(t, tt.handler.TerminateBatch(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)

			}
		})
	}
}
