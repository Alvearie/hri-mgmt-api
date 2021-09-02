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
	"github.com/Alvearie/hri-mgmt-api/common/auth"
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
	integratorId := "integratorId"
	batchId := "batch123"
	batchName := "batchName"
	batchDataType := "batchDataType"
	topicBase := "batchTopic"
	inputTopic := topicBase + inputSuffix
	batchMetadata := map[string]interface{}{"operation": "update"}
	batchInvalidThreshold := 10

	validArgs := map[string]interface{}{
		path.ParamOwPath:       fmt.Sprintf("/hri/tenants/%s/batches", tenantId),
		param.Name:             batchName,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Metadata:         batchMetadata,
		param.InvalidThreshold: batchInvalidThreshold,
	}

	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

	validBatchMetadata := map[string]interface{}{
		param.BatchId:          batchId,
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        "ignored",
		param.Metadata:         batchMetadata,
		param.InvalidThreshold: batchInvalidThreshold,
	}

	elasticIndexRequestBody, err := json.Marshal(map[string]interface{}{
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        test.DatePattern,
		param.Metadata:         batchMetadata,
		param.InvalidThreshold: batchInvalidThreshold,
	})
	if err != nil {
		t.Fatal("Unable to marshal expected elastic Index request body")
	}

	elasticIndexRequestBodyDefaultThreshold, err := json.Marshal(map[string]interface{}{
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        test.DatePattern,
		param.Metadata:         batchMetadata,
		param.InvalidThreshold: -1,
	})

	invalidThresholdBody := map[string]interface{}{
		param.BatchId:          batchId,
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        "ignored",
		param.Metadata:         batchMetadata,
		param.InvalidThreshold: -1,
	}

	badParamResponse := map[string]interface{}{"bad": "param"}
	elasticErrMsg := "elasticErrMsg"

	testCases := []struct {
		name              string
		args              map[string]interface{}
		validatorResponse map[string]interface{}
		claims            auth.HriClaims
		transport         *test.FakeTransport
		writerError       error
		expected          map[string]interface{}
		kafkaValue        map[string]interface{}
	}{
		{
			name:      "unauthorized",
			args:      validArgs,
			claims:    auth.HriClaims{Scope: auth.HriConsumer, Subject: integratorId},
			transport: test.NewFakeTransport(t),
			expected:  response.Error(http.StatusUnauthorized, fmt.Sprintf(auth.MsgIntegratorRoleRequired, "create")),
		},
		{
			name:      "empty-subject",
			args:      validArgs,
			claims:    auth.HriClaims{Scope: auth.HriIntegrator, Subject: ""},
			transport: test.NewFakeTransport(t),
			expected:  response.Error(http.StatusUnauthorized, "JWT access token 'sub' claim must be populated"),
		},
		{
			name:              "bad-param",
			args:              validArgs,
			claims:            validClaims,
			transport:         test.NewFakeTransport(t),
			validatorResponse: badParamResponse,
			expected:          badParamResponse,
			kafkaValue:        validBatchMetadata,
		},
		{
			name: "missing-path",
			args: map[string]interface{}{
				param.Name:     batchName,
				param.Topic:    inputTopic,
				param.DataType: batchDataType,
			},
			claims:    validClaims,
			transport: test.NewFakeTransport(t),
			expected: response.Error(
				http.StatusBadRequest,
				"Required parameter '__ow_path' is missing"),
			kafkaValue: validBatchMetadata,
		},
		{
			name:   "bad-response",
			args:   validArgs,
			claims: validClaims,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody: string(elasticIndexRequestBody),
					ResponseErr: errors.New(elasticErrMsg),
				},
			),
			expected: response.Error(
				http.StatusInternalServerError,
				fmt.Sprintf("Batch creation failed: elasticsearch client error: %s", elasticErrMsg),
			),
			kafkaValue: validBatchMetadata,
		},
		{
			name:   "writer-error",
			args:   validArgs,
			claims: validClaims,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody:  string(elasticIndexRequestBody),
					ResponseBody: fmt.Sprintf(`{"%s": "%s"}`, esparam.EsDocId, batchId),
				},
			).AddCall(
				fmt.Sprintf("/%s-batches/_doc/%s", tenantId, batchId),
				test.ElasticCall{},
			),
			writerError: errors.New("Unable to write to Kafka"),
			expected:    response.Error(http.StatusInternalServerError, "Unable to write to Kafka"),
			kafkaValue:  validBatchMetadata,
		},
		{
			name:   "good-request",
			args:   validArgs,
			claims: validClaims,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody:  string(elasticIndexRequestBody),
					ResponseBody: fmt.Sprintf(`{"%s": "%s"}`, esparam.EsDocId, batchId),
				},
			),
			expected:   response.Success(http.StatusCreated, map[string]interface{}{param.BatchId: batchId}),
			kafkaValue: validBatchMetadata,
		},
		{
			name: "missing-invalid-threshold",
			args: map[string]interface{}{
				path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s/batches", tenantId),
				param.Name:       batchName,
				param.Topic:      inputTopic,
				param.DataType:   batchDataType,
				param.Metadata:   batchMetadata,
			},
			claims: validClaims,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody:  string(elasticIndexRequestBodyDefaultThreshold),
					ResponseBody: fmt.Sprintf(`{"%s": "%s"}`, esparam.EsDocId, batchId),
				},
			),
			expected:   response.Success(http.StatusCreated, map[string]interface{}{param.BatchId: batchId}),
			kafkaValue: invalidThresholdBody,
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
			ExpectedValue: tc.kafkaValue,
			Error:         tc.writerError,
		}

		t.Run(tc.name, func(t *testing.T) {
			if actual := Create(tc.args, validator, tc.claims, client, writer); !reflect.DeepEqual(tc.expected, actual) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tc.expected, actual))
			}
			tc.transport.VerifyCalls()
		})
	}
}
