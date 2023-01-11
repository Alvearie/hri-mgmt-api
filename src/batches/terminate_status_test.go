package batches

import (
	"errors"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// func Test_getTerminateUpdateScript(t *testing.T) {
// 	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

// 	tests := []struct {
// 		name    string
// 		request *model.TerminateRequest
// 		claims  auth.HriClaims
// 		// Note that the following chars must be escaped because expectedScript is used as a regex pattern: ., ), (, [, ]
// 		expectedRequest map[string]interface{}
// 		metadata        bool
// 	}{
// 		{
// 			name:    "UpdateScript returns expected script without metadata",
// 			request: getTestTerminateRequest(nil),
// 			claims:  validClaims,
// 			expectedRequest: map[string]interface{}{
// 				"script": map[string]interface{}{
// 					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}`,
// 				},
// 			},
// 			metadata: false,
// 		},
// 		{
// 			name:    "UpdateScript returns expected script with metadata",
// 			request: getTestTerminateRequest(map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5}),
// 			claims:  validClaims,
// 			expectedRequest: map[string]interface{}{
// 				"script": map[string]interface{}{
// 					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `'; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}`,
// 					"lang":   "painless",
// 					"params": map[string]interface{}{"metadata": map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5}},
// 				},
// 			},
// 			metadata: true,
// 		},
// 		{
// 			name:    "Missing claim.Subject",
// 			request: getTestTerminateRequest(nil),
// 			claims:  auth.HriClaims{},
// 			expectedRequest: map[string]interface{}{
// 				"script": map[string]interface{}{
// 					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == ''\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200;} else {ctx\.op = 'none'}`,
// 				},
// 			},
// 			metadata: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			request := getTerminateUpdateScript(tt.request, tt.claims.Subject)

// 			if err := RequestCompareScriptTest(tt.expectedRequest, request); err != nil {
// 				t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", request, tt.expectedRequest, err)
// 			} else if tt.metadata {
// 				if err := RequestCompareWithMetadataTest(tt.expectedRequest, request); err != nil {
// 					t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", request, tt.expectedRequest, err)
// 				}
// 			}
// 		})
// 	}
// }

// func TestUpdateStatus_Terminate(t *testing.T) {
// 	const (
// 		scriptTerminate        = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
// 		scriptTerminateWrongId = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'wrong id'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
// 		currentStatus          = status.SendCompleted
// 	)

// 	logwrapper.Initialize("error", os.Stdout)
// 	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

// 	terminatedBatch := map[string]interface{}{
// 		param.BatchId:             test.ValidBatchId,
// 		param.Name:                batchName,
// 		param.IntegratorId:        integratorId,
// 		param.Topic:               batchTopic,
// 		param.DataType:            batchDataType,
// 		param.Status:              status.Terminated.String(),
// 		param.StartDate:           batchStartDate,
// 		param.RecordCount:         batchExpectedRecordCount,
// 		param.ExpectedRecordCount: batchExpectedRecordCount,
// 	}
// 	terminatedJSON, err := json.Marshal(terminatedBatch)
// 	if err != nil {
// 		t.Errorf("Unable to create batch JSON string: %s", err.Error())
// 	}

// 	tests := []struct {
// 		name                 string
// 		request              *model.TerminateRequest
// 		claims               auth.HriClaims
// 		ft                   *test.FakeTransport
// 		writerError          error
// 		expectedNotification map[string]interface{}
// 		expectedCode         int
// 		expectedResponse     interface{}
// 	}{
// 		{
// 			name:    "successful Terminate",
// 			request: getTestTerminateRequest(nil),
// 			claims:  validClaims,
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptTerminate,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, terminatedJSON),
// 				},
// 			),
// 			expectedNotification: terminatedBatch,
// 			expectedCode:         http.StatusOK,
// 			expectedResponse:     nil,
// 		},
// 		{
// 			name:             "401 Unauthorized when Data Integrator scope is missing",
// 			request:          getTestTerminateRequest(nil),
// 			claims:           auth.HriClaims{},
// 			expectedCode:     http.StatusUnauthorized,
// 			expectedResponse: response.NewErrorDetail(requestId, "Must have hri_data_integrator role to terminate a batch"),
// 		},
// 		{
// 			name:    "terminate fails on Elastic error",
// 			request: getTestTerminateRequest(nil),
// 			claims:  validClaims,
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptTerminate,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, terminatedJSON),
// 					ResponseErr: errors.New("timeout"),
// 				},
// 			),
// 			expectedCode:     http.StatusInternalServerError,
// 			expectedResponse: response.NewErrorDetail(requestId, "could not update the status of batch test-batch: [500] elasticsearch client error: timeout"),
// 		},
// 		{
// 			name:    "terminate fails on wrong integrator Id",
// 			request: getTestTerminateRequest(nil),
// 			claims:  auth.HriClaims{Scope: auth.HriIntegrator, Subject: "wrong id"},
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptTerminateWrongId,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "noop",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, terminatedJSON),
// 				},
// 			),
// 			expectedNotification: terminatedBatch,
// 			expectedCode:         http.StatusUnauthorized,
// 			expectedResponse:     response.NewErrorDetail(requestId, "terminate requested by 'wrong id' but owned by 'integratorId'"),
// 		},
// 		{
// 			name:    "Terminate fails on terminated batch",
// 			request: getTestTerminateRequest(nil),
// 			claims:  validClaims,
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptTerminate,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "noop",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, terminatedJSON),
// 				},
// 			),
// 			expectedCode:     http.StatusConflict,
// 			expectedResponse: response.NewErrorDetail(requestId, "terminate failed, batch is in 'terminated' state"),
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			esClient, err := elastic.ClientFromTransport(tt.ft)
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			writer := test.FakeWriter{
// 				T:             t,
// 				ExpectedTopic: InputTopicToNotificationTopic(batchTopic),
// 				ExpectedKey:   test.ValidBatchId,
// 				ExpectedValue: tt.expectedNotification,
// 				Error:         tt.writerError,
// 			}

// 			code, result := Terminate(requestId, tt.request, tt.claims, esClient, writer, currentStatus)

// 			if tt.ft != nil {
// 				tt.ft.VerifyCalls()
// 			}
// 			if code != tt.expectedCode {
// 				t.Errorf("Terminate() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
// 			}
// 			if !reflect.DeepEqual(result, tt.expectedResponse) {
// 				t.Errorf("Terminate() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
// 			}
// 		})
// 	}
// }

// func Test_TerminateNoAuth(t *testing.T) {
// 	logwrapper.Initialize("error", os.Stdout)

// 	const (
// 		requestId                       = "requestNoAuth3"
// 		batchName                       = "BatchToTerminate"
// 		batchTopic               string = "bobo.batch.in"
// 		batchExpectedRecordCount        = float64(30)
// 		currentStatus                   = status.SendCompleted
// 	)

// 	terminatedBatch := map[string]interface{}{
// 		param.BatchId:             test.ValidBatchId,
// 		param.Name:                batchName,
// 		param.IntegratorId:        "",
// 		param.Topic:               batchTopic,
// 		param.DataType:            batchDataType,
// 		param.Status:              status.Terminated.String(),
// 		param.StartDate:           batchStartDate,
// 		param.RecordCount:         batchExpectedRecordCount,
// 		param.ExpectedRecordCount: batchExpectedRecordCount,
// 	}
// 	terminatedJSON, err := json.Marshal(terminatedBatch)
// 	if err != nil {
// 		t.Errorf("Unable to create batch JSON string: %s", err.Error())
// 	}

// 	const (
// 		scriptTerminate = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'NoAuthUnkIntegrator'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
// 	)

// 	tests := []struct {
// 		name                 string
// 		request              *model.TerminateRequest
// 		ft                   *test.FakeTransport
// 		writerError          error
// 		expectedNotification map[string]interface{}
// 		expectedCode         int
// 		expectedResponse     interface{}
// 	}{
// 		{
// 			name:    "successful NoAuth Terminate",
// 			request: getTestTerminateRequest(nil),
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptTerminate,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidTenantId, terminatedJSON),
// 				},
// 			),
// 			expectedNotification: terminatedBatch,
// 			expectedCode:         http.StatusOK,
// 			expectedResponse:     nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			esClient, err := elastic.ClientFromTransport(tt.ft)
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			writer := test.FakeWriter{
// 				T:             t,
// 				ExpectedTopic: InputTopicToNotificationTopic(batchTopic),
// 				ExpectedKey:   test.ValidBatchId,
// 				ExpectedValue: tt.expectedNotification,
// 				Error:         tt.writerError,
// 			}

// 			var emptyClaims = auth.HriClaims{}
// 			code, result := TerminateNoAuth(requestId, tt.request, emptyClaims, esClient, writer, currentStatus)

// 			tt.ft.VerifyCalls()
// 			if code != tt.expectedCode {
// 				t.Errorf("TerminateNoAuth() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
// 			}
// 			if !reflect.DeepEqual(result, tt.expectedResponse) {
// 				t.Errorf("TerminateNoAuth() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
// 			}
// 		})
// 	}
// }

func getTestTerminateRequest(metadata map[string]interface{}) *model.TerminateRequest {
	request := model.TerminateRequest{
		TenantId: test.ValidTenantId,
		BatchId:  test.ValidBatchId,
		Metadata: metadata,
	}
	return &request
}

func TestTerminate200(t *testing.T) {
	expectedCode := 200
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	mdata := bson.M{"compression": "gzip", "finalRecordCount": 20}

	writer := test.FakeWriter{
		T:             t,
		ExpectedTopic: "ingest.pentest.claims.notification",
		ExpectedKey:   "batchid1",
		ExpectedValue: map[string]interface{}{"dataType": "rspec-batch", "id": "batchid1", "integratorId": "8b1e7a81-7f4a-41b0-a170-ae19f843f27c", "invalidThreshold": 5, "metadata": mdata, "name": "rspec-pentest-batch", "startDate": "2022-11-29T09:52:07Z", "status": "started", "topic": "ingest.pentest.claims.in"},
		Error:         nil,
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{"topic", "ingest.pentest.claims.in"},
			{"dataType", "rspec-batch"},
			{"invalidThreshold", 5},
			{"metadata", i},
			{"id", "batchid1"},
			{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{"status", "started"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(second, killCursors2)

	})

	code, _ := TerminateBatch(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestTerminateUnauthorizedRole(t *testing.T) {
	expectedCode := 401
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
	}
	claims := auth.HriAzClaims{
		Subject: "ClaimSubjNotEqualIntegratorId",
		Roles:   []string{"hri_data_integrator"},
	}

	writer := test.FakeWriter{}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	code, _ := TerminateBatch(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestTerminateNoAuth200(t *testing.T) {
	expectedCode := 200
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
		Metadata: map[string]interface{}{"compression": "gzip", "finalRecordCount": 20},
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	mdata := bson.M{"compression": "gzip", "finalRecordCount": 20}

	writer := test.FakeWriter{
		T:             t,
		ExpectedTopic: "ingest.pentest.claims.notification",
		ExpectedKey:   "batchid1",
		ExpectedValue: map[string]interface{}{"dataType": "rspec-batch", "id": "batchid1", "integratorId": "8b1e7a81-7f4a-41b0-a170-ae19f843f27c", "invalidThreshold": 5, "metadata": mdata, "name": "rspec-pentest-batch", "startDate": "2022-11-29T09:52:07Z", "status": "started", "topic": "ingest.pentest.claims.in"},
		Error:         nil,
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{"topic", "ingest.pentest.claims.in"},
			{"dataType", "rspec-batch"},
			{"invalidThreshold", 5},
			{"metadata", i},
			{"id", "batchid1"},
			{"integratorId", "NoAuthUnkIntegrator"},
			{"status", "started"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(second, killCursors2)

	})

	code, _ := TerminateBatchNoAuth(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestTerminategetBatchMetaDataError(t *testing.T) {
	expectedCode := 404
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	writer := test.FakeWriter{}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{"topic", "ingest.pentest.claims.in"},
			{"dataType", "rspec-batch"},
			{"invalidThreshold", 5},
			{"metadata", i},
			{"id", "batchid1"},
			{"integratorId", "NoAuthUnkIntegrator"},
			{"status", "started"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch,
			detailsMap,
		)

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := TerminateBatch(requestId, &request, claims, writer)

	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestTerminateStatusNotStarted(t *testing.T) {
	expectedCode := 409
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	writer := test.FakeWriter{}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{"topic", "ingest.pentest.claims.in"},
			{"dataType", "rspec-batch"},
			{"invalidThreshold", 5},
			{"metadata", i},
			{"id", "batchid1"},
			{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{"status", "completed"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := TerminateBatch(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendTerminateClaimSubjNotEqualIntegratorId(t *testing.T) {
	expectedCode := 401
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
	}
	claims := auth.HriAzClaims{
		Subject: "ClaimSubjNotEqualIntegratorId",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	writer := test.FakeWriter{}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{"topic", "ingest.pentest.claims.in"},
			{"dataType", "rspec-batch"},
			{"invalidThreshold", 5},
			{"metadata", i},
			{"id", "batchid1"},
			{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{"status", "started"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := TerminateBatch(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestTerminateupdateBatchStatusErr(t *testing.T) {
	expectedCode := 500
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	writer := test.FakeWriter{}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{"topic", "ingest.pentest.claims.in"},
			{"dataType", "rspec-batch"},
			{"invalidThreshold", 5},
			{"metadata", i},
			{"id", "batchid1"},
			{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{"status", "started"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

		mt.AddMockResponses(bson.D{
			{"ok", 1},
		})

	})

	code, _ := TerminateBatch(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestTerminateKafkaErr(t *testing.T) {
	expectedCode := 500
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	//mdata := bson.M{"compression": "gzip", "finalRecordCount": 20}

	writer := test.FakeWriter{
		T:             t,
		Error:         errors.New("Kafka Error"),
		ExpectedTopic: "ingest.pentest.claims.notification",
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{"topic", "ingest.pentest.claims.in"},
			{"dataType", "rspec-batch"},
			{"invalidThreshold", 5},
			{"metadata", i},
			{"id", "batchid1"},
			{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{"status", "started"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(second, killCursors2)

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"nModified", 1},
		})

	})

	code, _ := TerminateBatch(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestTerminateModifiedCountZero(t *testing.T) {
	expectedCode := 500
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
		Metadata: map[string]interface{}{"compression": "gzip", "finalRecordCount": 20},
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	mdata := bson.M{"compression": "gzip", "finalRecordCount": 20}

	writer := test.FakeWriter{
		T:             t,
		ExpectedTopic: "ingest.pentest.claims.notification",
		ExpectedKey:   "batchid1",
		ExpectedValue: map[string]interface{}{"dataType": "rspec-batch", "id": "batchid1", "integratorId": "8b1e7a81-7f4a-41b0-a170-ae19f843f27c", "invalidThreshold": 5, "metadata": mdata, "name": "rspec-pentest-batch", "startDate": "2022-11-29T09:52:07Z", "status": "started", "topic": "ingest.pentest.claims.in"},
		Error:         nil,
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{"topic", "ingest.pentest.claims.in"},
			{"dataType", "rspec-batch"},
			{"invalidThreshold", 5},
			{"metadata", i},
			{"id", "batchid1"},
			{"integratorId", "NoAuthUnkIntegrator"},
			{"status", "started"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			//{"nModified", 1},
		})

		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(second, killCursors2)

	})

	code, _ := TerminateBatchNoAuth(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestTerminateExtractBatchStatus(t *testing.T) {
	expectedCode := 500
	request := model.TerminateRequest{
		TenantId: "tid1",
		BatchId:  "batchid1",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	writer := test.FakeWriter{}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{"topic", "ingest.pentest.claims.in"},
			{"dataType", "rspec-batch"},
			{"invalidThreshold", 5},
			{"metadata", i},
			{"id", "batchid1"},
			{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{"status", "complete"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := TerminateBatch(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
