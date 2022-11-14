/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	// "fmt"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/elastic/go-elasticsearch/v7"
)

func GetById(requestId string, tenantId string, client *elasticsearch.Client) (int, interface{}) {
	prefix := "tenant/GetById"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Tenants Get By ID")

	// Query elastic for information on the tenant
	index := elastic.IndexFromTenantId(tenantId)
	var res, err2 = client.Cat.Indices(client.Cat.Indices.WithIndex(index),
		client.Cat.Indices.WithFormat("json"))

	resultBody, elasticErr := elastic.DecodeBodyFromJsonArray(res, err2)
	if elasticErr != nil {
		if elasticErr.Code == http.StatusNotFound {
			msg := "Tenant: " + tenantId + " not found"
			return http.StatusNotFound, elasticErr.LogAndBuildErrorDetail(requestId, logger, msg)
		}

		msg := fmt.Sprintf("Could not retrieve tenant '%s'", tenantId)
		return elasticErr.Code, elasticErr.LogAndBuildErrorDetail(requestId, logger, msg)
	}

	return http.StatusOK, resultBody[0]
}

func GetTenantById(
	requestId string,
	tenantId string,
	mongoClient *mongo.Collection) (int, interface{}) {

	prefix := "tenants/GetTenantById"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	var ctx = context.Background()
	var filter = bson.M{"tenantId": mongoApi.IndexFromTenantId(tenantId)}
	var returnTenetResult model.GetTenantDetail
	var tenantResponse model.TenatGetResponse

	mongoClient.FindOne(ctx, filter).Decode(&returnTenetResult)

	if (model.GetTenantDetail{}) == returnTenetResult {
		msg := "Tenant: " + tenantId + " not found"
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, msg)
	}

	healthOk, datasize := mongoApi.DatabaseHealthCheck(mongoClient)
	tenantResponse.Index = returnTenetResult.TenantId
	tenantResponse.Uuid = returnTenetResult.Uuid
	tenantResponse.DocsCount = returnTenetResult.Docs_count
	tenantResponse.DocsDeleted = strconv.Itoa(returnTenetResult.Docs_deleted)
	tenantResponse.Size = datasize
	if healthOk == "1" {
		tenantResponse.Health = "green"
		tenantResponse.Status = "open"
	} else {
		tenantResponse.Health = "red"
		tenantResponse.Status = "close"
	}

	return http.StatusOK, tenantResponse
}
