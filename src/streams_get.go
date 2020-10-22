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
	"github.com/Alvearie/hri-mgmt-api/streams"
	"log"
	"os"
	"time"
)

func main() {
	actionloopmin.Main(getStreamsMain)
}

func getStreamsMain(params map[string]interface{}) map[string]interface{} {
	logger := log.New(os.Stdout, "streams/get: ", log.Llongfile)
	start := time.Now()
	logger.Printf("start getStreamsMain, %s \n", start)

	service, err := eventstreams.CreateService(params)
	if err != nil {
		return err
	}

	resp := streams.Get(params, service)
	logger.Printf("processing time getStreamsMain, %d milliseconds \n", time.Since(start).Milliseconds())
	return resp
}
