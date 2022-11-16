/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

// func TestGetById(t *testing.T) {
// 	logwrapper.Initialize("error", os.Stdout)

// 	requestId := "reqZx01010"
// 	subject := "dataIntegrator1"

// 	testCases := []struct {
// 		name         string
// 		tenantId     string
// 		batchId      string
// 		claims       auth.HriClaims
// 		transport    *test.FakeTransport
// 		expectedCode int
// 		expectedBody interface{}
// 	}{
// 		{
// 			name:     "success-case",
// 			tenantId: test.ValidTenantId,
// 			batchId:  test.ValidBatchId,
// 			claims:   auth.HriClaims{Scope: auth.HriConsumer},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index" : "%s-batches",
// 							"_type" : "_doc",
// 							"_id" : "%s",
// 							"_version" : 1,
// 							"_seq_no" : 0,
// 							"_primary_term" : 1,
// 							"found" : true,
// 							"_source" : {
// 								"name" : "monkeyBatch",
// 								"topic" : "ingest-test",
// 								"dataType" : "claims",
// 								"status" : "started",
// 								"recordCount" : 1,
// 								"startDate" : "2019-12-13"
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId),
// 				},
// 			),
// 			expectedCode: http.StatusOK,
// 			expectedBody: map[string]interface{}{"id": test.ValidBatchId, "name": "monkeyBatch", "status": "started", "startDate": "2019-12-13", "dataType": "claims", "topic": "ingest-test", "recordCount": float64(1), "expectedRecordCount": float64(1)},
// 		},
// 		{
// 			name:     "batch not found",
// 			tenantId: test.ValidTenantId,
// 			batchId:  "batch-no-existo",
// 			claims:   auth.HriClaims{Scope: auth.HriConsumer},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/batch-no-existo`, test.ValidTenantId),
// 				test.ElasticCall{
// 					ResponseStatusCode: http.StatusNotFound,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "batch-no-existo",
// 							"found": false
// 						}`, test.ValidTenantId),
// 				},
// 			),
// 			expectedCode: http.StatusNotFound,
// 			expectedBody: response.NewErrorDetail(requestId, fmt.Sprintf(`The document for tenantId: %s with document (batch) ID: batch-no-existo was not found`, test.ValidTenantId)),
// 		},
// 		{
// 			name:         "no role set in Claim error",
// 			tenantId:     test.ValidTenantId,
// 			batchId:      test.ValidBatchId,
// 			claims:       auth.HriClaims{},
// 			transport:    test.NewFakeTransport(t),
// 			expectedCode: http.StatusUnauthorized,
// 			expectedBody: response.NewErrorDetail(requestId, auth.MsgAccessTokenMissingScopes),
// 		},
// 		{
// 			name:     "bad tenantId",
// 			tenantId: "bad-tenant",
// 			batchId:  test.ValidBatchId,
// 			claims:   auth.HriClaims{Scope: auth.HriConsumer},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/bad-tenant-batches/_doc/%s`, test.ValidBatchId),
// 				test.ElasticCall{
// 					ResponseStatusCode: http.StatusNotFound,
// 					ResponseBody: `
// 						{
// 							"error": {
// 								"root_cause": [
// 									{
// 										"type" : "index_not_found_exception",
// 										"reason" : "no such index",
// 										"resource.type" : "index_or_alias",
// 										"resource.id" : "bad-tenant-batches",
// 										"index_uuid" : "_na_",
// 										"index" : "bad-tenant-batches"
// 									}
// 								],
// 								"type" : "index_not_found_exception",
// 								"reason" : "no such index",
// 								"resource.type" : "index_or_alias",
// 								"resource.id" : "bad-tenant-batches",
// 								"index_uuid" : "_na_",
// 								"index" : "bad-tenant-batches"
// 							},
// 							"status" : 404
// 						}`,
// 				},
// 			),
// 			expectedCode: http.StatusNotFound,
// 			expectedBody: response.NewErrorDetail(requestId, fmt.Sprintf(`The document for tenantId: bad-tenant with document (batch) ID: %s was not found`, test.ValidBatchId)),
// 		},
// 		{
// 			name:     "integrator role integrator id matches sub claim",
// 			tenantId: test.ValidTenantId,
// 			batchId:  test.ValidBatchId,
// 			claims:   auth.HriClaims{Scope: auth.HriIntegrator, Subject: subject},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					ResponseBody: fmt.Sprintf(`
// 					{
// 						"_index" : "%s-batches",
// 						"_type" : "_doc",
// 						"_id" : "%s",
// 						"_version" : 1,
// 						"_seq_no" : 0,
// 						"_primary_term" : 1,
// 						"found" : true,
// 						"_source" : {
// 							"name" : "monkeyBatch",
// 							"topic" : "ingest-test",
// 							"dataType" : "claims",
// 							"integratorId" : "dataIntegrator1",
// 							"status" : "started",
// 							"recordCount" : 1,
// 							"startDate" : "2019-12-13"
// 						}
// 					}`, test.ValidTenantId, test.ValidBatchId),
// 				},
// 			),
// 			expectedCode: http.StatusOK,
// 			expectedBody: map[string]interface{}{"id": test.ValidBatchId, "integratorId": "dataIntegrator1", "name": "monkeyBatch", "status": "started", "startDate": "2019-12-13", "dataType": "claims", "topic": "ingest-test", "recordCount": float64(1), "expectedRecordCount": float64(1)},
// 		},
// 		{
// 			name:     "integrator role integrator id Does NOT Match sub claim",
// 			tenantId: test.ValidTenantId,
// 			batchId:  test.ValidBatchId,
// 			claims:   auth.HriClaims{Scope: auth.HriIntegrator, Subject: "no_match_integrator"},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					ResponseBody: fmt.Sprintf(`
// 					{
// 						"_index" : "%s-batches",
// 						"_type" : "_doc",
// 						"_id" : "%s",
// 						"_version" : 1,
// 						"_seq_no" : 0,
// 						"_primary_term" : 1,
// 						"found" : true,
// 						"_source" : {
// 							"name" : "monkeyBatch",
// 							"topic" : "ingest-test",
// 							"dataType" : "claims",
// 							"integratorId" : "dataIntegrator1",
// 							"status" : "started",
// 							"recordCount" : 1,
// 							"startDate" : "2019-12-13"
// 						}
// 					}`, test.ValidTenantId, test.ValidBatchId),
// 				},
// 			),
// 			expectedCode: http.StatusUnauthorized,
// 			expectedBody: response.NewErrorDetail(requestId, "The token's sub claim (clientId): no_match_integrator does not match the data integratorId: dataIntegrator1"),
// 		},
// 		{
// 			name:     "elastic-error-response",
// 			tenantId: test.ValidTenantId,
// 			batchId:  test.ValidBatchId,
// 			claims:   auth.HriClaims{Scope: auth.HriConsumer},
// 			transport: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					ResponseBody: fmt.Sprintf(`
// 					{
// 						"_index" : "%s-batches",
// 						"_type" : "_doc",
// 						"_id" : "%s",
// 						"_version" : 1,
// 						"_seq_no" : 0,
// 						"_primary_term" : 1,
// 						"found" : true,
// 						"_source" : {
// 							"name" : "monkeyBatch",
// 							"topic" : "ingest-test",
// 							"dataType" : "claims",
// 							"integratorId" : "dataIntegrator1",
// 							"status" : "started",
// 							"recordCount" : 1,
// 							"startDate" : "2019-12-13"
// 						}
// 					}`, test.ValidTenantId, test.ValidBatchId),
// 					ResponseErr: errors.New("elasticErrMsg"),
// 				},
// 			),
// 			expectedCode: http.StatusInternalServerError,
// 			expectedBody: response.NewErrorDetail(requestId,
// 				fmt.Sprintf("Get batch by ID failed: [500] elasticsearch client error: elasticErrMsg"),
// 			),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		client, err := elastic.ClientFromTransport(tc.transport)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		t.Run(tc.name, func(t *testing.T) {
// 			actualCode, actualBody := GetById(requestId, getTestGetByIdBatch(tc.tenantId, tc.batchId), tc.claims, client)
// 			if actualCode != tc.expectedCode || !reflect.DeepEqual(tc.expectedBody, actualBody) {
// 				t.Errorf("GetById()\n   actual: %v,%v\n expected: %v,%v", actualCode, actualBody, tc.expectedCode, tc.expectedBody)
// 			}
// 		})
// 	}
// }

// func TestGetByIdNoAuth(t *testing.T) {
// 	logwrapper.Initialize("error", os.Stdout)

// 	requestId := "requestNoAuth"

// 	testCases := []struct {
// 		name         string
// 		tenantId     string
// 		batchId      string
// 		transport    *test.FakeTransport
// 		expectedCode int
// 		expectedBody interface{}
// 	}{
// 		{
// 			name:     "success-case",
// 			tenantId: test.ValidTenantId,
// 			batchId:  test.ValidBatchId,
// 			transport: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					ResponseBody: fmt.Sprintf(`
// 				{
// 					"_index" : "%s-batches",
// 					"_type" : "_doc",
// 					"_id" : "%s",
// 					"_version" : 1,
// 					"_seq_no" : 0,
// 					"_primary_term" : 1,
// 					"found" : true,
// 					"_source" : {
// 						"name" : "monkeyBatch75",
// 						"topic" : "ingest-test",
// 						"dataType" : "claims",
// 						"status" : "started",
// 						"recordCount" : 1,
// 						"startDate" : "2019-12-13"
// 					}
// 				}`, test.ValidTenantId, test.ValidBatchId),
// 				},
// 			),
// 			expectedCode: http.StatusOK,
// 			expectedBody: map[string]interface{}{"id": test.ValidBatchId, "name": "monkeyBatch75", "status": "started", "startDate": "2019-12-13", "dataType": "claims", "topic": "ingest-test", "recordCount": float64(1), "expectedRecordCount": float64(1)},
// 		},
// 		{
// 			name:     "batch not found",
// 			tenantId: test.ValidTenantId,
// 			batchId:  "batch-no-existo2",
// 			transport: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/batch-no-existo2`, test.ValidTenantId),
// 				test.ElasticCall{
// 					ResponseStatusCode: http.StatusNotFound,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "batch-no-existo2",
// 							"found": false
// 						}`, test.ValidTenantId),
// 				},
// 			),
// 			expectedCode: http.StatusNotFound,
// 			expectedBody: response.NewErrorDetail(requestId, fmt.Sprintf(`The document for tenantId: %s with document (batch) ID: batch-no-existo2 was not found`, test.ValidTenantId)),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		client, err := elastic.ClientFromTransport(tc.transport)
// 		if err != nil {
// 			t.Error(err)
// 		}

// 		var emptyClaims = auth.HriClaims{}
// 		t.Run(tc.name, func(t *testing.T) {
// 			actualCode, actualBody := GetByIdNoAuth(requestId, getTestGetByIdBatch(tc.tenantId, tc.batchId), emptyClaims, client)

// 			tc.transport.VerifyCalls()
// 			if actualCode != tc.expectedCode || !reflect.DeepEqual(tc.expectedBody, actualBody) {
// 				t.Errorf("GetByIdNoAuth()\n   actual: %v,%v\n expected: %v,%v", actualCode, actualBody, tc.expectedCode, tc.expectedBody)
// 			}
// 		})
// 	}
// }

// func TestDocumentNotFound(t *testing.T) {
// 	testCases := []struct {
// 		name             string
// 		shouldBeNotFound bool
// 		elasticErr       *elastic.ResponseError
// 		resultBody       map[string]interface{}
// 	}{
// 		{
// 			name:             "document cannot be found, elastic response contains no err",
// 			shouldBeNotFound: true,
// 			elasticErr:       &elastic.ResponseError{},
// 			resultBody:       map[string]interface{}{"found": false},
// 		},
// 		{
// 			name:             "document cannot be found, elastic response contains 404 err",
// 			shouldBeNotFound: true,
// 			elasticErr: &elastic.ResponseError{
// 				ErrorObj: fmt.Errorf("index_not_found_exception: no such index"), Code: http.StatusNotFound},
// 		},
// 		{
// 			name:             "document was found",
// 			shouldBeNotFound: false,
// 			elasticErr:       &elastic.ResponseError{},
// 			resultBody:       map[string]interface{}{"found": true},
// 		},
// 		{
// 			name:             "inconclusive, elastic gives error that does not have to do with a missing doc",
// 			shouldBeNotFound: false,
// 			elasticErr: &elastic.ResponseError{
// 				ErrorObj: fmt.Errorf("this is some other error"), Code: http.StatusInternalServerError},
// 		},
// 		{
// 			name:             "inconclusive, no elastic error but result body is missing",
// 			shouldBeNotFound: false,
// 			elasticErr:       &elastic.ResponseError{},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			notFound := documentNotFound(tc.elasticErr, tc.resultBody)
// 			assert.Equal(t, tc.shouldBeNotFound, notFound)
// 		})
// 	}
// }

// func TestCheckBatchAuthorization(t *testing.T) {
// 	subject := "dataIntegrator1"
// 	requestId := "reqZx01010"

// 	testCases := []struct {
// 		name              string
// 		claims            auth.HriClaims
// 		resultBody        map[string]interface{}
// 		expectedErrDetail *response.ErrorDetailResponse
// 	}{
// 		{
// 			name:   "empty_claim_scope_return_error",
// 			claims: auth.HriClaims{},
// 			resultBody: map[string]interface{}{
// 				"_index":        test.ValidTenantId + "-batches",
// 				"_type":         "_doc",
// 				"_id":           test.ValidBatchId,
// 				"_version":      1,
// 				"_seq_no":       0,
// 				"_primary_term": 1,
// 				"found":         true,
// 				"_source": map[string]interface{}{
// 					"name":         "monkeyBatch",
// 					"topic":        "ingest-test",
// 					"dataType":     "claims",
// 					"integratorId": subject,
// 					"status":       "started",
// 					"recordCount":  1,
// 					"startDate":    "2019-12-13",
// 				},
// 			},
// 			expectedErrDetail: response.NewErrorDetailResponse(
// 				http.StatusUnauthorized, requestId, auth.MsgAccessTokenMissingScopes),
// 		},
// 		{
// 			name:              "consumer_role_returns_authorized",
// 			claims:            auth.HriClaims{Scope: auth.HriConsumer},
// 			resultBody:        map[string]interface{}{},
// 			expectedErrDetail: nil,
// 		},
// 		{
// 			name:   "missing_result_source_error",
// 			claims: auth.HriClaims{Scope: auth.HriIntegrator, Subject: subject},
// 			resultBody: map[string]interface{}{
// 				"_index":        test.ValidTenantId + "-batches",
// 				"_type":         "_doc",
// 				"_id":           test.ValidBatchId,
// 				"_version":      1,
// 				"_seq_no":       0,
// 				"_primary_term": 1,
// 				"found":         true,
// 			},
// 			expectedErrDetail: response.NewErrorDetailResponse(
// 				http.StatusInternalServerError, requestId, msgMissingStatusElem),
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			actualErrDetail := checkBatchAuthorization(requestId, &tc.claims, tc.resultBody)
// 			if !reflect.DeepEqual(tc.expectedErrDetail, actualErrDetail) {
// 				t.Errorf("GetById() = %v, expected %v", actualErrDetail, tc.expectedErrDetail)
// 			}
// 		})
// 	}
// }

// func getTestGetByIdBatch(tenantId string, batchId string) model.GetByIdBatch {
// 	return model.GetByIdBatch{
// 		TenantId: tenantId,
// 		BatchId:  batchId,
// 	}
// }
