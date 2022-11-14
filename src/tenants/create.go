/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/elastic/go-elasticsearch/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Create(
	requestId string,
	tenantId string,
	esClient *elasticsearch.Client) (int, interface{}) {

	prefix := "tenants/Create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	//create new index
	//esClient.Indices.Get()
	indexRes, err := esClient.Indices.Create(elastic.IndexFromTenantId(tenantId))

	// parse the response
	_, elasticErr := elastic.DecodeBody(indexRes, err)
	if elasticErr != nil {
		return elasticErr.Code, elasticErr.LogAndBuildErrorDetail(requestId, logger,
			fmt.Sprintf("Unable to create new tenant [%s]", tenantId))
	}

	// return the ID of the newly created tenant
	respBody := map[string]interface{}{param.TenantId: tenantId}
	return http.StatusCreated, respBody
}

func CreateTenant(
	requestId string,
	tenantId string,
	mongoClient *mongo.Collection) (int, interface{}) {

	prefix := "tenants/CreateTenant"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	var ctx = context.Background()
	var filter = bson.M{"tenantId": mongoApi.IndexFromTenantId(tenantId)}
	var returnResult model.CreateTenantRequest

	//create new tenant In azure cosmos- mongo API
	//As it is a new tenant creation passing docCount and docDeleted as 0
	createTenantRequest := mongoApi.ConvertToJSON(mongoApi.IndexFromTenantId(tenantId), "0", 0)

	//check if duplicate tenantId or not
	res := mongoClient.FindOne(ctx, filter).Decode(&returnResult)
	fmt.Println("find by id result ", res)

	if (model.CreateTenantRequest{}) != returnResult {
		return http.StatusBadRequest, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusBadRequest, logger,
			fmt.Sprintf("Unable to create new tenant as it already exists[%s]", tenantId))
	}

	// Insert one
	_, mongoErr := mongoClient.InsertOne(ctx, createTenantRequest)

	if mongoErr != nil {
		return http.StatusBadRequest, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusBadRequest, logger,
			fmt.Sprintf("Unable to create new tenant [%s]", tenantId))
	}

	// return the ID of the newly created tenant
	respBody := map[string]interface{}{param.TenantId: tenantId}
	return http.StatusCreated, respBody
}
