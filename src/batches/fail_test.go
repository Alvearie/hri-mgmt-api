/*
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"encoding/json"
	"errors"
	"fmt"
	"ibm.com/watson/health/foundation/hri/batches/status"
	"ibm.com/watson/health/foundation/hri/common/auth"
	"ibm.com/watson/health/foundation/hri/common/elastic"
	"ibm.com/watson/health/foundation/hri/common/logwrapper"
	"ibm.com/watson/health/foundation/hri/common/model"
	"ibm.com/watson/health/foundation/hri/common/param"
	"ibm.com/watson/health/foundation/hri/common/response"
	"ibm.com/watson/health/foundation/hri/common/test"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestFail_GetUpdateScript(t *testing.T) {
	tests := []struct {
		name    string
		request *model.FailRequest
		// Note that the following chars must be escaped because expectedScript is used as a regex pattern: ., ), (, [, ]
		expectedRequest map[string]interface{}
	}{
		{
			name:    "success",
			request: getTestFailRequest(10, 2, "Batch Failed"),
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status != 'terminated' && ctx\._source\.status != 'failed'\) {ctx\._source\.status = 'failed'; ctx\._source\.actualRecordCount = 10; ctx\._source\.invalidRecordCount = 2; ctx\._source\.failureMessage = 'Batch Failed'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := getFailUpdateScript(tt.request)
			if err := RequestCompareScriptTest(tt.expectedRequest, request); err != nil {
				t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", request, tt.expectedRequest, err)
			}
		})
	}
}

func TestUpdateStatus_Fail(t *testing.T) {
	const scriptFail = `{"script":{"source":"if \(ctx\._source\.status != 'terminated' && ctx\._source\.status != 'failed'\) {ctx\._source\.status = 'failed'; ctx\._source\.actualRecordCount = 10; ctx\._source\.invalidRecordCount = 2; ctx\._source\.failureMessage = 'Batch Failed'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
	const currentStatus = status.Started

	logwrapper.Initialize("error", os.Stdout)
	validClaims := auth.HriClaims{Scope: auth.HriInternal}

	failedBatch := map[string]interface{}{
		param.BatchId:             test.ValidBatchId,
		param.Name:                batchName,
		param.IntegratorId:        integratorId,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.Failed.String(),
		param.StartDate:           batchStartDate,
		param.RecordCount:         batchExpectedRecordCount,
		param.ExpectedRecordCount: batchExpectedRecordCount,
		param.ActualRecordCount:   batchActualRecordCount,
		param.InvalidThreshold:    batchInvalidThreshold,
		param.InvalidRecordCount:  batchInvalidRecordCount,
	}
	failedJSON, err := json.Marshal(failedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	terminatedBatch := map[string]interface{}{
		param.BatchId:             test.ValidBatchId,
		param.Name:                batchName,
		param.IntegratorId:        integratorId,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.Terminated.String(),
		param.StartDate:           batchStartDate,
		param.ExpectedRecordCount: batchExpectedRecordCount,
		param.InvalidThreshold:    batchInvalidThreshold,
	}
	terminatedJSON, err := json.Marshal(terminatedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	tests := []struct {
		name                 string
		request              *model.FailRequest
		claims               auth.HriClaims
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedCode         int
		expectedResponse     interface{}
	}{
		{
			name:    "successful request",
			request: getTestFailRequest(int(batchActualRecordCount), int(batchInvalidRecordCount), batchFailureMessage),
			claims:  validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptFail,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidBatchId, failedJSON),
				},
			),
			expectedNotification: failedBatch,
			expectedCode:         http.StatusOK,
			expectedResponse:     nil,
		},
		{
			name:             "401 unauthorized when internal scope is missing",
			request:          getTestFailRequest(int(batchActualRecordCount), int(batchInvalidRecordCount), batchFailureMessage),
			claims:           auth.HriClaims{},
			expectedCode:     http.StatusUnauthorized,
			expectedResponse: response.NewErrorDetail(requestId, "Must have hri_internal role to mark a batch as failed"),
		},
		{
			name:    "500 on Elastic failure",
			request: getTestFailRequest(int(batchActualRecordCount), int(batchInvalidRecordCount), batchFailureMessage),
			claims:  validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptFail,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidBatchId, failedJSON),
					ResponseErr: errors.New("timeout"),
				},
			),
			expectedNotification: failedBatch,
			expectedCode:         http.StatusInternalServerError,
			expectedResponse:     response.NewErrorDetail(requestId, "could not update the status of batch test-batch: [500] elasticsearch client error: timeout"),
		},
		{
			name:    "'fail' fails on terminated batch",
			request: getTestFailRequest(int(batchActualRecordCount), int(batchInvalidRecordCount), batchFailureMessage),
			claims:  validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptFail,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "noop",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidBatchId, terminatedJSON),
				},
			),
			expectedCode:     http.StatusConflict,
			expectedResponse: response.NewErrorDetail(requestId, "'fail' failed, batch is in 'terminated' state"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esClient, err := elastic.ClientFromTransport(tt.ft)
			if err != nil {
				t.Error(err)
			}
			writer := test.FakeWriter{
				T:             t,
				ExpectedTopic: InputTopicToNotificationTopic(batchTopic),
				ExpectedKey:   test.ValidBatchId,
				ExpectedValue: tt.expectedNotification,
				Error:         tt.writerError,
			}

			code, result := Fail(requestId, tt.request, tt.claims, esClient, writer, currentStatus)

			if tt.ft != nil {
				tt.ft.VerifyCalls()
			}
			if code != tt.expectedCode {
				t.Errorf("Fail() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
			}
			if !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("Fail() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
			}
		})
	}
}

func TestFailNoAuth(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)
	const currentStatus = status.Started

	const (
		requestId                       = "requestNoAuth"
		batchName                       = "MonkeyBatch"
		batchTopic               string = "catz.batch.in"
		batchExpectedRecordCount        = float64(85)
		batchActualRecordCount          = float64(84)
		batchInvalidThreshold           = float64(5)
		batchInvalidRecordCount         = float64(1)
	)

	const failRequestBody = `{"script":{"source":"if \(ctx\._source\.status != 'terminated' && ctx\._source\.status != 'failed'\) {ctx\._source\.status = 'failed'; ctx\._source\.actualRecordCount = 84; ctx\._source\.invalidRecordCount = 1; ctx\._source\.failureMessage = 'Batch Failed'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"

	failedBatch := map[string]interface{}{
		param.BatchId:             test.ValidBatchId,
		param.Name:                batchName,
		param.IntegratorId:        integratorId,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.Failed.String(),
		param.StartDate:           batchStartDate,
		param.RecordCount:         batchExpectedRecordCount,
		param.ExpectedRecordCount: batchExpectedRecordCount,
		param.ActualRecordCount:   batchActualRecordCount,
		param.InvalidThreshold:    batchInvalidThreshold,
		param.InvalidRecordCount:  batchInvalidRecordCount,
	}
	failedJSON, err := json.Marshal(failedBatch)
	if err != nil {
		t.Errorf("Unable to create (Failed) batch JSON string: %s", err.Error())
	}

	tests := []struct {
		name                 string
		request              *model.FailRequest
		ft                   *test.FakeTransport
		expectedNotification map[string]interface{}
		expectedCode         int
		expectedResponse     interface{}
	}{
		{
			name:    "successful No Auth request",
			request: getTestFailRequest(int(batchActualRecordCount), int(batchInvalidRecordCount), batchFailureMessage),
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  failRequestBody,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidBatchId, failedJSON),
				},
			),
			expectedNotification: failedBatch,
			expectedCode:         http.StatusOK,
			expectedResponse:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esClient, err := elastic.ClientFromTransport(tt.ft)
			if err != nil {
				t.Error(err)
			}
			writer := test.FakeWriter{
				T:             t,
				ExpectedTopic: InputTopicToNotificationTopic(batchTopic),
				ExpectedKey:   test.ValidBatchId,
				ExpectedValue: tt.expectedNotification,
			}

			var emptyClaims = auth.HriClaims{}
			code, result := FailNoAuth(requestId, tt.request, emptyClaims, esClient, writer, currentStatus)

			tt.ft.VerifyCalls()
			if code != tt.expectedCode {
				t.Errorf("FailNoAuth() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
			}
			if !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("FailNoAuth() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
			}
		})
	}
}

func getTestFailRequest(actualRecordCount int, invalidRecordCount int, failureMsg string) *model.FailRequest {
	request := model.FailRequest{
		ProcessingCompleteRequest: model.ProcessingCompleteRequest{
			TenantId:           test.ValidTenantId,
			BatchId:            test.ValidBatchId,
			ActualRecordCount:  &actualRecordCount,
			InvalidRecordCount: &invalidRecordCount,
		},
		FailureMessage: failureMsg,
	}
	return &request
}
