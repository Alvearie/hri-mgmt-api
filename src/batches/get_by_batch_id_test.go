package batches

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var request = model.GetByIdBatch{
	TenantId: "tid1",
	BatchId:  "batchid1",
}

var i = map[string]interface{}{"compression": "gzip", "finalRecordCount": 20}

var detailsMap = bson.D{
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

func TestGetByBatchIdNoAuth200(t *testing.T) {
	expectedCode := 200

	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator", "hri_consumer", "hri_tenant_tid1_data_consumer"},
	}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)
	})

	code, _ := GetByBatchId(requestId, request, claims)
	if code != expectedCode {
		t.Errorf("GetByBatchId() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestGetByBatchIdBadRoles(t *testing.T) {
	expectedCode := 401

	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
	}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)
	})

	code, _ := GetByBatchId(requestId, request, claims)
	if code != expectedCode {
		t.Errorf("GetByBatchId() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func TestGetByBatchIdcheckBatchAuth(t *testing.T) {
	expectedCode := 401

	claims := auth.HriAzClaims{
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_consumer"},
	}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll

		array1 := []bson.D{detailsMap}

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "batch", Value: array1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(first, killCursors)

	})

	code, _ := GetByBatchId(requestId, request, claims)
	if code != expectedCode {
		t.Errorf("GetByBatchId() = \n\t%v,\nexpected: \n\t%v", code, expectedCode)
	}
}

func Test_checkBatchAuth401(t *testing.T) {
	expetedResponse := response.ErrorDetailResponse{
		Code: 401,
	}
	requestId := "rid1"
	claims := auth.HriAzClaims{
		Roles: []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
	}
	resultBody := map[string]interface{}{"integratorId": "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"}
	tenantId := "tid1"

	response := checkBatchAuth(requestId, &claims, resultBody, tenantId)
	if response.Code != expetedResponse.Code {
		t.Errorf("checkBatchAuth() = \n\t%v,\nexpected: \n\t%v", response.Code, expetedResponse.Code)
	}

}

func Test_checkBatchAuthNoIntegratorId(t *testing.T) {
	expetedResponse := response.ErrorDetailResponse{
		Code: 500,
	}
	requestId := "rid1"
	claims := auth.HriAzClaims{
		Roles:   []string{"hri_data_integrator", "hri_tenant_tid1_data_integrator"},
		Subject: "8b1e7a81-7f4a-41b0-a170-ae19f843f27c",
	}
	resultBody := map[string]interface{}{}
	tenantId := "tid1"

	response := checkBatchAuth(requestId, &claims, resultBody, tenantId)
	if response.Code != expetedResponse.Code {
		t.Errorf("checkBatchAuth() = \n\t%v,\nexpected: \n\t%v", response.Code, expetedResponse.Code)
	}

}

func Test_checkBatchAuth401_noclame(t *testing.T) {
	expetedResponse := response.ErrorDetailResponse{
		Code: 401,
	}
	requestId := "rid1"
	claims := auth.HriAzClaims{}
	resultBody := map[string]interface{}{"integratorId": "8b1e7a81-7f4a-41b0-a170-ae19f843f27c"}
	tenantId := "tid1"

	response := checkBatchAuth(requestId, &claims, resultBody, tenantId)
	if response.Code != expetedResponse.Code {
		t.Errorf("checkBatchAuth() = \n\t%v,\nexpected: \n\t%v", response.Code, expetedResponse.Code)
	}

}
