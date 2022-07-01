/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/elastic/go-elasticsearch/v7"
	"net/http"
)

func Create(
	requestId string,
	tenantId string,
	esClient *elasticsearch.Client) (int, interface{}) {

	prefix := "tenants/Create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Infof("Start Create Tenant ")
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
	logger.Infof("TenantId [%s] ", tenantId)
	logger.Infof("End Create Tenant ")
	return http.StatusCreated, respBody
}
