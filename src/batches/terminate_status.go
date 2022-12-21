/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package batches

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func TerminateBatch(
	requestId string,
	request *model.TerminateRequest,
	claims auth.HriAzClaims,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/terminate"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Terminate")

	// Only Integrators can call terminate
	if !claims.HasRole(auth.HriIntegrator) || !claims.HasRole(auth.GetAuthRole(request.TenantId, auth.HriIntegrator)) {
		msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "terminate")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	var subject = claims.Subject
	return terminateBatch(requestId, request, subject, logger, writer, currentStatus)
}

func TerminateBatchNoAuth(
	requestId string,
	request *model.TerminateRequest,
	_ auth.HriAzClaims,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/TerminateNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Terminate (No Auth)")

	var subject = auth.NoAuthFakeIntegrator
	return terminateBatch(requestId, request, subject, logger, writer, currentStatus)
}

func terminateBatch(requestId string, request *model.TerminateRequest,
	claimsSubject string,
	logger logrus.FieldLogger,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	var claims = auth.HriAzClaims{}
	var getBatchRequest = model.GetByIdBatch{TenantId: request.TenantId, BatchId: request.BatchId}
	var updateRequest = map[string]interface{}{}

	getByIdCode, getByIdBody := GetByBatchIdNoAuth(requestId, getBatchRequest, claims)

	if getByIdCode != 200 {
		return getByIdCode, getByIdBody
	}

	batchDetail, ok := getByIdBody.(map[string]interface{})

	if status.Started.String() != batchDetail[param.Status] {
		errMsg := fmt.Sprintf("terminate failed, batch is in '%s' state", batchDetail[param.Status].(string))
		logger.Errorln(errMsg)
		return http.StatusConflict, response.NewErrorDetail(requestId, errMsg)
	}

	if ok && claimsSubject == batchDetail[param.IntegratorId] {
		updateRequest = getTerminateUpdateRequest(request, claimsSubject, batchDetail)
	} else {
		errMsg := fmt.Sprintf("terminate requested by '%s' but owned by '%s'", claimsSubject, batchDetail[param.IntegratorId])
		logger.Errorln(errMsg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, errMsg)
	}

	errResp := updateBatchStatus(requestId, request.TenantId, request.BatchId, updateRequest, writer, currentStatus)
	if errResp != nil {
		return errResp.Code, errResp.Body
	}

	return http.StatusOK, nil
}

func getTerminateUpdateRequest(request *model.TerminateRequest, claimsSubject string, batch map[string]interface{}) map[string]interface{} {
	currentTime := time.Now().UTC()

	var updateRequest map[string]interface{}
	if request.Metadata == nil {
		updateRequest = bson.M{
			"$set": bson.M{
				"batch.$.status":  status.Terminated.String(),
				"batch.$.endDate": currentTime.Format(mongoApi.DateTimeFormat),
			},
			"$inc": bson.M{"docs_deleted": 1},
		}
	} else {
		updateRequest = bson.M{
			"$set": bson.M{
				"batch.$.metadata": request.Metadata,
				"batch.$.status":   status.Terminated.String(),
				"batch.$.endDate":  currentTime.Format(mongoApi.DateTimeFormat),
			},
			"$inc": bson.M{"docs_deleted": 1},
		}
	}
	return updateRequest
}
