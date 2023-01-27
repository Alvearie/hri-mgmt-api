package tenants

import (
	"context"
	"strings"

	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
)

func GetTenants(
	requestId string) (int, interface{}) {

	prefix := "tenants/GetTenants"
	logger := logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Tenants Get (All)")

	tenantsMap := make(map[string]interface{})

	var tenantsArray []model.GetTenants

	projection := bson.D{
		{Key: "tenantId", Value: 1},
		{Key: "_id", Value: 0},
	}

	cursor, err := mongoApi.HriCollection.Find(
		context.TODO(),
		bson.D{},
		options.Find().SetProjection(projection),
	)

	if err != nil {
		return http.StatusBadRequest, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, "Could not retrieve tenants")
	}

	if err = cursor.All(context.TODO(), &tenantsArray); err != nil {
		return http.StatusBadRequest, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, "Could not retrieve tenants")
	}

	tenenatsIdList := []model.GetTenantId{}

	for _, tenantRecord := range tenantsArray {

		if strings.Contains(tenantRecord.TenantId, "-batches") {
			tenenatsIdList = append(tenenatsIdList, model.GetTenantId{TenantId: mongoApi.TenantIdWithSuffix(tenantRecord.TenantId)})
		}
	}

	tenantsMap["results"] = tenenatsIdList

	return http.StatusOK, tenantsMap

}
