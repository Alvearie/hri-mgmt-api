/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"net/http"
	"os"
)

const msgMissingStatusElem = "Error: Elastic Search Result body does Not have the expected '_source' Element"
const msgDocNotFound string = "The document for tenantId: %s with document (batch) ID: %s was not found"

func GetById(params map[string]interface{}, claims auth.HriClaims, client *elasticsearch.Client) map[string]interface{} {
	logger := log.New(os.Stdout, "batches/GetById: ", log.Llongfile)

	if !claims.HasScope(auth.HriIntegrator) && !claims.HasScope(auth.HriConsumer) {
		errMsg := auth.MsgAccessTokenMissingScopes
		logger.Println(errMsg)
		return response.Error(http.StatusUnauthorized, errMsg)
	}

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

	res, err := client.Get(index, batchId)

	resultBody, elasticErr := elastic.DecodeBody(res, err)
	if elasticErr != nil {
		if elasticErr.ErrorObj == nil && resultBody != nil && docNotFound(resultBody) {
			msg := fmt.Sprintf(msgDocNotFound, tenantId, batchId)
			logger.Println(msg)
			return response.Error(http.StatusNotFound, msg)
		}

		return elasticErr.LogAndBuildApiResponse(logger,
			fmt.Sprintf("Could not retrieve batch with id: %s", batchId))
	}

	errResponse := checkBatchAuthorization(claims, resultBody)
	if errResponse != nil {
		return errResponse
	}

	return response.Success(http.StatusOK, EsDocToBatch(resultBody))
}

func docNotFound(resultBody map[string]interface{}) bool {
	found, ok := resultBody["found"].(bool)
	return ok && !found
}

// Data Integrators and Consumers can call this endpoint, but the behavior is slightly different. Consumers can see
// all Batches, but Data Integrators are only allowed to see Batches they created.
func checkBatchAuthorization(claims auth.HriClaims, resultBody map[string]interface{}) map[string]interface{} {
	if claims.HasScope(auth.HriConsumer) { //= Always Authorized
		return nil // return nil Error for Authorized
	}

	if claims.HasScope(auth.HriIntegrator) {
		if sourceBody, ok := resultBody["_source"].(map[string]interface{}); ok {
			integratorId := sourceBody[param.IntegratorId]
			//if claims.Subject from the token does NOT match the previously saved batch.IntegratorId, user NOT Authorized
			if claims.Subject != integratorId {
				errMsg := fmt.Sprintf(auth.MsgIntegratorSubClaimNoMatch, claims.Subject, integratorId)
				return response.Error(http.StatusUnauthorized, errMsg)
			}
		} else { //_source elem does Not exist - Internal Server Error
			return response.Error(http.StatusInternalServerError, msgMissingStatusElem)
		}
	} else { //No Scope was provided -> Unauthorized - we should never reach here
		errMsg := auth.MsgAccessTokenMissingScopes
		return response.Error(http.StatusUnauthorized, errMsg)
	}

	return nil //Default Return: we are Authorized  => nil error
}
