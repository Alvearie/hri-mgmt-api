/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/param/esparam"
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

	tenantId := "tenant123"
	docId := "batch123"
	batchId := EsDocIdToBatchId(docId)
	batchName := "batchName"
	batchDataType := "batchDataType"
	topicBase := "batchTopic"
	inputTopic := topicBase + inputSuffix
	batchMetadata := map[string]interface{}{"operation": "update"}

	validArgs := map[string]interface{}{
		path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s/batches", tenantId),
		param.Name:       batchName,
		param.Topic:      inputTopic,
		param.DataType:   batchDataType,
		param.Metadata:   batchMetadata,
	}

	validBatchMetadata := map[string]interface{}{
		param.BatchId:   batchId,
		param.Name:      batchName,
		param.Topic:     inputTopic,
		param.DataType:  batchDataType,
		param.Status:    status.Started.String(),
		param.StartDate: "ignored",
		param.Metadata:  batchMetadata,
	}

	elasticIndexRequestBody, err := json.Marshal(map[string]interface{}{
		param.Name:      batchName,
		param.Topic:     inputTopic,
		param.DataType:  batchDataType,
		param.Status:    status.Started.String(),
		param.StartDate: test.DatePattern,
		param.Metadata:  batchMetadata,
	})
	if err != nil {
		t.Fatal("Unable to marshal expected elastic Index request body")
	}

	badParamResponse := map[string]interface{}{"bad": "param"}
	elasticErrMsg := "elasticErrMsg"

	testCases := []struct {
		name              string
		args              map[string]interface{}
		validatorResponse map[string]interface{}
		transport         *test.FakeTransport
		writerError       error
		expected          map[string]interface{}
	}{
		{
			name:              "bad-param",
			args:              validArgs,
			transport:         test.NewFakeTransport(t),
			validatorResponse: badParamResponse,
			expected:          badParamResponse,
		},
		{
			name: "missing-path",
			args: map[string]interface{}{
				param.Name:     batchName,
				param.Topic:    inputTopic,
				param.DataType: batchDataType,
			},
			transport: test.NewFakeTransport(t),
			expected: response.Error(
				http.StatusBadRequest,
				"Required parameter '__ow_path' is missing"),
		},
		{
			name: "bad-response",
			args: validArgs,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody: string(elasticIndexRequestBody),
					ResponseErr: errors.New(elasticErrMsg),
				},
			),
			expected: response.Error(
				http.StatusInternalServerError,
				fmt.Sprintf("Elastic client error: %s", elasticErrMsg),
			),
		},
		{
			name: "writer-error",
			args: validArgs,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody:  string(elasticIndexRequestBody),
					ResponseBody: fmt.Sprintf(`{"%s": "%s"}`, esparam.EsDocId, docId),
				},
			).AddCall(
				fmt.Sprintf("/%s-batches/_doc/%s", tenantId, docId),
				test.ElasticCall{},
			),
			writerError: errors.New("Unable to write to Kafka"),
			expected:    response.Error(http.StatusInternalServerError, "Unable to write to Kafka"),
		},
		{
			name: "good-request",
			args: validArgs,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody:  string(elasticIndexRequestBody),
					ResponseBody: fmt.Sprintf(`{"%s": "%s"}`, esparam.EsDocId, docId),
				},
			),
			expected: response.Success(http.StatusCreated, map[string]interface{}{param.BatchId: batchId}),
		},
	}

	for _, tc := range testCases {
		validator := test.FakeValidator{
			T: t,
			Required: []param.Info{
				param.Info{param.Name, reflect.String},
				param.Info{param.Topic, reflect.String},
				param.Info{param.DataType, reflect.String},
			},
			Response: tc.validatorResponse,
		}

		client, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		writer := test.FakeWriter{
			T:             t,
			ExpectedTopic: topicBase + notificationSuffix,
			ExpectedKey:   batchId,
			ExpectedValue: validBatchMetadata,
			Error:         tc.writerError,
		}

		t.Run(tc.name, func(t *testing.T) {
			if actual := Create(tc.args, validator, client, writer); !reflect.DeepEqual(tc.expected, actual) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tc.expected, actual))
			}
			tc.transport.VerifyCalls()
		})
	}
}
