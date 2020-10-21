/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	// "fmt"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v6"
	"log"
	"net/http"
	"os"
)

func GetById(params map[string]interface{}, client *elasticsearch.Client) map[string]interface{} {
	logger := log.New(os.Stdout, "batches/GetById: ", log.Llongfile)

	// validate that required Two input params are present ONLY in the PATH param
	tenantId, err := path.ExtractParam(params, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	batchId, err := path.ExtractParam(params, param.BatchIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	logger.Printf("params_tenantID: %v, batchID: %v", tenantId, batchId)

	index := elastic.IndexFromTenantId(tenantId)
	logger.Printf("index: %v", index)

	res, err := client.Get(index, BatchIdToEsDocId(batchId))

	resultBody, errResp := elastic.DecodeBody(res, err, tenantId, logger)
	if errResp != nil {
		return errResp
	}

	return response.Success(http.StatusOK, EsDocToBatch(resultBody))
}
