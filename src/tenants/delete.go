/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
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
	if err2 != nil {
		logger.Println(err2.Error())
		return response.Error(http.StatusInternalServerError, err2.Error())
	}

	_, errResp := elastic.DecodeBody(res, err2, tenantId, logger)

	if errResp != nil {
		logger.Printf("Error in decode: %v", errResp)
		return errResp
	}

	return map[string]interface{}{
		"statusCode": http.StatusOK,
	}
}
