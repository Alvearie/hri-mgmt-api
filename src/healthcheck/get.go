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
	esp "github.com/Alvearie/hri-mgmt-api/common/param/esparam"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7"
	"log"
	"net/http"
	"os"
	"strconv"
)

const statusAllGood string = "green"
const serviceUnavailableMsg string = "HRI Service Temporarily Unavailable | error Detail: %v"
const kafkaConnFail string = "Kafka status: Kafka Connection/Read Partition failed"
const notReported string = "NotReported"
const noStatusReported string = "NONE/" + notReported

func Get(params map[string]interface{}, client *elasticsearch.Client, partReader kafka.PartitionReader) map[string]interface{} {
	logger := log.New(os.Stdout, "healthcheck/get: ", log.Llongfile)
	logger.Printf("Prepare HealthCheck - ElasticSearch (No Input Params)")

	//1. Do ElasticSearch healthCheck call
	resp, err := client.Cat.Health(client.Cat.Health.WithFormat("json"))
	respBody, elasticErr := elastic.DecodeFirstArrayElement(resp, err)
	if elasticErr != nil {
		return elasticErr.LogAndBuildApiResponse(logger, "Could not perform elasticsearch health check")
	}

	var isErr bool = false
	var esErrMsg string = ""
	var healthResponse map[string]interface{}
	status, ok := respBody[esp.EsStatus].(string)
	if !ok {
		isErr = true
		esErrMsg = getESErrorDetail(respBody, noStatusReported)
		logger.Println(esErrMsg)
	} else if status != statusAllGood {
		isErr = true
		esErrMsg = getESErrorDetail(respBody, status)
		logger.Println(esErrMsg)
	} else {
		unixTimestamp := getReturnedTimestamp(respBody)
		logger.Printf("ElasticSearch Health Success! ES Healthchk status: %v and timestamp: %v & \n", status, unixTimestamp)
	}

	//2. Do Kakfa Conn healthCheck
	isAvailable, err := kafka.CheckConnection(partReader)
	logger.Println("Kafka HealthCheck Result: " + strconv.FormatBool(isAvailable))
	var kaErrMsg string = ""
	if err != nil || isAvailable == false {
		isErr = true
		kaErrMsg = printKafkaErrDetail(logger)
		logger.Println(kaErrMsg)
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
		healthResponse = response.Error(http.StatusServiceUnavailable, errMessage)
	} else { //All Good for BOTH ElasticSearch AND Kafka Healthcheck
		emptyRespBody := map[string]interface{}{}
		healthResponse = response.Success(http.StatusOK, emptyRespBody)
	}

	return healthResponse
}

func printKafkaErrDetail(logger *log.Logger) string {
	errMessage := fmt.Sprintf(serviceUnavailableMsg, kafkaConnFail)
	logger.Println(errMessage)
	return errMessage
}

func getESErrorDetail(decodedResultBody map[string]interface{}, status string) string {
	unixTimestamp := getReturnedTimestamp(decodedResultBody)
	var clusterId string = notReported
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
