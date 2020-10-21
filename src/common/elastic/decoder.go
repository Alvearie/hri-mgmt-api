/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package elastic

import (
	"encoding/json"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v6/esapi"
	"log"
	"net/http"
)

const msgClientErr string = "Elastic client error: %s"
const MsgNilResponse string = "Elastic client returned nil response without an error"
const msgParseErr string = "Error parsing the Elastic search response body: %s"
const msgDocNotFound string = "The document for tenantId: %s with document (batch) ID: %s was not found"
const msgResponseErr string = "%s: %s"
const msgUnexpectedErr string = "Unexpected Elastic Error - %s"
const msgTransformResult string = "Error transforming result -> %s"

func handleNotFound(body map[string]interface{}, tenantId string, logger *log.Logger) (string, map[string]interface{}) {
	var msg string
	var rtnBody map[string]interface{}
	if foundFlag, ok := body["found"]; ok {
		if foundFlag == false { //handle HTTP 404/NotFound errors
			logger.Println("got 'found=false' -> Find batchId...")
			batchId := body["_id"]
			msg = fmt.Sprintf(msgDocNotFound, tenantId, batchId)
			rtnBody = map[string]interface{}{"error": fmt.Sprintf(msgDocNotFound, tenantId, batchId)}
		}
	}
	return msg, rtnBody
}

// If error is nil, attempts to parse the json body of the Response.
// If successful, returns parsed body and optionally an error map, otherwise returns nil body and an error map.
// Note that this function will close the Response Body before returning.
func DecodeBody(res *esapi.Response, err error, tenantId string, logger *log.Logger) (map[string]interface{}, map[string]interface{}) {
	respErr := handleRespErr(err, logger, res)
	if respErr != nil {
		return nil, respErr
	}

	defer res.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		msg := fmt.Sprintf(msgParseErr, err.Error())
		logger.Println(msg)
		return nil, response.Error(http.StatusInternalServerError, msg)
	}

	if res.IsError() {
		var msg string
		if errBody, ok := body["error"].(map[string]interface{}); ok {
			if rootCause, ok := errBody["root_cause"].(map[string]interface{}); ok {
				msg = fmt.Sprintf(msgResponseErr, rootCause["type"], rootCause["reason"])
			} else {
				msg = fmt.Sprintf(msgResponseErr, errBody["type"], errBody["reason"])
			}
		} else if body["found"] != nil {
			msg, body = handleNotFound(body, tenantId, logger)
		} else { //Default: err handler -> No "error" element in body
			msg = fmt.Sprintf(msgUnexpectedErr, body)
			return nil, response.Error(http.StatusInternalServerError, msg)
		}

		logger.Println(msg)
		return body, response.Error(res.StatusCode, msg)
	}

	return body, nil
}

func DecodeBodyFromJsonArray(res *esapi.Response, err error, logger *log.Logger) ([]map[string]interface{}, map[string]interface{}) {
	respErr := handleRespErr(err, logger, res)
	var body []map[string]interface{}
	if respErr != nil {
		return body, respErr
	}

	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		msg := fmt.Sprintf(msgParseErr, err.Error())
		logger.Println(msg)
		return nil, response.Error(http.StatusInternalServerError, msg)
	}

	return body, nil
}

func DecodeFirstArrayElement(res *esapi.Response, err error, logger *log.Logger) (map[string]interface{}, map[string]interface{}) {
	decoded, errResp := DecodeBodyFromJsonArray(res, err, logger)
	if errResp != nil {
		return nil, errResp
	}

	if len(decoded) == 0 {
		msg := fmt.Sprintf(msgTransformResult, "Uh-Oh: we got no Object inside this ElasticSearch Results Array")
		logger.Println(msg)
		return nil, response.Error(http.StatusInternalServerError, msg)
	}
	return decoded[0], errResp
}

func handleRespErr(err error, logger *log.Logger, res *esapi.Response) map[string]interface{} {
	if err != nil {
		msg := fmt.Sprintf(msgClientErr, err.Error())
		logger.Println(msg)
		return response.Error(http.StatusInternalServerError, msg)
	} else if res == nil {
		logger.Print(MsgNilResponse)
		return response.Error(http.StatusInternalServerError, MsgNilResponse)
	}
	return nil
}
