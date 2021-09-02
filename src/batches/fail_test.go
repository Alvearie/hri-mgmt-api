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

func TestFail_AuthCheck(t *testing.T) {
	tests := []struct {
		name        string
		claims      auth.HriClaims
		expectedErr string
	}{
		{
			name:   "With internal role, should return nil",
			claims: auth.HriClaims{Scope: auth.HriInternal},
		},
		{
			name:   "With internal & Consumer role, should return nil",
			claims: auth.HriClaims{Scope: auth.HriInternal + " " + auth.HriConsumer},
		},
		{
			name:        "Without internal role, should return error",
			claims:      auth.HriClaims{Scope: auth.HriIntegrator},
			expectedErr: "Must have hri_internal role to mark a batch as failed",
		},
	}

	fail := Fail{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := fail.CheckAuth(tt.claims)
			if (err == nil && tt.expectedErr != "") || (err != nil && err.Error() != tt.expectedErr) {
				t.Errorf("GetAuth() = '%v', expected '%v'", err, tt.expectedErr)
			}
		})
	}
}

func TestFail_GetUpdateScript(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		claims auth.HriClaims
		// Note that the following chars must be escaped because expectedScript is used as a regex pattern: ., ), (
		expectedRequest map[string]interface{}
		expectedErr     map[string]interface{}
	}{
		{
			name: "success",
			params: map[string]interface{}{
				param.ActualRecordCount:  float64(10),
				param.InvalidRecordCount: float64(2),
				param.FailureMessage:     "Batch Failed",
			},
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status != 'terminated' && ctx\._source\.status != 'failed'\) {ctx\._source\.status = 'failed'; ctx\._source\.actualRecordCount = 10; ctx\._source\.invalidRecordCount = 2; ctx\._source\.failureMessage = 'Batch Failed'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}`,
				},
			},
		},
		{
			name: "Missing Actual Record Count param",
			params: map[string]interface{}{
				param.InvalidRecordCount: float64(2),
				param.FailureMessage:     "Batch Failed",
			},
			expectedErr: response.MissingParams(param.ActualRecordCount),
		},
		{
			name: "Missing Invalid Record Count param",
			params: map[string]interface{}{
				param.ActualRecordCount: float64(10),
				param.FailureMessage:    "Batch Failed",
			},
			expectedErr: response.MissingParams(param.InvalidRecordCount),
		},
		{
			name: "Missing Failure Message param",
			params: map[string]interface{}{
				param.ActualRecordCount:  float64(10),
				param.InvalidRecordCount: float64(2),
			},
			expectedErr: response.MissingParams(param.FailureMessage),
		},
	}

	fail := Fail{}
	logger := log.New(os.Stdout, fmt.Sprintf("batches/%s: ", fail.GetAction()), log.Llongfile)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request, errResp := fail.GetUpdateScript(tt.params, param.ParamValidator{}, tt.claims, logger)
			if !reflect.DeepEqual(errResp, tt.expectedErr) {
				t.Errorf("GetUpdateScript().errResp = '%v', expected '%v'", errResp, tt.expectedErr)
			} else if tt.expectedRequest != nil {
				if err := RequestCompareScriptTest(tt.expectedRequest, request); err != nil {
					t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", request, tt.expectedRequest, err)
				}
			}

		})
	}

}

func TestUpdateStatus_Fail(t *testing.T) {
	activationId := "activationId"
	_ = os.Setenv(response.EnvOwActivationId, activationId)

	const (
		scriptFail string = `{"script":{"source":"if \(ctx\._source\.status != 'terminated' && ctx\._source\.status != 'failed'\) {ctx\._source\.status = 'failed'; ctx\._source\.actualRecordCount = 10; ctx\._source\.invalidRecordCount = 2; ctx\._source\.failureMessage = 'Batch Failed'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
	)

	validClaims := auth.HriClaims{Scope: auth.HriInternal, Subject: "internalId"}

	failedBatch := map[string]interface{}{
		param.BatchId:             batchId,
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
		param.BatchId:             batchId,
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
		params               map[string]interface{}
		claims               auth.HriClaims
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedResponse     map[string]interface{}
	}{
		{
			name: "simple fail",
			params: map[string]interface{}{
				path.ParamOwPath:         "/hri/tenants/1234/batches/test-batch/action/fail",
				param.ActualRecordCount:  batchActualRecordCount,
				param.InvalidRecordCount: batchInvalidRecordCount,
				param.FailureMessage:     batchFailureMessage,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptFail,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, failedJSON),
				},
			),
			expectedNotification: failedBatch,
			expectedResponse:     response.Success(http.StatusOK, map[string]interface{}{}),
		},
		{
			name: "'fail' fails on terminated batch",
			params: map[string]interface{}{
				path.ParamOwPath:         "/hri/tenants/1234/batches/test-batch/action/fail",
				param.ActualRecordCount:  batchActualRecordCount,
				param.InvalidRecordCount: batchInvalidRecordCount,
				param.FailureMessage:     batchFailureMessage,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptFail,
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
			expectedResponse: response.Error(http.StatusConflict, "The 'fail' endpoint failed, batch is in 'terminated' state"),
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

			if result := UpdateStatus(tt.params, param.ParamValidator{}, tt.claims, Fail{}, esClient, writer); !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("UpdateStatus() = \n\t%v, expected: \n\t%v", result, tt.expectedResponse)
			}
		})
	}
}
