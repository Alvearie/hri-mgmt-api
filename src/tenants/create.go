/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v6"
	"log"
	"net/http"
	"os"
)

func Create(
	args map[string]interface{},
	validator param.Validator,
	esClient *elasticsearch.Client) map[string]interface{} {

	logger := log.New(os.Stdout, "tenants/create: ", log.Llongfile)

	// extract tenantId path param from URL
	tenantId, err := path.ExtractParam(args, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	if err := param.TenantIdCheck(tenantId); err != nil {
		return response.Error(http.StatusBadRequest, err.Error())
	}

	//create new index
	indexRes, err := esClient.Indices.Create(elastic.IndexFromTenantId(tenantId))

	if err != nil {
		logger.Printf("Unable to publish new tenant [%s]. %s", tenantId, err.Error())
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	// parse the response
	_, errRes := elastic.DecodeBody(indexRes, err, tenantId, logger)
	if errRes != nil {
		return errRes
	}

	// return the ID of the newly created tenant
	respBody := map[string]interface{}{param.TenantId: tenantId}
	return response.Success(http.StatusCreated, respBody)
}
