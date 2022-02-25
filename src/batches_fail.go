//go:build !tests
// +build !tests

/*
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/Alvearie/hri-mgmt-api/batches"
	"github.com/Alvearie/hri-mgmt-api/common/actionloopmin"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/kafka"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/coreos/go-oidc"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	actionloopmin.Main(failMain)
}

func failMain(params map[string]interface{}) map[string]interface{} {
	logger := log.New(os.Stdout, "batches/fail: ", log.Llongfile)
	start := time.Now()
	logger.Printf("start failMain, %s \n", start)

	claims, errResp := auth.GetValidatedClaims(params, auth.AuthValidator{}, oidc.NewProvider)
	if errResp != nil {
		return errResp
	}

	esClient, err := elastic.ClientFromParams(params)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	kafkaWriter, err := kafka.NewWriterFromParams(params)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}

	resp := batches.UpdateStatus(params, param.ParamValidator{}, claims, batches.Fail{}, esClient, kafkaWriter)

	logger.Printf("processing time failMain, %d milliseconds \n", time.Since(start).Milliseconds())
	return resp
}
