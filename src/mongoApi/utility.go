package mongoApi

import (
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ConvertToJSON(tenantId, testShardId string) model.CreateTenantRequest {
	result := model.CreateTenantRequest{
		ID:            primitive.NewObjectID(),
		Test_shard_id: testShardId,
		TenantId:      tenantId,
	}
	return result
}
