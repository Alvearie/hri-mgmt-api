/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package batches

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const msgMissingStatusElem = "Error: Cosmos Search Result body does Not have the expected 'integratorId' Element"

const docNotFoundMsg string = "The document for tenantId: %s with document (batch) ID: %s was not found"

func GetByBatchId(requestId string, batch model.GetByIdBatch, claims auth.HriAzClaims) (int, interface{}) {
	prefix := "batches/GetByBatchId"
	logger := logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Tenants Get By ID(Metadata)")
	// validate that caller has sufficient permissions
	if !claims.HasRole(auth.HriIntegrator) || !claims.HasRole(auth.GetAuthRole(batch.TenantId, auth.HriIntegrator)) || !claims.HasRole(auth.HriConsumer) || !claims.HasRole(auth.GetAuthRole(batch.TenantId, auth.HriConsumer)) {
		errMsg := auth.MsgAccessTokenMissingScopes
		logger.Errorln(errMsg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, errMsg)
	}

	logger.Debugf("params_tenantID: %v, batchID: %v", batch.TenantId, batch.BatchId)
	var noAuthFlag = false
	return getByBatchId(requestId, batch, noAuthFlag, logger, &claims)

}

func GetByBatchIdNoAuth(requestId string, params model.GetByIdBatch,
	_ auth.HriAzClaims) (int, interface{}) {

	prefix := "batches/GetByBatchIdNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch GetById (No Auth)")

	var noAuthFlag = true
	return getByBatchId(requestId, params, noAuthFlag, logger, nil)
}
func getByBatchId(requestId string, batch model.GetByIdBatch,
	noAuthFlag bool, logger logrus.FieldLogger,
	claims *auth.HriAzClaims) (int, interface{}) {

	//Apending "-batches" to tenants id
	index := mongoApi.GetTenantWithBatchesSuffix(batch.TenantId)
	logger.Debugf("index: %v", index)

	var details []bson.M

	// Query Cosmos for information on the tenant
	projection := bson.D{
		{"batch", bson.D{
			{"$elemMatch", bson.D{
				{"id", batch.BatchId},
			}}}},
		{"_id", 0},
	}

	cursor, err := mongoApi.HriCollection.Find(
		context.TODO(),
		bson.D{
			{"tenantId", index},
		},
		options.Find().SetProjection(projection),
	)

	if err != nil {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, "Get batch by ID failed")
	}
	if err = cursor.All(context.TODO(), &details); err != nil {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, "Get batch by ID failed")
	}
	msg := fmt.Sprintf(docNotFoundMsg, batch.TenantId, batch.BatchId)

	if details == nil {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetailWithoutStatusCode(requestId, logger, msg)
	}
	batchMap, ok := details[0]["batch"].(primitive.A)
	batchMapSlice := []interface{}(batchMap)

	if len(batchMapSlice) == 0 || !ok {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetailWithoutStatusCode(requestId, logger, msg)
	}

	mapResponseBody, _ := batchMapSlice[0].(primitive.M)
	mapResponse := map[string]interface{}(mapResponseBody)

	if !noAuthFlag {
		errDetailResponse := checkBatchAuth(requestId, claims, mapResponse, batch.TenantId)
		if errDetailResponse != nil {
			return errDetailResponse.Code, errDetailResponse.Body
		}
	}

	return http.StatusOK, NormalizeBatchRecordCountValues(mapResponse)

}

// Data Integrators and Consumers can call this endpoint, but the behavior is slightly different. Consumers can see
// all Batches, but Data Integrators are only allowed to see Batches they created.
func checkBatchAuth(requestId string, claims *auth.HriAzClaims, resultBody map[string]interface{}, tenantId string) *response.ErrorDetailResponse {
	if claims.HasRole(auth.HriConsumer) && claims.HasRole(auth.GetAuthRole(tenantId, auth.HriConsumer)) { //= Always Authorized
		return nil // return nil Error for Authorized
	}
	if claims.HasRole(auth.HriIntegrator) && claims.HasRole(auth.GetAuthRole(tenantId, auth.HriIntegrator)) {
		if integratorId, ok := resultBody[param.IntegratorId].(string); ok {
			//if claims.Subject from the token does NOT match the previously saved batch.IntegratorId, user NOT Authorized
			if claims.Subject != integratorId {
				errMsg := fmt.Sprintf(auth.MsgIntegratorSubClaimNoMatch, claims.Subject, integratorId)
				return response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, errMsg)
			}
		} else { //integratorId elem does Not exist - Internal Server Error
			return response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msgMissingStatusElem)
		}
	} else { //No Scope was provided -> Unauthorized - we should never reach here
		errMsg := auth.MsgAccessTokenMissingScopes
		return response.NewErrorDetailResponse(http.StatusUnauthorized, requestId, errMsg)
	}

	return nil //Default Return: we are Authorized  => nil error
}
