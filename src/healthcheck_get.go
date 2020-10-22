// +build !tests

/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/Alvearie/hri-mgmt-api/common/actionloopmin"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/healthcheck"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	actionloopmin.Main(healthcheckMain)
}

func healthcheckMain(params map[string]interface{}) map[string]interface{} {
	logger := log.New(os.Stdout, "healthcheck/get: ", log.Llongfile)
	start := time.Now()
	logger.Printf("start healthcheckMain, %s \n", start)

	esClient, err := elastic.ClientFromParams(params)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	dialer, err := kafka.CreateDialer(params)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	kafkaConn, err := kafka.ConnFromParams(params, dialer)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	defer kafkaConn.Close()

	resp := healthcheck.Get(params, esClient, kafkaConn)
	logger.Printf("processing time healthcheckMain, %d milliseconds \n", time.Since(start).Milliseconds())
	return resp
}
