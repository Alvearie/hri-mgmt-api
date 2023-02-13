/*
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func intPtr(val int) *int {
	return &val
}

// func Test_getSendCompleteUpdateScript(t *testing.T) {
// 	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

// 	tests := []struct {
// 		name    string
// 		request model.SendCompleteRequest
// 		claims  auth.HriClaims
// 		// Note that the following chars must be escaped because expectedScript is used as a regex pattern: ., ), (, [, ]
// 		expectedRequest map[string]interface{}
// 		metadata        bool
// 	}{
// 		{
// 			name: "no metadata with validation",
// 			request: model.SendCompleteRequest{
// 				ExpectedRecordCount: intPtr(200),
// 				Validation:          true,
// 			},
// 			claims: validClaims,
// 			expectedRequest: map[string]interface{}{
// 				"script": map[string]interface{}{
// 					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200;} else {ctx\.op = 'none'}`,
// 				},
// 			},
// 			metadata: false,
// 		},
// 		{
// 			name: "no metadata without validation",
// 			request: model.SendCompleteRequest{
// 				ExpectedRecordCount: intPtr(200),
// 				Validation:          false,
// 			},
// 			claims: validClaims,
// 			expectedRequest: map[string]interface{}{
// 				"script": map[string]interface{}{
// 					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'completed'; ctx\._source\.expectedRecordCount = 200; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}`,
// 				},
// 			},
// 			metadata: false,
// 		},
// 		{
// 			name: "metadata with validation",
// 			request: model.SendCompleteRequest{
// 				ExpectedRecordCount: intPtr(200),
// 				Metadata:            map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5},
// 				Validation:          false,
// 			},
// 			claims: validClaims,
// 			expectedRequest: map[string]interface{}{
// 				"script": map[string]interface{}{
// 					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}`,
// 					"lang":   "painless",
// 					"params": map[string]interface{}{"metadata": map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5}},
// 				},
// 			},
// 			metadata: true,
// 		},
// 		{
// 			name: "metadata without validation",
// 			request: model.SendCompleteRequest{
// 				ExpectedRecordCount: intPtr(200),
// 				Metadata:            map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": 3},
// 				Validation:          false,
// 			},
// 			claims: validClaims,
// 			expectedRequest: map[string]interface{}{
// 				"script": map[string]interface{}{
// 					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'completed'; ctx\._source\.expectedRecordCount = 200; ctx\._source\.endDate = '` + test.DatePattern + `'; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}`,
// 					"lang":   "painless",
// 					"params": map[string]interface{}{"metadata": map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": 3}},
// 				},
// 			},
// 			metadata: true,
// 		},
// 		{
// 			name: "with deprecated recordCount field",
// 			request: model.SendCompleteRequest{
// 				RecordCount: intPtr(200),
// 				Validation:  true,
// 			},
// 			claims: validClaims,
// 			expectedRequest: map[string]interface{}{
// 				"script": map[string]interface{}{
// 					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200;} else {ctx\.op = 'none'}`,
// 				},
// 			},
// 			metadata: false,
// 		},
// 		{
// 			name: "Missing claim.Subject",
// 			request: model.SendCompleteRequest{
// 				ExpectedRecordCount: intPtr(200),
// 				Validation:          true,
// 			},
// 			claims: auth.HriClaims{},
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
// 			updateRequest := getSendCompleteUpdateScript(&tt.request, tt.claims.Subject)

// 			if err := RequestCompareScriptTest(tt.expectedRequest, updateRequest); err != nil {
// 				t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", updateRequest, tt.expectedRequest, err)
// 			} else if tt.metadata {
// 				if err := RequestCompareWithMetadataTest(tt.expectedRequest, updateRequest); err != nil {
// 					t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", updateRequest, tt.expectedRequest, err)
// 				}
// 			}
// 		})
// 	}
// }

func TestSendComplete401(t *testing.T) {
	expectedCode := 401
	e := 12
	r := 23
	request := model.SendCompleteRequest{
		TenantId:            "tid1",
		BatchId:             "batchid1",
		ExpectedRecordCount: &e,
		RecordCount:         &r,
	}
	claims := auth.HriAzClaims{}
	writer := test.FakeWriter{
		T:             t,
		ExpectedTopic: InputTopicToNotificationTopic(batchTopic),
		ExpectedKey:   test.ValidBatchId,
		//ExpectedValue: map[string]interface{}{},
		Error: nil,
	}

	code, _ := SendStatusComplete(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestSendComplete200(t *testing.T) {
	expectedCode := 200
	e := 12
	r := 23
	request := model.SendCompleteRequest{
		TenantId:            "tid1",
		BatchId:             "batchid1",
		ExpectedRecordCount: &e,
		RecordCount:         &r,
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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendCompleteNoAuth200(t *testing.T) {
	expectedCode := 200
	e := 12
	r := 23
	request := model.SendCompleteRequest{
		TenantId:            "tid1",
		BatchId:             "batchid1",
		ExpectedRecordCount: &e,
		RecordCount:         &r,
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}

	//m := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

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

	code, _ := SendStatusCompleteNoAuth(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestSendCompletegetBatchMetaDataError(t *testing.T) {
	expectedCode := 404
	e := 12
	r := 23
	request := model.SendCompleteRequest{
		TenantId:            "tid1",
		BatchId:             "batchid1",
		ExpectedRecordCount: &e,
		RecordCount:         &r,
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

	code, _ := SendStatusCompleteNoAuth(requestId, &request, claims, writer)

	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestSendCompleteupdateBatchStatusErr(t *testing.T) {
	expectedCode := 500
	e := 12
	r := 23
	request := model.SendCompleteRequest{
		TenantId:            "tid1",
		BatchId:             "batchid1",
		ExpectedRecordCount: &e,
		RecordCount:         &r,
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

	code, _ := SendStatusCompleteNoAuth(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestSendCompletClaimSubjNotEqualIntegratorId(t *testing.T) {
	expectedCode := 401
	e := 12
	r := 23
	request := model.SendCompleteRequest{
		TenantId:            "tid1",
		BatchId:             "batchid1",
		ExpectedRecordCount: &e,
		RecordCount:         &r,
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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendCompletStatusNotStarted(t *testing.T) {
	expectedCode := 409
	e := 12
	r := 23
	request := model.SendCompleteRequest{
		TenantId:            "tid1",
		BatchId:             "batchid1",
		ExpectedRecordCount: &e,
		RecordCount:         &r,
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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendCompleteExpectedRecordCountNil(t *testing.T) {
	expectedCode := 200
	r := 23
	request := model.SendCompleteRequest{
		TenantId:    "tid1",
		BatchId:     "batchid1",
		RecordCount: &r,
		Validation:  true,
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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendCompleteValidationTrueMatadataNotNil(t *testing.T) {
	expectedCode := 200
	r := 23
	request := model.SendCompleteRequest{
		TenantId:    "tid1",
		BatchId:     "batchid1",
		RecordCount: &r,
		Validation:  true,
		Metadata:    map[string]interface{}{"compression": "gzip", "finalRecordCount": 20},
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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendCompleteValidationFalseMatadataNotNil(t *testing.T) {
	expectedCode := 200
	r := 23
	request := model.SendCompleteRequest{
		TenantId:    "tid1",
		BatchId:     "batchid1",
		RecordCount: &r,
		Metadata:    map[string]interface{}{"compression": "gzip", "finalRecordCount": 20},
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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendCompletExtractBatchStatus(t *testing.T) {
	expectedCode := 500
	e := 12
	r := 23
	request := model.SendCompleteRequest{
		TenantId:            "tid1",
		BatchId:             "batchid1",
		ExpectedRecordCount: &e,
		RecordCount:         &r,
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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

// func TestSendComplete(t *testing.T) {
// 	const (
// 		requestId = "requestId"
// 	)

// 	const (
// 		scriptSendComplete             = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 10;} else {ctx\.op = 'none'}"}}` + "\n"
// 		scriptSendCompleteWithMetadata = `{"script":{"lang":"painless","params":{"metadata":{"compression":"gzip","userMetaField1":"metadata","userMetaField2":-5}},"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 10; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}"}}` + "\n"
// 		scriptSendCompleteWrongId      = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'wrong id'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 10;} else {ctx\.op = 'none'}"}}` + "\n"
// 	)

// 	logwrapper.Initialize("error", os.Stdout)
// 	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

// 	sendCompletedBatch := map[string]interface{}{
// 		param.BatchId:             test.ValidBatchId,
// 		param.Name:                batchName,
// 		param.IntegratorId:        integratorId,
// 		param.Topic:               batchTopic,
// 		param.DataType:            batchDataType,
// 		param.Status:              status.SendCompleted.String(),
// 		param.StartDate:           batchStartDate,
// 		param.RecordCount:         batchExpectedRecordCount,
// 		param.ExpectedRecordCount: batchExpectedRecordCount,
// 		param.InvalidThreshold:    batchInvalidThreshold,
// 	}
// 	sendCompletedJSON, err := json.Marshal(sendCompletedBatch)
// 	if err != nil {
// 		t.Errorf("Unable to create batch JSON string: %s", err.Error())
// 	}

// 	sendCompletedBatchWithMetadata := map[string]interface{}{
// 		param.BatchId:             test.ValidBatchId,
// 		param.Name:                batchName,
// 		param.IntegratorId:        integratorId,
// 		param.Topic:               batchTopic,
// 		param.DataType:            batchDataType,
// 		param.Status:              status.SendCompleted.String(),
// 		param.StartDate:           batchStartDate,
// 		param.RecordCount:         batchExpectedRecordCount,
// 		param.ExpectedRecordCount: batchExpectedRecordCount,
// 		param.InvalidThreshold:    batchInvalidThreshold,
// 	}
// 	sendCompletedBatchWithMetadataJSON, err := json.Marshal(sendCompletedBatchWithMetadata)
// 	if err != nil {
// 		t.Errorf("Unable to create batch JSON string: %s", err.Error())
// 	}

// 	terminatedBatch := map[string]interface{}{
// 		param.BatchId:             test.ValidBatchId,
// 		param.Name:                batchName,
// 		param.IntegratorId:        integratorId,
// 		param.Topic:               batchTopic,
// 		param.DataType:            batchDataType,
// 		param.Status:              status.Terminated.String(),
// 		param.StartDate:           batchStartDate,
// 		param.ExpectedRecordCount: batchExpectedRecordCount,
// 		param.InvalidThreshold:    batchInvalidThreshold,
// 		param.Metadata:            map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5},
// 	}
// 	terminatedJSON, err := json.Marshal(terminatedBatch)
// 	if err != nil {
// 		t.Errorf("Unable to create batch JSON string: %s", err.Error())
// 	}

// 	tests := []struct {
// 		name                 string
// 		request              *model.SendCompleteRequest
// 		claims               auth.HriClaims
// 		ft                   *test.FakeTransport
// 		writerError          error
// 		expectedNotification map[string]interface{}
// 		expectedCode         int
// 		expectedResponse     interface{}
// 		currentStatus        status.BatchStatus
// 	}{
// 		{
// 			name:    "successful sendComplete, with validation",
// 			request: getTestSendCompleteRequest(intPtr(int(batchExpectedRecordCount)), nil, nil, true),
// 			claims:  validClaims,
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptSendComplete,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, sendCompletedJSON),
// 				},
// 			),
// 			expectedNotification: sendCompletedBatch,
// 			expectedCode:         http.StatusOK,
// 			expectedResponse:     nil,
// 		},
// 		{
// 			name: "successful sendComplete, with validation and metadata",
// 			request: getTestSendCompleteRequest(
// 				intPtr(int(batchExpectedRecordCount)),
// 				nil,
// 				map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5},
// 				true,
// 			),
// 			claims: validClaims,
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptSendCompleteWithMetadata,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, sendCompletedBatchWithMetadataJSON),
// 				},
// 			),
// 			expectedNotification: sendCompletedBatchWithMetadata,
// 			expectedCode:         http.StatusOK,
// 			expectedResponse:     nil,
// 		},
// 		{
// 			name:             "401 Unauthorized when Data Integrator scope is missing",
// 			request:          getTestSendCompleteRequest(intPtr(int(batchExpectedRecordCount)), nil, nil, true),
// 			claims:           auth.HriClaims{Scope: auth.HriConsumer, Subject: integratorId},
// 			expectedCode:     http.StatusUnauthorized,
// 			expectedResponse: response.NewErrorDetail(requestId, `Must have hri_data_integrator role to initiate sendComplete on a batch`),
// 		},
// 		{
// 			name:    "sendComplete fails on Elastic error",
// 			request: getTestSendCompleteRequest(intPtr(int(batchExpectedRecordCount)), nil, nil, true),
// 			claims:  validClaims,
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptSendComplete,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, sendCompletedJSON),
// 					ResponseErr: errors.New("timeout"),
// 				},
// 			),
// 			expectedCode:     http.StatusInternalServerError,
// 			expectedResponse: response.NewErrorDetail(requestId, "could not update the status of batch test-batch: [500] elasticsearch client error: timeout"),
// 		},
// 		{
// 			name:    "sendComplete fails on terminated batch",
// 			request: getTestSendCompleteRequest(intPtr(int(batchExpectedRecordCount)), nil, nil, true),
// 			claims:  validClaims,
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptSendComplete,
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
// 			expectedResponse: response.NewErrorDetail(requestId, "sendComplete failed, batch is in 'terminated' state"),
// 		},
// 		{
// 			name:    "sendComplete fails on wrong integrator Id",
// 			request: getTestSendCompleteRequest(intPtr(int(batchExpectedRecordCount)), nil, nil, true),
// 			claims:  auth.HriClaims{Scope: auth.HriIntegrator, Subject: "wrong id"},
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptSendCompleteWrongId,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "noop",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, sendCompletedJSON),
// 				},
// 			),
// 			expectedCode:     http.StatusUnauthorized,
// 			expectedResponse: response.NewErrorDetail(requestId, "sendComplete requested by 'wrong id' but owned by 'integratorId'"),
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

// 			code, result := SendComplete(requestId, tt.request, tt.claims, esClient, writer, tt.currentStatus)

// 			if tt.ft != nil {
// 				tt.ft.VerifyCalls()
// 			}

// 			if code != tt.expectedCode {
// 				t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
// 			}
// 			if !reflect.DeepEqual(result, tt.expectedResponse) {
// 				t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
// 			}
// 		})
// 	}
// }

// func TestSendCompleteNoAuth(t *testing.T) {
// 	logwrapper.Initialize("error", os.Stdout)

// 	const (
// 		requestId                string = "req99X5"
// 		batchName                string = "porcipino"
// 		batchDataType            string = "claims"
// 		batchStartDate           string = "IgnoredNoDate"
// 		integratorId                    = auth.NoAuthFakeIntegrator
// 		batchExpectedRecordCount        = float64(14)
// 		batchInvalidThreshold           = float64(5)
// 		currentStatus                   = status.Started
// 	)

// 	const (
// 		scriptSendComplete             = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'NoAuthUnkIntegrator'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 14;} else {ctx\.op = 'none'}"}}` + "\n"
// 		scriptSendCompleteWithMetadata = `{"script":{"lang":"painless","params":{"metadata":{"compression":"gzip","userMetaField1":"metadataUno","userMetaField2":-30}},"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'NoAuthUnkIntegrator'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 14; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}"}}` + "\n"
// 	)

// 	sendCompletedBatch := map[string]interface{}{
// 		param.BatchId:             test.ValidBatchId,
// 		param.Name:                batchName,
// 		param.Topic:               batchTopic,
// 		param.DataType:            batchDataType,
// 		param.IntegratorId:        integratorId,
// 		param.Status:              status.SendCompleted.String(),
// 		param.StartDate:           batchStartDate,
// 		param.RecordCount:         batchExpectedRecordCount,
// 		param.ExpectedRecordCount: batchExpectedRecordCount,
// 		param.InvalidThreshold:    batchInvalidThreshold,
// 	}
// 	sendCompletedJSON, err := json.Marshal(sendCompletedBatch)
// 	if err != nil {
// 		t.Errorf("Unable to create batch JSON string: %s", err.Error())
// 	}

// 	sendCompletedBatchWithMetadata := map[string]interface{}{
// 		param.BatchId:             test.ValidBatchId,
// 		param.Name:                batchName,
// 		param.Topic:               batchTopic,
// 		param.DataType:            batchDataType,
// 		param.IntegratorId:        integratorId,
// 		param.Status:              status.SendCompleted.String(),
// 		param.StartDate:           batchStartDate,
// 		param.RecordCount:         batchExpectedRecordCount,
// 		param.ExpectedRecordCount: batchExpectedRecordCount,
// 		param.InvalidThreshold:    batchInvalidThreshold,
// 	}
// 	sendCompletedBatchWithMetadataJSON, err := json.Marshal(sendCompletedBatchWithMetadata)
// 	if err != nil {
// 		t.Errorf("Unable to create batch JSON string: %s", err.Error())
// 	}

// 	tests := []struct {
// 		name                 string
// 		request              *model.SendCompleteRequest
// 		claims               auth.HriClaims
// 		ft                   *test.FakeTransport
// 		writerError          error
// 		expectedNotification map[string]interface{}
// 		expectedCode         int
// 		expectedResponse     interface{}
// 	}{
// 		{
// 			name:    "successful sendCompleteNoAuth, with validation",
// 			request: getTestSendCompleteRequest(intPtr(int(batchExpectedRecordCount)), nil, nil, true),
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptSendComplete,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, sendCompletedJSON),
// 				},
// 			),
// 			expectedNotification: sendCompletedBatch,
// 			expectedCode:         http.StatusOK,
// 			expectedResponse:     nil,
// 		},
// 		{
// 			name: "successful sendCompleteNoAuth, with validation and metadata",
// 			request: getTestSendCompleteRequest(
// 				intPtr(int(batchExpectedRecordCount)),
// 				nil,
// 				map[string]interface{}{"compression": "gzip", "userMetaField1": "metadataUno", "userMetaField2": -30},
// 				true,
// 			),
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  scriptSendCompleteWithMetadata,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, sendCompletedBatchWithMetadataJSON),
// 				},
// 			),
// 			expectedNotification: sendCompletedBatchWithMetadata,
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
// 			code, result := SendCompleteNoAuth(requestId, tt.request, emptyClaims, esClient, writer, currentStatus)

// 			tt.ft.VerifyCalls()

// 			if code != tt.expectedCode {
// 				t.Errorf("SendCompleteNoAuth() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
// 			}
// 			if !reflect.DeepEqual(result, tt.expectedResponse) {
// 				t.Errorf("SendCompleteNoAuth() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
// 			}
// 		})
// 	}
// }

func getTestSendCompleteRequest(expectedRecCount *int, recCount *int, metadata map[string]interface{}, validation bool) *model.SendCompleteRequest {
	request := model.SendCompleteRequest{
		TenantId:            test.ValidTenantId,
		BatchId:             test.ValidBatchId,
		ExpectedRecordCount: expectedRecCount,
		RecordCount:         recCount,
		Metadata:            metadata,
		Validation:          validation,
	}
	return &request
}
