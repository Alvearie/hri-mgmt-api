/*
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package healthcheck

import (
	"fmt"
	"net/http"

	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/mongoApi"
	"go.mongodb.org/mongo-driver/mongo"
)

const hriServiceUnavailableMsg string = "HRI Service Temporarily Unavailable | error Detail: %v"

func GetCheck(requestId string, client *mongo.Collection, healthChecker kafka.HealthChecker) (int, *response.ErrorDetail) {
	prefix := "hrihealthcheck/getCheck"
	var logger = logwrapper.GetMyLogger(requestId, prefix)
	logger.Infof("Prepare HRI HealthCheck - CosmosDB (No Input Params)")

	//1. Do Mongo healthCheck call
	health_status, _, err := mongoApi.HriDatabaseHealthCheck(client)

	if err != nil {
		return http.StatusServiceUnavailable, mongoApi.LogAndBuildErrorDetail(requestId, http.StatusServiceUnavailable, logger, "Could not perform Cosmos health check")
	}

	var isErr = false
	var mErrMsg = ""
	if health_status != "1" {
		isErr = true
		mErrMsg = hriServiceUnavailableMsg
		logger.Errorln(mErrMsg)

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
		if len(mErrMsg) > 0 && len(kaErrMsg) > 0 {
			errMessage = fmt.Sprintf(serviceUnavailableMsg, mErrMsg+" | "+kaErrMsg)
		} else if len(kaErrMsg) > 0 {
			errMessage = fmt.Sprintf(serviceUnavailableMsg, kaErrMsg)
		} else {
			errMessage = fmt.Sprintf(serviceUnavailableMsg, mErrMsg)
		}
		return http.StatusServiceUnavailable, response.NewErrorDetail(requestId, errMessage)
	} else { //All Good for BOTH Mongo DB AND Kafka Healthcheck
		return http.StatusOK, nil
	}
}
