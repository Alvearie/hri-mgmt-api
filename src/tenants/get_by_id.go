/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"fmt"
	// "fmt"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"net/http"
	"os"
)

func GetById(params map[string]interface{}, client *elasticsearch.Client) map[string]interface{} {
	logger := log.New(os.Stdout, "tenants/GetById: ", log.Llongfile)

	// validate that required input param is present ONLY in the PATH param

	tenantId, err := path.ExtractParam(params, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	// Query elastic for information on the tenant
	index := elastic.IndexFromTenantId(tenantId)
	var res, err2 = client.Cat.Indices(client.Cat.Indices.WithIndex(index), client.Cat.Indices.WithFormat("json"))

	resultBody, elasticErr := elastic.DecodeBodyFromJsonArray(res, err2)
	if elasticErr != nil {
		if elasticErr.Code == http.StatusNotFound {
			msg := "Tenant: " + tenantId + " not found"
			logger.Println(msg)
			return response.Error(http.StatusNotFound, msg)
		}

		return elasticErr.LogAndBuildApiResponse(logger, fmt.Sprintf("Could not retrieve tenant %s", tenantId))
	}

	return response.Success(http.StatusOK, resultBody[0])
}
