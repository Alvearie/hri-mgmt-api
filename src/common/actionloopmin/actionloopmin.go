/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Modified from: https://github.com/apache/openwhisk-runtime-go/blob/master/common/gobuild.py.launcher.go
 */

package actionloopmin

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Main is the actionloop wrapper for a cloud function to make it work with the actionloop container
func Main(fn func(params map[string]interface{}) map[string]interface{}) {
	out := os.NewFile(3, "pipe")
	startLoop(fn, out)
}

func startLoop(fn func(params map[string]interface{}) map[string]interface{}, out *os.File) {
	defer recoverFromPanicAndCleanup(fn, out)

	reader := bufio.NewReader(os.Stdin)

	// send ack
	// note that it depends on the runtime,
	// go 1.13+ requires an ack, past versions does not
	fmt.Fprintf(out, `{ "ok": true}%s`, "\n")

	// read-eval-print loop
	for {
		// read one line
		inbuf, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}

		// parse one line
		var input map[string]interface{}
		err = json.Unmarshal(inbuf, &input)
		if err != nil {
			log.Println(err.Error())
			fmt.Fprintf(out, "{ error: %q}\n", err.Error())
			continue
		}

		// set environment variables
		for k, v := range input {
			if k == "value" {
				continue
			}
			if s, ok := v.(string); ok {
				os.Setenv("__OW_"+strings.ToUpper(k), s)
			}
		}
		// get payload if not empty
		var payload map[string]interface{}
		if value, ok := input["value"].(map[string]interface{}); ok {
			payload = value
		}
		// process the request
		result := fn(payload)
		// encode the answer
		encodeResult(result, out)
	}
}

func recoverFromPanicAndCleanup(fn func(params map[string]interface{}) map[string]interface{}, out *os.File) {
	if r := recover(); r != nil {
		panicMsg := fmt.Sprint(r)
		encodeResult(response.Error(http.StatusInternalServerError, panicMsg), out)

		// restart loop to listen for subsequent action invocations
		startLoop(fn, out)
	} else {
		out.Close()
	}
}

func encodeResult(result map[string]interface{}, out *os.File) {
	output, err := json.Marshal(&result)
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintf(out, "{ error: %q}\n", err.Error())
	} else {
		output = bytes.Replace(output, []byte("\n"), []byte(""), -1)
		fmt.Fprintf(out, "%s\n", output)
	}
}
