/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package elastic

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"
)

func TestDecodeBody(t *testing.T) {
	errText := "errText"
	errType := "errType"
	errReason := "errReason"
	errStatus := 300
	tenantId := "tenant123"
	batchId := "batch_no_exist"
	indexName := tenantId + "-" + batchId
	logger := log.New(ioutil.Discard, "responses/test: ", log.Llongfile)

	testCases := []struct {
		name         string
		res          *esapi.Response
		err          error
		expectedBody map[string]interface{}
		expectedErr  map[string]interface{}
	}{
		{
			name:        "has-err",
			err:         errors.New(errText),
			expectedErr: response.Error(http.StatusInternalServerError, fmt.Sprintf(msgClientErr, errText)),
		},
		{
			name:        "nil-response",
			expectedErr: response.Error(http.StatusInternalServerError, MsgNilResponse),
		},
		{
			name:        "bad-json",
			res:         &esapi.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(`{bad json: "`)))},
			expectedErr: response.Error(http.StatusInternalServerError, fmt.Sprintf(msgParseErr, "invalid character 'b' looking for beginning of object key string")),
		},
		{
			name:         "has-root-cause",
			res:          &esapi.Response{StatusCode: errStatus, Body: ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf(`{"error": {"root_cause": {"type": "%s", "reason": "%s"}}}`, errType, errReason))))},
			expectedErr:  response.Error(errStatus, fmt.Sprintf(msgResponseErr, errType, errReason)),
			expectedBody: map[string]interface{}{"error": map[string]interface{}{"root_cause": map[string]interface{}{"type": errType, "reason": errReason}}},
		},
		{
			name:         "no-root-cause",
			res:          &esapi.Response{StatusCode: errStatus, Body: ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf(`{"error": {"type": "%s", "reason": "%s"}}`, errType, errReason))))},
			expectedErr:  response.Error(errStatus, fmt.Sprintf(msgResponseErr, errType, errReason)),
			expectedBody: map[string]interface{}{"error": map[string]interface{}{"type": errType, "reason": errReason}},
		},
		{
			name:         "batch-not-found",
			res:          &esapi.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf(`{"_index": "%s", "_type":"_doc", "_id": "%s", "found":false}`, indexName, batchId))))},
			expectedErr:  response.Error(http.StatusNotFound, fmt.Sprintf(msgDocNotFound, tenantId, batchId)),
			expectedBody: map[string]interface{}{"error": fmt.Sprintf(msgDocNotFound, tenantId, batchId)},
		},
		{
			name:        "unexpected-err",
			res:         &esapi.Response{StatusCode: http.StatusServiceUnavailable, Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"unexpected response": " we did not expect this response:"}`)))},
			expectedErr: response.Error(http.StatusInternalServerError, fmt.Sprintf(msgUnexpectedErr, map[string]interface{}{"unexpected response: we did not expect this response": ""})),
		},
		{
			res:          &esapi.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(`{"good": "json"}`)))},
			expectedBody: map[string]interface{}{"good": "json"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := DecodeBody(tc.res, tc.err, tenantId, logger)
			if !reflect.DeepEqual(body, tc.expectedBody) {
				t.Errorf("Unexpected Body. Expected: [%v], Actual: [%v]", tc.expectedBody, body)
			}
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("Unexpected Error. Expected: [%v], Actual: [%v]", tc.expectedErr, err)
			}
		})
	}
}

func TestDecodeBodyFromJsonArray(t *testing.T) {
	errText := "arrayErrText"
	logger := log.New(ioutil.Discard, "decodeBodyFromArray/test: ", log.Llongfile)
	var validResult []map[string]interface{}
	validValuesMap := map[string]interface{}{
		"valid":  "json-array",
		"status": "green",
	}
	validResult = append(validResult, validValuesMap)

	testCases := []struct {
		name         string
		res          *esapi.Response
		err          error
		expectedBody []map[string]interface{}
		expectedErr  map[string]interface{}
	}{
		{
			name:         "success-case",
			res:          &esapi.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(`[{"valid": "json-array", "status": "green"}]`)))},
			expectedBody: validResult,
		},
		{
			name:        "has-err",
			err:         errors.New(errText),
			expectedErr: response.Error(http.StatusInternalServerError, fmt.Sprintf(msgClientErr, errText)),
		},
		{
			name:        "nil-response",
			expectedErr: response.Error(http.StatusInternalServerError, MsgNilResponse),
		},
		{
			name:        "decoder-error",
			res:         &esapi.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(`{bad json:[]][ "`)))},
			expectedErr: response.Error(http.StatusInternalServerError, fmt.Sprintf(msgParseErr, "invalid character 'b' looking for beginning of object key string")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := DecodeBodyFromJsonArray(tc.res, tc.err, logger)
			if !reflect.DeepEqual(body, tc.expectedBody) {
				t.Errorf("Unexpected Body. Expected: [%v], Actual: [%v]", tc.expectedBody, body)
			}
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("Unexpected Error. Expected: [%v], Actual: [%v]", tc.expectedErr, err)
			}
		})
	}
}

func TestDecodeFirstArrayElement(t *testing.T) {
	logger := log.New(ioutil.Discard, "decodeBodyFromArrayAsMap/test: ", log.Llongfile)

	testCases := []struct {
		name         string
		res          *esapi.Response
		err          error
		expectedBody map[string]interface{}
		expectedErr  map[string]interface{}
	}{
		{
			name:         "success-case",
			res:          &esapi.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(`[{"valid": "json-map-from-array", "status": "porcupine"}]`)))},
			expectedBody: map[string]interface{}{"valid": "json-map-from-array", "status": "porcupine"},
		},
		{
			name:        "empty-array",
			res:         &esapi.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(`[]`)))},
			expectedErr: response.Error(http.StatusInternalServerError, "Error transforming result -> Uh-Oh: we got no Object inside this ElasticSearch Results Array"),
		},
		{
			name:        "decoded-body-from-array-error",
			res:         &esapi.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte(`{bad json:[]][ "`)))},
			expectedErr: response.Error(http.StatusInternalServerError, fmt.Sprintf(msgParseErr, "invalid character 'b' looking for beginning of object key string")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := DecodeFirstArrayElement(tc.res, tc.err, logger)
			if !reflect.DeepEqual(body, tc.expectedBody) {
				t.Errorf("Unexpected Body. Expected: [%v], Actual: [%v]", tc.expectedBody, body)
			}
			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("Unexpected Error. Expected: [%v], Actual: [%v]", tc.expectedErr, err)
			}
		})
	}
}
