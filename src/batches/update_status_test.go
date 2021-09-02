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
	"net/http"
	"os"
	"reflect"
	"testing"
)

const (
	batchId          string  = "test-batch"
	batchName        string  = "batchName"
	batchTopic       string  = "test.batch.in"
	batchDataType    string  = "batchDataType"
	batchStartDate   string  = "ignored"
	batchRecordCount float64 = float64(1)
	integratorId     string  = "integratorId"
	// Note that the following chars must be escaped because RequestBody is used as a regex pattern: ., ), (
	scriptSendComplete         string = `{"script":{"source":"if \(ctx\._source\.status == 'started' \\u0026\\u0026 ctx._source.integratorId == '` + integratorId + `'\) {ctx\._source\.status = 'completed'; ctx\._source\.recordCount = 1; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
	scriptSendCompleteMetadata string = `{"script":{"lang":"painless","params":{"metadata":{"compression":"gzip","userMetaField1":"metadata"}},"source":"if \(ctx._source.status == 'started' \\u0026\\u0026 ctx._source.integratorId == '` + integratorId + `'\) {ctx._source.status = 'completed'; ctx._source.recordCount = 1; ctx._source.endDate = '` + test.DatePattern + `'; ctx\._source\.metadata = params\.metadata;} else {ctx.op = 'none'}"}}` + "\n"
	// Note that the following chars must be escaped because RequestBody is used as a regex pattern: ., ), (
	scriptTerminated         string = `{"script":{"source":"if \(ctx._source.status == 'started' \\u0026\\u0026 ctx._source.integratorId == '` + integratorId + `'\) {ctx._source.status = 'terminated'; ctx._source.endDate = '` + test.DatePattern + `';} else {ctx.op = 'none'}"}}` + "\n"
	scriptTerminatedMetadata string = `{"script":{"lang":"painless","params":{"metadata":{"compression":"gzip","userMetaField1":"metadata"}},"source":"if \(ctx\._source\.status == 'started' \\u0026\\u0026 ctx._source.integratorId == '` + integratorId + `'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `'; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}"}}` + "\n"
	transportQueryParams     string = "_source=true"
)

var batchMetadata = map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata"}

func TestUpdateStatus(t *testing.T) {
	activationId := "activationId"
	_ = os.Setenv(response.EnvOwActivationId, activationId)

	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: integratorId}

	// create some example batches in different states
	sendCompleteBatch := createBatch(status.Completed)
	sendCompleteJSON, err := json.Marshal(sendCompleteBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	terminatedBatch := createBatch(status.Terminated)
	terminatedJSON, err := json.Marshal(terminatedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	failedBatch := createBatch(status.Failed)
	failedJSON, err := json.Marshal(failedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	startedBatch := createBatch(status.Started)
	startedJSON, err := json.Marshal(startedBatch)
	if err != nil {
		t.Errorf("Unable to create batch JSON string: %s", err.Error())
	}

	tests := []struct {
		name                 string
		targetStatus         status.BatchStatus
		params               map[string]interface{}
		claims               auth.HriClaims
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedResponse     map[string]interface{}
	}{
		{
			name:         "invalid openwhisk path",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234",
				param.RecordCount: batchRecordCount,
			},
			claims:           validClaims,
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants 1234], requested index: 5"),
		},
		{
			name:         "return error for missing tenantId path param",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants",
				param.RecordCount: batchRecordCount,
			},
			claims:           validClaims,
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants], requested index: 3"),
		},
		{
			name:         "simple sendComplete",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
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
						}`, sendCompleteJSON),
				},
			),
			expectedNotification: sendCompleteBatch,
			expectedResponse:     response.Success(http.StatusOK, map[string]interface{}{}),
		},
		{
			name:         "sendComplete with metadata",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/" + batchId + "/action/sendComplete",
				param.RecordCount: batchRecordCount,
				param.Metadata:    batchMetadata,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptSendCompleteMetadata,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, sendCompleteJSON),
				},
			),
			expectedNotification: sendCompleteBatch,
			expectedResponse:     response.Success(http.StatusOK, map[string]interface{}{}),
		},
		{
			name:         "sendComplete fails on terminated batch",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
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
			expectedResponse: response.Error(http.StatusConflict, "Batch status was not updated to 'completed', batch is already in 'terminated' state"),
		},
		{
			name:         "sendComplete fails on missing record count parameter",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/sendComplete",
			},
			claims:           validClaims,
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.Error(http.StatusBadRequest, "Missing required parameter(s): [recordCount]"),
		},
		{
			name:         "sendComplete fails on bad metadata parameter",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/" + batchId + "/action/sendComplete",
				param.RecordCount: batchRecordCount,
				param.Metadata:    "not json object",
			},
			claims:           validClaims,
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.Error(http.StatusBadRequest, "Invalid parameter type(s): [metadata must be a map, got string instead.]"),
		},
		{
			name:         "sendComplete fails when batch already in completed state",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
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
						}`, sendCompleteJSON),
				},
			),
			expectedResponse: response.Error(http.StatusConflict, "Batch status was not updated to 'completed', batch is already in 'completed' state"),
		},
		{
			name:         "fail when update result not returned by elastic",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
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
							"get": {
								"_source": %s
							}
						}`, sendCompleteJSON),
				},
			),
			expectedResponse: response.Error(http.StatusInternalServerError, "Update result not returned in Elastic response"),
		},
		{
			name:         "fail when updated document not returned by elastic",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptSendComplete,
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
			name:         "fail when elastic result is unrecognized or invalid",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
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
							"result": "MOnkeez-bad-result",
							"get": {
								"_source": %s
							}
						}`, sendCompleteJSON),
				},
			),
			expectedResponse: response.Error(http.StatusInternalServerError, "An unexpected error occurred updating the batch, Elastic update returned result 'MOnkeez-bad-result'"),
		},
		{
			name:         "invalid elastic response",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptSendComplete,
					ResponseBody: `{"_index": "1234-batches",`,
				},
			),
			expectedResponse: response.Error(http.StatusInternalServerError, "Error parsing the Elastic search response body: unexpected EOF"),
		},
		{
			name:         "fail on nonexistent tenant",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/tenant-that-doesnt-exist/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/tenant-that-doesnt-exist-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery:       transportQueryParams,
					RequestBody:        scriptSendComplete,
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
			expectedResponse: response.Error(http.StatusNotFound, "index_not_found_exception: no such index and [action.auto_create_index] is [false]"),
		},
		{
			name:         "fail when updating nonexistent batch",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/batch-that-doesnt-exist/action/sendComplete",
				param.RecordCount: batchRecordCount,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/batch-that-doesnt-exist/_update",
				test.ElasticCall{
					RequestQuery:       transportQueryParams,
					RequestBody:        scriptSendComplete,
					ResponseStatusCode: http.StatusNotFound,
					ResponseBody: `
						{
							"error": {
								"type": "document_missing_exception",
								"reason": "[_doc][batch-that-doesnt-exist]: document missing",
								"index": "1234-batches"
							},
							"status": 404
						}`,
				},
			),
			expectedResponse: response.Error(http.StatusNotFound, "document_missing_exception: [_doc][batch-that-doesnt-exist]: document missing"),
		},
		{
			name:         "fail when unable to send notification",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
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
						}`, sendCompleteJSON),
				},
			),
			expectedNotification: sendCompleteBatch,
			writerError:          errors.New("Unable to write to Kafka"),
			expectedResponse:     response.Error(http.StatusInternalServerError, "Unable to write to Kafka"),
		},
		{
			name:         "simple terminate",
			targetStatus: status.Terminated,
			params:       map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/terminate"},
			claims:       validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminated,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, terminatedJSON),
				},
			),
			expectedNotification: terminatedBatch,
			expectedResponse:     response.Success(http.StatusOK, map[string]interface{}{}),
		},
		{
			name:         "terminate with metadata",
			targetStatus: status.Terminated,
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/terminate",
				param.Metadata:   batchMetadata,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminatedMetadata,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "updated",
							"get": {
								"_source": %s
							}
						}`, terminatedJSON),
				},
			),
			expectedNotification: terminatedBatch,
			expectedResponse:     response.Success(http.StatusOK, map[string]interface{}{}),
		},
		{
			name:         "terminate fails on failed batch",
			targetStatus: status.Terminated,
			params:       map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/terminate"},
			claims:       validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminated,
					ResponseBody: fmt.Sprintf(`
						{
							"_index": "1234-batches",
							"_type": "_doc",
							"_id": "test-batch",
							"result": "noop",
							"get": {
								"_source": %s
							}
						}`, failedJSON),
				},
			),
			expectedResponse: response.Error(http.StatusConflict, "Batch status was not updated to 'terminated', batch is already in 'failed' state"),
		},
		{
			name:         "terminate fails on previously terminated batch",
			targetStatus: status.Terminated,
			params:       map[string]interface{}{path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/terminate"},
			claims:       validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminated,
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
			expectedResponse: response.Error(http.StatusConflict, "Batch status was not updated to 'terminated', batch is already in 'terminated' state"),
		},
		{
			name:         "return error response for Unknown status",
			targetStatus: status.Unknown,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/blargBlarg",
				param.RecordCount: batchRecordCount,
			},
			claims:           validClaims,
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.Error(http.StatusUnprocessableEntity, "Cannot update batch to status 'unknown', only 'completed' and 'terminated' are acceptable"),
		},
		{
			name:         "invalid role",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
			},
			claims:           auth.HriClaims{Scope: auth.HriConsumer, Subject: integratorId},
			ft:               test.NewFakeTransport(t),
			expectedResponse: response.Error(http.StatusUnauthorized, fmt.Sprintf(auth.MsgIntegratorRoleRequired, "update")),
		},
		{
			name:         "complete fails on bad integrator id",
			targetStatus: status.Completed,
			params: map[string]interface{}{
				path.ParamOwPath:  "/hri/tenants/1234/batches/test-batch/action/sendComplete",
				param.RecordCount: batchRecordCount,
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
						}`, startedJSON),
				},
			),
			expectedResponse: response.Error(http.StatusUnauthorized, fmt.Sprintf("Batch status was not updated to 'completed'. Requested by 'bad integrator' but owned by '%s'", integratorId)),
		},
		{
			name:         "terminate fails on bad integrator id",
			targetStatus: status.Terminated,
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/sendComplete",
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
						}`, startedJSON),
				},
			),
			expectedResponse: response.Error(http.StatusUnauthorized, fmt.Sprintf("Batch status was not updated to 'terminated'. Requested by 'bad integrator' but owned by '%s'", integratorId)),
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

			if result := UpdateStatus(tt.params, param.ParamValidator{}, tt.claims, tt.targetStatus, esClient, writer); !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("UpdateStatus() = %v, expected %v", result, tt.expectedResponse)
			}
		})
	}
}

func createBatch(status status.BatchStatus) map[string]interface{} {
	return map[string]interface{}{
		param.BatchId:      batchId,
		param.Name:         batchName,
		param.Topic:        batchTopic,
		param.DataType:     batchDataType,
		param.Status:       status.String(),
		param.StartDate:    batchStartDate,
		param.RecordCount:  batchRecordCount,
		param.Metadata:     batchMetadata,
		param.IntegratorId: integratorId,
	}
}
