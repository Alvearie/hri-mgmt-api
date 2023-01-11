/*
 * (C) Copyright IBM Corp. 2020
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

func ProcessingCompleteBatch(
	requestId string,
	request *model.ProcessingCompleteRequest,
	claims auth.HriAzClaims,
	writer kafka.Writer) (int, interface{}) {

	prefix := "batches/ProcessingComplete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Processing Complete")

	if !claims.HasRole(auth.HriInternal) || !claims.HasRole(auth.GetAuthRole(request.TenantId, auth.HriInternal)) {
		msg := fmt.Sprintf(auth.MsgInternalRoleRequired, "processingComplete")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	return processingStatusComplete(requestId, request, writer, logger)
}

func ProcessingCompleteBatchNoAuth(requestId string,
	request *model.ProcessingCompleteRequest,
	_ auth.HriAzClaims,
	writer kafka.Writer) (int, interface{}) {

	prefix := "batches/ProcessingCompleteNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Processing Complete (No Auth)")

	return processingStatusComplete(requestId, request, writer, logger)
}

func processingStatusComplete(requestId string,
	request *model.ProcessingCompleteRequest,
	writer kafka.Writer,
	logger logrus.FieldLogger) (int, interface{}) {

	// get the Current Batch Status --> Need current batch Status for potential "revert Status operation" in updateBatchStatus()
	// Note: this call will Always use the empty claims (NoAuth) option for calling getTenantByIdNoAuth()

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
	if batch_metaData[param.Status] == status.SendCompleted.String() {
		updateRequest := getProcessingCompleteUpdate(request)
		errResp := updateBatchStatus(requestId, request.TenantId, request.BatchId, updateRequest, writer, currentStatus)
		if errResp != nil {
			return errResp.Code, errResp.Body
		}
	} else {
		// update resulted in no-op, due to previous batch status
		errMsg := fmt.Sprintf("processingComplete failed, batch is in '%s' state", batch_metaData[param.Status].(string))
		logger.Errorln(errMsg)
		return http.StatusConflict, response.NewErrorDetail(requestId, errMsg)
	}

	return http.StatusOK, nil
}

func getProcessingCompleteUpdate(request *model.ProcessingCompleteRequest) map[string]interface{} {
	currentTime := time.Now().UTC()

	updateRequest := bson.M{
		"$set": bson.M{
			"batch.$.status":             status.Completed.String(),
			"batch.$.actualRecordCount":  request.ActualRecordCount,
			"batch.$.invalidRecordCount": request.InvalidRecordCount,
			"batch.$.endDate":            currentTime.Format(mongoApi.DateTimeFormat),
		},
		"$inc": bson.M{"docs_deleted": 1},
	}
	return updateRequest
}
