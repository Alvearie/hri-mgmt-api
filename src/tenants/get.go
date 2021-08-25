/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"github.com/elastic/go-elasticsearch/v7"
	"ibm.com/watson/health/foundation/hri/common/elastic"
	"ibm.com/watson/health/foundation/hri/common/logwrapper"
	"net/http"
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
