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
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func SendStatusComplete(requestId string,
	request *model.SendCompleteRequest,
	claims auth.HriAzClaims,
	writer kafka.Writer, currentStatus status.BatchStatus, integratorId string) (int, interface{}) {
	prefix := "batches/sendComplete"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	// Only Integrators can call sendComplete
	if !claims.HasRole(auth.HriIntegrator) || !claims.HasRole(auth.GetAuthRole(request.TenantId, auth.HriIntegrator)) {
		msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "initiate sendComplete on")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	//We know claims Must be Non-nil because the handler checks for that before we reach this point
	var claimSubj = claims.Subject

	return sendStatusComplete(requestId, request, claimSubj, writer, logger, currentStatus, integratorId)
}

func SendStatusCompleteNoAuth(
	requestId string,
	request *model.SendCompleteRequest,
	_ auth.HriAzClaims,
	writer kafka.Writer, currentStatus status.BatchStatus, integratorId string) (int, interface{}) {
	prefix := "batches/sendStatusCompleteNoAuth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)

	//claims == nil --> NO Auth (Auth is NOT Enabled)
	var subject = auth.NoAuthFakeIntegrator

	return sendStatusComplete(requestId, request, subject, writer, logger, currentStatus, integratorId)
}

func sendStatusComplete(
	requestId string,
	request *model.SendCompleteRequest,
	claimSubj string,
	kafkaWriter kafka.Writer,
	logger logrus.FieldLogger, currentStatus status.BatchStatus, integratorId string) (int, interface{}) {

	// Can only transition if the batch is in the 'started' state
	if currentStatus == status.Started && integratorId == claimSubj {
		updateRequest := getSendCompleteUpdateRequest(request, claimSubj, requestId)

		errResp := updateBatchStatus(requestId, request.TenantId, request.BatchId,
			updateRequest, kafkaWriter, currentStatus)
		if errResp != nil {
			return errResp.Code, errResp.Body
		}
	} else if claimSubj != integratorId {
		// update resulted in no-op, due to insufficient permissions
		errMsg := fmt.Sprintf("sendComplete requested by '%s' but owned by '%s'", claimSubj,
			integratorId)
		logger.Errorln(errMsg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, errMsg)
	} else {
		// update resulted in no-op, due to previous batch status
		origBatchStatus := currentStatus.String()
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

func logNoUpdateToBatchStatus(origBatchStatus string, logger logrus.FieldLogger, requestId string) (int, interface{}) {
	errMsg := fmt.Sprintf("sendComplete failed, batch is in '%s' state", origBatchStatus)
	logger.Errorln(errMsg)
	return http.StatusConflict, response.NewErrorDetail(requestId, errMsg)
}
