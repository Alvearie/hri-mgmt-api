/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"context"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"
)

const (
	elasticResultKey           string = "result"
	elasticResultUpdated       string = "updated"
	elasticResultNoop          string = "noop"
	msgUpdateResultNotReturned string = "Update result not returned in Elastic response"
)

func UpdateStatus(
	params map[string]interface{},
	validator param.Validator,
	claims auth.HriClaims,
	targetStatus status.BatchStatus,
	client *elasticsearch.Client,
	kafkaWriter kafka.Writer) map[string]interface{} {

	logger := log.New(os.Stdout, fmt.Sprintf("batches/%s: ", targetStatus), log.Llongfile)

	// validate that caller has sufficient permissions
	if !claims.HasScope(auth.HriIntegrator) {
		msg := fmt.Sprintf(fmt.Sprintf(auth.MsgIntegratorRoleRequired, "update"))
		logger.Printf(msg)
		return response.Error(http.StatusUnauthorized, msg)
	}

	// validate that required input params are present
	tenantId, err := path.ExtractParam(params, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}
	batchId, err := path.ExtractParam(params, param.BatchIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	errResp := validator.ValidateOptional(
		params,
		param.Info{param.Metadata, reflect.Map},
	)
	if errResp != nil {
		logger.Printf("Bad input optional params: %s", errResp)
		return errResp
	}
	metadata := params[param.Metadata]

	index := elastic.IndexFromTenantId(tenantId)

	// Elastic conditional update query
	var updateScript string
	currentTime := time.Now().UTC()
	if targetStatus == status.Completed {
		// recordCount is required for Completed
		errResp := validator.Validate(
			params,
			// golang receives numeric JSON values as Float64
			param.Info{param.RecordCount, reflect.Float64},
		)
		if errResp != nil {
			logger.Printf("Bad input params: %s", errResp)
			return errResp
		}
		recordCount := int(params[param.RecordCount].(float64))

		// NOTE: whenever the Elastic document is NOT updated, set the ctx.op = 'none'
		// flag. Elastic will use this flag in the response so we can check if the update took place.
		if metadata == nil {
			updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.recordCount = %d; ctx._source.endDate = '%s';} else {ctx.op = 'none'}",
				status.Started, claims.Subject, targetStatus, recordCount, currentTime.Format(elastic.DateTimeFormat))
		} else {
			updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.recordCount = %d; ctx._source.endDate = '%s'; ctx._source.metadata = params.metadata;} else {ctx.op = 'none'}",
				status.Started, claims.Subject, targetStatus, recordCount, currentTime.Format(elastic.DateTimeFormat))
		}

	} else if targetStatus == status.Terminated {

		if metadata == nil {
			updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.endDate = '%s';} else {ctx.op = 'none'}",
				status.Started, claims.Subject, targetStatus, currentTime.Format(elastic.DateTimeFormat))
		} else {
			updateScript = fmt.Sprintf("if (ctx._source.status == '%s' && ctx._source.integratorId == '%s') {ctx._source.status = '%s'; ctx._source.endDate = '%s'; ctx._source.metadata = params.metadata;} else {ctx.op = 'none'}",
				status.Started, claims.Subject, targetStatus, currentTime.Format(elastic.DateTimeFormat))
		}

	} else {
		// this method was somehow invoked with an invalid batch status
		errMsg := fmt.Sprintf("Cannot update batch to status '%s', only '%s' and '%s' are acceptable", targetStatus, status.Completed, status.Terminated)
		logger.Println(errMsg)
		return response.Error(http.StatusUnprocessableEntity, errMsg)
	}

	var updateRequest map[string]interface{}
	if metadata == nil {
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
				"params": map[string]interface{}{"metadata": metadata},
			},
		}
	}

	encodedQuery, err := elastic.EncodeQueryBody(updateRequest)
	if err != nil {
		errMsg := fmt.Sprintf("Error encoding Elastic query: %s", err.Error())
		logger.Println(errMsg)
		return response.Error(http.StatusInternalServerError, errMsg)
	}

	updateResponse, updateErr := client.Update(
		index,
		batchId,
		encodedQuery, // request body
		client.Update.WithContext(context.Background()),
		client.Update.WithSource("true"), // return updated batch in response
	)

	decodedUpdateResponse, elasticResponseErr := elastic.DecodeBody(updateResponse, updateErr, tenantId, logger)
	if elasticResponseErr != nil {
		return elasticResponseErr
	}
	// read elastic response and verify the batch was updated
	updateResult, hasUpdateResult := decodedUpdateResponse[elasticResultKey].(string)
	if !hasUpdateResult {
		logger.Println(msgUpdateResultNotReturned)
		return response.Error(http.StatusInternalServerError, msgUpdateResultNotReturned)
	}
	updatedBatch, err := param.ExtractValues(decodedUpdateResponse, "get", "_source")
	if err != nil {
		errMsg := fmt.Sprintf("Updated document not returned in Elastic response: %s", err.Error())
		logger.Println(errMsg)
		return response.Error(http.StatusInternalServerError, errMsg)
	}

	if updateResult == elasticResultUpdated {
		// successful update; publish update notification to Kafka
		updatedBatch[param.BatchId] = batchId
		notificationTopic := InputTopicToNotificationTopic(updatedBatch[param.Topic].(string))
		err = kafkaWriter.Write(notificationTopic, batchId, updatedBatch)
		if err != nil {
			logger.Println(err.Error())
			return response.Error(http.StatusInternalServerError, err.Error())
		}
		return response.Success(http.StatusOK, map[string]interface{}{})
	} else if updateResult == elasticResultNoop {
		statusCode, errMsg := determineCause(targetStatus, claims.Subject, updatedBatch)
		logger.Println(errMsg)
		return response.Error(statusCode, errMsg)
	} else {
		errMsg := fmt.Sprintf("An unexpected error occurred updating the batch, Elastic update returned result '%s'", updateResult)
		logger.Println(errMsg)
		return response.Error(http.StatusInternalServerError, errMsg)
	}
}

func determineCause(targetStatus status.BatchStatus, subject string, batch map[string]interface{}) (int, string) {
	if subject != batch[param.IntegratorId] {
		// update resulted in no-op, due to insuffient permissions
		return http.StatusUnauthorized, fmt.Sprintf(
			"Batch status was not updated to '%s'. Requested by '%s' but owned by '%s'",
			targetStatus,
			subject,
			batch[param.IntegratorId],
		)
	} else {
		// update resulted in no-op, due to previous batch status
		return http.StatusConflict, fmt.Sprintf(
			"Batch status was not updated to '%s', batch is already in '%s' state",
			targetStatus,
			batch[param.Status],
		)
	}
}
