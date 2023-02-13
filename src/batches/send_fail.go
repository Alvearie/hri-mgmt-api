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
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func SendFail(
	requestId string,
	request *model.FailRequest,
	claims auth.HriAzClaims,
	writer kafka.Writer,
) (int, interface{}) {

	prefix := "batches/sendfail"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Fail")

	// Only Integrators can call fail
	if !claims.HasRole(auth.GetAuthRole(request.TenantId, auth.HriInternal)) || !claims.HasRole(auth.HriInternal) {
		msg := fmt.Sprintf(auth.MsgInternalRoleRequired, "failed")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	return sendFail(requestId, request, logger, writer)
}

func SendFailNoAuth(requestId string, request *model.FailRequest,
	_ auth.HriAzClaims,
	writer kafka.Writer,
) (int, interface{}) {

	prefix := "batches/FailNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Fail (No Auth)")

	return sendFail(requestId, request, logger, writer)
}

func sendFail(requestId string, request *model.FailRequest,
	logger logrus.FieldLogger,
	writer kafka.Writer) (int, interface{}) {

	batch_metaData, err := getBatchMetaData(requestId, request.TenantId, request.BatchId, logger)
	if err != nil {
		return err.Code, response.NewErrorDetail(requestId, err.Body.ErrorDescription)
	}
	currentStatus, extractErr := ExtractBatchStatus(batch_metaData)
	if extractErr != nil {
		errMsg := fmt.Sprintf(msgGetByIdErr, extractErr)
		logger.Errorln(errMsg)
		return http.StatusInternalServerError, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, errMsg)
	}

	// Can't fail a batch if status is already 'terminated' or 'failed'

	if (status.Failed.String() != batch_metaData[param.Status]) && status.Terminated.String() != (batch_metaData[param.Status]) {
		updateRequest := getBatchFailUpdateRequest(request)

		errResp := updateBatchStatus(requestId, request.TenantId, request.BatchId, updateRequest, writer, currentStatus)
		if errResp != nil {
			return errResp.Code, errResp.Body
		}
	} else {
		// update resulted in no-op, due to previous batch status
		errMsg := fmt.Sprintf("'fail' failed, batch is in '%s' state", batch_metaData[param.Status].(string))
		logger.Errorln(errMsg)
		return http.StatusConflict, response.NewErrorDetail(requestId, errMsg)

	}

	return http.StatusOK, nil

}

func getBatchFailUpdateRequest(request *model.FailRequest) map[string]interface{} {
	currentTime := time.Now().UTC()

	updateRequest := bson.M{
		"$set": bson.M{
			"batch.$.status":             status.Failed.String(),
			"batch.$.invalidRecordCount": *request.InvalidRecordCount,
			"batch.$.actualRecordCount":  *request.ActualRecordCount,
			"batch.$.failureMessage":     request.FailureMessage,
			"batch.$.endDate":            currentTime,
		},
		"$inc": bson.M{"docs_deleted": 1},
	}

	return updateRequest
}
