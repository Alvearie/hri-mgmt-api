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

const (
	requestId                string = "the-request-id"
	batchName                string = "batchName"
	integratorId             string = "integratorId"
	batchTopic               string = "test.batch.in"
	batchDataType            string = "batchDataType"
	batchStartDate           string = "ignored"
	batchExpectedRecordCount        = float64(10)
	batchActualRecordCount          = float64(10)
	batchInvalidThreshold           = float64(5)
	batchInvalidRecordCount         = float64(2)
	batchFailureMessage      string = "Batch Failed"
	transportQueryParams     string = "_source=true"
	currentStatus                   = status.Started
)

var eCount = 12
var rCount = 23

var sendCompleteRequest = model.SendCompleteRequest{
	TenantId:            "tid1",
	BatchId:             "batchid1",
	ExpectedRecordCount: &eCount,
	RecordCount:         &rCount,
}

func TestSendComplete401(t *testing.T) {
	expectedCode := 401

	claims := auth.HriAzClaims{}
	writer := test.FakeWriter{
		T:             t,
		ExpectedTopic: InputTopicToNotificationTopic(batchTopic),
		ExpectedKey:   test.ValidBatchId,
		Error:         nil,
	}

	code, _ := SendStatusComplete(requestId, &sendCompleteRequest, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestSendComplete200(t *testing.T) {
	expectedCode := 200

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
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batchid1"},
			{Key: "integratorId", Value: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{Key: "status", Value: "started"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

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

	code, _ := SendStatusComplete(requestId, &sendCompleteRequest, claims, writer)

	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendCompleteNoAuth200(t *testing.T) {
	expectedCode := 200

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

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

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

	code, _ := SendStatusCompleteNoAuth(requestId, &sendCompleteRequest, claims, writer)

	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestSendCompletegetBatchMetaDataError(t *testing.T) {
	expectedCode := 404

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
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batchid1"},
			{Key: "integratorId", Value: "NoAuthUnkIntegrator"},
			{Key: "status", Value: "started"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch,
			detailsMap,
		)

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := SendStatusCompleteNoAuth(requestId, &sendCompleteRequest, claims, writer)

	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestSendCompleteupdateBatchStatusErr(t *testing.T) {
	expectedCode := 500

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

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
		})

	})

	code, _ := SendStatusCompleteNoAuth(requestId, &sendCompleteRequest, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}

}

func TestSendCompletClaimSubjNotEqualIntegratorId(t *testing.T) {
	expectedCode := 401

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
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batchid1"},
			{Key: "integratorId", Value: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{Key: "status", Value: "started"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := SendStatusComplete(requestId, &sendCompleteRequest, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendCompletStatusNotStarted(t *testing.T) {
	expectedCode := 409

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
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batchid1"},
			{Key: "integratorId", Value: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{Key: "status", Value: "completed"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := SendStatusComplete(requestId, &sendCompleteRequest, claims, writer)
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
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batchid1"},
			{Key: "integratorId", Value: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{Key: "status", Value: "started"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)

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
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batchid1"},
			{Key: "integratorId", Value: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{Key: "status", Value: "started"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)

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
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batchid1"},
			{Key: "integratorId", Value: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{Key: "status", Value: "started"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

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

	code, _ := SendStatusComplete(requestId, &request, claims, writer)

	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestSendCompletExtractBatchStatus(t *testing.T) {
	expectedCode := 500

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
			{Key: "topic", Value: "ingest.pentest.claims.in"},
			{Key: "dataType", Value: "rspec-batch"},
			{Key: "invalidThreshold", Value: 5},
			{Key: "metadata", Value: i},
			{Key: "id", Value: "batchid1"},
			{Key: "integratorId", Value: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"},
			{Key: "status", Value: "complete"},
			{Key: "startDate", Value: "2022-11-29T09:52:07Z"},
		}

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := SendStatusComplete(requestId, &sendCompleteRequest, claims, writer)
	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
