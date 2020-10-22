/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"errors"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"net/http"
	"os"
	"reflect"
	"testing"
)

const activationId string = "activationId"

func TestGetById(t *testing.T) {
	_ = os.Setenv(response.EnvOwActivationId, activationId)

	docId := "batch7j3"
	batchId := EsDocIdToBatchId(docId)
	tenantId := "tenant12x"
	validPath := "/hri/tenants/" + tenantId + "/batches/" + batchId
	validPathArg := map[string]interface{}{
		path.ParamOwPath: validPath,
	}

	testCases := []struct {
		name      string
		args      map[string]interface{}
		transport *test.FakeTransport
		expected  map[string]interface{}
	}{
		{
			name: "success-case",
			args: validPathArg,
			transport: test.NewFakeTransport(t).AddCall(
				"/tenant12x-batches/_doc/batch7j3",
				test.ElasticCall{
					ResponseBody: `
						{ 
							"_index" : "tenant12x-batches",
							"_type" : "_doc",
							"_id" : "batch7j3",
							"_version" : 1,
							"_seq_no" : 0,
							"_primary_term" : 1,
							"found" : true,
							"_source" : {
								"name" : "monkeyBatch",
								"topic" : "ingest-test",
								"dataType" : "claims",
								"status" : "started",
								"recordCount" : 1,
								"startDate" : "2019-12-13"
							}
						}`,
				},
			),
			expected: response.Success(http.StatusOK, map[string]interface{}{"id": batchId, "name": "monkeyBatch", "status": "started", "startDate": "2019-12-13", "dataType": "claims", "topic": "ingest-test", "recordCount": float64(1)}),
		},
		{
			name: "batch not found",
			args: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/tenant12x/batches/batch-no-existo",
			},
			transport: test.NewFakeTransport(t).AddCall(
				"/tenant12x-batches/_doc/batch-no-existo",
				test.ElasticCall{
					ResponseStatusCode: http.StatusNotFound,
					ResponseBody: `
						{
							"_index": "tenant12x-batches",
							"_type": "_doc",
							"_id": "batch-no-existo",
							"found": false
						}`,
				},
			),
			expected: response.Error(http.StatusNotFound, "The document for tenantId: tenant12x with document (batch) ID: batch-no-existo was not found"),
		},
		{
			name:      "missing open whisk path param",
			args:      map[string]interface{}{}, //Missing Path Param
			transport: test.NewFakeTransport(t),
			expected: response.Error(
				http.StatusBadRequest,
				"Required parameter '__ow_path' is missing"),
		},
		{
			name: "bad open whisk path param tenant",
			args: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants",
			},
			transport: test.NewFakeTransport(t),
			expected: response.Error(
				http.StatusBadRequest,
				"The path is shorter than the requested path parameter; path: [ hri tenants], requested index: 3"),
		},
		{
			name: "bad path param batchId",
			args: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/tenant12x/batchId",
			},
			transport: test.NewFakeTransport(t),
			expected: response.Error(
				http.StatusBadRequest,
				"The path is shorter than the requested path parameter; path: [ hri tenants tenant12x batchId], requested index: 5"),
		},
		{
			name: "bad tenantId",
			args: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/bad-tenant/batches/" + batchId,
			},
			transport: test.NewFakeTransport(t).AddCall(
				"/bad-tenant-batches/_doc/batch7j3",
				test.ElasticCall{
					ResponseStatusCode: http.StatusNotFound,
					ResponseBody: `
						{
							"error": {
								"root_cause": [
									{
										"type" : "index_not_found_exception",
										"reason" : "no such index",
										"resource.type" : "index_or_alias",
										"resource.id" : "bad-tenant-batches",
										"index_uuid" : "_na_",
										"index" : "bad-tenant-batches"
									}
								],
								"type" : "index_not_found_exception",
								"reason" : "no such index",
								"resource.type" : "index_or_alias",
								"resource.id" : "bad-tenant-batches",
								"index_uuid" : "_na_",
								"index" : "bad-tenant-batches"
							},
							"status" : 404
						}`,
				},
			),
			expected: response.Error(
				http.StatusNotFound,
				"index_not_found_exception: no such index"),
		},
		{
			name: "bad-ES-response-body-EOF",
			args: validPathArg,
			transport: test.NewFakeTransport(t).AddCall(
				"/tenant12x-batches/_doc/batch7j3",
				test.ElasticCall{
					ResponseStatusCode: http.StatusNotFound,
					ResponseBody:       ``,
				},
			),
			expected: response.Error(
				http.StatusInternalServerError,
				"Error parsing the Elastic search response body: EOF"),
		},
		{
			name: "body decode error on ES OK Response",
			args: validPathArg,
			transport: test.NewFakeTransport(t).AddCall(
				"/tenant12x-batches/_doc/batch7j3",
				test.ElasticCall{
					ResponseBody: `{bad json message : "`,
				},
			),
			expected: response.Error(
				http.StatusInternalServerError,
				"Error parsing the Elastic search response body: invalid character 'b' looking for beginning of object key string"),
		},
		{
			name: "body decode error on ES Response: 400 Bad Request",
			args: validPathArg,
			transport: test.NewFakeTransport(t).AddCall(
				"/tenant12x-batches/_doc/batch7j3",
				test.ElasticCall{
					ResponseStatusCode: http.StatusBadRequest,
					ResponseBody:       `{bad json message : "`,
				},
			),
			expected: response.Error(
				http.StatusInternalServerError,
				"Error parsing the Elastic search response body: invalid character 'b' looking for beginning of object key string"),
		},
		{
			name: "client error",
			args: validPathArg,
			transport: test.NewFakeTransport(t).AddCall(
				"/tenant12x-batches/_doc/batch7j3",
				test.ElasticCall{
					ResponseErr: errors.New("some client error"),
				},
			),
			expected: response.Error(
				http.StatusInternalServerError,
				"Elastic client error: some client error"),
		},
	}

	for _, tc := range testCases {

		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			if actual := GetById(tc.args, client); !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("GetById() = %v, expected %v", actual, tc.expected)
			}
		})
	}
}
