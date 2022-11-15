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
	"go.mongodb.org/mongo-driver/mongo"
)

func SendStatusComplete(requestId string,
	request *model.SendCompleteRequest,
	claims auth.HriAzClaims,
	mongoClient *mongo.Collection,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {
	prefix := "batches/sendComplete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// Only Integrators can call sendComplete
	if !claims.HasRole(auth.HriIntegrator) {
		msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	//We know claims Must be Non-nil because the handler checks for that before we reach this point
	var claimSubj = claims.Subject

	return sendStatusComplete(requestId, request, claimSubj, mongoClient, writer, logger, currentStatus)
}

func SendStatusCompleteNoAuth(
	requestId string,
	request *model.SendCompleteRequest,
	_ auth.HriAzClaims,
	client *mongo.Collection,
	writer kafka.Writer,
	currentStatus status.BatchStatus) (int, interface{}) {
	prefix := "batches/sendStatusCompleteNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	//claims == nil --> NO Auth (Auth is NOT Enabled)
	var subject = auth.NoAuthFakeIntegrator

	return sendStatusComplete(requestId, request, subject, client, writer, logger, currentStatus)
}

func sendStatusComplete(
	requestId string,
	request *model.SendCompleteRequest,
	claimSubj string,
	client *mongo.Collection,
	kafkaWriter kafka.Writer,
	logger logrus.FieldLogger,
	currentStatus status.BatchStatus) (int, interface{}) {

	batch_metaData, err := getBatchMetaData(requestId, request.TenantId, request.BatchId, client, logger)
	if err != nil {
		return err.Code, response.NewErrorDetail(requestId, err.Body.ErrorDescription)
	}
	// Can only transition if the batch is in the 'started' state
	if batch_metaData[param.Status] == status.Started.String() && batch_metaData[param.IntegratorId] == claimSubj {
		updateRequest := getSendCompleteUpdateRequest(request, claimSubj, requestId)

		errResp := updateBatchStatus(requestId, request.TenantId, request.BatchId,
			updateRequest, client, kafkaWriter, currentStatus)
		if errResp != nil {
			return errResp.Code, errResp.Body
		}
	} else if claimSubj != batch_metaData[param.IntegratorId] {
		// update resulted in no-op, due to insufficient permissions
		errMsg := fmt.Sprintf("sendComplete requested by '%s' but owned by '%s'", claimSubj,
			batch_metaData[param.IntegratorId].(string))
		logger.Errorln(errMsg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, errMsg)
	} else {
		// update resulted in no-op, due to previous batch status
		origBatchStatus := batch_metaData[param.Status].(string)
		return logNoUpdateToBatchStatus(origBatchStatus, logger, requestId)
	}

	return http.StatusOK, nil
}
func getSendCompleteUpdateRequest(request *model.SendCompleteRequest, claimSubj string, requestId string) map[string]interface{} {
	var expectedRecordCount int
	if request.ExpectedRecordCount != nil {
		expectedRecordCount = *request.ExpectedRecordCount
	} else {
		expectedRecordCount = *request.RecordCount
	}

	var updateRequest map[string]interface{}
	// When validation is enabled
	//   - change the status to 'sendCompleted'
	//   - set the record count
	if request.Validation {
		if request.Metadata == nil {

			updateRequest = bson.M{
				"$set": bson.M{
					"batch.$.status":              status.SendCompleted.String(),
					"batch.$.expectedRecordCount": expectedRecordCount,
				},
				"$inc": bson.M{"docs_deleted": 1},
			}

		} else {

			updateRequest = bson.M{
				"$set": bson.M{
					"batch.$.metadata":            request.Metadata,
					"batch.$.status":              status.SendCompleted.String(),
					"batch.$.expectedRecordCount": expectedRecordCount,
				},
				"$inc": bson.M{"docs_deleted": 1},
			}
		}
	} else {
		// When validation is not enabled:
		//   - change the status to 'completed'
		//   - set the record count
		//   - set the end date
		currentTime := time.Now().UTC()

		if request.Metadata == nil {

			updateRequest = bson.M{
				"$set": bson.M{
					"batch.$.status":              status.Completed.String(),
					"batch.$.expectedRecordCount": expectedRecordCount,
					"batch.$.endDate":             currentTime,
				},
				"$inc": bson.M{"docs_deleted": 1},
			}
		} else {

			updateRequest = bson.M{
				"$set": bson.M{
					"batch.$.metadata":            request.Metadata,
					"batch.$.status":              status.Completed.String(),
					"batch.$.expectedRecordCount": expectedRecordCount,
					"batch.$.endDate":             currentTime,
				},
				"$inc": bson.M{"docs_deleted": 1},
			}

		}
	}
	return updateRequest

}

func getBatchMetaData(requestId string, tenantId string, batchId string, mongoClient *mongo.Collection, logger logrus.FieldLogger) (map[string]interface{}, *response.ErrorDetailResponse) {
	//Always use the empty claims (NoAuth) option
	var claims = auth.HriAzClaims{}
	getBatchRequest := model.GetByIdBatch{
		TenantId: tenantId,
		BatchId:  batchId,
	}
	getByIdCode, batch := GetByBatchIdNoAuth(requestId, getBatchRequest, claims, mongoClient)
	if getByIdCode != http.StatusOK { //error getting current Batch Info
		newErrMsg := fmt.Sprintf(msgGetByIdErr, " Error getting current Batch Info")
		logger.Errorln(newErrMsg)
		return nil, response.NewErrorDetailResponse(getByIdCode, requestId, newErrMsg)
	}
	batchMap := batch.(map[string]interface{})
	return batchMap, nil
}
func logNoUpdateToBatchStatus(origBatchStatus string, logger logrus.FieldLogger, requestId string) (int, interface{}) {
	errMsg := fmt.Sprintf("sendComplete failed, batch is in '%s' state", origBatchStatus)
	logger.Errorln(errMsg)
	return http.StatusConflict, response.NewErrorDetail(requestId, errMsg)
}
