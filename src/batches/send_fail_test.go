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

func intPtr1(val int) *int {
	return &val
}

func TestFail401(t *testing.T) {
	expectedCode := 401
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}
	request := model.FailRequest{
		ProcessingCompleteRequest: processingCompleteRequest,
		FailureMessage:            "Failed Message",
	}
	claims := auth.HriAzClaims{}
	writer := test.FakeWriter{
		T:             t,
		ExpectedTopic: InputTopicToNotificationTopic(batchTopic),
		ExpectedKey:   test.ValidBatchId,
		//ExpectedValue: map[string]interface{}{},
		Error: nil,
	}

	code, _ := SendFail(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendFail() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestFail200(t *testing.T) {
	expectedCode := 200

	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}
	request := model.FailRequest{
		ProcessingCompleteRequest: processingCompleteRequest,
		FailureMessage:            "Failed Message",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_internal", "hri_tenant_tid1_data_internal"},
		Scope:   "hri_internal",
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

	code, _ := SendFail(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendFail() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestFailBatchStatusError(t *testing.T) {
	expectedCode := 500

	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}
	request := model.FailRequest{
		ProcessingCompleteRequest: processingCompleteRequest,
		FailureMessage:            "Failed Message",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_internal", "hri_tenant_tid1_data_internal"},
		Scope:   "hri_internal",
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
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

		// mt.AddMockResponses(bson.D{
		// 	{"ok", 1},
		// 	{"nModified", 1},
		// })

		// second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
		// 	{"batch", array1},
		// })

		// killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		// mt.AddMockResponses(second, killCursors2)

	})

	code, _ := SendFail(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendFail() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
func TestFailNoAuth200(t *testing.T) {
	expectedCode := 200
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}
	request := model.FailRequest{
		ProcessingCompleteRequest: processingCompleteRequest,
		FailureMessage:            "Failed Message",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_internal", "hri_tenant_tid1_data_internal"},
		Scope:   "hri_internal",
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

	code, _ := SendFailNoAuth(requestId, &request, claims, writer)
	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
	if code != expectedCode {
		t.Errorf("SendFailNoAuth() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

// func TestSendCompletegetBatchMetaDataError(t *testing.T) {
// 	expectedCode := 404
// 	e := 12
// 	r := 23
// 	request := model.SendCompleteRequest{
// 		TenantId:            "tid1",
// 		BatchId:             "batchid1",
// 		ExpectedRecordCount: &e,
// 		RecordCount:         &r,
// 	}
// 	claims := auth.HriAzClaims{
// 		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
// 		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
// 	}

// 	writer := test.FakeWriter{}

// 	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
// 	defer mt.Close()

// 	mt.Run("success", func(mt *mtest.T) {
// 		mongoApi.HriCollection = mt.Coll

// 		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

// 		detailsMap := bson.D{
// 			{Key: "name", Value: "rspec-pentest-batch"},
// 			{"topic", "ingest.pentest.claims.in"},
// 			{"dataType", "rspec-batch"},
// 			{"invalidThreshold", 5},
// 			{"metadata", i},
// 			{"id", "batchid1"},
// 			{"integratorId", "NoAuthUnkIntegrator"},
// 			{"status", "started"},
// 			{"startDate", "2022-11-29T09:52:07Z"},
// 		}

// 		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch,
// 			detailsMap,
// 		)

// 		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
// 		mt.AddMockResponses(first, killCursors)

// 	})

// 	code, _ := SendStatusCompleteNoAuth(requestId, &request, claims, writer)

// 	if code != expectedCode {
// 		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
// 	}

// }

// func TestSendCompleteupdateBatchStatusErr(t *testing.T) {
// 	expectedCode := 500
// 	e := 12
// 	r := 23
// 	request := model.SendCompleteRequest{
// 		TenantId:            "tid1",
// 		BatchId:             "batchid1",
// 		ExpectedRecordCount: &e,
// 		RecordCount:         &r,
// 	}
// 	claims := auth.HriAzClaims{
// 		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
// 		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
// 	}

// 	writer := test.FakeWriter{}
// 	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
// 	defer mt.Close()

// 	mt.Run("success", func(mt *mtest.T) {
// 		mongoApi.HriCollection = mt.Coll

// 		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

// 		detailsMap := bson.D{
// 			{Key: "name", Value: "rspec-pentest-batch"},
// 			{"topic", "ingest.pentest.claims.in"},
// 			{"dataType", "rspec-batch"},
// 			{"invalidThreshold", 5},
// 			{"metadata", i},
// 			{"id", "batchid1"},
// 			{"integratorId", "NoAuthUnkIntegrator"},
// 			{"status", "started"},
// 			{"startDate", "2022-11-29T09:52:07Z"},
// 		}

// 		array1 := []bson.D{detailsMap}

// 		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
// 			{"batch", array1},
// 		})

// 		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
// 		mt.AddMockResponses(first, killCursors)

// 		mt.AddMockResponses(bson.D{
// 			{"ok", 1},
// 		})

// 	})

// 	code, _ := SendStatusCompleteNoAuth(requestId, &request, claims, writer)
// 	if code != expectedCode {
// 		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
// 	}

// }

func TestSendFailClaimSubjNotEqualIntegratorId(t *testing.T) {
	expectedCode := 500
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}
	request := model.FailRequest{
		ProcessingCompleteRequest: processingCompleteRequest,
		FailureMessage:            "Failed Message",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-fgt-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_internal", "hri_tenant_tid1_data_internal"},
		Scope:   "hri_internal",
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

	code, _ := SendFail(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendFail() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
func TestSendFailStatusTerminateOrFail1(t *testing.T) {
	expectedCode := 409
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}
	request := model.FailRequest{
		ProcessingCompleteRequest: processingCompleteRequest,
		FailureMessage:            "Failed Message",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-fgt-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_internal", "hri_tenant_tid1_data_internal"},
		Scope:   "hri_internal",
	}

	writer := test.FakeWriter{}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		// i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		// detailsMap := bson.D{
		// 	{Key: "name", Value: "rspec-pentest-batch"},
		// 	{"topic", "ingest.pentest.claims.in"},
		// 	{"dataType", "rspec-batch"},
		// 	{"invalidThreshold", 5},
		// 	{"metadata", i},
		// 	{"id", "batchid1"},
		// 	{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
		// 	{"status", "failed"},
		// 	{"startDate", "2022-11-29T09:52:07Z"},
		// }

		// array1 := []bson.D{detailsMap}

		// first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
		// 	{"batch", array1},
		// })

		// killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		// mt.AddMockResponses(first, killCursors)

	})

	code, _ := SendFail(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendFail() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
func TestSendFailStatusTerminateOrFail(t *testing.T) {
	expectedCode := 409
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}
	request := model.FailRequest{
		ProcessingCompleteRequest: processingCompleteRequest,
		FailureMessage:            "Failed Message",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-fgt-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_internal", "hri_tenant_tid1_data_internal"},
		Scope:   "hri_internal",
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
			{"status", "failed"},
			{"startDate", "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := SendFail(requestId, &request, claims, writer)
	if code != expectedCode {
		t.Errorf("SendFail() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

// func TestSendFailExpectedRecordCountNil(t *testing.T) {
// 	expectedCode := 200
// 	// r := 23
// 	// request := model.SendCompleteRequest{
// 	// 	TenantId:    "tid1",
// 	// 	BatchId:     "batchid1",
// 	// 	RecordCount: &r,
// 	// 	Validation:  true,
// 	// }
// 	// claims := auth.HriAzClaims{
// 	// 	Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
// 	// 	Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
// 	// }
// 	arc := 12
// 	// irc := 23
// 	processingCompleteRequest := model.ProcessingCompleteRequest{
// 		TenantId:          "tid1",
// 		BatchId:           "batchid1",
// 		ActualRecordCount: &arc,
// 	}
// 	request := model.FailRequest{
// 		ProcessingCompleteRequest: processingCompleteRequest,
// 		FailureMessage:            "Failed Message",
// 	}
// 	claims := auth.HriAzClaims{
// 		Subject: "8b1e7a81-fgt-41b0-a170-ae19f843f27c",
// 		Roles:   []string{"hri_data_internal", "hri_tenant_tid1_data_internal"},
// 		Scope:   "hri_internal",
// 	}
// 	mdata := bson.M{"compression": "gzip", "finalRecordCount": 20}

// 	writer := test.FakeWriter{
// 		T:             t,
// 		ExpectedTopic: "ingest.pentest.claims.notification",
// 		ExpectedKey:   "batchid1",
// 		ExpectedValue: map[string]interface{}{"dataType": "rspec-batch", "id": "batchid1", "integratorId": "8b1e7a81-7f4a-41b0-a170-ae19f843f27c", "invalidThreshold": 5, "metadata": mdata, "name": "rspec-pentest-batch", "startDate": "2022-11-29T09:52:07Z", "status": "started", "topic": "ingest.pentest.claims.in"},
// 		Error:         nil,
// 	}
// 	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
// 	defer mt.Close()

// 	mt.Run("success", func(mt *mtest.T) {
// 		mongoApi.HriCollection = mt.Coll

// 		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

// 		detailsMap := bson.D{
// 			{Key: "name", Value: "rspec-pentest-batch"},
// 			{"topic", "ingest.pentest.claims.in"},
// 			{"dataType", "rspec-batch"},
// 			{"invalidThreshold", 5},
// 			{"metadata", i},
// 			{"id", "batchid1"},
// 			{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
// 			{"status", "started"},
// 			{"startDate", "2022-11-29T09:52:07Z"},
// 		}

// 		array1 := []bson.D{detailsMap}

// 		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
// 			{"batch", array1},
// 		})

// 		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
// 		mt.AddMockResponses(first, killCursors)

// 		mt.AddMockResponses(bson.D{
// 			{"ok", 1},
// 			{"nModified", 1},
// 		})

// 		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
// 			{"batch", array1},
// 		})

// 		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
// 		mt.AddMockResponses(second, killCursors2)

// 	})

// 	code, _ := SendFail(requestId, &request, claims, writer)
// 	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
// 	if code != expectedCode {
// 		t.Errorf("SendFail() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
// 	}
// }

// func TestSendCompleteValidationTrueMatadataNotNil(t *testing.T) {
// 	expectedCode := 200
// 	r := 23
// 	request := model.SendCompleteRequest{
// 		TenantId:    "tid1",
// 		BatchId:     "batchid1",
// 		RecordCount: &r,
// 		Validation:  true,
// 		Metadata:    map[string]interface{}{"compression": "gzip", "finalRecordCount": 20},
// 	}
// 	claims := auth.HriAzClaims{
// 		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
// 		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
// 	}

// 	mdata := bson.M{"compression": "gzip", "finalRecordCount": 20}

// 	writer := test.FakeWriter{
// 		T:             t,
// 		ExpectedTopic: "ingest.pentest.claims.notification",
// 		ExpectedKey:   "batchid1",
// 		ExpectedValue: map[string]interface{}{"dataType": "rspec-batch", "id": "batchid1", "integratorId": "8b1e7a81-7f4a-41b0-a170-ae19f843f27c", "invalidThreshold": 5, "metadata": mdata, "name": "rspec-pentest-batch", "startDate": "2022-11-29T09:52:07Z", "status": "started", "topic": "ingest.pentest.claims.in"},
// 		Error:         nil,
// 	}
// 	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
// 	defer mt.Close()

// 	mt.Run("success", func(mt *mtest.T) {
// 		mongoApi.HriCollection = mt.Coll

// 		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

// 		detailsMap := bson.D{
// 			{Key: "name", Value: "rspec-pentest-batch"},
// 			{"topic", "ingest.pentest.claims.in"},
// 			{"dataType", "rspec-batch"},
// 			{"invalidThreshold", 5},
// 			{"metadata", i},
// 			{"id", "batchid1"},
// 			{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
// 			{"status", "started"},
// 			{"startDate", "2022-11-29T09:52:07Z"},
// 		}

// 		array1 := []bson.D{detailsMap}

// 		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
// 			{"batch", array1},
// 		})

// 		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
// 		mt.AddMockResponses(first, killCursors)

// 		mt.AddMockResponses(bson.D{
// 			{"ok", 1},
// 			{"nModified", 1},
// 		})

// 		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
// 			{"batch", array1},
// 		})

// 		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
// 		mt.AddMockResponses(second, killCursors2)

// 	})

// 	code, _ := SendStatusComplete(requestId, &request, claims, writer)
// 	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
// 	if code != expectedCode {
// 		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
// 	}
// }

// func TestSendCompleteValidationFalseMatadataNotNil(t *testing.T) {
// 	expectedCode := 200
// 	r := 23
// 	request := model.SendCompleteRequest{
// 		TenantId:    "tid1",
// 		BatchId:     "batchid1",
// 		RecordCount: &r,
// 		Metadata:    map[string]interface{}{"compression": "gzip", "finalRecordCount": 20},
// 	}
// 	claims := auth.HriAzClaims{
// 		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
// 		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
// 	}

// 	mdata := bson.M{"compression": "gzip", "finalRecordCount": 20}

// 	writer := test.FakeWriter{
// 		T:             t,
// 		ExpectedTopic: "ingest.pentest.claims.notification",
// 		ExpectedKey:   "batchid1",
// 		ExpectedValue: map[string]interface{}{"dataType": "rspec-batch", "id": "batchid1", "integratorId": "8b1e7a81-7f4a-41b0-a170-ae19f843f27c", "invalidThreshold": 5, "metadata": mdata, "name": "rspec-pentest-batch", "startDate": "2022-11-29T09:52:07Z", "status": "started", "topic": "ingest.pentest.claims.in"},
// 		Error:         nil,
// 	}
// 	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
// 	defer mt.Close()

// 	mt.Run("success", func(mt *mtest.T) {
// 		mongoApi.HriCollection = mt.Coll

// 		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

// 		detailsMap := bson.D{
// 			{Key: "name", Value: "rspec-pentest-batch"},
// 			{"topic", "ingest.pentest.claims.in"},
// 			{"dataType", "rspec-batch"},
// 			{"invalidThreshold", 5},
// 			{"metadata", i},
// 			{"id", "batchid1"},
// 			{"integratorId", "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
// 			{"status", "started"},
// 			{"startDate", "2022-11-29T09:52:07Z"},
// 		}

// 		array1 := []bson.D{detailsMap}

// 		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
// 			{"batch", array1},
// 		})

// 		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
// 		mt.AddMockResponses(first, killCursors)

// 		mt.AddMockResponses(bson.D{
// 			{"ok", 1},
// 			{"nModified", 1},
// 		})

// 		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
// 			{"batch", array1},
// 		})

// 		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
// 		mt.AddMockResponses(second, killCursors2)

// 	})

// 	code, _ := SendStatusComplete(requestId, &request, claims, writer)
// 	//msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
// 	if code != expectedCode {
// 		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
// 	}
// }

// func getTestSendCompleteRequest(expectedRecCount *int, recCount *int, metadata map[string]interface{}, validation bool) *model.SendCompleteRequest {
// 	request := model.SendCompleteRequest{
// 		TenantId:            test.ValidTenantId,
// 		BatchId:             test.ValidBatchId,
// 		ExpectedRecordCount: expectedRecCount,
// 		RecordCount:         recCount,
// 		Metadata:            metadata,
// 		Validation:          validation,
// 	}
// 	return &request
// }
