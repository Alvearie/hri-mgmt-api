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
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	requestId := "request_id_1"
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
		name         string
		ft           *test.FakeTransport
		expectedBody interface{}
		expectedCode int
	}{
		{
			name: "bad-response",
			ft: test.NewFakeTransport(t).AddCall(
				"/_cat/indices",
				test.ElasticCall{
					ResponseErr: errors.New(elasticErrMsg),
				},
			),
			expectedBody: response.NewErrorDetail(requestId, fmt.Sprintf("Could not retrieve tenants: [500] elasticsearch client error: %s", elasticErrMsg)),
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "simple",
			ft: test.NewFakeTransport(t).AddCall(
				"/_cat/indices",
				test.ElasticCall{
					RequestQuery: "format=json&h=index",
					ResponseBody: `[{"index":"pi001-batches"},{"index":"searchguard"},{"index":"pi002-batches"},{"index":"qatenant-batches"}]`,
				},
			),
			expectedBody: map[string]interface{}{"results": indices},
			expectedCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esClient, err := elastic.ClientFromTransport(tt.ft)
			if err != nil {
				t.Error(err)
			}

			code, body := Get(requestId, esClient)
			if code != tt.expectedCode {
				t.Errorf("Get() = %d, expected %d", code, tt.expectedCode)
			}
			if !reflect.DeepEqual(body, tt.expectedBody) {
				t.Errorf("Get() = %v, expected %v", body, tt.expectedBody)
			}
		})
	}
}
