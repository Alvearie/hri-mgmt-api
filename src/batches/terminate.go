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

func Terminate(
	requestId string,
	request *model.TerminateRequest,
	claims auth.HriClaims,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/terminate"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Terminate")

	// Only Integrators can call terminate
	if !claims.HasScope(auth.HriIntegrator) {
		msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "terminate")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	var subject = claims.Subject
	return terminate(requestId, request, subject, logger, esClient, writer, currentStatus)
}

func TerminateNoAuth(
	requestId string,
	request *model.TerminateRequest,
	_ auth.HriClaims,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	prefix := "batches/TerminateNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Terminate (No Auth)")

	var subject = auth.NoAuthFakeIntegrator
	return terminate(requestId, request, subject, logger, esClient, writer, currentStatus)
}

func terminate(requestId string, request *model.TerminateRequest,
	claimsSubject string,
	logger logrus.FieldLogger,
	esClient *elasticsearch.Client,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {

	updateRequest := getTerminateUpdateScript(request, claimsSubject)

	origBatch, errResp := updateStatus(requestId, request.TenantId, request.BatchId, updateRequest, esClient, writer, currentStatus)
	if errResp != nil {
		return errResp.Code, errResp.Body
	}
	if origBatch != nil {
		if claimsSubject != origBatch[param.IntegratorId] {
			// update resulted in no-op, due to insufficient permissions
			errMsg := fmt.Sprintf("terminate requested by '%s' but owned by '%s'", claimsSubject, origBatch[param.IntegratorId])
			logger.Errorln(errMsg)
			return http.StatusUnauthorized, response.NewErrorDetail(requestId, errMsg)
		} else {
			// update resulted in no-op, due to previous batch status
			errMsg := fmt.Sprintf("terminate failed, batch is in '%s' state", origBatch[param.Status].(string))
			logger.Errorln(errMsg)
			return http.StatusConflict, response.NewErrorDetail(requestId, errMsg)
		}
	}

	return http.StatusOK, nil
}

func getTerminateUpdateScript(request *model.TerminateRequest, claimsSubject string) map[string]interface{} {
	currentTime := time.Now().UTC()

	// Can only transition if the batch is in the 'started' state
	if request.Metadata == nil {
		updateScript := fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.endDate = '%s';} else {ctx.op = 'none'}",
			status.Started, claimsSubject, status.Terminated, currentTime.Format(elastic.DateTimeFormat))

		updateRequest := map[string]interface{}{
			"script": map[string]interface{}{
				"source": updateScript,
			},
		}
		return updateRequest
	} else {
		updateScript := fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.endDate = '%s'; ctx._source.metadata = params.metadata;} else {ctx.op = 'none'}",
			status.Started, claimsSubject, status.Terminated, currentTime.Format(elastic.DateTimeFormat))

		updateRequest := map[string]interface{}{
			"script": map[string]interface{}{
				"source": updateScript,
				"lang":   "painless",
				"params": map[string]interface{}{"metadata": request.Metadata},
			},
		}
		return updateRequest
	}
}
