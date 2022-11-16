/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package tenants

/*func TestNewHandler(t *testing.T) {
	config := config.Config{}

	handler := NewHandler(config).(*theHandler)
	assert.Equal(t, config, handler.config)
	// This asserts that they are the same function by memory address
	assert.Equal(t, reflect.ValueOf(elastic.CheckElasticIAM), reflect.ValueOf(handler.checkElasticIAM))
	assert.Equal(t, reflect.ValueOf(Create), reflect.ValueOf(handler.create))
	assert.Equal(t, reflect.ValueOf(Get), reflect.ValueOf(handler.get))
	assert.Equal(t, reflect.ValueOf(Delete), reflect.ValueOf(handler.delete))
	assert.Equal(t, reflect.ValueOf(GetById), reflect.ValueOf(handler.getById))
}

func Test_myHandler_Create(t *testing.T) {
	conf := config.Config{
		ElasticUrl:        "https://elastic.url",
		ElasticUsername:   "myElasticUser",
		ElasticPassword:   "myElasticPassword",
		ElasticCert:       "bXlFbGFzdGljQ2VydA==", // myElasticCert
		ElasticServiceCrn: "myElasticCrn",
	}
	badConf := config.Config{
		ElasticUrl:        "https://elastic.invalid  .url",
		ElasticUsername:   "myElasticUser",
		ElasticPassword:   "myElasticPassword",
		ElasticCert:       "bXlFbGFzdGljQ2VydA==", // myElasticCert
		ElasticServiceCrn: "myElasticCrn",
	}

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
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
				create: func(string, string, *elasticsearch.Client) (int, interface{}) {
					return http.StatusCreated, map[string]interface{}{"tenantId": "1_a-tenant-id"}
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusCreated,
			expectedBody: "{\"tenantId\":\"1_a-tenant-id\"}\n",
		},
		{
			name: "invalid tenant id exclamation mark",
			handler: theHandler{
				config: conf,
			},
			tenantId:     "invalid-tenant-id!",
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'\"}\n",
		},
		{
			name: "invalid tenant id uppercase",
			handler: theHandler{
				config: conf,
			},
			tenantId:     "INVALID-tenant-id",
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"invalid request arguments:\\n- tenantId (url path parameter) may only contain lower-case alpha-numeric chars and the following 2 special chars: '-', '_'\"}\n",
		},
		{
			name: "400 on create",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
				create: func(string, string, *elasticsearch.Client) (int, interface{}) {
					return http.StatusBadRequest, map[string]interface{}{"errorEventId": "test-request-id", "errorDescription": "Unable to create tenant"}
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorDescription\":\"Unable to create tenant\",\"errorEventId\":\"test-request-id\"}\n",
		},
		{
			name: "401 on iam check",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 401, errors.New("unauthorized")
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusUnauthorized,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"unauthorized\"}\n",
		},
		{
			name: "500 on iam check",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 500, errors.New("500 internal server error")
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"500 internal server error\"}\n",
		},
		{
			name: "500 on bad config invalid elastic url",
			handler: theHandler{
				config: badConf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"cannot create client: cannot parse url: parse \\\"https://elastic.invalid  .url\\\": invalid character \\\" \\\" in host name\"}\n",
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

			if assert.NoError(t, tt.handler.Create(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func Test_myHandler_Get(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	conf := config.Config{
		ElasticUrl:        "https://elastic.url",
		ElasticUsername:   "myElasticUser",
		ElasticPassword:   "myElasticPassword",
		ElasticCert:       "bXlFbGFzdGljQ2VydA==", // myElasticCert
		ElasticServiceCrn: "myElasticCrn",
	}
	badConf := config.Config{
		ElasticUrl:        "https://elastic.invalid  .url",
		ElasticUsername:   "myElasticUser",
		ElasticPassword:   "myElasticPassword",
		ElasticCert:       "bXlFbGFzdGljQ2VydA==", // myElasticCert
		ElasticServiceCrn: "myElasticCrn",
	}

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

	tests := []struct {
		name         string
		handler      theHandler
		expectedCode int
		expectedBody interface{}
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
				get: func(string, *elasticsearch.Client) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"results": indices}
				},
			},
			expectedCode: http.StatusOK,
			expectedBody: "{\"results\":[{\"id\":\"pi001\"},{\"id\":\"pi002\"},{\"id\":\"qatenant\"}]}\n",
		},
		{
			name: "401 on iam check",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 401, errors.New("unauthorized")
				},
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"unauthorized\"}\n",
		},
		{
			name: "500 on iam check",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 500, errors.New("500 internal server error")
				},
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"500 internal server error\"}\n",
		},
		{
			name: "500 on bad config invalid elastic url",
			handler: theHandler{
				config: badConf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"cannot create client: cannot parse url: parse \\\"https://elastic.invalid  .url\\\": invalid character \\\" \\\" in host name\"}\n",
		},
		{
			name: "500 on get",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
				get: func(string, *elasticsearch.Client) (int, interface{}) {
					return http.StatusInternalServerError, map[string]interface{}{"errorEventId": "test-request-id", "errorDescription": "Could not retrieve tenants"}
				},
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorDescription\":\"Could not retrieve tenants\",\"errorEventId\":\"test-request-id\"}\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/hri/tenants", nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			request.Header.Set(echo.HeaderXRequestID, "test-request-id")
			request.Header.Set(echo.HeaderAuthorization, "Bearer 123456789")
			context.SetPath("/hri/tenants")

			if assert.NoError(t, tt.handler.Get(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func Test_myHandler_Delete(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	conf := config.Config{
		ElasticUrl:        "https://elastic.url",
		ElasticUsername:   "myElasticUser",
		ElasticPassword:   "myElasticPassword",
		ElasticCert:       "bXlFbGFzdGljQ2VydA==", // myElasticCert
		ElasticServiceCrn: "myElasticCrn",
	}
	badConf := config.Config{
		ElasticUrl:        "https://elastic.invalid  .url",
		ElasticUsername:   "myElasticUser",
		ElasticPassword:   "myElasticPassword",
		ElasticCert:       "bXlFbGFzdGljQ2VydA==", // myElasticCert
		ElasticServiceCrn: "myElasticCrn",
	}

	tests := []struct {
		name         string
		handler      theHandler
		tenantId     string
		expectedCode int
		expectedBody interface{}
	}{
		{
			name: "happy path",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
				delete: func(string, string, *elasticsearch.Client) (int, interface{}) {
					return http.StatusOK, nil
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
		{
			name: "Unauthorized error 401 on iam check",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 401, errors.New("unauthorized")
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusUnauthorized,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"unauthorized\"}\n",
		},
		{
			name: "500 on iam check",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 500, errors.New("500 internal server error")
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"500 internal server error\"}\n",
		},
		{
			name: "500 on bad config invalid elastic url",
			handler: theHandler{
				config: badConf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"test-request-id\",\"errorDescription\":\"cannot create client: cannot parse url: parse \\\"https://elastic.invalid  .url\\\": invalid character \\\" \\\" in host name\"}\n",
		},
		{
			name: "Unable to Delete error 400 on delete",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
				delete: func(string, string, *elasticsearch.Client) (int, interface{}) {
					return http.StatusBadRequest, map[string]interface{}{"errorEventId": "test-request-id", "errorDescription": "Unable to delete tenant"}
				},
			},
			tenantId:     "1_a-tenant-id",
			expectedCode: http.StatusBadRequest,
			expectedBody: "{\"errorDescription\":\"Unable to delete tenant\",\"errorEventId\":\"test-request-id\"}\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/hri/tenants/"+tt.tenantId, nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			request.Header.Set(echo.HeaderXRequestID, "test-request-id")
			request.Header.Set(echo.HeaderAuthorization, "Bearer 123456789")
			context.SetPath("/hri/tenants/:tenantId")
			context.SetParamNames(param.TenantId)
			context.SetParamValues(tt.tenantId)

			if assert.NoError(t, tt.handler.Delete(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}

func Test_myHandler_GetById(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	conf := config.Config{
		ElasticUrl:        "https://elastic.url",
		ElasticUsername:   "JayZ",
		ElasticPassword:   "Roc-A-Fella",
		ElasticCert:       "bXlFbGFzdGljQ2VydA==", // myElasticCert
		ElasticServiceCrn: "myElasticCrn",
	}
	badConf := config.Config{
		ElasticUrl:        "https://some-invalid-es  .url",
		ElasticUsername:   "monkeyUser",
		ElasticPassword:   "monkeyPassword",
		ElasticCert:       "NO_CERT", // myElasticCert
		ElasticServiceCrn: "badElasticCrn",
	}

	validTenantId := "valid-tenant-id"
	requestId := "req-id-129"
	id1 := make(map[string]interface{})
	var indices []interface{}
	id1["id"] = validTenantId
	indices = append(indices, id1)

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
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
				getById: func(string, string, *elasticsearch.Client) (int, interface{}) {
					return http.StatusOK, map[string]interface{}{"results": indices}
				},
			},
			tenantId:     validTenantId,
			expectedCode: http.StatusOK,
			expectedBody: "{\"results\":[{\"id\":\"" + validTenantId + "\"}]}\n",
		},
		{
			name: "unauthorized error 401 on iam check",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 401, errors.New("elastic IAM authentication returned 401")
				},
			},
			tenantId:     validTenantId,
			expectedCode: http.StatusUnauthorized,
			expectedBody: "{\"errorEventId\":\"" + requestId + "\",\"errorDescription\":\"elastic IAM authentication returned 401\"}\n",
		},
		{
			name: "Internal Server error 500 on iam check",
			handler: theHandler{
				config: conf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 500, errors.New("500 internal server error")
				},
			},
			tenantId:     validTenantId,
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"" + requestId + "\",\"errorDescription\":\"500 internal server error\"}\n",
		},
		{
			name: "Bad config invalid elastic url returns 500 Internal Server Error",
			handler: theHandler{
				config: badConf,
				checkElasticIAM: func(string, string, elastic.ResourceControllerService) (int, error) {
					return 200, nil
				},
			},
			tenantId:     validTenantId,
			expectedCode: http.StatusInternalServerError,
			expectedBody: "{\"errorEventId\":\"" + requestId + "\",\"errorDescription\":\"cannot create client: cannot parse url: parse \\\"" +
				badConf.ElasticUrl + "\\\": invalid character \\\" \\\" in host name\"}\n",
		},
	}

	e := test.GetTestServer()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/hri/tenants/"+tt.tenantId, nil)
			context, recorder := test.PrepareHeadersContextRecorder(request, e)
			request.Header.Set(echo.HeaderXRequestID, requestId)
			request.Header.Set(echo.HeaderAuthorization, "Bearer 123456789")
			context.SetPath("/hri/tenants/:tenantId")
			context.SetParamNames(param.TenantId)
			context.SetParamValues(tt.tenantId)

			if assert.NoError(t, tt.handler.GetById(context)) {
				assert.Equal(t, tt.expectedCode, recorder.Code)
				assert.Equal(t, tt.expectedBody, recorder.Body.String())
			}
		})
	}
}*/
