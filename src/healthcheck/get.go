/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package healthcheck

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"ibm.com/watson/health/foundation/hri/common/elastic"
	"ibm.com/watson/health/foundation/hri/common/kafka"
	"ibm.com/watson/health/foundation/hri/common/logwrapper"
	esp "ibm.com/watson/health/foundation/hri/common/param/esparam"
	"ibm.com/watson/health/foundation/hri/common/response"
	"net/http"
)

const statusAllGood string = "green"
const serviceUnavailableMsg string = "HRI Service Temporarily Unavailable | error Detail: %v"
const notReported string = "NotReported"
const noStatusReported = "NONE/" + notReported

func Get(requestId string, client *elasticsearch.Client, healthChecker kafka.HealthChecker) (int, *response.ErrorDetail) {
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

	//2. Do Kafka healthCheck
	err = healthChecker.Check()
	logger.Infof("Kafka HealthCheck error: %v", err)
	var kaErrMsg = ""
	if err != nil {
		isErr = true
		kaErrMsg = err.Error()
		logger.Errorln(kaErrMsg)
	}

	var errMessage string
	if isErr {
		if len(esErrMsg) > 0 && len(kaErrMsg) > 0 {
			errMessage = fmt.Sprintf(serviceUnavailableMsg, esErrMsg+" | "+kaErrMsg)
		} else if len(kaErrMsg) > 0 {
			errMessage = fmt.Sprintf(serviceUnavailableMsg, kaErrMsg)
		} else {
			errMessage = fmt.Sprintf(serviceUnavailableMsg, esErrMsg)
		}
		return http.StatusServiceUnavailable, response.NewErrorDetail(requestId, errMessage)
	} else { //All Good for BOTH ElasticSearch AND Kafka Healthcheck
		return http.StatusOK, nil
	}
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
	return "ElasticSearch status: " + status + ", clusterId: " + clusterId + ", unixTimestamp: " + epoch
}
