package tenants

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func CreateTenant(
	requestId string,
	tenantId string) (int, interface{}) {

	prefix := "tenants/CreateTenant"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	var ctx = context.Background()
	var filter = bson.M{"tenantId": mongoApi.GetTenantWithBatchesSuffix(tenantId)}
	var returnResult model.CreateTenantRequest

	update := bson.M{
		"$set": bson.M{"tenantId": mongoApi.GetTenantWithBatchesSuffix(tenantId)},
	}
	upsert := true
	opt := options.FindOneAndUpdateOptions{
		Upsert: &upsert,
	}
	result := mongoApi.HriCollection.FindOneAndUpdate(ctx, filter, update, &opt).Decode(&returnResult)

	if result != nil && !strings.Contains(result.Error(), "no documents in result") {
		return http.StatusBadRequest, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusBadRequest, logger,
			fmt.Sprintf("Unable to create new tenant [%s]", tenantId+" - "+result.Error()))
	}

	if returnResult.TenantId == mongoApi.GetTenantWithBatchesSuffix(tenantId) {
		return http.StatusBadRequest, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusBadRequest, logger,
			fmt.Sprintf("Unable to create new tenant as it already exists[%s]", tenantId))
	}

	// return the ID of the newly created tenant
	respBody := map[string]interface{}{param.TenantId: tenantId}
	return http.StatusCreated, respBody
}
