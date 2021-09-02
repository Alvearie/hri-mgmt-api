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

func Fail(
	requestId string,
	request *model.FailRequest,
	claims auth.HriClaims,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/fail"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Fail")

	// Only Integrators can call fail
	if !claims.HasScope(auth.HriInternal) {
		msg := fmt.Sprintf(auth.MsgInternalRoleRequired, "failed")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	return fail(requestId, request, logger, esClient, writer, currentStatus)
}

func FailNoAuth(requestId string, request *model.FailRequest,
	_ auth.HriClaims,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/FailNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Fail (No Auth)")

	return fail(requestId, request, logger, esClient, writer, currentStatus)
}

func fail(requestId string, request *model.FailRequest,
	logger logrus.FieldLogger,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	updateRequest := getFailUpdateScript(request)

	origBatch, errResp := updateStatus(requestId, request.TenantId, request.BatchId, updateRequest, esClient, writer, currentStatus)
	if errResp != nil {
		return errResp.Code, errResp.Body
	}

	if origBatch != nil {
		// update resulted in no-op, due to previous batch status
		errMsg := fmt.Sprintf("'fail' failed, batch is in '%s' state", origBatch[param.Status].(string))
		logger.Errorln(errMsg)
		return http.StatusConflict, response.NewErrorDetail(requestId, errMsg)
	}

	return http.StatusOK, nil
}

func getFailUpdateScript(request *model.FailRequest) map[string]interface{} {
	currentTime := time.Now().UTC()

	// Can't fail a batch if status is already 'terminated' or 'failed'
	updateScript := fmt.Sprintf("if (ctx._source.status != '%s' && ctx._source.status != '%s') {ctx._source.status = '%s'; ctx._source.actualRecordCount = %d; ctx._source.invalidRecordCount = %d; ctx._source.failureMessage = '%s'; ctx._source.endDate = '%s';} else {ctx.op = 'none'}",
		status.Terminated, status.Failed, status.Failed, *request.ActualRecordCount, *request.InvalidRecordCount, request.FailureMessage, currentTime.Format(elastic.DateTimeFormat))

	updateRequest := map[string]interface{}{
		"script": map[string]interface{}{
			"source": updateScript,
		},
	}

	return updateRequest
}
