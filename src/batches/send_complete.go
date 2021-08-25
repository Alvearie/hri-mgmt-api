/*
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"ibm.com/watson/health/foundation/hri/batches/status"
	"ibm.com/watson/health/foundation/hri/common/auth"
	"ibm.com/watson/health/foundation/hri/common/elastic"
	"ibm.com/watson/health/foundation/hri/common/kafka"
	"ibm.com/watson/health/foundation/hri/common/logwrapper"
	"ibm.com/watson/health/foundation/hri/common/model"
	"ibm.com/watson/health/foundation/hri/common/param"
	"ibm.com/watson/health/foundation/hri/common/response"
	"net/http"
	"time"
)

func SendComplete(requestId string,
	request *model.SendCompleteRequest,
	claims auth.HriClaims,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/sendComplete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// Only Integrators can call sendComplete
	if !claims.HasScope(auth.HriIntegrator) {
		msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	//We know claims Must be Non-nil because the handler checks for that before we reach this point
	var claimSubj = claims.Subject
	return sendComplete(requestId, request, claimSubj, esClient, writer, logger, currentStatus)
}

func SendCompleteNoAuth(
	requestId string,
	request *model.SendCompleteRequest,
	_ auth.HriClaims,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/sendCompleteNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	//claims == nil --> NO Auth (Auth is NOT Enabled)
	var subject = auth.NoAuthFakeIntegrator

	return sendComplete(requestId, request, subject, esClient, writer, logger, currentStatus)
}

func sendComplete(
	requestId string,
	request *model.SendCompleteRequest,
	claimSubj string,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	logger logrus.FieldLogger,
	currentStatus status.BatchStatus) (int, interface{}) {

	updateRequest := getSendCompleteUpdateScript(request, claimSubj)

	origBatch, errResp := updateStatus(requestId, request.TenantId, request.BatchId,
		updateRequest, esClient, writer, currentStatus)
	if errResp != nil {
		return errResp.Code, errResp.Body
	}
	if origBatch != nil {
		if claimSubj != origBatch[param.IntegratorId] {
			// update resulted in no-op, due to insufficient permissions
			errMsg := fmt.Sprintf("sendComplete requested by '%s' but owned by '%s'", claimSubj,
				origBatch[param.IntegratorId])
			logger.Errorln(errMsg)
			return http.StatusUnauthorized, response.NewErrorDetail(requestId, errMsg)
		} else {
			// update resulted in no-op, due to previous batch status
			origBatchStatus := origBatch[param.Status].(string)
			return logNoUpdateToBatchStatus(origBatchStatus, logger, requestId)
		}
	}

	return http.StatusOK, nil
}

func getSendCompleteUpdateScript(request *model.SendCompleteRequest, claimSubj string) map[string]interface{} {
	var expectedRecordCount int
	if request.ExpectedRecordCount != nil {
		expectedRecordCount = *request.ExpectedRecordCount
	} else {
		expectedRecordCount = *request.RecordCount
	}

	var updateScript string
	if request.Validation {
		updateScript = getUpdateScriptWithValidation(request, claimSubj, expectedRecordCount)
	} else {
		updateScript = getUpdateScriptNoValidation(request, claimSubj, expectedRecordCount)
	}

	var updateRequest = map[string]interface{}{}
	if request.Metadata == nil {
		updateRequest = map[string]interface{}{
			"script": map[string]interface{}{
				"source": updateScript,
			},
		}
	} else {
		updateRequest = map[string]interface{}{
			"script": map[string]interface{}{
				"source": updateScript,
				"lang":   "painless",
				"params": map[string]interface{}{"metadata": request.Metadata},
			},
		}
	}

	return updateRequest
}

// When validation is not enabled:
//   - change the status to 'completed'
//   - set the record count
//   - set the end date
func getUpdateScriptNoValidation(request *model.SendCompleteRequest, subject string, expectedRecordCount int) string {
	currentTime := time.Now().UTC()

	var updateScript string

	// Can only transition if the batch is in the 'started' state
	if request.Metadata == nil {
		updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.expectedRecordCount = %d; ctx._source.endDate = '%s';} else {ctx.op = 'none'}",
			status.Started, subject, status.Completed, expectedRecordCount, currentTime.Format(elastic.DateTimeFormat))
	} else {
		updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.expectedRecordCount = %d; ctx._source.endDate = '%s'; ctx._source.metadata = params.metadata;} else {ctx.op = 'none'}",
			status.Started, subject, status.Completed, expectedRecordCount, currentTime.Format(elastic.DateTimeFormat))
	}
	return updateScript
}

// When validation is enabled
//   - change the status to 'sendCompleted'
//   - set the record count
func getUpdateScriptWithValidation(request *model.SendCompleteRequest, subject string, expectedRecCount int) string {
	var updateScript string

	if request.Metadata == nil {
		updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.expectedRecordCount = %d;} else {ctx.op = 'none'}",
			status.Started, subject, status.SendCompleted, expectedRecCount)
	} else {
		updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.expectedRecordCount = %d; ctx._source.metadata = params.metadata;} else {ctx.op = 'none'}",
			status.Started, subject, status.SendCompleted, expectedRecCount)
	}
	return updateScript
}

func logNoUpdateToBatchStatus(origBatchStatus string, logger logrus.FieldLogger, requestId string) (int, interface{}) {
	errMsg := fmt.Sprintf("sendComplete failed, batch is in '%s' state", origBatchStatus)
	logger.Errorln(errMsg)
	return http.StatusConflict, response.NewErrorDetail(requestId, errMsg)
}
