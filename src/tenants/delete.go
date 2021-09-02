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

func Delete(params map[string]interface{}, client *elasticsearch.Client) map[string]interface{} {
	logger := log.New(os.Stdout, "tenants/Delete: ", log.Llongfile)

	// validate that required input param is present ONLY in the PATH param
	tenantId, err := path.ExtractParam(params, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	index := []string{elastic.IndexFromTenantId(tenantId)}

	//make call to elastic to delete tenant
	res, err2 := client.Indices.Delete(index)

	_, elasticErr := elastic.DecodeBody(res, err2)
	if elasticErr != nil {
		return elasticErr.LogAndBuildApiResponse(logger, fmt.Sprintf("Could not delete tenant [%s]", tenantId))
	}

	return map[string]interface{}{
		"statusCode": http.StatusOK,
	}
}
