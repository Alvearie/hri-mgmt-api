/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package healthcheck

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	esp "github.com/Alvearie/hri-mgmt-api/common/param/esparam"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

const statusAllGood string = "green"
const serviceUnavailableMsg string = "HRI Service Temporarily Unavailable | error Detail: %v"
const kafkaConnFail string = "Kafka status: Kafka Connection/Read Partition failed"
const notReported string = "NotReported"
const noStatusReported = "NONE/" + notReported

func Get(requestId string, client *elasticsearch.Client, partReader kafka.PartitionReader) (int, *response.ErrorDetail) {
	prefix := "healthcheck/get"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Infof("Prepare HealthCheck - ElasticSearch (No Input Params)")

	//1. Do ElasticSearch healthCheck call
	resp, err := client.Cat.Health(client.Cat.Health.WithFormat("json"))
	respBody, elasticErr := elastic.DecodeFirstArrayElement(resp, err)
	if elasticErr != nil {
		return http.StatusServiceUnavailable, elasticErr.LogAndBuildErrorDetail(
			requestId, logger, "Could not perform elasticsearch health check")
	}

	var isErr = false
	var esErrMsg = ""
	status, ok := respBody[esp.EsStatus].(string)
	if !ok {
		isErr = true
		esErrMsg = getESErrorDetail(respBody, noStatusReported)
		logger.Errorln(esErrMsg)
	} else if status != statusAllGood {
		isErr = true
		esErrMsg = getESErrorDetail(respBody, status)
		logger.Errorln(esErrMsg)
	} else {
		unixTimestamp := getReturnedTimestamp(respBody)
		logger.Infof("ElasticSearch Health Success! ES Healthcheck status: %v and timestamp: %v & \n",
			status, unixTimestamp)
	}

	//2. Do Kafka Conn healthCheck
	isAvailable, err := kafka.CheckConnection(partReader)
	logger.Infoln("Kafka HealthCheck Result: " + strconv.FormatBool(isAvailable))
	var kaErrMsg = ""
	if err != nil || isAvailable == false {
		isErr = true
		kaErrMsg = printKafkaErrDetail(logger)
		logger.Errorln(kaErrMsg)
	}

	var errMessage string
	if isErr {
		if len(esErrMsg) > 0 && len(kaErrMsg) > 0 {
			errMessage = esErrMsg + "| " + kafkaConnFail
		} else if len(kaErrMsg) > 0 {
			errMessage = kaErrMsg
		} else {
			errMessage = esErrMsg
		}
		return http.StatusServiceUnavailable, response.NewErrorDetail(requestId, errMessage)
	} else { //All Good for BOTH ElasticSearch AND Kafka Healthcheck
		return http.StatusOK, nil
	}
}

func printKafkaErrDetail(logger logrus.FieldLogger) string {
	errMessage := fmt.Sprintf(serviceUnavailableMsg, kafkaConnFail)
	logger.Errorln(errMessage)
	return errMessage
}

func getESErrorDetail(decodedResultBody map[string]interface{}, status string) string {
	unixTimestamp := getReturnedTimestamp(decodedResultBody)
	var clusterId = notReported
	_, ok := decodedResultBody[esp.Cluster].(string)
	if ok {
		clusterId = decodedResultBody[esp.Cluster].(string)
	}

	errMessage := createESErrMsg(unixTimestamp, clusterId, status)
	return errMessage
}

func getReturnedTimestamp(decodedResultBody map[string]interface{}) string {
	_, ok := decodedResultBody[esp.Epoch].(string)
	if ok {
		return decodedResultBody[esp.Epoch].(string)
	}
	return notReported
}

func createESErrMsg(epoch string, clusterId string, status string) string {
	errDetails := "ElasticSearch status: " + status + ", clusterId: " + clusterId + ", unixTimestamp: " + epoch
	return fmt.Sprintf(serviceUnavailableMsg, errDetails)
}
