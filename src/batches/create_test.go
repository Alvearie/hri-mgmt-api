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
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/param/esparam"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestCreate(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	requestId := "reqZxQ9706"
	tenantId := "tenant123"
	integratorId := "integratorId"
	batchId := "batch654"
	batchName := "monkeeName"
	batchDataType := "pikachu"
	topicBase := "batchFunTopic"
	inputTopic := topicBase + inputSuffix
	batchMetadata := map[string]interface{}{"batchContact": "Samuel L. Jackson", "finalRecordCount": 200}
	batchInvalidThreshold := 10

	validBatch := model.CreateBatch{
		TenantId:         tenantId,
		Name:             batchName,
		Topic:            inputTopic,
		DataType:         batchDataType,
		InvalidThreshold: batchInvalidThreshold,
		Metadata:         batchMetadata,
	}

	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}
	var elasticErrMsg = "elasticErrMsg"

	validBatchKafkaMetadata := map[string]interface{}{
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

	testCases := []struct {
		name         string
		requestId    string
		batch        model.CreateBatch
		claims       auth.HriClaims
		transport    *test.FakeTransport
		writerError  error
		expectedCode int
		expectedBody interface{}
		kafkaValue   map[string]interface{}
	}{
		{
			name:         "unauthorized",
			requestId:    requestId,
			batch:        validBatch,
			claims:       auth.HriClaims{Scope: auth.HriConsumer, Subject: integratorId},
			transport:    test.NewFakeTransport(t),
			expectedCode: http.StatusUnauthorized,
			expectedBody: response.NewErrorDetail(requestId, fmt.Sprintf(auth.MsgIntegratorRoleRequired, "create")),
		},
		{
			name:         "empty-subject",
			requestId:    requestId,
			batch:        validBatch,
			claims:       auth.HriClaims{Scope: auth.HriIntegrator, Subject: ""},
			transport:    test.NewFakeTransport(t),
			expectedCode: http.StatusUnauthorized,
			expectedBody: response.NewErrorDetail(requestId, "JWT access token 'sub' claim must be populated."),
		},
		{
			name:      "elastic-error-response",
			requestId: requestId,
			batch:     validBatch,
			claims:    validClaims,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody: string(elasticIndexRequestBody),
					ResponseErr: errors.New(elasticErrMsg),
				},
			),
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewErrorDetail(requestId,
				fmt.Sprintf("Batch creation failed: [500] elasticsearch client error: %s", elasticErrMsg),
			),
			kafkaValue: validBatchKafkaMetadata,
		},
		{
			name:      "writer-error",
			requestId: requestId,
			batch:     validBatch,
			claims:    validClaims,
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
			writerError:  errors.New("Unable to write to Kafka"),
			expectedCode: http.StatusInternalServerError,
			expectedBody: response.NewErrorDetail(requestId, "Unable to write to Kafka"),
			kafkaValue:   validBatchKafkaMetadata,
		},
		{
			name:      "happy-path-good-request",
			requestId: requestId,
			batch:     validBatch,
			claims:    validClaims,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody:  string(elasticIndexRequestBody),
					ResponseBody: fmt.Sprintf(`{"%s": "%s"}`, esparam.EsDocId, batchId),
				},
			),
			expectedCode: http.StatusCreated,
			expectedBody: map[string]interface{}{param.BatchId: batchId},
			kafkaValue:   validBatchKafkaMetadata,
		},
	}

	for _, tc := range testCases {
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
			actualCode, actualBody := Create(tc.requestId, tc.batch, tc.claims, client, writer)
			if actualCode != tc.expectedCode || !reflect.DeepEqual(tc.expectedBody, actualBody) {
				//notify/print error event as test result
				t.Errorf("Batches-Create()\n   actual: %v,%v\n expected: %v,%v", actualCode, actualBody, tc.expectedCode, tc.expectedBody)
			}
			tc.transport.VerifyCalls()
		})
	}
}

func TestCreateNoAuth(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	requestId := "reqAc887"
	tenantId := "tenant03"
	batchId := "batch20"
	batchName := "hallNOates"
	batchDataType := "Snorkel"
	topicBase := "batchSadTopic"
	integratorId := auth.NoAuthFakeIntegrator
	inputTopic := topicBase + inputSuffix
	batchMetadata := map[string]interface{}{"batchContact": "Sergio Leone", "finalRecordCount": 200}
	batchInvalidThreshold := 5

	validBatch := model.CreateBatch{
		Name:             batchName,
		TenantId:         tenantId,
		Topic:            inputTopic,
		DataType:         batchDataType,
		InvalidThreshold: batchInvalidThreshold,
		Metadata:         batchMetadata,
	}

	validBatchKafkaMetadata := map[string]interface{}{
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

	testCases := []struct {
		name         string
		requestId    string
		batch        model.CreateBatch
		transport    *test.FakeTransport
		writerError  error
		expectedCode int
		expectedBody interface{}
		kafkaValue   map[string]interface{}
	}{
		{
			name:      "happy-path-good-request",
			requestId: requestId,
			batch:     validBatch,
			transport: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf("/%s-batches/_doc", tenantId),
				test.ElasticCall{
					RequestBody:  string(elasticIndexRequestBody),
					ResponseBody: fmt.Sprintf(`{"%s": "%s"}`, esparam.EsDocId, batchId),
				},
			),
			expectedCode: http.StatusCreated,
			expectedBody: map[string]interface{}{param.BatchId: batchId},
			kafkaValue:   validBatchKafkaMetadata,
		},
	}

	for _, tc := range testCases {
		eClient, err := elastic.ClientFromTransport(tc.transport)
		if err != nil {
			t.Error(err)
		}

		kWriter := test.FakeWriter{
			T:             t,
			ExpectedTopic: topicBase + notificationSuffix,
			ExpectedKey:   batchId,
			ExpectedValue: tc.kafkaValue,
			Error:         tc.writerError,
		}

		t.Run(tc.name, func(t *testing.T) {
			actualCode, actualBody := CreateNoAuth(tc.requestId, tc.batch, auth.HriClaims{}, eClient, kWriter)
			if actualCode != tc.expectedCode || !reflect.DeepEqual(tc.expectedBody, actualBody) {
				//print error event as test result
				t.Errorf("Batches-CreateNoAuth()\n  actual: %v,%v\n expected: %v,%v",
					actualCode, actualBody, tc.expectedCode, tc.expectedBody)
			}
			tc.transport.VerifyCalls()
		})
	}
}

func TestBuildBatchInfo(t *testing.T) {
	integratorId := "integratorId3"
	batchName := "monkeeBatch"
	batchDataType := "porcipine"
	topicBase := "batchFunTopic"
	inputTopic := topicBase + inputSuffix
	validBatchMetadata := map[string]interface{}{"batchContact": "The_Village_People", "finalRecordCount": 18}
	batchInvalidThreshold := 42

	validBatch := model.CreateBatch{
		Name:             batchName,
		Topic:            inputTopic,
		DataType:         batchDataType,
		InvalidThreshold: batchInvalidThreshold,
		Metadata:         validBatchMetadata,
	}

	validBatchNoThreshold := model.CreateBatch{
		Name:     batchName,
		Topic:    inputTopic,
		DataType: batchDataType,
		Metadata: validBatchMetadata,
	}

	validBatchNoMetadata := model.CreateBatch{
		Name:             batchName,
		Topic:            inputTopic,
		DataType:         batchDataType,
		InvalidThreshold: batchInvalidThreshold,
	}

	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

	validBatchInfo := map[string]interface{}{
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        test.DatePattern,
		param.Metadata:         validBatchMetadata,
		param.InvalidThreshold: validBatch.InvalidThreshold,
	}

	batchInfoNoThreshold := map[string]interface{}{
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        test.DatePattern,
		param.Metadata:         validBatchMetadata,
		param.InvalidThreshold: -1,
	}

	batchInfoNoMetadata := map[string]interface{}{
		param.Name:             batchName,
		param.IntegratorId:     integratorId,
		param.Topic:            inputTopic,
		param.DataType:         batchDataType,
		param.Status:           status.Started.String(),
		param.StartDate:        test.DatePattern,
		param.Metadata:         nil,
		param.InvalidThreshold: validBatchNoMetadata.InvalidThreshold,
	}

	testCases := []struct {
		name              string
		batch             model.CreateBatch
		claims            auth.HriClaims
		expectedBatchInfo map[string]interface{}
		expErr            error
	}{
		{
			name:              "Success valid Batch",
			batch:             validBatch,
			claims:            auth.HriClaims{Scope: auth.HriConsumer, Subject: integratorId},
			expectedBatchInfo: validBatchInfo,
		},
		{
			name:              "No Invalid Threshold Batch",
			batch:             validBatchNoThreshold,
			claims:            auth.HriClaims{Scope: auth.HriConsumer, Subject: integratorId},
			expectedBatchInfo: batchInfoNoThreshold,
		},
		{
			name:              "Nil Metadata Batch Success",
			batch:             validBatchNoMetadata,
			claims:            auth.HriClaims{Scope: auth.HriConsumer, Subject: integratorId},
			expectedBatchInfo: batchInfoNoMetadata,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualBatchInfo := buildBatchInfo(validBatch, validClaims.Subject)
			if batchInfoNotEqual(tc.expectedBatchInfo, actualBatchInfo) {
				//notify/print error event as test result
				t.Errorf("BuildBatchInfo()\n   actual: %v,\n expected: %v", actualBatchInfo, tc.expectedBatchInfo)
			}
		})
	}
}

func batchInfoNotEqual(expectedBatchInfo map[string]interface{},
	actualBatchInfo map[string]interface{}) bool {

	if expectedBatchInfo[param.Name] != actualBatchInfo[param.Name] ||
		expectedBatchInfo[param.IntegratorId] != actualBatchInfo[param.IntegratorId] ||
		expectedBatchInfo[param.Topic] != actualBatchInfo[param.Topic] ||
		expectedBatchInfo[param.DataType] != actualBatchInfo[param.DataType] ||
		expectedBatchInfo[param.Status] != actualBatchInfo[param.Status] ||
		expectedBatchInfo[param.InvalidThreshold] != actualBatchInfo[param.InvalidThreshold] ||
		actualBatchInfo[param.StartDate] == nil {

		var actualMD = actualBatchInfo[param.Metadata]
		var expectedMD = expectedBatchInfo[param.Metadata]
		if actualMD == nil && expectedMD != nil {
			return true
		} else if actualMD != nil && expectedMD == nil {
			return true
		} else if actualMD != nil && expectedMD != nil {
			if len(actualMD.(map[string]interface{})) != len(expectedMD.(map[string]interface{})) {
				//This is probably good enough; don't really need to Verify that the Metadata map was copied over exactly
				// We can assume it is if length is same
				return true
			}
		}
	}
	return false
}
