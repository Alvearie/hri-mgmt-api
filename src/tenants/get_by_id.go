/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"context"
	"strconv"

	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"net/http"
)

func GetTenantById(
	requestId string,
	tenantId string,
	mongoClient *mongo.Collection) (int, interface{}) {

	prefix := "tenants/GetTenantById"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	var ctx = context.Background()
	var filter = bson.M{"tenantId": mongoApi.GetTenantWithBatchesSuffix(tenantId)}
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
