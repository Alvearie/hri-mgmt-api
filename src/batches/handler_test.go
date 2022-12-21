package batches

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

const (
	validToken   = "BEaRer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNjUyMTA4MTQ0LCJleHAiOjI1NTIxMTE3NDR9.XxTTNBtgjX48iCM4FaV_hhhGenzhzrUaTWn6ooepK14" // expires in 2050
	validAztoken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IjJaUXBKM1VwYmpBWVhZR2FYRUpsOGxWMFRPSSIsImtpZCI6IjJaUXBKM1VwYmpBWVhZR2FYRUpsOGxWMFRPSSJ9.eyJhdWQiOiJjMzNhYzRkYS0yMWM2LTQyNmItYWJjYy0yN2UyNGZmMWNjZjkiLCJpc3MiOiJodHRwczovL3N0cy53aW5kb3dzLm5ldC9jZWFhNjNhYS01ZDVjLTRjN2QtOTRiMC0wMmY5YTNhYjZhOGMvIiwiaWF0IjoxNjYzNzQyMTM0LCJuYmYiOjE2NjM3NDIxMzQsImV4cCI6MTY2Mzc0NjAzNCwiYWlvIjoiRTJaZ1lGaHdablhvSG84elJvOHpQUlpNMU9VNENBQT0iLCJhcHBpZCI6ImMzM2FjNGRhLTIxYzYtNDI2Yi1hYmNjLTI3ZTI0ZmYxY2NmOSIsImFwcGlkYWNyIjoiMSIsImlkcCI6Imh0dHBzOi8vc3RzLndpbmRvd3MubmV0L2NlYWE2M2FhLTVkNWMtNGM3ZC05NGIwLTAyZjlhM2FiNmE4Yy8iLCJvaWQiOiI4YjFlN2E4MS03ZjRhLTQxYjAtYTE3MC1hZTE5Zjg0M2YyN2MiLCJyaCI6IjAuQVZBQXFtT3F6bHhkZlV5VXNBTDVvNnRxak5yRU9zUEdJV3RDcTh3bjRrX3h6UGxfQUFBLiIsInJvbGVzIjpbImhyaS5ocmlfaW50ZXJuYWwiLCJ0ZW5hbnRfcHJvdmlkZXIxMjM0IiwidGVuYW50X3BlbnRlc3QiLCJ0ZXN0X3JvbGUiLCJ0ZXN0IiwiaHJpX2NvbnN1bWVyIiwiaHJpX2RhdGFfaW50ZWdyYXRvciIsInByb3ZpZGVyMTIzNCJdLCJzdWIiOiI4YjFlN2E4MS03ZjRhLTQxYjAtYTE3MC1hZTE5Zjg0M2YyN2MiLCJ0aWQiOiJjZWFhNjNhYS01ZDVjLTRjN2QtOTRiMC0wMmY5YTNhYjZhOGMiLCJ1dGkiOiJnaUJlZUliWk9rS0ZYbGFIaHNfZ0FBIiwidmVyIjoiMS4wIn0.LdwhQpf5M1LSprQ9gk9abisbucKhNQtDnYEN1GLw_SqJ23DIFlfevlLikw075rVYvwf-4p_MJN3-7QZ2gMzTsqQ-G2x9IH4BO-oULlXeoHBQllDtmnYQFEesGogM0OjtXvoIAzUXCTPyxbjzTX3sPvghXuCSWPfu9ehVn8mRVtXuH0LWaU47XjTYzDE-RIFM2S80UCv7ZQErLrshC91OI0rNyc8ARPEc-TlnIK-KQ8HgehjFaapO6VL15s3YLO0zGA1v4RLnxbd36SdFfGxE_Vlv7WSLR5nB_n403FbiUUpwdIORaFRdMBEtNDbuI2RwHesUIEL6lrBrDxXPuaLIsA"
)

// Fake for the auth.Validator interface; just returns the desired values
type fakeBatchAuthValidator struct {
	errResp *response.ErrorDetailResponse
}

func (f fakeBatchAuthValidator) GetValidatedClaimsForTenant(_ string, _ string) *response.ErrorDetailResponse {
	return f.errResp
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
		// {
		// 	name: "jwtValidator error",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 		jwtValidator: fakeAuthValidator{
		// 			errResp: response.NewErrorDetailResponse(http.StatusBadRequest, "test-request-id", "jwtValidator error"),
		// 		},
		// 	},
		// 	tenantId:     "1_a-tenant-id",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"jwtValidator error\"}\n",
		// },
		// {
		// 	name: "invalid tenant id exclamation mark",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 	},
		// 	tenantId:     "invalid-tenant-id!",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'\"}\n",
		// },

		// {
		// 	name: "invalid tenant id uppercase",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 	},
		// 	tenantId:     "INVALID-tenant-id",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'\"}\n",
		// },
		// {
		// 	name: "400 on create",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 		jwtValidator: fakeAuthValidator{
		// 			errResp: nil,
		// 		},
		// 		createTenant: func(string, string) (int, interface{}) {
		// 			return http.StatusBadRequest, map[string]interface{}{"errorEventId": "test-request-id", "errorDescription": "Unable to create tenant"}
		// 		},
		// 	},
		// 	tenantId:     "1_a-tenant-id",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorDescription\":\"Unable to create tenant\",\"errorEventId\":\"test-request-id\"}\n",
		// },
		// {
		// 	name: "400 on create with _",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 		jwtValidator: fakeAuthValidator{
		// 			errResp: nil,
		// 		},
		// 		createTenant: func(string, string) (int, interface{}) {
		// 			return http.StatusBadRequest, map[string]interface{}{"errorEventId": "test-request-id", "errorDescription": "Unable to create tenant"}
		// 		},
		// 	},
		// 	tenantId:     "_",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"Unable to create a new tenant[_]:[400]\"}\n",
		// },
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
				// assert.Equal(t, tt.expectedBody, recorder.Body.String())
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
		name     string
		handler  theHandler
		tenantId string
		// batchId      string
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
			tenantId: "1_a-tenant-id",
			// batchId:      "test-batch-id",
			requestBody:  validReqBody,
			expectedCode: 201,
		},
		// {
		// 	name: "jwtValidator error",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 		jwtValidator: fakeAuthValidator{
		// 			errResp: response.NewErrorDetailResponse(http.StatusBadRequest, "test-request-id", "jwtValidator error"),
		// 		},
		// 	},
		// 	tenantId:     "1_a-tenant-id",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"jwtValidator error\"}\n",
		// },
		// {
		// 	name: "invalid tenant id exclamation mark",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 	},
		// 	tenantId:     "invalid-tenant-id!",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'\"}\n",
		// },

		// {
		// 	name: "invalid tenant id uppercase",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 	},
		// 	tenantId:     "INVALID-tenant-id",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'\"}\n",
		// },
		// {
		// 	name: "400 on create",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 		jwtValidator: fakeAuthValidator{
		// 			errResp: nil,
		// 		},
		// 		createTenant: func(string, string) (int, interface{}) {
		// 			return http.StatusBadRequest, map[string]interface{}{"errorEventId": "test-request-id", "errorDescription": "Unable to create tenant"}
		// 		},
		// 	},
		// 	tenantId:     "1_a-tenant-id",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorDescription\":\"Unable to create tenant\",\"errorEventId\":\"test-request-id\"}\n",
		// },
		// {
		// 	name: "400 on create with _",
		// 	handler: theHandler{
		// 		config: validConfig,
		// 		jwtValidator: fakeAuthValidator{
		// 			errResp: nil,
		// 		},
		// 		createTenant: func(string, string) (int, interface{}) {
		// 			return http.StatusBadRequest, map[string]interface{}{"errorEventId": "test-request-id", "errorDescription": "Unable to create tenant"}
		// 		},
		// 	},
		// 	tenantId:     "_",
		// 	expectedCode: http.StatusBadRequest,
		// 	expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"Unable to create a new tenant[_]:[400]\"}\n",
		// },
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
				// assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func Test_myHandler_GetByBatchId(t *testing.T) {
	validConfig := config.Config{}

	logwrapper.Initialize("error", os.Stdout)

	tests := []struct {
		name     string
		handler  theHandler
		tenantId string
		batchId  string
		// requestBody  string
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
			tenantId: "1_a-tenant-id",
			batchId:  "test-batch-id",
			// requestBody:  validReqBody,
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
				// assert.Equal(t, tt.expectedBody, recorder.Body.String())
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
				// assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}
