/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"bytes"
	"encoding/json"
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/param/esparam"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v6"
	"log"
	"net/http"
	"os"
	"reflect"
	"time"
)

func Create(
	args map[string]interface{},
	validator param.Validator,
	esClient *elasticsearch.Client,
	kafkaWriter kafka.Writer) map[string]interface{} {

	logger := log.New(os.Stdout, "batches/create: ", log.Llongfile)
	//TODO validate caller?

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

	// extract tenantId path param from URL (we know it's there because we validated the params)
	tenantId, err := path.ExtractParam(args, param.TenantIndex)
	if err != nil {
		logger.Println(err.Error())
		return response.Error(http.StatusBadRequest, err.Error())
	}

	batchInfo := buildBatchInfo(args)
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
	body, errRes := elastic.DecodeBody(indexRes, err, tenantId, logger)
	if errRes != nil {
		return errRes
	}
	batchId := EsDocIdToBatchId(body[esparam.EsDocId].(string))

	// add batchId to info and publish to the notification topic
	batchInfo[param.BatchId] = batchId
	notificationTopic := InputTopicToNotificationTopic(args[param.Topic].(string))
	err = kafkaWriter.Write(notificationTopic, batchId, batchInfo)
	if err != nil {
		logger.Printf("Unable to publish to topic [%s] about new batch [%s]. %s", notificationTopic, batchId, err.Error())

		// cleanup the elastic document
		esClient.Delete(elastic.IndexFromTenantId(tenantId), BatchIdToEsDocId(batchId))
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	// return the ID of the newly created batch
	respBody := map[string]interface{}{param.BatchId: batchId}
	return response.Success(http.StatusCreated, respBody)
}

func buildBatchInfo(args map[string]interface{}) map[string]interface{} {
	currentTime := time.Now().UTC()

	info := map[string]interface{}{
		param.Name:      args[param.Name].(string),
		param.Topic:     args[param.Topic].(string),
		param.DataType:  args[param.DataType].(string),
		param.Status:    status.Started.String(),
		param.StartDate: currentTime.Format(elastic.DateTimeFormat),
	}

	if val, ok := args[param.Metadata]; ok {
		info[param.Metadata] = val
	}

	return info
}
