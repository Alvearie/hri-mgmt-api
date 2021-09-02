/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"

	// "fmt"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/elastic/go-elasticsearch/v7"
	"net/http"
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
