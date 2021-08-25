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

func Test_getTerminateUpdateScript(t *testing.T) {
	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

	tests := []struct {
		name    string
		request *model.TerminateRequest
		claims  auth.HriClaims
		// Note that the following chars must be escaped because expectedScript is used as a regex pattern: ., ), (, [, ]
		expectedRequest map[string]interface{}
		metadata        bool
	}{
		{
			name:    "UpdateScript returns expected script without metadata",
			request: getTestTerminateRequest(nil),
			claims:  validClaims,
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}`,
				},
			},
			metadata: false,
		},
		{
			name:    "UpdateScript returns expected script with metadata",
			request: getTestTerminateRequest(map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5}),
			claims:  validClaims,
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `'; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}`,
					"lang":   "painless",
					"params": map[string]interface{}{"metadata": map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5}},
				},
			},
			metadata: true,
		},
		{
			name:    "Missing claim.Subject",
			request: getTestTerminateRequest(nil),
			claims:  auth.HriClaims{},
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == ''\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200;} else {ctx\.op = 'none'}`,
				},
			},
			metadata: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := getTerminateUpdateScript(tt.request, tt.claims.Subject)

			if err := RequestCompareScriptTest(tt.expectedRequest, request); err != nil {
				t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", request, tt.expectedRequest, err)
			} else if tt.metadata {
				if err := RequestCompareWithMetadataTest(tt.expectedRequest, request); err != nil {
					t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", request, tt.expectedRequest, err)
				}
			}
		})
	}
}

func TestUpdateStatus_Terminate(t *testing.T) {
	const (
		scriptTerminate        = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
		scriptTerminateWrongId = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'wrong id'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
		currentStatus          = status.SendCompleted
	)

	logwrapper.Initialize("error", os.Stdout)
	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

	terminatedBatch := map[string]interface{}{
		param.BatchId:             test.ValidBatchId,
		param.Name:                batchName,
		param.IntegratorId:        integratorId,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.Terminated.String(),
		param.StartDate:           batchStartDate,
		param.RecordCount:         batchExpectedRecordCount,
		param.ExpectedRecordCount: batchExpectedRecordCount,
	}
	terminatedJSON, err := json.Marshal(terminatedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	tests := []struct {
		name                 string
		request              *model.TerminateRequest
		claims               auth.HriClaims
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedCode         int
		expectedResponse     interface{}
	}{
		{
			name:    "successful Terminate",
			request: getTestTerminateRequest(nil),
			claims:  validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminate,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidBatchId, terminatedJSON),
				},
			),
			expectedNotification: terminatedBatch,
			expectedCode:         http.StatusOK,
			expectedResponse:     nil,
		},
		{
			name:             "401 Unauthorized when Data Integrator scope is missing",
			request:          getTestTerminateRequest(nil),
			claims:           auth.HriClaims{},
			expectedCode:     http.StatusUnauthorized,
			expectedResponse: response.NewErrorDetail(requestId, "Must have hri_data_integrator role to terminate a batch"),
		},
		{
			name:    "terminate fails on Elastic error",
			request: getTestTerminateRequest(nil),
			claims:  validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminate,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidBatchId, terminatedJSON),
					ResponseErr: errors.New("timeout"),
				},
			),
			expectedCode:     http.StatusInternalServerError,
			expectedResponse: response.NewErrorDetail(requestId, "could not update the status of batch test-batch: [500] elasticsearch client error: timeout"),
		},
		{
			name:    "terminate fails on wrong integrator Id",
			request: getTestTerminateRequest(nil),
			claims:  auth.HriClaims{Scope: auth.HriIntegrator, Subject: "wrong id"},
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminateWrongId,
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
			expectedNotification: terminatedBatch,
			expectedCode:         http.StatusUnauthorized,
			expectedResponse:     response.NewErrorDetail(requestId, "terminate requested by 'wrong id' but owned by 'integratorId'"),
		},
		{
			name:    "Terminate fails on terminated batch",
			request: getTestTerminateRequest(nil),
			claims:  validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminate,
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
			expectedResponse: response.NewErrorDetail(requestId, "terminate failed, batch is in 'terminated' state"),
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

			code, result := Terminate(requestId, tt.request, tt.claims, esClient, writer, currentStatus)

			if tt.ft != nil {
				tt.ft.VerifyCalls()
			}
			if code != tt.expectedCode {
				t.Errorf("Terminate() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
			}
			if !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("Terminate() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
			}
		})
	}
}

func Test_TerminateNoAuth(t *testing.T) {
	logwrapper.Initialize("error", os.Stdout)

	const (
		requestId                       = "requestNoAuth3"
		batchName                       = "BatchToTerminate"
		batchTopic               string = "bobo.batch.in"
		batchExpectedRecordCount        = float64(30)
		currentStatus                   = status.SendCompleted
	)

	terminatedBatch := map[string]interface{}{
		param.BatchId:             test.ValidBatchId,
		param.Name:                batchName,
		param.IntegratorId:        "",
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.Terminated.String(),
		param.StartDate:           batchStartDate,
		param.RecordCount:         batchExpectedRecordCount,
		param.ExpectedRecordCount: batchExpectedRecordCount,
	}
	terminatedJSON, err := json.Marshal(terminatedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	const (
		scriptTerminate = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'NoAuthUnkIntegrator'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
	)

	tests := []struct {
		name                 string
		request              *model.TerminateRequest
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedCode         int
		expectedResponse     interface{}
	}{
		{
			name:    "successful NoAuth Terminate",
			request: getTestTerminateRequest(nil),
			ft: test.NewFakeTransport(t).AddCall(
				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminate,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "%s-batches",
							"_type": "_doc",
							"_id": "%s",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, test.ValidTenantId, test.ValidTenantId, terminatedJSON),
				},
			),
			expectedNotification: terminatedBatch,
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
			code, result := TerminateNoAuth(requestId, tt.request, emptyClaims, esClient, writer, currentStatus)

			tt.ft.VerifyCalls()
			if code != tt.expectedCode {
				t.Errorf("TerminateNoAuth() = \n\t%v,\nexpected: \n\t%v", code, tt.expectedCode)
			}
			if !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("TerminateNoAuth() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
			}
		})
	}
}

func getTestTerminateRequest(metadata map[string]interface{}) *model.TerminateRequest {
	request := model.TerminateRequest{
		TenantId: test.ValidTenantId,
		BatchId:  test.ValidBatchId,
		Metadata: metadata,
	}
	return &request
}
