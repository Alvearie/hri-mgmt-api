/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"github.com/Alvearie/hri-mgmt-api/batches/status"
)

const (
	requestId                string = "the-request-id"
	batchName                string = "batchName"
	integratorId             string = "integratorId"
	batchTopic               string = "test.batch.in"
	batchDataType            string = "batchDataType"
	batchStartDate           string = "ignored"
	batchExpectedRecordCount        = float64(10)
	batchActualRecordCount          = float64(10)
	batchInvalidThreshold           = float64(5)
	batchInvalidRecordCount         = float64(2)
	batchFailureMessage      string = "Batch Failed"
	transportQueryParams     string = "_source=true"
	currentStatus                   = status.Started
)

// func Test_updateStatus(t *testing.T) {
// 	const requestId = "a-request-id"
// 	var updateRequest = map[string]interface{}{"script": map[string]interface{}{"source": "update script"}}
// 	const currentStatus = status.Started
// 	var revertScript = fmt.Sprintf("{ctx._source.status = '%s';}", currentStatus)
// 	logwrapper.Initialize("info", os.Stdout)

// 	batch := map[string]interface{}{
// 		param.BatchId:             test.ValidBatchId,
// 		param.Name:                batchName,
// 		param.IntegratorId:        integratorId,
// 		param.Topic:               batchTopic,
// 		param.DataType:            batchDataType,
// 		param.Status:              status.Completed.String(),
// 		param.StartDate:           batchStartDate,
// 		param.RecordCount:         batchExpectedRecordCount,
// 		param.ExpectedRecordCount: batchExpectedRecordCount,
// 		param.ActualRecordCount:   batchActualRecordCount,
// 		param.InvalidThreshold:    batchInvalidThreshold,
// 		param.InvalidRecordCount:  batchInvalidRecordCount,
// 	}
// 	completedJSON, err := json.Marshal(batch)
// 	if err != nil {
// 		t.Errorf("Unable to create batch JSON string: %s", err.Error())
// 	}

// 	tests := []struct {
// 		name                 string
// 		ft                   *test.FakeTransport
// 		writerError          error
// 		expectedNotification map[string]interface{}
// 		expectedBatch        map[string]interface{}
// 		expectedError        *response.ErrorDetailResponse
// 		currentStatus        status.BatchStatus
// 	}{
// 		{
// 			name: "success",
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  "update script",
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, completedJSON),
// 				},
// 			),
// 			expectedNotification: batch,
// 			expectedBatch:        nil,
// 			expectedError:        nil,
// 			currentStatus:        currentStatus,
// 		},
// 		// Can't find a way to make the script encoding fail -> elastic.EncodeQueryBody(updateRequest)
// 		// So there's no unit test for the failure check.
// 		{
// 			name: "fail when update result not returned by elastic",
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, completedJSON),
// 				},
// 			),
// 			expectedBatch: nil,
// 			expectedError: response.NewErrorDetailResponse(http.StatusInternalServerError, requestId,
// 				"update result not returned in Elastic response"),
// 		},
// 		{
// 			name: "fail when updated document not returned by elastic",
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated"
// 						}`, test.ValidTenantId, test.ValidBatchId),
// 				},
// 			),
// 			expectedBatch: nil,
// 			expectedError: response.NewErrorDetailResponse(http.StatusInternalServerError, requestId,
// 				"updated document not returned in Elastic response: error extracting the get section of the JSON"),
// 		},
// 		{
// 			name: "fail when elastic result is unrecognized or invalid",
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "MOnkeez-bad-result",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, completedJSON),
// 				},
// 			),
// 			expectedBatch: nil,
// 			expectedError: response.NewErrorDetailResponse(http.StatusInternalServerError, requestId,
// 				"an unexpected error occurred updating the batch, Elastic update returned result 'MOnkeez-bad-result'"),
// 		},
// 		{
// 			name: "fail on nonexistent tenant",
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery:       transportQueryParams,
// 					ResponseStatusCode: http.StatusNotFound,
// 					ResponseBody: `
// 						{
// 							"error": {
// 								"type": "index_not_found_exception",
// 								"reason": "no such index and [action.auto_create_index] is [false]",
// 								"index": "tenant-that-doesnt-exist-batches"
// 							},
// 							"status": 404
// 						}`,
// 				},
// 			),
// 			expectedBatch: nil,
// 			expectedError: response.NewErrorDetailResponse(http.StatusNotFound, requestId,
// 				"could not update the status of batch test-batch: [404] index_not_found_exception: no such index and [action.auto_create_index] is [false]"),
// 		},
// 		{
// 			name: "original batch is returned when result is noop",
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  "update script",
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "noop",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, completedJSON),
// 				},
// 			),
// 			expectedBatch: batch,
// 			currentStatus: currentStatus,
// 		},
// 		{
// 			name: "revert Batch Status (in Elastic) Succeeds when (Kafka) notification msg send fails",
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  "update script",
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, completedJSON),
// 				},
// 			).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestBody: revertScript,
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, completedJSON),
// 				},
// 			),
// 			expectedNotification: batch,
// 			writerError:          errors.New("unable to write to Kafka"),
// 			expectedError: response.NewErrorDetailResponse(http.StatusInternalServerError, requestId,
// 				"error writing batch notification to kafka: unable to write to Kafka"),
// 			currentStatus: currentStatus,
// 		},
// 		{
// 			name: "when notification msg send fails AND 1st Status Revert attempt fails, retry Revert action up to 5 times",
// 			ft: test.NewFakeTransport(t).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestQuery: transportQueryParams,
// 					RequestBody:  "update script",
// 					ResponseBody: fmt.Sprintf(`
// 						{
// 							"_index": "%s-batches",
// 							"_type": "_doc",
// 							"_id": "%s",
// 							"result": "updated",
// 							"get": {
// 								"_source": %s
// 							}
// 						}`, test.ValidTenantId, test.ValidBatchId, completedJSON),
// 				},
// 			).AddCall( // HERE we Add the calls for the *Five (5)* Attempts to Revert the Batch Status to currentStatus = status.Started (in Elastic)
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestBody:        revertScript,
// 					ResponseStatusCode: http.StatusForbidden,
// 					ResponseBody: `
// 						{
// 							"error": {
// 								"type": "some_random_exception_1",
// 								"reason": "Blah, blah, no such index and [action.auto_create_index] is [false]",
// 								"index": "tenant-that-doesnt-exist-batches"
// 							},
// 							"status": 403
// 						}`,
// 				},
// 			).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestBody:        revertScript,
// 					ResponseStatusCode: http.StatusTeapot,
// 					ResponseBody: `
// 						{
// 							"error": {
// 								"type": "some_random_exception_2",
// 								"reason": "Blah, blah, no such index and [action.auto_create_index] is [false]",
// 								"index": "tenant-that-doesnt-exist-batches"
// 							},
// 							"status": 418
// 						}`,
// 				},
// 			).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestBody:        revertScript,
// 					ResponseStatusCode: http.StatusInternalServerError,
// 					ResponseBody: `
// 						{
// 							"error": {
// 								"type": "some_random_exception_3",
// 								"reason": "Blah, blah, no such index and [action.auto_create_index] is [false]",
// 								"index": "tenant-that-doesnt-exist-batches"
// 							},
// 							"status": 500
// 						}`,
// 				},
// 			).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestBody:        revertScript,
// 					ResponseStatusCode: http.StatusConflict,
// 					ResponseBody: `
// 						{
// 							"error": {
// 								"type": "some_random_exception_4",
// 								"reason": "Blah, blah, no such index and [action.auto_create_index] is [false]",
// 								"index": "tenant-that-doesnt-exist-batches"
// 							},
// 							"status": 409
// 						}`,
// 				},
// 			).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestBody:        revertScript,
// 					ResponseStatusCode: http.StatusMethodNotAllowed,
// 					ResponseBody: `
// 						{
// 							"error": {
// 								"type": "some_random_exception_5",
// 								"reason": "Blah, blah, no such index and [action.auto_create_index] is [false]",
// 								"index": "tenant-that-doesnt-exist-batches"
// 							},
// 							"status": 405
// 						}`,
// 				},
// 			).AddCall(
// 				fmt.Sprintf(`/%s-batches/_doc/%s/_update`, test.ValidTenantId, test.ValidBatchId),
// 				test.ElasticCall{
// 					RequestBody:        revertScript,
// 					ResponseStatusCode: http.StatusTeapot,
// 					ResponseBody: `
// 						{
// 							"error": {
// 								"type": "some_random_exception_6",
// 								"reason": "Blah, blah, no such index and [action.auto_create_index] is [false]",
// 								"index": "tenant-that-doesnt-exist-batches"
// 							},
// 							"status": 418
// 						}`,
// 				},
// 			),
// 			expectedNotification: batch,
// 			writerError:          errors.New("unable to write to Kafka"),
// 			expectedError: response.NewErrorDetailResponse(http.StatusInternalServerError, requestId,
// 				"error writing batch notification to kafka: unable to write to Kafka"),
// 			currentStatus: currentStatus,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			esClient, err := elastic.ClientFromTransport(tt.ft)
// 			if err != nil {
// 				t.Error(err)
// 			}
// 			writer := test.FakeWriter{
// 				T:             t,
// 				ExpectedTopic: InputTopicToNotificationTopic(batchTopic),
// 				ExpectedKey:   test.ValidBatchId,
// 				ExpectedValue: tt.expectedNotification,
// 				Error:         tt.writerError,
// 			}

// 			result, errResp := updateStatus(requestId, test.ValidTenantId, test.ValidBatchId, updateRequest, esClient, writer, tt.currentStatus)

// 			tt.ft.VerifyCalls()

// 			if !reflect.DeepEqual(result, tt.expectedBatch) {
// 				t.Errorf("Returned Batch did not match\n\texpected: %v, \n\tactual:   %v", tt.expectedBatch, result)
// 			}
// 			if !reflect.DeepEqual(errResp, tt.expectedError) {
// 				t.Errorf("Returned ErrorDetailResponse did not match\nexpected: %v -> %v\nactual:   %v -> %v",
// 					tt.expectedError, tt.expectedError.Body, errResp, errResp.Body)
// 			}
// 		})
// 	}
// }
