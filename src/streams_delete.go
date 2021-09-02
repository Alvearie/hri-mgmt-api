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
	actionloopmin.Main(deleteStreamMain)
}

func deleteStreamMain(params map[string]interface{}) map[string]interface{} {
	logger := log.New(os.Stdout, "streams/delete: ", log.Llongfile)
	start := time.Now()
	logger.Printf("start deleteStreamMain, %s \n", start)

	service, err := eventstreams.CreateService(params)
	if err != nil {
		return err
	}

	resp := streams.Delete(params, param.ParamValidator{}, service)
	logger.Printf("processing time deleteStreamMain, %d milliseconds \n", time.Since(start).Milliseconds())
	return resp
}
