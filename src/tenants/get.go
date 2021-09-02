/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"net/http"
	"os"
)

func Get(client *elasticsearch.Client) map[string]interface{} {
	logger := log.New(os.Stdout, "tenants/get: ", log.Llongfile)

	//Use elastic to return the list of indices
	res, err := client.Cat.Indices(client.Cat.Indices.WithH("index"), client.Cat.Indices.WithFormat("json"))
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	body, errResp := elastic.DecodeBodyFromJsonArray(res, err, logger)
	if errResp != nil {
		return errResp
	}

	//sort through result for tenantIds and add to an array
	tenantsMap := elastic.TenantsFromIndices(body)

	return response.Success(http.StatusOK, tenantsMap)
}
