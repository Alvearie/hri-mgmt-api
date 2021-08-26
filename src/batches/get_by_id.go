/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"ibm.com/watson/health/foundation/hri/common/auth"
	"ibm.com/watson/health/foundation/hri/common/elastic"
	"ibm.com/watson/health/foundation/hri/common/logwrapper"
	"ibm.com/watson/health/foundation/hri/common/model"
	"ibm.com/watson/health/foundation/hri/common/param"
	"ibm.com/watson/health/foundation/hri/common/response"
	"net/http"
)

const msgMissingStatusElem = "Error: Elastic Search Result body does Not have the expected '_source' Element"
const msgDocNotFound string = "The document for tenantId: %s with document (batch) ID: %s was not found"

func GetById(requestId string, batch model.GetByIdBatch, claims auth.HriClaims, client *elasticsearch.Client) (int, interface{}) {
	prefix := "batches/getById"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch GetById")

	if !claims.HasScope(auth.HriIntegrator) && !claims.HasScope(auth.HriConsumer) {
		errMsg := auth.MsgAccessTokenMissingScopes
		logger.Errorln(errMsg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, errMsg)
	}

	logger.Debugf("params_tenantID: %v, batchID: %v", batch.TenantId, batch.BatchId)

	var noAuthFlag = false
	return getById(requestId, batch, noAuthFlag, logger, &claims, client)
}

func GetByIdNoAuth(requestId string, params model.GetByIdBatch,
	_ auth.HriClaims, client *elasticsearch.Client) (int, interface{}) {

	prefix := "batches/GetByIdNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch GetById (No Auth)")

	var noAuthFlag = true
	return getById(requestId, params, noAuthFlag, logger, nil, client)
}

func getById(requestId string, batch model.GetByIdBatch,
	noAuthFlag bool, logger logrus.FieldLogger,
	claims *auth.HriClaims, client *elasticsearch.Client) (int, interface{}) {

	index := elastic.IndexFromTenantId(batch.TenantId)
	logger.Debugf("index: %v", index)

	res, err := client.Get(index, batch.BatchId)

	resultBody, elasticErr := elastic.DecodeBody(res, err)
	if elasticErr != nil {
		if documentNotFound(elasticErr, resultBody) {
			msg := fmt.Sprintf(msgDocNotFound, batch.TenantId, batch.BatchId)
			logger.Errorln(msg)
			return http.StatusNotFound, response.NewErrorDetail(requestId, msg)
		}

		return http.StatusInternalServerError, elasticErr.LogAndBuildErrorDetail(requestId, logger,
			"Get batch by ID failed")
	}

	if !noAuthFlag {
		errDetailResponse := checkBatchAuthorization(requestId, claims, resultBody)
		if errDetailResponse != nil {
			return errDetailResponse.Code, errDetailResponse.Body
		}
	}

	return http.StatusOK, EsDocToBatch(resultBody)
}

// Check Elastic Decode Body error + response to determine whether the issue is a 404
// This function is called when it has already been determined there is some sort of non-nil elasticErr
// (note - not elasticErr.ErrorObj but specifically elasticErr).
func documentNotFound(elasticErr *elastic.ResponseError, resultBody map[string]interface{}) bool {
	// No error from Elastic, & Doc Not Found is indicated by "found" attr in non-nil result body
	if elasticErr.ErrorObj == nil && resultBody != nil {
		return noDocInResultBody(resultBody)
	}
	// Doc Not Found indicated by 404 error code in Elastic response
	if elasticErr.ErrorObj != nil && elasticErr.Code == http.StatusNotFound {
		return true
	}
	// The error is not related to the document not being found, or it is inconclusive as to whether it is related or not
	return false
}

func noDocInResultBody(resultBody map[string]interface{}) bool {
	found, ok := resultBody["found"].(bool)
	return ok && !found
}

// Data Integrators and Consumers can call this endpoint, but the behavior is slightly different. Consumers can see
// all Batches, but Data Integrators are only allowed to see Batches they created.
func checkBatchAuthorization(requestId string, claims *auth.HriClaims, resultBody map[string]interface{}) *response.ErrorDetailResponse {
	if claims.HasScope(auth.HriConsumer) { //= Always Authorized
		return nil // return nil Error for Authorized
	}

	if claims.HasScope(auth.HriIntegrator) {
		if sourceBody, ok := resultBody["_source"].(map[string]interface{}); ok {
			integratorId := sourceBody[param.IntegratorId]
			//if claims.Subject from the token does NOT match the previously saved batch.IntegratorId, user NOT Authorized
			if claims.Subject != integratorId {
				errMsg := fmt.Sprintf(auth.MsgIntegratorSubClaimNoMatch, claims.Subject, integratorId)
				return response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, errMsg)
			}
		} else { //_source elem does Not exist - Internal Server Error
			return response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msgMissingStatusElem)
		}
	} else { //No Scope was provided -> Unauthorized - we should never reach here
		errMsg := auth.MsgAccessTokenMissingScopes
		return response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, errMsg)
	}

	return nil //Default Return: we are Authorized  => nil error
}
