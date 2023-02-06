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

func intPtr(val int) *int {
	return &val
}
func TestProcessComplete401(t *testing.T) {
	expectedCode := 401
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}

	claims := auth.HriAzClaims{}
	writer := test.FakeWriter{
		T:             t,
		ExpectedTopic: InputTopicToNotificationTopic(batchTopic),
		ExpectedKey:   test.ValidBatchId,
		Error:         nil,
	}

	code, _ := ProcessingCompleteBatch(requestId, &processingCompleteRequest, claims, writer, status.SendCompleted)

	if code != expectedCode {
		t.Errorf("SendProcessingComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestProcessingComplete200(t *testing.T) {
	expectedCode := 200

	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
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
			{Key: "integratorId", Value: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{Key: "status", Value: "sendCompleted"},
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

	code, _ := ProcessingCompleteBatch(requestId, &processingCompleteRequest, claims, writer, status.SendCompleted)
	if code != expectedCode {
		t.Errorf("ProcessingCompleteBatch() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
func TestProcessingCompleteBatchStatusError(t *testing.T) {
	expectedCode := 500

	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
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

	code, _ := ProcessingCompleteBatch(requestId, &processingCompleteRequest, claims, writer, status.SendCompleted)
	if code != expectedCode {
		t.Errorf("ProcessingCompleteBatch() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
func TestProcessingCompleteRequestNoAuth200(t *testing.T) {
	expectedCode := 200
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
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
			{Key: "status", Value: "sendCompleted"},
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

	code, _ := ProcessingCompleteBatchNoAuth(requestId, &processingCompleteRequest, claims, writer, status.SendCompleted)
	if code != expectedCode {
		t.Errorf("ProcessingCompleteBatchNoAuth() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestProcessingCompleteRequestStatusTerminateOrFail(t *testing.T) {
	statusConflict := 409
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
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

	code, _ := ProcessingCompleteBatch(requestId, &processingCompleteRequest, claims, writer, status.Terminated)
	if code != statusConflict {
		t.Errorf("SendProcessingComplete() = \n\t%v,\nexpected: \n\t%v", code, statusConflict)
	}
}

func TestProcessingCompleteRequestFail(t *testing.T) {
	expectedCode := 409
	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
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

	code, _ := ProcessingCompleteBatch(requestId, &processingCompleteRequest, claims, writer, status.Failed)
	if code != expectedCode {
		t.Errorf("SendProcessingComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestProcessingCompleteupdateBatchStatusErr(t *testing.T) {
	expectedCode := 500

	arc := 12
	irc := 23
	processingCompleteRequest := model.ProcessingCompleteRequest{
		TenantId:           "tid1",
		BatchId:            "batchid1",
		ActualRecordCount:  &arc,
		InvalidRecordCount: &irc,
	}

	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_internal", "hri_tenant_tid1_data_internal"},
		Scope:   "hri_internal",
	}

	writer := test.FakeWriter{}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
		})

	})

	code, _ := ProcessingCompleteBatch(requestId, &processingCompleteRequest, claims, writer, status.SendCompleted)
	if code != expectedCode {
		t.Errorf("ProcessingCompleteBatch() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
