/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"ibm.com/watson/health/foundation/hri/common/elastic"
	"ibm.com/watson/health/foundation/hri/common/logwrapper"
	"ibm.com/watson/health/foundation/hri/common/param"
	"net/http"
)

func Create(
	requestId string,
	tenantId string,
	esClient *elasticsearch.Client) (int, interface{}) {

	prefix := "tenants/Create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	//create new index
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
