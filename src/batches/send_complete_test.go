/*
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"encoding/json"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/batches/status"
	"github.com/Alvearie/hri-mgmt-api/common/auth"
	"github.com/Alvearie/hri-mgmt-api/common/elastic"
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestSendComplete_AuthCheck(t *testing.T) {
	tests := []struct {
		name        string
		claims      auth.HriClaims
		expectedErr string
	}{
		{
			name:   "With DI role, should return nil",
			claims: auth.HriClaims{Scope: auth.HriIntegrator},
		},
		{
			name:   "With DI & Consumer role, should return nil",
			claims: auth.HriClaims{Scope: auth.HriIntegrator + " " + auth.HriConsumer},
		},
		{
			name:        "Without DI role, should return error",
			claims:      auth.HriClaims{},
			expectedErr: "Must have hri_data_integrator role to update a batch",
		},
	}

	sendComplete := SendComplete{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := sendComplete.CheckAuth(tt.claims)
			if (err == nil && tt.expectedErr != "") || (err != nil && err.Error() != tt.expectedErr) {
				t.Errorf("GetAuth() = '%v', expected '%v'", err, tt.expectedErr)
			}
		})
	}
}

func TestSendComplete_GetUpdateScript(t *testing.T) {

	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: "integratorId"}

	tests := []struct {
		name   string
		params map[string]interface{}
		claims auth.HriClaims
		// Note that the following chars must be escaped because expectedScript is used as a regex pattern: ., ), (, [, ]
		expectedRequest map[string]interface{}
		expectedErr     map[string]interface{}
		metadata        bool
	}{
		{
			name: "no metadata with validation",
			params: map[string]interface{}{
				param.Validation:          true,
				param.ExpectedRecordCount: float64(200),
			},
			claims: validClaims,
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200;} else {ctx\.op = 'none'}`,
				},
			},
			metadata: false,
		},
		{
			name: "no metadata without validation",
			params: map[string]interface{}{
				param.Validation:          false,
				param.ExpectedRecordCount: float64(200),
			},
			claims: validClaims,
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'completed'; ctx\._source\.expectedRecordCount = 200; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}`,
				},
			},
			metadata: false,
		},
		{
			name: "metadata with validation",
			params: map[string]interface{}{
				param.Validation:          true,
				param.ExpectedRecordCount: float64(200),
				param.Metadata:            map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5},
			},
			claims: validClaims,
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}`,
					"lang":   "painless",
					"params": map[string]interface{}{"metadata": map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5}},
				},
			},
			metadata: true,
		},
		{
			name: "metadata without validation",
			params: map[string]interface{}{
				param.Validation:          false,
				param.ExpectedRecordCount: float64(200),
				param.Metadata:            map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": 3},
			},
			claims: validClaims,
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'completed'; ctx\._source\.expectedRecordCount = 200; ctx\._source\.endDate = '` + test.DatePattern + `'; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}`,
					"lang":   "painless",
					"params": map[string]interface{}{"metadata": map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": 3}},
				},
			},
			metadata: true,
		},
		{
			name: "with deprecated recordCount field",
			params: map[string]interface{}{
				param.Validation:  true,
				param.RecordCount: float64(200),
			},
			claims: validClaims,
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200;} else {ctx\.op = 'none'}`,
				},
			},
			metadata: false,
		},
		{
			name: "Missing Validation param",
			params: map[string]interface{}{
				param.ExpectedRecordCount: float64(200),
			},
			claims:      validClaims,
			expectedErr: response.MissingParams(param.Validation),
		},
		{
			name: "Missing Record Count param",
			params: map[string]interface{}{
				param.Validation: false,
			},
			claims:      validClaims,
			expectedErr: response.MissingParams(param.ExpectedRecordCount),
		},
		{
			name: "Bad Metadata type",
			params: map[string]interface{}{
				param.Validation:          false,
				param.ExpectedRecordCount: float64(200),
				param.Metadata:            "nil",
			},
			claims:      validClaims,
			expectedErr: response.InvalidParams("metadata must be a map, got string instead."),
		},
		{
			name: "Missing claim.Subject",
			params: map[string]interface{}{
				param.Validation:          true,
				param.ExpectedRecordCount: float64(200),
			},
			claims: auth.HriClaims{},
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == ''\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200;} else {ctx\.op = 'none'}`,
				},
			},
			metadata: false,
		},
	}

	sendComplete := SendComplete{}
	logger := log.New(os.Stdout, fmt.Sprintf("batches/%s: ", sendComplete.GetAction()), log.Llongfile)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			updateRequest, errResp := sendComplete.GetUpdateScript(tt.params, param.ParamValidator{}, tt.claims, logger)
			if !reflect.DeepEqual(errResp, tt.expectedErr) {
				t.Errorf("GetUpdateScript().errResp = '%v', expected '%v'", errResp, tt.expectedErr)
			} else if tt.expectedRequest != nil {
				if err := RequestCompareScriptTest(tt.expectedRequest, updateRequest); err != nil {
					t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", updateRequest, tt.expectedRequest, err)
				} else if tt.metadata {
					if err := RequestCompareWithMetadataTest(tt.expectedRequest, updateRequest); err != nil {
						t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", updateRequest, tt.expectedRequest, err)
					}
				}
			}

		})
	}

}

func TestUpdateStatus_SendComplete(t *testing.T) {
	activationId := "activationId"
	_ = os.Setenv(response.EnvOwActivationId, activationId)

	const (
		scriptSendComplete             string = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 10;} else {ctx\.op = 'none'}"}}` + "\n"
		scriptSendCompleteWithMetadata string = `{"script":{"lang":"painless","params":{"metadata":{"compression":"gzip","userMetaField1":"metadata","userMetaField2":-5}},"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 10; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}"}}` + "\n"
	)

	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: "integratorId"}

	sendCompletedBatch := map[string]interface{}{
		param.BatchId:             batchId,
		param.Name:                batchName,
		param.IntegratorId:        integratorId,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.SendCompleted.String(),
		param.StartDate:           batchStartDate,
		param.RecordCount:         batchExpectedRecordCount,
		param.ExpectedRecordCount: batchExpectedRecordCount,
		param.InvalidThreshold:    batchInvalidThreshold,
	}
	sendCompletedJSON, err := json.Marshal(sendCompletedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	sendCompletedBatchWithMetadata := map[string]interface{}{
		param.BatchId:             batchId,
		param.Name:                batchName,
		param.IntegratorId:        integratorId,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.SendCompleted.String(),
		param.StartDate:           batchStartDate,
		param.RecordCount:         batchExpectedRecordCount,
		param.ExpectedRecordCount: batchExpectedRecordCount,
		param.InvalidThreshold:    batchInvalidThreshold,
	}

	sendCompletedBatchWithMetadataJSON, err := json.Marshal(sendCompletedBatchWithMetadata)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	terminatedBatch := map[string]interface{}{
		param.BatchId:             batchId,
		param.Name:                batchName,
		param.IntegratorId:        integratorId,
		param.Topic:               batchTopic,
		param.DataType:            batchDataType,
		param.Status:              status.Terminated.String(),
		param.StartDate:           batchStartDate,
		param.ExpectedRecordCount: batchExpectedRecordCount,
		param.InvalidThreshold:    batchInvalidThreshold,
		param.Metadata:            map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5},
	}
	terminatedJSON, err := json.Marshal(terminatedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	tests := []struct {
		name                 string
		params               map[string]interface{}
		claims               auth.HriClaims
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedResponse     map[string]interface{}
	}{
		{
			name: "simple sendComplete, with validation",
			params: map[string]interface{}{
				path.ParamOwPath:          "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.Validation:          true,
				param.ExpectedRecordCount: batchExpectedRecordCount,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptSendComplete,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, sendCompletedJSON),
				},
			),
			expectedNotification: sendCompletedBatch,
			expectedResponse:     response.Success(http.StatusOK, map[string]interface{}{}),
		},
		{
			name: "simple sendComplete, with validation and metadata",
			params: map[string]interface{}{
				path.ParamOwPath:          "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.Validation:          true,
				param.ExpectedRecordCount: batchExpectedRecordCount,
				param.Metadata:            map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5},
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptSendCompleteWithMetadata,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, sendCompletedBatchWithMetadataJSON),
				},
			),
			expectedNotification: sendCompletedBatchWithMetadata,
			expectedResponse:     response.Success(http.StatusOK, map[string]interface{}{}),
		},
		{
			name: "sendComplete fails on terminated batch",
			params: map[string]interface{}{
				path.ParamOwPath:          "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.Validation:          true,
				param.ExpectedRecordCount: batchExpectedRecordCount,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptSendComplete,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "noop",
							"get": {
								"_source": %s
							}
						}`, terminatedJSON),
				},
			),
			expectedResponse: response.Error(http.StatusConflict, "The 'sendComplete' endpoint failed, batch is in 'terminated' state"),
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
				ExpectedKey:   batchId,
				ExpectedValue: tt.expectedNotification,
				Error:         tt.writerError,
			}

			if result := UpdateStatus(tt.params, param.ParamValidator{}, tt.claims, SendComplete{}, esClient, writer); !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("UpdateStatus() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
			}
		})
	}
}
