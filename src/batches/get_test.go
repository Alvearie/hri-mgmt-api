/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

// func TestGet(t *testing.T) {
// 	logwrapper.Initialize("error", os.Stdout)

// 	requestId := "reqZx01010"
// 	validTenantId := "tenant12x"
// 	validBatchName := "niceBatch"
// 	validStatusName := "started"
// 	validGte := "01/01/2020"
// 	validLte := "01/01/2021"
// 	validSize := defaultSize
// 	validFrom := defaultFrom

// 	tests := []struct {
// 		name         string
// 		params       model.GetBatch
// 		claims       auth.HriClaims
// 		transport    *test.FakeTransport
// 		expectedCode int
// 		expectedBody interface{}
// 	}{
// 		{
// 			name: "success with minimal params",
// 			params: model.GetBatch{
// 				TenantId: validTenantId,
// 			},
// 			claims: auth.HriClaims{Scope: auth.HriConsumer},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				"/"+validTenantId+"-batches/_search",
// 				test.ElasticCall{
// 					RequestQuery: "from=0&size=10&track_total_hits=true",
// 					ResponseBody: `
// 						{
// 							"hits":{
// 								"total" : {
// 								  "value" : 1,
// 								  "relation" : "eq"
// 								},
// 								"hits":[
// 									{
// 										"_id":"uuid",
// 										"_source":{
// 											"dataType" : "rspec-batch",
// 											"invalidThreshold" : -1,
// 											"metadata" : {
// 												"rspec1" : "test1"
// 											},
// 											"name" : "mybatch",
// 											"startDate" : "2021-02-24T18:08:36Z",
// 											"status" : "started",
// 											"topic" : "ingest.test.claims.in",
// 											"integratorId" : "modified-integrator-id"
// 										}
// 									}
// 								]
// 							}
// 						}`,
// 				},
// 			),
// 			expectedCode: http.StatusOK,
// 			expectedBody: map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "2021-02-24T18:08:36Z", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}},
// 		},
// 		{
// 			name: "success with all params",
// 			params: model.GetBatch{
// 				TenantId: validTenantId,
// 				Name:     &validBatchName,
// 				Status:   &validStatusName,
// 				GteDate:  &validGte,
// 				LteDate:  &validLte,
// 				Size:     &validSize,
// 				From:     &validFrom,
// 			},
// 			claims: auth.HriClaims{Scope: auth.HriConsumer},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				"/"+validTenantId+"-batches/_search",
// 				test.ElasticCall{
// 					RequestQuery: "from=0&size=10&track_total_hits=true",
// 					// Note that ] and [ must be escaped because RequestBody is used as a regex pattern
// 					RequestBody: `{"query":{"bool":{"must":\[{"term":{"name":"niceBatch"}},{"term":{"status":"started"}},{"range":{"startDate":{"gte":"01/01/2020","lte":"01/01/2021"}}}\]}}}` + "\n",
// 					ResponseBody: `
// 						{
// 							"hits":{
// 								"total" : {
// 								  "value" : 1,
// 								  "relation" : "eq"
// 								},
// 								"hits":[
// 									{
// 										"_id":"uuid",
// 										"_source":{
// 											"dataType" : "rspec-batch",
// 											"invalidThreshold" : -1,
// 											"metadata" : {
// 												"rspec1" : "test1"
// 											},
// 											"name" : "niceBatch",
// 											"startDate" : "01/02/2020",
// 											"status" : "started",
// 											"topic" : "ingest.test.claims.in",
// 											"integratorId" : "modified-integrator-id"
// 										}
// 									}
// 								]
// 							}
// 						}`,
// 				},
// 			),
// 			expectedCode: http.StatusOK,
// 			expectedBody: map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": validBatchName, "startDate": "01/02/2020", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}},
// 		},
// 		{
// 			name: "elastic client error",
// 			params: model.GetBatch{
// 				TenantId: validTenantId,
// 			},
// 			claims: auth.HriClaims{Scope: auth.HriConsumer},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				"/"+validTenantId+"-batches/_search",
// 				test.ElasticCall{
// 					RequestQuery: "from=0&size=10&track_total_hits=true",
// 					ResponseErr:  errors.New("client error"),
// 				},
// 			),
// 			expectedCode: http.StatusInternalServerError,
// 			expectedBody: response.NewErrorDetail(requestId, "Get batch failed: [500] elasticsearch client error: client error"),
// 		},
// 		{
// 			name:         "missing scopes no role set in Claim",
// 			claims:       auth.HriClaims{},
// 			transport:    test.NewFakeTransport(t),
// 			expectedCode: http.StatusUnauthorized,
// 			expectedBody: response.NewErrorDetail(requestId, auth.MsgAccessTokenMissingScopes),
// 		},
// 		{
// 			name: "integrator filter",
// 			params: model.GetBatch{
// 				TenantId: validTenantId,
// 			},
// 			claims: auth.HriClaims{Scope: auth.HriIntegrator, Subject: "clientId"},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				"/"+validTenantId+"-batches/_search",
// 				test.ElasticCall{
// 					RequestQuery: "from=0&size=10&track_total_hits=true",
// 					RequestBody:  `{"query":{"bool":{"must":\[{"term":{"integratorId":"clientId"}}\]}}}` + "\n",
// 					ResponseBody: `
// 						{
// 							"hits":{
// 								"total" : {
// 								  "value" : 1,
// 								  "relation" : "eq"
// 								},
// 								"hits":[
// 									{
// 										"_id":"uuid",
// 										"_source":{
// 											"dataType" : "rspec-batch",
// 											"invalidThreshold" : -1,
// 											"metadata" : {
// 												"rspec1" : "test1"
// 											},
// 											"name" : "mybatch",
// 											"startDate" : "2021-02-24T18:08:36Z",
// 											"status" : "started",
// 											"topic" : "ingest.test.claims.in",
// 											"integratorId" : "modified-integrator-id"
// 										}
// 									}
// 								]
// 							}
// 						}`,
// 				},
// 			),
// 			expectedCode: http.StatusOK,
// 			expectedBody: map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "2021-02-24T18:08:36Z", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}},
// 		},
// 		{
// 			name: "consumer & integrator no filter",
// 			params: model.GetBatch{
// 				TenantId: validTenantId,
// 			},
// 			claims: auth.HriClaims{Scope: auth.HriIntegrator + " " + auth.HriConsumer, Subject: "clientId"},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				"/"+validTenantId+"-batches/_search",
// 				test.ElasticCall{
// 					RequestQuery: "from=0&size=10&track_total_hits=true",
// 					ResponseBody: `
// 						{
// 							"hits":{
// 								"total" : {
// 								  "value" : 1,
// 								  "relation" : "eq"
// 								},
// 								"hits":[
// 									{
// 										"_id":"uuid",
// 										"_source":{
// 											"dataType" : "rspec-batch",
// 											"invalidThreshold" : -1,
// 											"metadata" : {
// 												"rspec1" : "test1"
// 											},
// 											"name" : "mybatch",
// 											"startDate" : "2021-02-24T18:08:36Z",
// 											"status" : "started",
// 											"topic" : "ingest.test.claims.in",
// 											"integratorId" : "modified-integrator-id"
// 										}
// 									}
// 								]
// 							}
// 						}`,
// 				},
// 			),
// 			expectedCode: http.StatusOK,
// 			expectedBody: map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "mybatch", "startDate": "2021-02-24T18:08:36Z", "status": "started", "topic": "ingest.test.claims.in", "integratorId": "modified-integrator-id", "metadata": map[string]interface{}{"rspec1": "test1"}}}},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			client, err := elastic.ClientFromTransport(tt.transport)
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			actualCode, actualBody := Get(requestId, tt.params, tt.claims, client)

// 			tt.transport.VerifyCalls()
// 			if actualCode != tt.expectedCode || !reflect.DeepEqual(tt.expectedBody, actualBody) {
// 				t.Errorf("Get()\n   actual: %v,%v\n expected: %v,%v", actualCode, actualBody, tt.expectedCode, tt.expectedBody)
// 			}
// 		})
// 	}
// }

// func TestGetNoAuth(t *testing.T) {
// 	logwrapper.Initialize("error", os.Stdout)

// 	requestId := "req109"
// 	validTenantId := "tenant24q"
// 	validBatchName := "naughtyBatch"
// 	validStatusName := "started"
// 	integratorId := auth.NoAuthFakeIntegrator
// 	validGte := "05/15/2020"
// 	validLte := "06/30/2021"
// 	validSize := defaultSize
// 	validFrom := defaultFrom

// 	tests := []struct {
// 		name         string
// 		params       model.GetBatch
// 		claims       auth.HriClaims
// 		transport    *test.FakeTransport
// 		expectedCode int
// 		expectedBody interface{}
// 	}{
// 		{
// 			name: "success with minimal params",
// 			params: model.GetBatch{
// 				TenantId: validTenantId,
// 			},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				"/"+validTenantId+"-batches/_search",
// 				test.ElasticCall{
// 					RequestQuery: "from=0&size=10&track_total_hits=true",
// 					ResponseBody: `
// 						{
// 							"hits":{
// 								"total" : {
// 								  "value" : 1,
// 								  "relation" : "eq"
// 								},
// 								"hits":[
// 									{
// 										"_id":"uuid",
// 										"_source":{
// 											"dataType" : "rspec-batch",
// 											"invalidThreshold" : -1,
// 											"metadata" : {
// 												"md1" : "porcupine"
// 											},
// 											"name" : "naughtyBatch",
// 											"startDate" : "2020-08-02T18:08:36Z",
// 											"status" : "started",
// 											"topic" : "ingest.test.claims.in",
// 											"integratorId" : "NoAuthUnkIntegrator"
// 										}
// 									}
// 								]
// 							}
// 						}`,
// 				},
// 			),
// 			expectedCode: http.StatusOK,
// 			expectedBody: map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": "naughtyBatch", "startDate": "2020-08-02T18:08:36Z", "status": "started", "topic": "ingest.test.claims.in", "integratorId": integratorId, "metadata": map[string]interface{}{"md1": "porcupine"}}}},
// 		},
// 		{
// 			name: "success with all params",
// 			params: model.GetBatch{
// 				TenantId: validTenantId,
// 				Name:     &validBatchName,
// 				Status:   &validStatusName,
// 				GteDate:  &validGte,
// 				LteDate:  &validLte,
// 				Size:     &validSize,
// 				From:     &validFrom,
// 			},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				"/"+validTenantId+"-batches/_search",
// 				test.ElasticCall{
// 					RequestQuery: "from=0&size=10&track_total_hits=true",
// 					// Note that ] and [ must be escaped because RequestBody is used as a regex pattern
// 					RequestBody: `{"query":{"bool":{"must":\[{"term":{"name":"naughtyBatch"}},{"term":{"status":"started"}},{"range":{"startDate":{"gte":"05/15/2020","lte":"06/30/2021"}}}\]}}}` + "\n",
// 					ResponseBody: `
// 					{
// 						"hits":{
// 							"total" : {
// 							  "value" : 1,
// 							  "relation" : "eq"
// 							},
// 							"hits":[
// 								{
// 									"_id":"uuid",
// 									"_source":{
// 										"dataType" : "rspec-batch",
// 										"invalidThreshold" : -1,
// 										"metadata" : {
// 											"md1" : "porcupine"
// 										},
// 										"name" : "naughtyBatch",
// 										"startDate" : "05/15/2020",
// 										"status" : "started",
// 										"topic" : "ingest.test.claims.in",
// 										"integratorId" : "NoAuthUnkIntegrator"
// 									}
// 								}
// 							]
// 						}
// 					}`,
// 				},
// 			),
// 			expectedCode: http.StatusOK,
// 			expectedBody: map[string]interface{}{"total": float64(1), "results": []interface{}{map[string]interface{}{"id": "uuid", "dataType": "rspec-batch", "invalidThreshold": float64(-1), "name": validBatchName, "startDate": validGte, "status": validStatusName, "topic": "ingest.test.claims.in", "integratorId": integratorId, "metadata": map[string]interface{}{"md1": "porcupine"}}}},
// 		},
// 	}

// 	var emptyClaims = auth.HriClaims{}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			client, err := elastic.ClientFromTransport(tt.transport)
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			actualCode, actualBody := GetNoAuth(requestId, tt.params, emptyClaims, client)
// 			if actualCode != tt.expectedCode || !reflect.DeepEqual(tt.expectedBody, actualBody) {
// 				t.Errorf("GetNoAuth()\n   actual: %v,%v\n expected: %v,%v", actualCode, actualBody, tt.expectedCode, tt.expectedBody)
// 			}
// 		})
// 	}
// }

// func TestBuildQuery(t *testing.T) {
// 	validBatchName := "niceBatch"
// 	validStatusName := "started"
// 	validGte := "01/01/2020"
// 	validLte := "01/01/2021"
// 	validSize := defaultSize
// 	validFrom := defaultFrom

// 	tests := []struct {
// 		name     string
// 		params   model.GetBatch
// 		claims   *auth.HriClaims
// 		expected map[string]interface{}
// 	}{
// 		{
// 			name:     "success empty query",
// 			params:   model.GetBatch{},
// 			claims:   &auth.HriClaims{},
// 			expected: nil,
// 		},
// 		{
// 			name: "success full query",
// 			params: model.GetBatch{
// 				Name:    &validBatchName,
// 				Status:  &validStatusName,
// 				GteDate: &validGte,
// 				LteDate: &validLte,
// 				Size:    &validSize,
// 				From:    &validFrom,
// 			},
// 			claims: &auth.HriClaims{Scope: auth.HriIntegrator, Subject: "subject"},
// 			expected: map[string]interface{}{"query": map[string]interface{}{"bool": map[string]interface{}{
// 				"must": []map[string]interface{}{
// 					{"term": map[string]interface{}{param.Name: validBatchName}},
// 					{"term": map[string]interface{}{param.Status: validStatusName}},
// 					{"range": map[string]interface{}{param.StartDate: map[string]interface{}{"gte": validGte, "lte": validLte}}},
// 					{"term": map[string]interface{}{param.IntegratorId: "subject"}},
// 				},
// 			}}},
// 		},
// 		{
// 			name: "success status only",
// 			params: model.GetBatch{
// 				Status: &validStatusName,
// 			},
// 			claims: &auth.HriClaims{Scope: auth.HriConsumer, Subject: "monkeez"},
// 			expected: map[string]interface{}{"query": map[string]interface{}{"bool": map[string]interface{}{
// 				"must": []map[string]interface{}{
// 					{"term": map[string]interface{}{param.Status: validStatusName}},
// 				},
// 			}}},
// 		},
// 		{
// 			name: "success name only",
// 			params: model.GetBatch{
// 				Name: &validBatchName,
// 			},
// 			claims: &auth.HriClaims{Scope: auth.HriConsumer, Subject: "bearZ"},
// 			expected: map[string]interface{}{"query": map[string]interface{}{"bool": map[string]interface{}{
// 				"must": []map[string]interface{}{
// 					{"term": map[string]interface{}{param.Name: validBatchName}},
// 				},
// 			}}},
// 		},
// 		{
// 			name: "success single date",
// 			params: model.GetBatch{
// 				GteDate: &validGte,
// 			},
// 			claims: &auth.HriClaims{Scope: auth.HriConsumer, Subject: "monkeez"},
// 			expected: map[string]interface{}{"query": map[string]interface{}{"bool": map[string]interface{}{
// 				"must": []map[string]interface{}{
// 					{"range": map[string]interface{}{param.StartDate: map[string]interface{}{"gte": validGte}}},
// 				},
// 			}}},
// 		},
// 		{
// 			name:   "success claims HriIntegrator",
// 			params: model.GetBatch{},
// 			claims: &auth.HriClaims{Scope: auth.HriIntegrator, Subject: "subject"},
// 			expected: map[string]interface{}{"query": map[string]interface{}{"bool": map[string]interface{}{
// 				"must": []map[string]interface{}{
// 					{"term": map[string]interface{}{param.IntegratorId: "subject"}},
// 				},
// 			}}},
// 		},
// 		{
// 			name:     "success claims HriConsumer",
// 			params:   model.GetBatch{},
// 			claims:   &auth.HriClaims{Scope: auth.HriConsumer, Subject: "subject"},
// 			expected: nil,
// 		},
// 		{
// 			name: "success path status and name NoAuth",
// 			params: model.GetBatch{
// 				Name:   &validBatchName,
// 				Status: &validStatusName,
// 			},
// 			claims: nil,
// 			expected: map[string]interface{}{"query": map[string]interface{}{"bool": map[string]interface{}{
// 				"must": []map[string]interface{}{
// 					{"term": map[string]interface{}{param.Name: validBatchName}},
// 					{"term": map[string]interface{}{param.Status: validStatusName}},
// 				},
// 			}}},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			query := buildQuery(tt.params, tt.claims)
// 			if !reflect.DeepEqual(tt.expected, query) {
// 				t.Errorf("buildQuery()\n  actual: %v\nexpected: %v", query, tt.expected)
// 			}
// 		})
// 	}
// }

// func TestGetClientSearchParams(t *testing.T) {
// 	anotherValidSize := 12
// 	anotherValidFrom := 14

// 	tests := []struct {
// 		name         string
// 		params       model.GetBatch
// 		expectedSize int
// 		expectedFrom int
// 	}{
// 		{
// 			name: "set values success",
// 			params: model.GetBatch{
// 				Size: &anotherValidSize,
// 				From: &anotherValidFrom,
// 			},
// 			expectedSize: 12,
// 			expectedFrom: 14,
// 		},
// 		{
// 			name:         "get default vals success",
// 			params:       model.GetBatch{},
// 			expectedSize: defaultSize,
// 			expectedFrom: defaultFrom,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			actualSize, actualFrom := getClientSearchParams(tt.params)
// 			assert.Equal(t, tt.expectedSize, actualSize)
// 			assert.Equal(t, tt.expectedFrom, actualFrom)
// 		})
// 	}
// }
