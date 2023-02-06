package batches

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestFail401(t *testing.T) {
	expectedCode := 401
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "test_tenant",
		BatchId:            "batch_id",
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
		Error:         nil,
	}

	code, _ := SendFail(requestId, &request, claims, writer, status.Started)

	if code != expectedCode {
		t.Errorf("SendFail() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestFail200(t *testing.T) {
	expectedCode := 200

	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "test_tenant",
		BatchId:            "batch_id",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}
	request := model.FailRequest{
		ProcessingCompleteRequest: processingCompleteRequest,
		FailureMessage:            "Failed Message",
	}
	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_internal", "hri_tenant_test_tenant_data_internal"},
		Scope:   "hri_internal",
	}

	mdata := bson.M{"compression": "gzip", "finalRecordCount": 20}

	writer := test.FakeWriter{
		T:             t,
		ExpectedTopic: "ingest.pentest.claims.notification",
		ExpectedKey:   "batch_id",
		ExpectedValue: map[string]interface{}{"dataType": "rspec-batch", "id": "batch_id", "integratorId": "8b1e7a81-7f4a-41b0-a170-ae19f843f27c", "invalidThreshold": 5, "metadata": mdata, "name": "rspec-pentest-batch", "startDate": "2022-11-29T09:52:07Z", "status": "started", "topic": "ingest.pentest.claims.in"},
		Error:         nil,
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("TestFail200", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		i := map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

		detailsMap := bson.D{
			{Key: "name", Value: "rspec-pentest-batch"},
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batch_id"},
			{Key: "integratorId", Value: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{Key: "status", Value: "started"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		detailsMapArray := []bson.D{detailsMap}

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
		})

		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: detailsMapArray},
		})

		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(second, killCursors2)

	})

	code, _ := SendFail(requestId, &request, claims, writer, status.Started)

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

	})

	code, _ := SendFail(requestId, &request, claims, writer, status.Started)

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
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batchid1"},
			{Key: "integratorId", Value: "NoAuthUnkIntegrator"},
			{Key: "status", Value: "started"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
		})

		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(second, killCursors2)

	})

	code, _ := SendFailNoAuth(requestId, &request, claims, writer, status.Started)

	if code != expectedCode {
		t.Errorf("SendFailNoAuth() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

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

	})

	code, _ := SendFail(requestId, &request, claims, writer, status.Started)
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

	})

	code, _ := SendFail(requestId, &request, claims, writer, status.Terminated)
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

	})

	code, _ := SendFail(requestId, &request, claims, writer, status.Failed)
	if code != expectedCode {
		t.Errorf("SendFail() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
