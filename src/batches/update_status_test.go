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
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
)

const (
	batchId                  string  = "test-batch"
	batchName                string  = "batchName"
	integratorId             string  = "integratorId"
	batchTopic               string  = "test.batch.in"
	batchDataType            string  = "batchDataType"
	batchStartDate           string  = "ignored"
	batchExpectedRecordCount float64 = float64(10)
	batchActualRecordCount   float64 = float64(10)
	batchInvalidThreshold    float64 = float64(5)
	batchInvalidRecordCount  float64 = float64(2)
	batchFailureMessage      string  = "Batch Failed"
	transportQueryParams     string  = "_source=true"
)

type TestStatusUpdater struct {
	UpdateRequest map[string]interface{}
	AuthResp      error
	ScriptErrResp map[string]interface{}
}

func (TestStatusUpdater) GetAction() string {
	return "testAction"
}

func (t TestStatusUpdater) CheckAuth(_ auth.HriClaims) error {
	return t.AuthResp
}

func (t TestStatusUpdater) GetUpdateScript(_ map[string]interface{}, _ param.Validator, _ auth.HriClaims, _ *log.Logger) (map[string]interface{}, map[string]interface{}) {
	return t.UpdateRequest, t.ScriptErrResp
}

func TestUpdateStatus(t *testing.T) {
	activationId := "activationId"
	_ = os.Setenv(response.EnvOwActivationId, activationId)

	batch := map[string]interface{}{
		param.BatchId:             batchId,
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
	completedJSON, err := json.Marshal(batch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

	tests := []struct {
		name                 string
		statusUpdater        StatusUpdater
		params               map[string]interface{}
		claims               auth.HriClaims
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedResponse     map[string]interface{}
	}{
		{
			name:          "invalid openwhisk path",
			statusUpdater: TestStatusUpdater{},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234",
			},
			claims:           validClaims,
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants 1234], requested index: 5"),
		},
		{
			name:          "return error for missing tenantId path param",
			statusUpdater: TestStatusUpdater{},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants",
			},
			claims:           validClaims,
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants], requested index: 3"),
		},
		{
			name:          "success",
			statusUpdater: TestStatusUpdater{UpdateRequest: map[string]interface{}{"script": map[string]interface{}{"source": "update script"}}},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/testAction",
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  "update script",
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, completedJSON),
				},
			),
			expectedNotification: batch,
			expectedResponse:     response.Success(http.StatusOK, map[string]interface{}{}),
		},
		{
			name:          "StatusUpdater error",
			statusUpdater: TestStatusUpdater{ScriptErrResp: response.MissingParams("test-param")},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/testAction",
			},
			claims:           validClaims,
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.MissingParams("test-param"),
		},
		// Can't find a way to make the script encoding fail -> elastic.EncodeQueryBody(updateRequest)
		// So there's no unit test for the failure check.
		{
			name:          "fail when update result not returned by elastic",
			statusUpdater: TestStatusUpdater{},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/testAction",
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"get": {
								"_source": %s
							}
						}`, completedJSON),
				},
			),
			expectedResponse: response.Error(http.StatusInternalServerError, "Update result not returned in Elastic response"),
		},
		{
			name:          "fail when updated document not returned by elastic",
			statusUpdater: TestStatusUpdater{},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/testAction",
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					ResponseBody: `
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated"
						}`,
				},
			),
			expectedResponse: response.Error(http.StatusInternalServerError, "Updated document not returned in Elastic response: error extracting the get section of the JSON"),
		},
		{
			name:          "fail when elastic result is unrecognized or invalid",
			statusUpdater: TestStatusUpdater{},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/testAction",
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "MOnkeez-bad-result",
							"get": {
								"_source": %s
							}
						}`, completedJSON),
				},
			),
			expectedResponse: response.Error(http.StatusInternalServerError, "An unexpected error occurred updating the batch, Elastic update returned result 'MOnkeez-bad-result'"),
		},
		{
			name:          "fail on nonexistent tenant",
			statusUpdater: TestStatusUpdater{},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/tenant-that-doesnt-exist/batches/test-batch/action/testAction",
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/tenant-that-doesnt-exist-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery:       transportQueryParams,
					ResponseStatusCode: http.StatusNotFound,
					ResponseBody: `
						{
							"error": {
								"type": "index_not_found_exception",
								"reason": "no such index and [action.auto_create_index] is [false]",
								"index": "tenant-that-doesnt-exist-batches"
							},
							"status": 404
						}`,
				},
			),
			expectedResponse: response.Error(http.StatusNotFound,
				"Could not update the status of batch test-batch: index_not_found_exception: no such index and [action.auto_create_index] is [false]"),
		},
		{
			name:          "fail when result is noop",
			statusUpdater: TestStatusUpdater{UpdateRequest: map[string]interface{}{"script": map[string]interface{}{"source": "update script"}}},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/testAction",
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  "update script",
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "noop",
							"get": {
								"_source": %s
							}
						}`, completedJSON),
				},
			),
			expectedNotification: batch,
			expectedResponse:     response.Error(http.StatusConflict, "The 'testAction' endpoint failed, batch is in 'completed' state"),
		},
		{
			name:          "fail when unable to send notification",
			statusUpdater: TestStatusUpdater{},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/testAction",
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, completedJSON),
				},
			),
			expectedNotification: batch,
			writerError:          errors.New("Unable to write to Kafka"),
			expectedResponse:     response.Error(http.StatusInternalServerError, "Unable to write to Kafka"),
		},
		{
			name:          "AuthCheck fails",
			statusUpdater: TestStatusUpdater{AuthResp: errors.New("not authorized")},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/testAction",
			},
			claims:           validClaims,
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.Error(http.StatusUnauthorized, "not authorized"),
		},
		{
			name:          "Update fails on bad integrator id",
			statusUpdater: TestStatusUpdater{},
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/testAction",
			},
			claims: auth.HriClaims{Scope: auth.HriIntegrator, Subject: "bad integrator"},
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "noop",
							"get": {
								"_source": %s
							}
						}`, completedJSON),
				},
			),
			expectedResponse: response.Error(http.StatusUnauthorized, fmt.Sprintf("Batch status was not updated to 'testAction'. Requested by 'bad integrator' but owned by '%s'", integratorId)),
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

			if result := UpdateStatus(tt.params, param.ParamValidator{}, tt.claims, tt.statusUpdater, esClient, writer); !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("UpdateStatus() = %v, expected %v", result, tt.expectedResponse)
			}
		})
	}
}
