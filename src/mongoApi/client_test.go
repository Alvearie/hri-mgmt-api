package mongoApi

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/stretchr/testify/assert"
)

func TestGetTenantWithBatchesSuffix(t *testing.T) {
	actual := GetTenantWithBatchesSuffix("tenants_id")
	expected := "tenants_id-batches"
	assert.Equal(t, expected, actual)

}

func TestTenantIdWithSuffix(t *testing.T) {
	actual := TenantIdWithSuffix("tenants_id-batches")
	expected := "tenants_id"
	assert.Equal(t, expected, actual)

}

func TestLogAndBuildErrorDetail(t *testing.T) {
	requestId := "request_id"
	prefix := "any"
	logger := logwrapper.GetMyLogger(requestId, prefix)
	actual := LogAndBuildErrorDetail(requestId, 200, logger, "ErrorMessage")

	assert.NotEmpty(t, actual)

}

func TestLogAndBuildErrorDetailWithoutStatusCode(t *testing.T) {
	requestId := "request_id"
	prefix := "any"
	logger := logwrapper.GetMyLogger(requestId, prefix)
	actual := LogAndBuildErrorDetailWithoutStatusCode(requestId, logger, "ErrorMessage")
	assert.NotEmpty(t, actual)

}
func TestGetMongoCollection(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	mt.Run("success", func(mt *mtest.T) {
		db = mt.DB
		res := GetMongoCollection("any")
		assert.NotNil(t, res)
	})
}

func TestHriDatabaseHealthCheck(t *testing.T) {

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	mt.Run("success", func(mt *mtest.T) {
		HriCollection = mt.Coll
		db = mt.DB
		res, res2, _ := HriDatabaseHealthCheck()

		assert.NotNil(t, res)
		assert.NotNil(t, res2)

	})
}
