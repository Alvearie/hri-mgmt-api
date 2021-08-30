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
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	elasticResultKey           string = "result"
	elasticResultUpdated       string = "updated"
	elasticResultNoop          string = "noop"
	msgUpdateResultNotReturned string = "update result not returned in Elastic response"
)

// Attempts to run the updateRequest on the specified batch
// On success return (nil, nil)
// If the update results in a 'noop', the original batch is returned: (batch, nil)
// On error returns (nil, error)
func updateStatus(requestId string,
	tenantId string,
	batchId string,
	updateRequest map[string]interface{},
	client *elasticsearch.Client,
	kafkaWriter kafka.Writer,
	currentStatus status.BatchStatus) (map[string]interface{}, *response.ErrorDetailResponse) {

	prefix := "batches/updateStatus"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Update Status")

	index := elastic.IndexFromTenantId(tenantId)

	encodedQuery, err := elastic.EncodeQueryBody(updateRequest)
	if err != nil {
		msg := fmt.Sprintf("error encoding Elastic query: %s", err.Error())
		logger.Errorln(msg)
		return nil, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msg)
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
		resp := elasticErr.LogAndBuildErrorDetail(requestId, logger,
			fmt.Sprintf("could not update the status of batch %s", batchId))
		return nil, &response.ErrorDetailResponse{Code: elasticErr.Code, Body: resp}
	}

	// read elastic response and verify the batch was updated
	updateResult, hasUpdateResult := decodedUpdateResponse[elasticResultKey].(string)
	if !hasUpdateResult {
		logger.Errorln(msgUpdateResultNotReturned)
		return nil, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msgUpdateResultNotReturned)
	}
	updatedBatch, err := param.ExtractValues(decodedUpdateResponse, "get", "_source")
	if err != nil {
		msg := fmt.Sprintf("updated document not returned in Elastic response: %s", err.Error())
		logger.Errorln(msg)
		return nil, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msg)
	}

	if updateResult == elasticResultUpdated {
		// successful update; publish update notification to Kafka
		updatedBatch[param.BatchId] = batchId
		notificationTopic := InputTopicToNotificationTopic(updatedBatch[param.Topic].(string))
		updatedBatch = NormalizeBatchRecordCountValues(updatedBatch)
		err = kafkaWriter.Write(notificationTopic, batchId, updatedBatch)
		if err != nil { //Write to Elastic Failed, try to Revert Batch Status
			kafkaErrMsg := fmt.Sprintf("error writing batch notification to kafka: %s", err.Error())
			logger.Errorln(kafkaErrMsg)

			encodeErr := revertBatchStatus(requestId, index, batchId, client, currentStatus, logger)
			if encodeErr != nil {
				return nil, encodeErr
			}

			return nil, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, kafkaErrMsg)
		}
		return nil, nil
	} else if updateResult == elasticResultNoop {
		return updatedBatch, nil
	} else {
		msg := fmt.Sprintf("an unexpected error occurred updating the batch, Elastic update returned result '%s'", updateResult)
		logger.Errorln(msg)
		return nil, response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msg)
	}
}

//Here we are reverting the Batch status to "currentStatus" in Elastic. "currentStatus" is the
//   status that the Batch had BEFORE the update operation
// If the Revert attempt in Elastic fails, we retry up to 5 times.
func revertBatchStatus(requestId string, idx string,
	batchId string,
	client *elasticsearch.Client,
	currentStatus status.BatchStatus,
	logger logrus.FieldLogger) *response.ErrorDetailResponse {

	var revertStatusScript = fmt.Sprintf("{ctx._source.status = '%s';}", currentStatus)
	var revertStatusRequest = map[string]interface{}{
		"script": map[string]interface{}{
			"source": revertStatusScript}}

	var revertErrMsg = "(Attempt # %d) Error Reverting batch Status back to %s; elasticResponseCode: %d, Elastic error: %s"
	var attemptNum = 1
	for attemptNum < 7 { //Retry Up to 5 Times (Total # attempts => 6)
		encodedQuery, err := elastic.EncodeQueryBody(revertStatusRequest)
		if err != nil {
			msg := fmt.Sprintf("error encoding (Revert Status) Elastic query: %s", err.Error())
			logger.Errorln(msg)
			logger.Debug(encodedQuery)
			return response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msg)
		}

		revertResponse, revertErr := client.Update(
			idx,
			batchId,
			encodedQuery, // Revert Status request body
			client.Update.WithContext(context.Background()),
		)

		decodedRevertResponse, elasticErr := elastic.DecodeBody(revertResponse, revertErr)

		if elasticErr != nil {
			msg := fmt.Sprintf(revertErrMsg, attemptNum, currentStatus,
				revertResponse.StatusCode, elasticErr.Error())
			logger.Errorln(msg)
			attemptNum += 1
		} else {
			var debugMsg = fmt.Sprintf("Revert batch Status back to %s succeeded: %s",
				currentStatus, decodedRevertResponse)
			logger.Debugln(debugMsg)
			return nil
		}
	}

	return nil
}
