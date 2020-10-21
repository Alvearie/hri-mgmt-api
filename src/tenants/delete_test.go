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

func TestDelete(t *testing.T) {
	os.Setenv(response.EnvOwActivationId, "activation123")

	tenantId := "tenant123"

	validArgs := map[string]interface{}{
		path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s", tenantId),
	}

	elasticErrMsg := "elasticErrMsg"

	testCases := []struct {
		name              string
		args              map[string]interface{}
		validatorResponse map[string]interface{}
		transport         *test.FakeTransport
		expected          map[string]interface{}
	}{
		{
			name:      "missing-path",
			transport: test.NewFakeTransport(t),
			expected: response.Error(
				http.StatusBadRequest,
				"Required parameter '__ow_path' is missing"),
		},
		{
			name: "bad-response",
			args: validArgs,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches", tenantId),
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
			args: validArgs,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches", tenantId),
				test.ElasticCall{
					ResponseBody: `{bad json message : "`,
				},
			),
			expected: response.Error(
				http.StatusInternalServerError,
				"Error parsing the Elastic search response body: invalid character 'b' looking for beginning of object key string"),
		},
		{
			name: "good-request",
			args: validArgs,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches", tenantId),
				test.ElasticCall{
					ResponseBody: fmt.Sprintf(`{"acknowledged":true,"shards_acknowledged":true,"index":"%s-batches"}`, tenantId),
				},
			),
			expected: map[string]interface{}{
				"statusCode": http.StatusOK,
			},
		},
	}
	for _, tc := range testCases {

		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			actual := Delete(tc.args, client)
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tc.expected, actual))
			}
			tc.transport.VerifyCalls()
		})
	}
}
