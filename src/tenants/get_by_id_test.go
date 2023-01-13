package tenants

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestGetTenantById(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	requestId := "request_id_1"
	tenantId := "test-batches"
	mt.Run("tenantNotFound", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		expectedTenant := model.TenatGetResponse{
			Health:      "red",
			Status:      "close",
			Index:       "test-batches",
			Uuid:        primitive.NewObjectID(),
			Size:        "270336",
			DocsCount:   "0",
			DocsDeleted: "0",
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "getTenantById", mtest.FirstBatch, bson.D{
			{Key: "health", Value: expectedTenant.Health},
			{Key: "_id", Value: expectedTenant.Uuid},
			{Key: "docs.count", Value: expectedTenant.DocsCount},
			{Key: "docs.deleted", Value: expectedTenant.DocsDeleted},
			{Key: "status", Value: expectedTenant.Status},
			{Key: "tenantId", Value: expectedTenant.Index},
			{Key: "size", Value: expectedTenant.Size},
		}))
		_, err := GetTenantById(requestId, tenantId)
		assert.NotNil(t, err)

	})

	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		expectedUser := model.GetTenantDetail{
			Uuid:         primitive.NewObjectID(),
			TenantId:     "test-batches",
			Docs_count:   "10",
			Docs_deleted: 0,
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedUser.Uuid},
			{Key: "tenantId", Value: expectedUser.TenantId},
			{Key: "docs_count", Value: expectedUser.Docs_count},
			{Key: "docs_deleted", Value: expectedUser.Docs_deleted},
		}))
		statusCode, response := GetTenantById(requestId, tenantId)
		assert.NotNil(t, response)
		assert.Equal(t, statusCode, 200)
	})

	mt.Run("success-EmptyDocCount", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		expectedUser := model.GetTenantDetail{
			Uuid:         primitive.NewObjectID(),
			TenantId:     "test-batches",
			Docs_count:   "",
			Docs_deleted: 0,
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedUser.Uuid},
			{Key: "tenantId", Value: expectedUser.TenantId},
			{Key: "docs_count", Value: expectedUser.Docs_count},
			{Key: "docs_deleted", Value: expectedUser.Docs_deleted},
		}))
		statusCode, response := GetTenantById(requestId, tenantId)
		assert.NotNil(t, response)
		assert.Equal(t, statusCode, 200)
	})
}
