/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"context"
	"fmt"
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
)

const (
	elasticResultKey           string = "result"
	elasticResultUpdated       string = "updated"
	elasticResultNoop          string = "noop"
	msgUpdateResultNotReturned string = "Update result not returned in Elastic response"
)

type StatusUpdater interface {
	GetAction() string

	// Return an error if the request is not authorized to perform this action
	CheckAuth(auth.HriClaims) error

	// NOTE: whenever the Elastic document is NOT updated, set the ctx.op = 'none'
	// flag. Elastic will use this flag in the response so we can check if the update took place.
	GetUpdateScript(map[string]interface{}, param.Validator, auth.HriClaims, *log.Logger) (map[string]interface{}, map[string]interface{})
}

func UpdateStatus(
	params map[string]interface{},
	validator param.Validator,
	claims auth.HriClaims,
	statusUpdater StatusUpdater,
	client *elasticsearch.Client,
	kafkaWriter kafka.Writer) map[string]interface{} {

	logger := log.New(os.Stdout, fmt.Sprintf("batches/%s: ", statusUpdater.GetAction()), log.Llongfile)

	if err := statusUpdater.CheckAuth(claims); err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusUnauthorized, err.Error())
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

	index := elastic.IndexFromTenantId(tenantId)

	updateRequest, errResp := statusUpdater.GetUpdateScript(params, validator, claims, logger)
	if errResp != nil {
		return errResp
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

	decodedUpdateResponse, elasticErr := elastic.DecodeBody(updateResponse, updateErr)
	if elasticErr != nil {
		return elasticErr.LogAndBuildApiResponse(logger,
			fmt.Sprintf("Could not update the status of batch %s", batchId))
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
		updatedBatch = NormalizeBatchRecordCountValues(updatedBatch)
		err = kafkaWriter.Write(notificationTopic, batchId, updatedBatch)
		if err != nil {
			logger.Println(err.Error())
			return response.Error(http.StatusInternalServerError, err.Error())
		}
		return response.Success(http.StatusOK, map[string]interface{}{})
	} else if updateResult == elasticResultNoop {
		statusCode, errMsg := determineCause(statusUpdater, claims.Subject, updatedBatch)
		logger.Println(errMsg)
		return response.Error(statusCode, errMsg)
	} else {
		errMsg := fmt.Sprintf("An unexpected error occurred updating the batch, Elastic update returned result '%s'", updateResult)
		logger.Println(errMsg)
		return response.Error(http.StatusInternalServerError, errMsg)
	}
}

func determineCause(statusUpdater StatusUpdater, subject string, batch map[string]interface{}) (int, string) {
	// The OAuth subject has to match the batch's integratorId only for the sendComplete and terminate endpoints.
	if subject != batch[param.IntegratorId] && statusUpdater.GetAction() != ProcessingCompleteAction && statusUpdater.GetAction() != FailAction {
		// update resulted in no-op, due to insufficient permissions
		errMsg := fmt.Sprintf("Batch status was not updated to '%s'. Requested by '%s' but owned by '%s'", statusUpdater.GetAction(), subject, batch[param.IntegratorId])
		return http.StatusUnauthorized, errMsg
	} else {
		// update resulted in no-op, due to previous batch status
		errMsg := fmt.Sprintf("The '%s' endpoint failed, batch is in '%s' state", statusUpdater.GetAction(), batch[param.Status].(string))
		return http.StatusConflict, errMsg
	}
}
