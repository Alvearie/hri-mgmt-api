package tenants

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/config"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

const (
	validToken   = "BEaRer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNjUyMTA4MTQ0LCJleHAiOjI1NTIxMTE3NDR9.XxTTNBtgjX48iCM4FaV_hhhGenzhzrUaTWn6ooepK14" // expires in 2050
	validAztoken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6IjJaUXBKM1VwYmpBWVhZR2FYRUpsOGxWMFRPSSIsImtpZCI6IjJaUXBKM1VwYmpBWVhZR2FYRUpsOGxWMFRPSSJ9.eyJhdWQiOiJjMzNhYzRkYS0yMWM2LTQyNmItYWJjYy0yN2UyNGZmMWNjZjkiLCJpc3MiOiJodHRwczovL3N0cy53aW5kb3dzLm5ldC9jZWFhNjNhYS01ZDVjLTRjN2QtOTRiMC0wMmY5YTNhYjZhOGMvIiwiaWF0IjoxNjYzNzQyMTM0LCJuYmYiOjE2NjM3NDIxMzQsImV4cCI6MTY2Mzc0NjAzNCwiYWlvIjoiRTJaZ1lGaHdablhvSG84elJvOHpQUlpNMU9VNENBQT0iLCJhcHBpZCI6ImMzM2FjNGRhLTIxYzYtNDI2Yi1hYmNjLTI3ZTI0ZmYxY2NmOSIsImFwcGlkYWNyIjoiMSIsImlkcCI6Imh0dHBzOi8vc3RzLndpbmRvd3MubmV0L2NlYWE2M2FhLTVkNWMtNGM3ZC05NGIwLTAyZjlhM2FiNmE4Yy8iLCJvaWQiOiI4YjFlN2E4MS03ZjRhLTQxYjAtYTE3MC1hZTE5Zjg0M2YyN2MiLCJyaCI6IjAuQVZBQXFtT3F6bHhkZlV5VXNBTDVvNnRxak5yRU9zUEdJV3RDcTh3bjRrX3h6UGxfQUFBLiIsInJvbGVzIjpbImhyaS5ocmlfaW50ZXJuYWwiLCJ0ZW5hbnRfcHJvdmlkZXIxMjM0IiwidGVuYW50X3BlbnRlc3QiLCJ0ZXN0X3JvbGUiLCJ0ZXN0IiwiaHJpX2NvbnN1bWVyIiwiaHJpX2RhdGFfaW50ZWdyYXRvciIsInByb3ZpZGVyMTIzNCJdLCJzdWIiOiI4YjFlN2E4MS03ZjRhLTQxYjAtYTE3MC1hZTE5Zjg0M2YyN2MiLCJ0aWQiOiJjZWFhNjNhYS01ZDVjLTRjN2QtOTRiMC0wMmY5YTNhYjZhOGMiLCJ1dGkiOiJnaUJlZUliWk9rS0ZYbGFIaHNfZ0FBIiwidmVyIjoiMS4wIn0.LdwhQpf5M1LSprQ9gk9abisbucKhNQtDnYEN1GLw_SqJ23DIFlfevlLikw075rVYvwf-4p_MJN3-7QZ2gMzTsqQ-G2x9IH4BO-oULlXeoHBQllDtmnYQFEesGogM0OjtXvoIAzUXCTPyxbjzTX3sPvghXuCSWPfu9ehVn8mRVtXuH0LWaU47XjTYzDE-RIFM2S80UCv7ZQErLrshC91OI0rNyc8ARPEc-TlnIK-KQ8HgehjFaapO6VL15s3YLO0zGA1v4RLnxbd36SdFfGxE_Vlv7WSLR5nB_n403FbiUUpwdIORaFRdMBEtNDbuI2RwHesUIEL6lrBrDxXPuaLIsA"
)

// Fake for the auth.Validator interface; just returns the desired values
type fakeAuthValidator struct {
	errResp *response.ErrorDetailResponse
}

func (f fakeAuthValidator) GetValidatedClaimsForTenant(_ string, _ string) *response.ErrorDetailResponse {
	return f.errResp
}

func TestNewHandler(t *testing.T) {
	config := config.Config{}

	handler := NewHandler(config).(*theHandler)
	assert.Equal(t, config, handler.config)
	// This asserts that they are the same function by memory address
	assert.Equal(t, reflect.ValueOf(CreateTenant), reflect.ValueOf(handler.createTenant))
	assert.Equal(t, reflect.ValueOf(GetTenants), reflect.ValueOf(handler.getTenants))
	assert.Equal(t, reflect.ValueOf(DeleteTenant), reflect.ValueOf(handler.deleteTenant))
	assert.Equal(t, reflect.ValueOf(GetTenantById), reflect.ValueOf(handler.getTenantById))
}

func Test_myHandler_Create(t *testing.T) {
	validConfig := config.Config{}
	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		expectedCode int
		expectedBody string
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: nil,
				},
				createTenant: func(string, string) (int, interface{}) {
					return http.StatusCreated, map[string]interface{}{"tenantId": "1_a-tenant-id"}
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusCreated,
			expectedBody: "{\"tenantId\":\"1_a-tenant-id\"}\n",
		},
		{
			name: "jwtValidator error",
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: response.NewErrorDetailResponse(http.StatusBadRequest, "test-request-id", "jwtValidator error"),
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"jwtValidator error\"}\n",
		},
		{
			name: "invalid tenant id exclamation mark",
			handler: theHandler{
				config: validConfig,
			},
			tenantId:     "invalid-tenant-id!",
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'\"}\n",
		},

		{
			name: "invalid tenant id uppercase",
			handler: theHandler{
				config: validConfig,
			},
			tenantId:     "INVALID-tenant-id",
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'\"}\n",
		},
		{
			name: "400 on create",
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: nil,
				},
				createTenant: func(string, string) (int, interface{}) {
					return http.StatusBadRequest, map[string]interface{}{"errorEventId": "test-request-id", "errorDescription": "Unable to create tenant"}
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorDescription\":\"Unable to create tenant\",\"errorEventId\":\"test-request-id\"}\n",
		},
		{
			name: "400 on create with _",
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: nil,
				},
				createTenant: func(string, string) (int, interface{}) {
					return http.StatusBadRequest, map[string]interface{}{"errorEventId": "test-request-id", "errorDescription": "Unable to create tenant"}
				},
			},
			tenantId:     "_",
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"Unable to create a new tenant[_]:[400]\"}\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/hri/tenants/"+tt.tenantId, nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			request.Header.Set(echo.HeaderXRequestID, "test-request-id")
			request.Header.Set(echo.HeaderAuthorization, "Bearer 123456789")
			context.SetPath("/hri/tenants/:tenantId")
			context.SetParamNames(param.TenantId)
			context.SetParamValues(tt.tenantId)
			if assert.NoError(t, tt.handler.CreateTenant(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func Test_myHandler_Get(t *testing.T) {
	validConfig := config.Config{}
	id1 := make(map[string]interface{})
	id2 := make(map[string]interface{})
	id3 := make(map[string]interface{})
	var indices []interface{}
	id1["id"] = "pi001"
	indices = append(indices, id1)
	id2["id"] = "pi002"
	indices = append(indices, id2)
	id3["id"] = "qatenant"
	indices = append(indices, id3)
	request_Id := "requestId1"
	tests := []struct {
		name         string
		handler      theHandler
		expectedCode int
		expectedBody string
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: nil,
				},
				getTenants: func(string) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"results": indices}
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: "{\"results\":[{\"id\":\"pi001\"},{\"id\":\"pi002\"},{\"id\":\"qatenant\"}]}\n",
		},
		{
			name: "errResp jwtValidator",
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: response.NewErrorDetailResponse(http.StatusBadRequest, request_Id, "jwtValidator error"),
				},
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"jwtValidator error\"}\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/hri/tenants/", nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			request.Header.Set(echo.HeaderXRequestID, "test-request-id")
			request.Header.Set(echo.HeaderAuthorization, "Bearer 123456789")
			context.SetPath("/hri/tenants/:tenantId")
			context.SetParamNames(param.TenantId)
			//context.SetParamValues(tt.tenantId)
			if assert.NoError(t, tt.handler.GetTenants(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func Test_myHandler_GetById(t *testing.T) {
	validConfig := config.Config{}
	validTenantId := "valid-tenant-id"
	requestId := "req-id-129"
	id1 := make(map[string]interface{})
	var indices []interface{}
	id1["id"] = validTenantId
	indices = append(indices, id1)
	tests := []struct {
		name         string
		handler      theHandler
		expectedCode int
		expectedBody string
		tenantId     string
	}{
		{
			name:     "happy path",
			tenantId: validTenantId,
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: nil,
				},
				getTenantById: func(string, string) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"results": indices}
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: "{\"results\":[{\"id\":\"" + validTenantId + "\"}]}\n",
		},
		{
			name:     "errResp jwtValidator",
			tenantId: validTenantId,
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: response.NewErrorDetailResponse(http.StatusBadRequest, requestId, "jwtValidator error"),
				},
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorEventId\":\"req-id-129\",\"errorDescription\":\"jwtValidator error\"}\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/hri/tenants/", nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			request.Header.Set(echo.HeaderXRequestID, "test-request-id")
			request.Header.Set(echo.HeaderAuthorization, "Bearer 123456789")
			context.SetPath("/hri/tenants/:tenantId")
			context.SetParamNames(param.TenantId)
			context.SetParamValues(tt.tenantId)
			if assert.NoError(t, tt.handler.GetTenantById(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func Test_myHandler_Delete(t *testing.T) {
	validConfig := config.Config{}
	validTenantId := "valid-tenant-id"
	requestId := "req-id-129"
	tests := []struct {
		name         string
		handler      theHandler
		expectedCode int
		expectedBody string
		tenantId     string
	}{
		{
			name:     "happy path",
			tenantId: validTenantId,
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: nil,
				},
				deleteTenant: func(string, string) (int, interface{}) {
					return http.StatusOK, nil
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name:     "errResp jwtValidator",
			tenantId: validTenantId,
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: response.NewErrorDetailResponse(http.StatusBadRequest, requestId, "jwtValidator error"),
				},
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorEventId\":\"req-id-129\",\"errorDescription\":\"jwtValidator error\"}\n",
		},
		{
			name:     "Delete Tenant error",
			tenantId: validTenantId,
			handler: theHandler{
				config: validConfig,
				jwtValidator: fakeAuthValidator{
					errResp: nil,
				},
				deleteTenant: func(string, string) (int, interface{}) {
					return http.StatusBadRequest, "Delete Tenant error"
				},
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: "\"Delete Tenant error\"\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/hri/tenants/", nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			request.Header.Set(echo.HeaderXRequestID, "test-request-id")
			request.Header.Set(echo.HeaderAuthorization, "Bearer 123456789")
			context.SetPath("/hri/tenants/:tenantId")
			context.SetParamNames(param.TenantId)
			context.SetParamValues(tt.tenantId)
			if assert.NoError(t, tt.handler.DeleteTenant(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}
