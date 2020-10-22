/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package tenants

import (
	"errors"
	"fmt"
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

	tenantId := "pi001"
	validPath := "/hri/tenants/" + tenantId
	validPathArg := map[string]interface{}{
		path.ParamOwPath: validPath,
	}

	elasticErrMsg := "elasticErrMsg"
	testCases := []struct {
		name      string
		args      map[string]interface{}
		transport *test.FakeTransport
		expected  map[string]interface{}
	}{
		{
			name:      "missing-path",
			transport: test.NewFakeTransport(t),
			expected: response.Error(
				http.StatusBadRequest,
				"Required parameter '__ow_path' is missing"),
		},
		{
			name: "bad path param tenantId",
			args: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants",
			},
			transport: test.NewFakeTransport(t),
			expected: response.Error(
				http.StatusBadRequest,
				"The path is shorter than the requested path parameter; path: [ hri tenants], requested index: 3"),
		},
		{
			name: "bad-response",
			args: validPathArg,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/_cat/indices/%s-batches", tenantId),
				test.ElasticCall{
					ResponseErr: errors.New(elasticErrMsg),
				},
			),
			expected: response.Error(
				http.StatusInternalServerError,
				fmt.Sprintf("%s", elasticErrMsg),
			),
		},
		{
			name: "Tenant not found",
			args: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/bad-tenant",
			},
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/indices/bad-tenant-batches",
				test.ElasticCall{
					ResponseStatusCode: http.StatusNotFound,
					ResponseBody: `
						{
  							"error" : {
    						"root_cause" : [
      							{
									"type" : "index_not_found_exception",
        							"reason" : "no such index",
        							"resource.type" : "index_or_alias",
        							"resource.id" : "pi001-batche",
        							"index_uuid" : "_na_",
        							"index" : "pi001-batche"
      							}
    						],
    						"type" : "index_not_found_exception",
    						"reason" : "no such index",
							"resource.type" : "index_or_alias",
    						"resource.id" : "pi001-batche",
							"index_uuid" : "_na_",
    						"index" : "pi001-batche"
  						},
  						"status" : 404
					}
				`,
				},
			),
			expected: response.Error(http.StatusNotFound, "Tenant: bad-tenant not found"),
		},
		{
			name: "body decode error on ES OK Response",
			args: validPathArg,
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/indices/pi001-batches",
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
				"/_cat/indices/pi001-batches",
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
			name: "success-case",
			args: validPathArg,
			transport: test.NewFakeTransport(t).AddCall(
				"/_cat/indices/pi001-batches",
				test.ElasticCall{
					ResponseBody: `
					[
						{
							"health" : "green",
    						"status" : "open",
    						"index" : "pi001-batches",
							"uuid" : "vTBmiZwhRcatGw4qAQqdRQ",
    						"pri" : "1",
    						"rep" : "1",
   							"docs.count" : "234",
    						"docs.deleted" : "2",
							"store.size" : "108.8kb",
    						"pri.store.size" : "54.4kb"
  						}
					]`,
				},
			),
			expected: response.Success(http.StatusOK, map[string]interface{}{
				"health":         "green",
				"status":         "open",
				"index":          "pi001-batches",
				"uuid":           "vTBmiZwhRcatGw4qAQqdRQ",
				"pri":            "1",
				"rep":            "1",
				"docs.count":     "234",
				"docs.deleted":   "2",
				"store.size":     "108.8kb",
				"pri.store.size": "54.4kb",
			}),
		},
	}

	for _, tc := range testCases {

		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			actual := GetById(tc.args, client)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("GetById() = %v, expected %v", actual, tc.expected)
			}
		})
	}
}
