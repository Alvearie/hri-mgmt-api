/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func buildBatchInfo(batch model.CreateBatch, integrator string) map[string]interface{} {
	currentTime := time.Now().UTC()

	info := map[string]interface{}{
		param.Name:             batch.Name,
		param.IntegratorId:     integrator,
		param.Topic:            batch.Topic,
		param.DataType:         batch.DataType,
		param.Status:           status.Started.String(),
		param.StartDate:        currentTime.Format(mongoApi.DateTimeFormat),
		param.InvalidThreshold: batch.InvalidThreshold,
	}

	if batch.InvalidThreshold == 0 {
		info[param.InvalidThreshold] = -1
	}

	if batch.Metadata != nil {
		info[param.Metadata] = batch.Metadata
	}

	return info
}

func UpdateBatchInfo(batch model.CreateBatch, integrator string, batchId string) model.CreateBatchBson {
	currentTime := time.Now().UTC()
	batchBson := model.CreateBatchBson{}
	batchBson.BatchId = batchId
	batchBson.DataType = batch.DataType
	batchBson.Metadata = batch.Metadata
	batchBson.Name = batch.Name
	batchBson.Topic = batch.Topic
	batchBson.IntegratorId = integrator
	batchBson.Status = status.Started.String()
	batchBson.StartDate = currentTime.Format(mongoApi.DateTimeFormat)
	batchBson.BatchId = batchId
	batchBson.InvalidThreshold = batch.InvalidThreshold
	if batch.InvalidThreshold == 0 {
		batchBson.InvalidThreshold = -1
	}

	return batchBson
}
func CreateBatchNoAuth(
	requestId string,
	batch model.CreateBatch,
	_ auth.HriAzClaims,
	kafkaWriter kafka.Writer) (int, interface{}) {

	var prefix = "batches/create_no_auth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Create (Without Auth)")

	var integratorId = auth.NoAuthFakeIntegrator
	return createBatch(requestId, batch, integratorId, kafkaWriter, logger)
}
func CreateBatch(
	requestId string,
	batch model.CreateBatch,
	claims auth.HriAzClaims,
	kafkaWriter kafka.Writer) (int, interface{}) {

	prefix := "batches/create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Create")

	// validate that caller has sufficient permissions
	if !claims.HasRole(auth.HriIntegrator) || !claims.HasRole(auth.GetAuthRole(batch.TenantId, auth.HriIntegrator)) {
		msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "create")
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	// validate that the Subject claim (integrator ID) is not missing
	if claims.Subject == "" {
		msg := fmt.Sprintf(auth.MsgSubClaimRequiredInJwt)
		logger.Errorln(msg)
		return http.StatusUnauthorized, response.NewErrorDetail(requestId, msg)
	}

	var subject = claims.Subject
	return createBatch(requestId, batch, subject, kafkaWriter, logger)
}

func createBatch(
	requestId string,
	batch model.CreateBatch,
	integratorId string,
	kafkaWriter kafka.Writer,
	logger logrus.FieldLogger) (int, interface{}) {

	var (
		ctx          = context.Background()
		filter       = bson.M{"tenantId": mongoApi.GetTenantWithBatchesSuffix(batch.TenantId)}
		returnResult model.CreateBatchRequestForTenant
	)

	errMsg := "Batch creation failed for Tenant " + batch.TenantId
	batchInfo := buildBatchInfo(batch, integratorId)

	logger.Debugf("Successfully built BatchInfo for batch name: %s", batch.Name)

	mongoApi.HriCollection.FindOne(ctx, filter).Decode(&returnResult)

	if returnResult.TenantId == "" {
		return http.StatusNotFound, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusNotFound,
			logger, errMsg)
	}
	//Generate batchId to add batchInfo.
	batchId := mongoApi.RandomString(20)

	if returnResult.Docs_count == "" {
		returnResult.Docs_count = "1"
	}
	total_count := len(returnResult.Batch) + 1
	returnResult.Docs_count = strconv.Itoa(total_count)
	//adding docs_deleted attribute
	if len(returnResult.Batch) == 0 {
		returnResult.Docs_deleted = 0
	}
	updatedBatchInfo := UpdateBatchInfo(batch, integratorId, batchId)

	if nil != returnResult.Batch {
		returnResult.Batch = append(returnResult.Batch, updatedBatchInfo)
	} else {
		batchArraymodel := []model.CreateBatchBson{updatedBatchInfo}
		returnResult.Batch = batchArraymodel

	}

	// add batchId to info and publish to the notification topic
	logger.Debugf("Sending Batch Info to Notification Topic")
	//Sending  batchid to Kafka
	batchInfo[param.BatchId] = batchId
	notificationTopic := InputTopicToNotificationTopic(batch.Topic)
	err := kafkaWriter.Write(notificationTopic, batchId, batchInfo)
	if err != nil {
		logger.Errorf("Unable to publish to topic [%s] about new batch [%s]. %s",
			notificationTopic, batchId, err.Error())
		return http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error())

	}

	//update batchInfo to tenantId after writing to kafkatopic successfully
	result, UpdateErr := mongoApi.HriCollection.UpdateOne(ctx, bson.M{"_id": returnResult.Uuid}, bson.M{"$set": returnResult})

	logger.Debugf("Updated result count", result.ModifiedCount)
	if UpdateErr != nil {
		return http.StatusInternalServerError, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusBadRequest,
			logger, errMsg)
	}

	// return the ID of the newly created batch
	respBody := map[string]interface{}{param.BatchId: batchId}
	return http.StatusCreated, respBody
}
