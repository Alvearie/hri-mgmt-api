package tenants

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestDeleteOneTenant(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	requestId := "request_id_1"
	tenantId := "test-batches"

	tenantId2 := "test"
	mt.Run("success", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}})
		statusCode, response := DeleteTenant(requestId, tenantId)
		assert.NotNil(t, statusCode)
		assert.Equal(t, statusCode, 200)
		assert.Nil(t, response)
	})

	mt.Run("no document deleted", func(mt *mtest.T) {
		mongoApi.HriCollection = mt.Coll
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 0}})
		statusCode, response := DeleteTenant(requestId, tenantId2)
		assert.NotNil(t, statusCode)
		assert.Equal(t, statusCode, 404)
		assert.NotNil(t, response)
	})
}
