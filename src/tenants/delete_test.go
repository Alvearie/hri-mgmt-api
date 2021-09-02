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

func TestDelete(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	requestId := "request_id_1"
	tenantId := "tenant123"

	elasticErrMsg := "elasticErrMsg"

	testCases := []struct {
		name              string
		validatorResponse map[string]interface{}
		transport         *test.FakeTransport
		expectedCode      int
		expectedBody      interface{}
	}{
		{
			name: "bad-response",
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches", tenantId),
				test.ElasticCall{
					ResponseErr: errors.New(elasticErrMsg),
				},
			),
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewErrorDetail(requestId, fmt.Sprintf("Could not delete tenant [%s]: [500] elasticsearch client error: %s", tenantId, elasticErrMsg)),
		},
		{
			name: "good-request",
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches", tenantId),
				test.ElasticCall{
					ResponseBody: fmt.Sprintf(`{"acknowledged":true,"shards_acknowledged":true,"index":"%s-batches"}`, tenantId),
				},
			),
			expectedCode: http.StatusOK,
			expectedBody: nil,
		},
	}

	for _, tc := range testCases {
		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		t.Run(tc.name, func(t *testing.T) {
			code, body := Delete(requestId, tenantId, client)
			if code != tc.expectedCode {
				t.Error(fmt.Sprintf("Incorrect HTTP code returned. Expected: [%v], actual: [%v]", tc.expectedCode, code))
			} else if !reflect.DeepEqual(tc.expectedBody, body) {
				t.Error(fmt.Sprintf("Incorrect HTTP response body returned. Expected: [%v], actual: [%v]", tc.expectedBody, body))
			}
			tc.transport.VerifyCalls()
		})
	}
}
