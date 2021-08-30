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
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/logwrapper"
	"github.com/Alvearie/hri-mgmt-api/common/model"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func Test_getProcessingCompleteUpdateScript(t *testing.T) {
	tests := []struct {
		name    string
		request model.ProcessingCompleteRequest
		// Note that the following chars must be escaped because expectedScript is used as a regex pattern: ., ), (, [, ]
		expectedRequest map[string]interface{}
	}{
		{
			name:    "success",
			request: *getValidTestProcessingCompleteRequest(),
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'sendCompleted'\) {ctx\._source\.status = 'completed'; ctx\._source\.actualRecordCount = 10; ctx\._source\.invalidRecordCount = 2; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}`,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := getProcessingCompleteUpdateScript(&tt.request)

			if err := RequestCompareScriptTest(tt.expectedRequest, request); err != nil {
				t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", request, tt.expectedRequest, err)
			}
		})
	}
}

func Test_ProcessingComplete(t *testing.T) {
	const scriptProcessingComplete = `{"script":{"source":"if \(ctx\._source\.status == 'sendCompleted'\) {ctx\._source\.status = 'completed'; ctx\._source\.actualRecordCount = 10; ctx\._source\.invalidRecordCount = 2; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
	const currentStatus = status.SendCompleted

	logwrapper.Initialize("error", os.Stdout)
	validClaims := auth.HriClaims{Scope: auth.HriInternal}

	completedBatch := map[string]interface{}{
		param.BatchId:             test.ValidBatchId,
		param.Name:                batchName,
		param.IntegratorId:        integratorId,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.Completed.String(),
		param.StartDate:           batchStartDate,
		param.RecordCount:         batchExpectedRecordCount,
		param.ExpectedRecordCount: batchExpectedRecordCount,
		param.ActualRecordCount:   batchActualRecordCount,
		param.InvalidThreshold:    batchInvalidThreshold,
		param.InvalidRecordCount:  batchInvalidRecordCount,
	}
	completedJSON, err := json.Marshal(completedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	failedBatch := map[string]interface{}{
		param.BatchId:             test.ValidBatchId,
		param.Name:                batchName,
		param.IntegratorId:        integratorId,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.Failed.String(),
		param.StartDate:           batchStartDate,
		param.ExpectedRecordCount: batchExpectedRecordCount,
		param.ActualRecordCount:   batchActualRecordCount,
		param.InvalidThreshold:    batchInvalidThreshold,
		param.InvalidRecordCount:  batchInvalidRecordCount,
		param.FailureMessage:      batchFailureMessage,
	}
	failedJSON, err := json.Marshal(failedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	tests := []struct {
		name                 string
		request              *model.ProcessingCompleteRequest
		claims               auth.HriClaims
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedCode         int
		expectedResponse     interface{}
	}{
		{
			name:    "successful processingComplete",
			request: getValidTestProcessingCompleteRequest(),
			claims:  validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptProcessingComplete,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidBatchId, completedJSON),
				},
			),
			expectedNotification: completedBatch,
			expectedCode:         http.StatusOK,
			expectedResponse:     nil,
		},
		{
			name:                 "processingComplete fails when missing internal scope",
			request:              getValidTestProcessingCompleteRequest(),
			claims:               auth.HriClaims{},
			expectedNotification: completedBatch,
			expectedCode:         http.StatusUnauthorized,
			expectedResponse:     response.NewErrorDetail(requestId, "Must have hri_internal role to mark a batch as processingComplete"),
		},
		{
			name:    "processingComplete fails on Elastic error",
			request: getValidTestProcessingCompleteRequest(),
			claims:  validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptProcessingComplete,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidBatchId, completedJSON),
					ResponseErr: errors.New("timeout"),
				},
			),
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: response.NewErrorDetail(requestId, "could not update the status of batch test-batch: [500] elasticsearch client error: timeout"),
		},
		{
			name:    "processingComplete fails on failed batch",
			request: getValidTestProcessingCompleteRequest(),
			claims:  validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptProcessingComplete,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "noop",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidBatchId, failedJSON),
				},
			),
			expectedCode:     http.StatusConflict,
			expectedResponse: response.NewErrorDetail(requestId, "processingComplete failed, batch is in 'failed' state"),
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

			code, result := ProcessingComplete(requestId, tt.request, tt.claims, esClient, writer, currentStatus)

			if tt.ft != nil {
				tt.ft.VerifyCalls()
			}

			if code != tt.expectedCode {
				t.Errorf("ProcessingComplete() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
			}
			if !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("ProcessingComplete() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
			}
		})
	}
}

func Test_ProcessingCompleteNoAuth(t *testing.T) {
	const scriptProcessingComplete = `{"script":{"source":"if \(ctx\._source\.status == 'sendCompleted'\) {ctx\._source\.status = 'completed'; ctx\._source\.actualRecordCount = 10; ctx\._source\.invalidRecordCount = 2; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
	const currentStatus = status.SendCompleted

	completedBatch := map[string]interface{}{
		param.BatchId:             test.ValidBatchId,
		param.Name:                batchName,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.Completed.String(),
		param.StartDate:           batchStartDate,
		param.RecordCount:         batchExpectedRecordCount,
		param.ExpectedRecordCount: batchExpectedRecordCount,
		param.ActualRecordCount:   batchActualRecordCount,
		param.InvalidThreshold:    batchInvalidThreshold,
		param.InvalidRecordCount:  batchInvalidRecordCount,
	}

	completedJSON, err := json.Marshal(completedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	logwrapper.Initialize("error", os.Stdout)

	tests := []struct {
		name                 string
		request              model.ProcessingCompleteRequest
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedCode         int
		expectedResponse     interface{}
	}{
		{
			name:    "successful processingComplete",
			request: *getValidTestProcessingCompleteRequest(),
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId), //"/tenantZzCat44-batches/_doc/batch789J/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptProcessingComplete,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "tenantZzCat44-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, completedJSON),
				},
			),
			expectedNotification: completedBatch,
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
				Error:         tt.writerError,
			}

			var emptyClaims = auth.HriClaims{}
			code, result := ProcessingCompleteNoAuth(requestId, &tt.request, emptyClaims, esClient, writer, currentStatus)

			tt.ft.VerifyCalls()

			if code != tt.expectedCode {
				t.Errorf("ProcessingCompleteNoAuth() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
			}
			if !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("ProcessingCompleteNoAuth() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
			}
		})
	}
}

func getTestProcessingCompleteRequest(actualRecCount int, invalidRecCount int) *model.ProcessingCompleteRequest {
	request := model.ProcessingCompleteRequest{
		TenantId:           test.ValidTenantId,
		BatchId:            test.ValidBatchId,
		ActualRecordCount:  &actualRecCount,
		InvalidRecordCount: &invalidRecCount,
	}
	return &request
}

func getValidTestProcessingCompleteRequest() *model.ProcessingCompleteRequest {
	return getTestProcessingCompleteRequest(int(batchActualRecordCount), int(batchInvalidRecordCount))
}
