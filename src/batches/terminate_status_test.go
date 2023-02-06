package batches

import (
	"errors"
	"testing"

	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

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

	code, _ := TerminateBatch(requestId, &request, claims, writer, status.Started, integratorIdtest)

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

	code, _ := TerminateBatch(requestId, &request, claims, writer, status.Started, integratorIdtest)
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

	code, _ := TerminateBatchNoAuth(requestId, &request, claims, writer, status.Started, "NoAuthUnkIntegrator")

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

	code, _ := TerminateBatch(requestId, &request, claims, writer, status.Completed, integratorIdtest)
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

	code, _ := TerminateBatch(requestId, &request, claims, writer, status.Started, "incorrectintegratorId")
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

		mt.AddMockResponses(bson.D{
			{"ok", 1},
		})

	})

	code, _ := TerminateBatch(requestId, &request, claims, writer, status.Started, integratorIdtest)
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

	code, _ := TerminateBatch(requestId, &request, claims, writer, status.Started, integratorIdtest)

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

		mt.AddMockResponses(bson.D{
			{"ok", 1},
		})

		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"batch", array1},
		})

		killCursors2 := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(second, killCursors2)

	})

	code, _ := TerminateBatchNoAuth(requestId, &request, claims, writer, status.Started, "NoAuthUnkIntegrator")

	if code != expectedCode {
		t.Errorf("SendComplete() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}
