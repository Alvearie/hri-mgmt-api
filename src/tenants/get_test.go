package tenants

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestFind(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	requestId := "request-Id1"
	tenantsMap := make(map[string]interface{})
	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		tenantId1 := "test-batches"
		tenantId2 := "provider1234-batches"
		tenantId3 := "pentest-batches"
		tenantId4 := "test123-batches"

		tenant1 := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"tenantId", tenantId1},
		})

		tenant2 := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, bson.D{
			{"tenantId", tenantId2},
		})
		tenant3 := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, bson.D{
			{"tenantId", tenantId3},
		})
		tenant4 := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, bson.D{
			{"tenantId", tenantId4},
		})
		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)
		mt.AddMockResponses(tenant1, tenant2, tenant3, tenant4, killCursors)

		tenants, response := GetTenants(requestId)
		tenenatsIdList := []model.GetTenantId{}

		tenantsMap["results"] = []model.GetTenantId{}
		tenenatsIdList = append(tenenatsIdList, model.GetTenantId{TenantId: mongoApi.TenantIdWithSuffix(tenantId1)})
		tenenatsIdList = append(tenenatsIdList, model.GetTenantId{TenantId: mongoApi.TenantIdWithSuffix(tenantId2)})
		tenenatsIdList = append(tenenatsIdList, model.GetTenantId{TenantId: mongoApi.TenantIdWithSuffix(tenantId3)})
		tenenatsIdList = append(tenenatsIdList, model.GetTenantId{TenantId: mongoApi.TenantIdWithSuffix(tenantId4)})

		tenantsMap["results"] = tenenatsIdList
		assert.NotNil(t, tenants)
		assert.NotNil(t, response)
		//assert.Equal(t, tenantsMap, response)
	})
}
