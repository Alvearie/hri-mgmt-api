//go:build !tests
// +build !tests

/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/Alvearie/hri-mgmt-api/common/actionloopmin"
	"github.com/Alvearie/hri-mgmt-api/common/eventstreams"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/streams"
	"log"
	"os"
	"time"
)

func main() {
	actionloopmin.Main(createStreamMain)
}

func createStreamMain(params map[string]interface{}) map[string]interface{} {
	logger := log.New(os.Stdout, "streams/create: ", log.Llongfile)
	start := time.Now()
	logger.Printf("start createStreamMain, %s \n", start)

	service, err := eventstreams.CreateService(params)
	if err != nil {
		return err
	}

	resp := streams.Create(params, param.ParamValidator{}, service)
	logger.Printf("processing time createStreamMain, %d milliseconds \n", time.Since(start).Milliseconds())
	return resp
}
