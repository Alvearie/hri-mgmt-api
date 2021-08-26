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
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestDecodeBody(t *testing.T) {
	testCases := []struct {
		name                  string
		res                   *esapi.Response
		clientError           error
		expectedBody          map[string]interface{}
		expectedResponseError *ResponseError
	}{
		{
			name: "200 OK Response",
			res: &esapi.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				{"good": "json"}
			`)))},
			expectedBody: map[string]interface{}{
				"good": "json",
			},
		},
		{
			name:        "Elastic Client Error",
			clientError: errors.New("client Error"),
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf(msgClientErr, errors.New("client Error")),
				Code:     http.StatusInternalServerError,
			},
		},
		{
			name: "Elastic No Response No Client Error",
			expectedResponseError: &ResponseError{
				ErrorObj: errors.New(MsgNilResponse),
				Code:     http.StatusInternalServerError,
			},
		},
		{
			name: "Bad Json Response",
			res: &esapi.Response{StatusCode: http.StatusBadRequest, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				{"bad": "json"
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf(msgParseErr, errors.New("unexpected EOF")),
				Code:     http.StatusInternalServerError,
			},
		},
		{
			name: "Elastic Error Response",
			res: &esapi.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				{
					"error": {
						"type": "index_not_found_exception",
						"reason": "no such index",
						"index_uuid": "_na_",
						"resource.type": "index_or_alias",
						"resource.id": "tenant-batches",
						"index": "tenant-batches"
					},
					"status": 404
				}
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj:  fmt.Errorf("%s: %s", "index_not_found_exception", "no such index"),
				Code:      http.StatusNotFound,
				ErrorType: "index_not_found_exception"},
		},
		{
			name: "Elastic Error Response with Identical Root Cause",
			res: &esapi.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				{
					"error": {
						"root_cause": [
							{
								"type": "index_not_found_exception",
								"reason": "no such index",
								"index_uuid": "_na_",
								"resource.type": "index_or_alias",
								"resource.id": "tenant-batches",
								"index": "tenant-batches"
							}
						],
						"type": "index_not_found_exception",
						"reason": "no such index",
						"index_uuid": "_na_",
						"resource.type": "index_or_alias",
						"resource.id": "tenant-batches",
						"index": "tenant-batches"
					},
					"status": 404
				}
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf("%s: %s", "index_not_found_exception", "no such index"),
				Code:     http.StatusNotFound, ErrorType: "index_not_found_exception"},
		},
		{
			name: "Elastic Error Response with Different Root Cause",
			res: &esapi.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				{
					"error": {
						"root_cause": [
							{
								"type": "some_other_error",
								"reason": "some other reason",
								"index_uuid": "_na_",
								"resource.type": "index_or_alias",
								"resource.id": "tenant-batches",
								"index": "tenant-batches"
							},{
								"type": "additional errors",
								"reason": "will be ignored"
							}
						],
						"type": "index_not_found_exception",
						"reason": "no such index",
						"index_uuid": "_na_",
						"resource.type": "index_or_alias",
						"resource.id": "tenant-batches",
						"index": "tenant-batches"
					},
					"status": 404
				}
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf("%v: %w",
					fmt.Errorf("%s: %s", "index_not_found_exception", "no such index"),
					fmt.Errorf("%s: %s", "some_other_error", "some other reason"),
				),
				Code:      http.StatusNotFound,
				ErrorType: "index_not_found_exception",
				RootCause: "some_other_error",
			},
		},
		{
			name: "Elastic Error Message",
			res: &esapi.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				{
					"error": "alias [test-batches] missing",
					"status": 404
				}
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf("alias [test-batches] missing"),
				Code:     http.StatusNotFound},
		},
		{
			name: "Elastic No Error Context",
			res: &esapi.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				{
					"_index": "test-batches",
					"_type": "_doc",
					"_id": "missing_doc_id",
					"found": false
				}
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj: nil,
				Code:     http.StatusNotFound,
			},
			expectedBody: map[string]interface{}{
				"_index": "test-batches",
				"_type":  "_doc",
				"_id":    "missing_doc_id",
				"found":  false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, responseErr := DecodeBody(tc.res, tc.clientError)
			if !reflect.DeepEqual(body, tc.expectedBody) {
				t.Errorf("Unexpected Body. Expected: [%v], Actual: [%v]", tc.expectedBody, body)
			}
			if !reflect.DeepEqual(responseErr, tc.expectedResponseError) {
				t.Errorf("Unexpected ResponseError. Expected: [%v], Actual: [%v]",
					tc.expectedResponseError, responseErr)
			}
		})
	}
}

func TestDecodeBodyFromJsonArray(t *testing.T) {
	testCases := []struct {
		name                  string
		res                   *esapi.Response
		clientError           error
		expectedBody          []map[string]interface{}
		expectedResponseError *ResponseError
	}{
		{
			name: "success-case",
			res: &esapi.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				[
					{"valid": "json-array", "status": "green"}
				]
			`)))},
			expectedBody: []map[string]interface{}{
				{
					"valid":  "json-array",
					"status": "green",
				},
			},
		},
		{
			name:        "Elastic Client Error",
			clientError: errors.New("client Error"),
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf(msgClientErr, errors.New("client Error")),
				Code:     http.StatusInternalServerError,
			},
		},
		{
			name: "Elastic No Response No Client Error",
			expectedResponseError: &ResponseError{
				ErrorObj: errors.New(MsgNilResponse),
				Code:     http.StatusInternalServerError,
			},
		},
		{
			name: "Bad Json Response",
			res: &esapi.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				[{"bad": "json"
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf(msgParseErr, errors.New("unexpected EOF")),
				Code:     http.StatusInternalServerError,
			},
		},
		{
			name: "Elastic Error Response",
			res: &esapi.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				{
					"error": {
						"root_cause": [
							{
								"type": "illegal_argument_exception",
								"reason": "request [/_cat/indices/*-batches] contains unrecognized parameter: [badParam]"
							}
						],
						"type": "illegal_argument_exception",
						"reason": "request [/_cat/indices/*-batches] contains unrecognized parameter: [badParam]"
					},
					"status": 400
				}
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf("%s: %s",
					"illegal_argument_exception",
					"request [/_cat/indices/*-batches] contains unrecognized parameter: [badParam]",
				),
				Code:      http.StatusNotFound,
				ErrorType: "illegal_argument_exception",
			},
		},
		{
			name: "Unrecognized Elastic Error Response",
			res: &esapi.Response{StatusCode: http.StatusBadRequest, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				[
					{"valid": "json-array", "status": "green"}
				]
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf(msgUnexpectedErr, http.StatusBadRequest),
				Code:     http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, responseErr := DecodeBodyFromJsonArray(tc.res, tc.clientError)
			if !reflect.DeepEqual(body, tc.expectedBody) {
				t.Errorf("Unexpected Body. Expected: [%v], Actual: [%v]", tc.expectedBody, body)
			}
			if !reflect.DeepEqual(responseErr, tc.expectedResponseError) {
				t.Errorf("Unexpected ResponseError. Expected: [%v], Actual: [%v]",
					tc.expectedResponseError, responseErr)
			}
		})
	}
}

func TestDecodeFirstArrayElementUpdateMe(t *testing.T) {
	testCases := []struct {
		name                  string
		res                   *esapi.Response
		clientError           error
		expectedBody          map[string]interface{}
		expectedResponseError *ResponseError
	}{
		{
			name: "success-case",
			res: &esapi.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				[
					{"valid": "json-array", "status": "green"}
				]
			`)))},
			expectedBody: map[string]interface{}{
				"valid":  "json-array",
				"status": "green",
			},
		},
		{
			name:        "Elastic Client Error",
			clientError: errors.New("client Error"),
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf(msgClientErr, errors.New("client Error")),
				Code:     http.StatusInternalServerError,
			},
		},
		{
			name: "Empty Result",
			res: &esapi.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewReader([]byte(`
				[]
			`)))},
			expectedResponseError: &ResponseError{
				ErrorObj: fmt.Errorf(msgEmptyResultErr),
				Code:     http.StatusInternalServerError,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, responseErr := DecodeFirstArrayElement(tc.res, tc.clientError)
			if !reflect.DeepEqual(body, tc.expectedBody) {
				t.Errorf("Unexpected Body. Expected: [%v], Actual: [%v]", tc.expectedBody, body)
			}
			if !reflect.DeepEqual(responseErr, tc.expectedResponseError) {
				t.Errorf("Unexpected ResponseError. Expected: [%v], Actual: [%v]",
					tc.expectedResponseError, responseErr)
			}
		})
	}
}
