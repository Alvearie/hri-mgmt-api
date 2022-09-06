package mongoApi

import (
	"github.com/Alvearie/hri-mgmt-api/common/model"
)

func ConvertToJSON(tenantId string, docCount string, docsDeleted string) model.CreateTenantRequest {
	result := model.CreateTenantRequest{
		//ID:            primitive.NewObjectID(),
		TenantId:     tenantId,
		Docs_count:   docCount,
		Docs_deleted: docsDeleted,
	}
	return result
}
