/*
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func ProcessingComplete(
	requestId string,
	request *model.ProcessingCompleteRequest,
	claims auth.HriClaims,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/ProcessingComplete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Processing Complete")

	if !claims.HasScope(auth.HriInternal) {
		msg := fmt.Sprintf(auth.MsgInternalRoleRequired, "processingComplete")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	return processingComplete(requestId, request, esClient, writer, logger, currentStatus)
}

func ProcessingCompleteNoAuth(requestId string,
	request *model.ProcessingCompleteRequest,
	_ auth.HriClaims, esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/ProcessingCompleteNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Processing Complete (No Auth)")

	return processingComplete(requestId, request, esClient, writer, logger, currentStatus)
}

func processingComplete(requestId string,
	request *model.ProcessingCompleteRequest,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	logger logrus.FieldLogger,
	currentStatus status.BatchStatus) (int, interface{}) {

	updateRequest := getProcessingCompleteUpdateScript(request)

	origBatch, errResp := updateStatus(requestId, request.TenantId, request.BatchId, updateRequest, esClient, writer, currentStatus)
	if errResp != nil {
		return errResp.Code, errResp.Body
	}
	if origBatch != nil {
		// update resulted in no-op, due to previous batch status
		errMsg := fmt.Sprintf("processingComplete failed, batch is in '%s' state", origBatch[param.Status].(string))
		logger.Errorln(errMsg)
		return http.StatusConflict, response.NewErrorDetail(requestId, errMsg)
	}

	return http.StatusOK, nil
}

func getProcessingCompleteUpdateScript(request *model.ProcessingCompleteRequest) map[string]interface{} {
	currentTime := time.Now().UTC()

	// Can only transition if the batch is in the 'sendCompleted' state
	updateScript := fmt.Sprintf("if (ctx._source.status == '%s') {ctx._source.status = '%s'; ctx._source.actualRecordCount = %d; ctx._source.invalidRecordCount = %d; ctx._source.endDate = '%s';} else {ctx.op = 'none'}",
		status.SendCompleted, status.Completed, *request.ActualRecordCount, *request.InvalidRecordCount, currentTime.Format(elastic.DateTimeFormat))

	updateRequest := map[string]interface{}{
		"script": map[string]interface{}{
			"source": updateScript,
		},
	}
	return updateRequest
}
