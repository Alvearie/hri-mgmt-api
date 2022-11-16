/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"context"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetBatch(requestId string, params model.GetBatch, claims auth.HriAzClaims, mongoClient *mongo.Collection) (int, interface{}) {
	prefix := "batches/get"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Get")

	// Data Integrators and Consumers can use this endpoint, so either scope allows access
	if !claims.HasRole(auth.HriConsumer) && !claims.HasRole(auth.HriIntegrator) {
		errMsg := auth.MsgAccessTokenMissingScopes
		logger.Errorln(errMsg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, errMsg)
	}

	return getBatch(requestId, params, false, &claims, mongoClient, logger)
}

func GetBatchNoAuth(requestId string, params model.GetBatch, _ auth.HriAzClaims, mongoClient *mongo.Collection) (int, interface{}) {
	prefix := "batches/getNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Get (No Auth)")

	var noAuthFlag = true
	return getBatch(requestId, params, noAuthFlag, nil, mongoClient, logger)
}

func getBatch(requestId string, params model.GetBatch, noAuthFlag bool, claims *auth.HriAzClaims,
	mongoClient *mongo.Collection, logger logrus.FieldLogger) (int, interface{}) {

	tenantId := params.TenantId

	var ctx = context.Background()
	var filter = bson.M{"tenantId": mongoApi.GetTenantWithBatchesSuffix(tenantId)}
	var returnTenetResult model.GetBatchTenantDetail
	//var tenantResponse model.TenatGetResponse

	mongoClient.FindOne(ctx, filter).Decode(&returnTenetResult)

	errMsg := "Tenant: " + tenantId + " not found"
	if returnTenetResult.TenantId == "" {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound, logger, errMsg)
	}

	return http.StatusOK, map[string]interface{}{
		"total":   len(returnTenetResult.Result),
		"results": returnTenetResult.Result,
	}

}
