package tenants

import (
	"context"
	"fmt"

	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"

	// "fmt"
	"net/http"
)

func DeleteTenant(requestId string, tenantId string) (int, interface{}) {
	prefix := "tenants/Delete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Tenant Delete")

	var ctx = context.Background()
	var filter = bson.M{"tenantId": mongoApi.GetTenantWithBatchesSuffix(tenantId)}

	//make call to elastic to delete tenant
	result, err := mongoApi.HriCollection.DeleteOne(ctx, filter)

	if err != nil {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound,
			logger, fmt.Sprintf("Could not delete tenant [%s]", tenantId))
	}

	if result.DeletedCount == 0 {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound,
			logger, fmt.Sprintf("Could not delete tenant [%s]", tenantId+"-batches"))
	}

	return http.StatusOK, nil
}
