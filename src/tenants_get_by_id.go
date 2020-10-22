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
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/tenants"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	actionloopmin.Main(getTenantByIdMain)
}

func getTenantByIdMain(params map[string]interface{}) map[string]interface{} {
	logger := log.New(os.Stdout, "tenants/GetById: ", log.Llongfile)
	start := time.Now()
	logger.Printf("start getTenantByIdMain, %s \n", start)

	// check bearer token
	service := elastic.CreateResourceControllerService()
	check, err := elastic.CheckElasticIAM(params, service)
	if err != nil || check == false {
		return response.Error(http.StatusUnauthorized, err.Error())
	}

	esClient, err := elastic.ClientFromParams(params)
	if err != nil {
		return response.Error(http.StatusInternalServerError, err.Error())
	}
	resp := tenants.GetById(params, esClient)
	logger.Printf("processing time getTenantByIdMain, %d milliseconds \n", time.Since(start).Milliseconds())
	return resp
}
