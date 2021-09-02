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
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/param/esparam"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"
)

func Create(
	args map[string]interface{},
	validator param.Validator,
	claims auth.HriClaims,
	esClient *elasticsearch.Client,
	kafkaWriter kafka.Writer) map[string]interface{} {

	logger := log.New(os.Stdout, "batches/create: ", log.Llongfile)

	// validate that caller has sufficient permissions
	if !claims.HasScope(auth.HriIntegrator) {
		msg := fmt.Sprintf(auth.MsgIntegratorRoleRequired, "create")
		logger.Printf(msg)
		return response.Error(http.StatusUnauthorized, msg)
	}

	// validate that the Subject claim (integrator ID) is not missing
	if claims.Subject == "" {
		logger.Printf(auth.MsgSubClaimRequiredInJwt)
		return response.Error(http.StatusUnauthorized, auth.MsgSubClaimRequiredInJwt)
	}

	// validate that required input params are present
	errResp := validator.Validate(
		args,
		param.Info{param.Name, reflect.String},
		param.Info{param.Topic, reflect.String},
		param.Info{param.DataType, reflect.String},
	)
	if errResp != nil {
		logger.Printf("Bad input params: %s", errResp)
		return errResp
	}

	// extract tenantId path param from URL
	tenantId, err := path.ExtractParam(args, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	batchInfo := buildBatchInfo(args, claims.Subject)
	jsonBatchInfo, err := json.Marshal(batchInfo)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	// add batch info to Elastic
	indexRes, err := esClient.Index(
		elastic.IndexFromTenantId(tenantId),
		bytes.NewReader(jsonBatchInfo),
	)

	// parse the response
	body, elasticErr := elastic.DecodeBody(indexRes, err)
	if elasticErr != nil {
		return elasticErr.LogAndBuildApiResponse(logger, "Batch creation failed")
	}

	batchId := body[esparam.EsDocId].(string)

	// add batchId to info and publish to the notification topic
	batchInfo[param.BatchId] = batchId
	notificationTopic := InputTopicToNotificationTopic(args[param.Topic].(string))
	err = kafkaWriter.Write(notificationTopic, batchId, batchInfo)
	if err != nil {
		logger.Printf("Unable to publish to topic [%s] about new batch [%s]. %s", notificationTopic, batchId, err.Error())

		// cleanup the elastic document
		esClient.Delete(elastic.IndexFromTenantId(tenantId), batchId)
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	// return the ID of the newly created batch
	respBody := map[string]interface{}{param.BatchId: batchId}
	return response.Success(http.StatusCreated, respBody)
}

func buildBatchInfo(args map[string]interface{}, integrator string) map[string]interface{} {
	currentTime := time.Now().UTC()

	info := map[string]interface{}{
		param.Name:             args[param.Name].(string),
		param.IntegratorId:     integrator,
		param.Topic:            args[param.Topic].(string),
		param.DataType:         args[param.DataType].(string),
		param.Status:           status.Started.String(),
		param.StartDate:        currentTime.Format(elastic.DateTimeFormat),
		param.InvalidThreshold: args[param.InvalidThreshold],
	}

	if info[param.InvalidThreshold] == nil {
		info[param.InvalidThreshold] = -1
	}

	if val, ok := args[param.Metadata]; ok {
		info[param.Metadata] = val
	}

	return info
}
