package batches

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

// Attempts to run the updateRequest on the specified batch
// On success return nil
// On error returns error
func updateBatchStatus(requestId string,
	tenantId string,
	batchId string,
	updateRequest map[string]interface{},

	kafkaWriter kafka.Writer,
	currentStatus status.BatchStatus) *response.ErrorDetailResponse {
	prefix := "batches/updateStatus"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Update Status")
	//appending "-batches"
	tenant_id := mongoApi.GetTenantWithBatchesSuffix(tenantId)

	filter := bson.D{
		{Key: "tenantId", Value: tenant_id},
		{Key: "batch.id", Value: batchId},
	}

	updateResponse, updateErr := mongoApi.HriCollection.UpdateMany(
		context.Background(),
		filter,
		updateRequest, // request body

	)

	if updateErr != nil {
		msg := fmt.Sprintf("could not update the status of batch %s", batchId)
		logger.Errorln(msg)
		return response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msg)

	}

	batchMap, errResp := getBatchMetaData(requestId, tenantId, batchId, logger)
	if errResp != nil {
		msg := fmt.Sprintf("updated document not returned in Cosmos response: %s", errResp.Body.ErrorDescription)
		logger.Errorln(msg)
		return response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msg)
	}

	notificationTopic := InputTopicToNotificationTopic(batchMap[param.Topic].(string))
	updatedBatch := NormalizeBatchRecordCountValues(batchMap)
	// // read cosmos response and verify the batch was updated: checking for updated data
	if updateResponse.ModifiedCount == 1 {
		// successful update; publish update notification to Kafka

		err := kafkaWriter.Write(notificationTopic, batchId, updatedBatch)

		if err != nil { //Write to Kafka Failed, try to Revert Batch Status
			kafkaErrMsg := fmt.Sprintf("error writing batch notification to kafka: %s", err.Error())
			logger.Errorln(kafkaErrMsg)
			revertErr := revertStatus(requestId, tenantId, batchId, currentStatus, logger)
			if revertErr != nil {
				return revertErr
			}
			return response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, kafkaErrMsg)
		}

		return nil

	} else {

		msg := fmt.Sprintf("an unexpected error occurred updating the batch, Cosmos update returned result '%s'", updatedBatch)
		logger.Errorln(msg)
		return response.NewErrorDetailResponse(http.StatusInternalServerError, requestId, msg)
	}

}

// Here we are reverting the Batch status to "currentStatus" in Cosmos. "currentStatus" is the
// status that the Batch had BEFORE the update operation//
// If the Revert attempt in Cosmos DB fails, we retry up to 5 times.
// TODO:enhacement {Reduce/eliminate retries}
func revertStatus(requestId string,
	tenantId string,
	batchId string,
	currentStatus status.BatchStatus,
	logger logrus.FieldLogger) *response.ErrorDetailResponse {

	tenant_id := mongoApi.GetTenantWithBatchesSuffix(tenantId)
	var revertErrMsg = "(Attempt # %d) Error Reverting batch Status back to %s; CosmosResponseCode: %d, Cosmos error: %s"
	var attemptNum = 1

	updateRequest := bson.M{
		"$set": bson.M{
			"batch.$.status": status.Started.String(),
		},
	}

	filter := bson.D{
		{Key: "tenantId", Value: tenant_id},
		{Key: "batch.id", Value: batchId},
	}

	for attemptNum < 7 { //Retry Up to 5 Times (Total # attempts => 6)

		updateResponse, updateErr := mongoApi.HriCollection.UpdateOne(
			context.Background(),
			filter,
			updateRequest,
		)

		if updateErr != nil || updateResponse.ModifiedCount == 0 {
			msg := fmt.Sprintf(revertErrMsg, attemptNum, currentStatus, updateErr.Error())
			logger.Errorln(msg)
			fmt.Println(msg)

			attemptNum += 1
		} else if updateResponse.ModifiedCount == 1 {
			batch, _ := getBatchMetaData(requestId, tenantId, batchId, logger)
			var debugMsg = fmt.Sprintf("Revert batch Status back to %s succeeded: %s",
				currentStatus, batch)
			logger.Debugln(debugMsg)
			return nil
		}
	}

	return nil
}
