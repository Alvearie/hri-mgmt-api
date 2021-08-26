/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"bytes"
	"encoding/json"
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
	"ibm.com/watson/health/foundation/hri/common/param/esparam"
	"ibm.com/watson/health/foundation/hri/common/response"
	"net/http"
	"time"
)

func Create(
	requestId string,
	batch model.CreateBatch,
	claims auth.HriClaims,
	esClient *elasticsearch.Client,
	kafkaWriter kafka.Writer) (int, interface{}) {

	prefix := "batches/create"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Create")

	// validate that caller has sufficient permissions
	if !claims.HasScope(auth.HriIntegrator) {
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
	return create(requestId, batch, subject, esClient, kafkaWriter, logger)
}

func CreateNoAuth(
	requestId string,
	batch model.CreateBatch,
	_ auth.HriClaims,
	esClient *elasticsearch.Client,
	kafkaWriter kafka.Writer) (int, interface{}) {

	var prefix = "batches/create_no_auth"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Debugln("Start Batch Create (Without Auth)")

	var integratorId = auth.NoAuthFakeIntegrator
	return create(requestId, batch, integratorId, esClient, kafkaWriter, logger)
}

func create(
	requestId string,
	batch model.CreateBatch,
	integratorId string,
	esClient *elasticsearch.Client,
	kafkaWriter kafka.Writer,
	logger logrus.FieldLogger) (int, interface{}) {

	batchInfo := buildBatchInfo(batch, integratorId)
	jsonBatchInfo, err := json.Marshal(batchInfo)
	if err != nil {
		//NOTE: This should Never happen because we are building the batchInfo map from the Statically-typed Batch struct
		return http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error())
	}
	logger.Debugf("Successfully built BatchInfo for batch name: %s", batch.Name)

	// add batch info to Elastic
	indexRes, err := esClient.Index(
		elastic.IndexFromTenantId(batch.TenantId),
		bytes.NewReader(jsonBatchInfo),
	)

	// parse the response
	body, elasticErr := elastic.DecodeBody(indexRes, err)
	if elasticErr != nil {
		return http.StatusInternalServerError, elasticErr.LogAndBuildErrorDetail(requestId,
			logger, "Batch creation failed")
	}

	batchId := body[esparam.EsDocId].(string)
	batchInfo[param.BatchId] = batchId

	// add batchId to info and publish to the notification topic
	logger.Debugf("Sending Batch Info to Notification Topic")

	notificationTopic := InputTopicToNotificationTopic(batch.Topic)
	err = kafkaWriter.Write(notificationTopic, batchId, batchInfo)
	if err != nil {
		logger.Errorf("Unable to publish to topic [%s] about new batch [%s]. %s",
			notificationTopic, batchId, err.Error())

		// cleanup the elastic document
		esClient.Delete(elastic.IndexFromTenantId(batch.TenantId), batchId)
		return http.StatusInternalServerError, response.NewErrorDetail(requestId, err.Error())
	}

	// return the ID of the newly created batch
	respBody := map[string]interface{}{param.BatchId: batchId}
	return http.StatusCreated, respBody
}

func buildBatchInfo(batch model.CreateBatch, integrator string) map[string]interface{} {
	currentTime := time.Now().UTC()

	info := map[string]interface{}{
		param.Name:             batch.Name,
		param.IntegratorId:     integrator,
		param.Topic:            batch.Topic,
		param.DataType:         batch.DataType,
		param.Status:           status.Started.String(),
		param.StartDate:        currentTime.Format(elastic.DateTimeFormat),
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
