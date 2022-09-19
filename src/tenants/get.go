/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"context"
	"strings"

	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/elastic/go-elasticsearch/v7"
)

func Get(requestId string, client *elasticsearch.Client) (int, interface{}) {
	prefix := "tenants/Get"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Tenants Get (All)")

	//Use elastic to return the list of indices
	res, err := client.Cat.Indices(client.Cat.Indices.WithH("index"), client.Cat.Indices.WithFormat("json"))

	body, elasticErr := elastic.DecodeBodyFromJsonArray(res, err)
	if elasticErr != nil {
		return elasticErr.Code, elasticErr.LogAndBuildErrorDetail(requestId,
			logger, "Could not retrieve tenants")
	}

	//sort through result for tenantIds and add to an array
	tenantsMap := elastic.TenantsFromIndices(body)

	return http.StatusOK, tenantsMap
}

func GetTenants(
	requestId string,
	mongoClient *mongo.Collection) (int, interface{}) {

	prefix := "tenants/GetTenants"
	logger := logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Tenants Get (All)")

	tenantsMap := make(map[string]interface{})
	var tenantsList []model.GetTenantId

	projection := bson.D{
		{"tenantid", 1},
		{"_id", 0},
	}

	cursor, err := mongoClient.Find(
		context.TODO(),
		bson.D{},
		options.Find().SetProjection(projection),
	)

	if err != nil {
		return http.StatusBadRequest, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, "Could not retrieve tenants")
	}

	if err = cursor.All(context.TODO(), &tenantsList); err != nil {
		return http.StatusBadRequest, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, "Could not retrieve tenants")
	}

	tenenatsIdList := []model.GetTenantId{}

	for _, tenantRecord := range tenantsList {

		if strings.Contains(tenantRecord.TenantId, "-batches") {
			tenenatsIdList = append(tenenatsIdList, model.GetTenantId{TenantId: mongoApi.TenantIdFromIndex(tenantRecord.TenantId)})
		}
	}

	tenantsMap["results"] = tenenatsIdList

	return http.StatusOK, tenantsMap

}
