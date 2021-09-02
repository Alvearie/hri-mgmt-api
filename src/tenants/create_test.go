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
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestCreate(t *testing.T) {
	os.Setenv(response.EnvOwActivationId, "activation123")

	tenantId := "tenant-123_"

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
			name:      "bad-params capital letters",
			args:      map[string]interface{}{path.ParamOwPath: fmt.Sprintf("/hri/tenants/BADTenant")},
			transport: test.NewFakeTransport(t),
			expected: response.Error(
				http.StatusBadRequest, "TenantId: BADTenant must be lower-case alpha-numeric, '-', or '_'. 'B' is not allowed.",
			),
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
				fmt.Sprintf(
					"Unable to publish new tenant [tenant-123_]: elasticsearch client error: %s", elasticErrMsg),
			),
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
			expected: response.Success(http.StatusCreated, map[string]interface{}{param.TenantId: tenantId}),
		},
	}
	for _, tc := range testCases {
		validator := test.FakeValidator{
			T: t,

			Response: tc.validatorResponse,
		}

		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			if actual := Create(tc.args, validator, client); !reflect.DeepEqual(tc.expected, actual) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tc.expected, actual))
			}
			tc.transport.VerifyCalls()
		})
	}
}
