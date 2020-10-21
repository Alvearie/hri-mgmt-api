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

func TestGet(t *testing.T) {
	_ = os.Setenv(response.EnvOwActivationId, activationId)
	elasticErrMsg := "elasticErrMsg"

	id := make(map[string]interface{})
	id2 := make(map[string]interface{})
	id3 := make(map[string]interface{})
	var indices []interface{}
	id["id"] = "pi001"
	indices = append(indices, id)
	id2["id"] = "pi002"
	indices = append(indices, id2)
	id3["id"] = "qatenant"
	indices = append(indices, id3)

	tests := []struct {
		name     string
		params   map[string]interface{}
		ft       *test.FakeTransport
		expected map[string]interface{}
	}{
		{
			name: "bad-response",
			ft: test.NewFakeTransport(t).AddCall(
				"/_cat/indices",
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
			name: "body decode error on ES OK Response",
			ft: test.NewFakeTransport(t).AddCall(
				"/_cat/indices",
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
			ft: test.NewFakeTransport(t).AddCall(
				"/_cat/indices",
				test.ElasticCall{
					ResponseStatusCode: http.StatusBadRequest,
					ResponseBody:       `{bad json message : "`,
				},
			),
			expected: response.Error(
				http.StatusInternalServerError,
				"Error parsing the Elastic search response body: invalid character 'b' looking for beginning of object key string"),
		},
		{"simple",
			map[string]interface{}{path.ParamOwPath: "/hri/tenants/"},
			test.NewFakeTransport(t).AddCall(
				"/_cat/indices",
				test.ElasticCall{
					RequestQuery: "format=json&h=index",
					ResponseBody: `[{"index":"pi001-batches"},{"index":"searchguard"},{"index":"pi002-batches"},{"index":"qatenant-batches"}]`,
				},
			),
			map[string]interface{}{
				"statusCode": http.StatusOK,
				"body":       map[string]interface{}{"results": indices},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esClient, err := elastic.ClientFromTransport(tt.ft)
			if err != nil {
				t.Error(err)
			}

			if got := Get(esClient); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Get() = %v, expected %v", got, tt.expected)
			}
		})
	}

}
