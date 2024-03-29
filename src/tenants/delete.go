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
