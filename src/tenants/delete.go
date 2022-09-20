/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"context"
	"fmt"

	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	// "fmt"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/elastic/go-elasticsearch/v7"
)

func Delete(requestId string, tenantId string, client *elasticsearch.Client) (int, interface{}) {
	prefix := "tenants/Delete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Tenant Delete")

	index := []string{elastic.IndexFromTenantId(tenantId)}

	//make call to elastic to delete tenant
	res, err2 := client.Indices.Delete(index)

	_, elasticErr := elastic.DecodeBody(res, err2)
	if elasticErr != nil {
		return elasticErr.Code, elasticErr.LogAndBuildErrorDetail(requestId,
			logger, fmt.Sprintf("Could not delete tenant [%s]", tenantId))
	}

	return http.StatusOK, nil
}

func DeleteTenant(requestId string, tenantId string, mongoClient *mongo.Collection) (int, interface{}) {
	prefix := "tenants/Delete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Tenant Delete")

	var ctx = context.Background()
	var filter = bson.M{"tenantid": mongoApi.IndexFromTenantId(tenantId)}

	//make call to elastic to delete tenant
	result, err := mongoClient.DeleteOne(ctx, filter)

	if err != nil {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound,
			logger, fmt.Sprintf("Could not delete tenant [%s]", tenantId))
	}

	if result.DeletedCount == 0 {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound,
			logger, fmt.Sprintf("index_not_found_exception: no such index [%s]", tenantId+"-batches"))
	}

	return http.StatusOK, nil
}
