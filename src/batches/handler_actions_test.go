// /*
//  * (C) Copyright IBM Corp. 2021
//  *
//  * SPDX-License-Identifier: Apache-2.0
//  */

package batches

// import (
// 	"reflect"
// 	"testing"

// 	"github.com/Alvearie/hri-mgmt-api/batches/status"
// 	"github.com/Alvearie/hri-mgmt-api/common/auth"
// 	"github.com/Alvearie/hri-mgmt-api/common/kafka"
// 	"github.com/Alvearie/hri-mgmt-api/common/model"
// 	"github.com/Alvearie/hri-mgmt-api/common/test"
// 	"github.com/elastic/go-elasticsearch/v7"
// )

// type fakeAction struct {
// 	t               *testing.T
// 	expectedRequest interface{}
// 	expectedStatus  status.BatchStatus
// 	code            int
// 	body            interface{}
// }

// func (fake fakeAction) sendComplete(_ string, request *model.SendCompleteRequest, _ auth.HriAzClaims, _ *elasticsearch.Client, _ kafka.Writer, currentStatus status.BatchStatus) (int, interface{}) {
// 	if !reflect.DeepEqual(fake.expectedRequest, request) {
// 		fake.t.Errorf("Request is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedRequest, request)
// 	}
// 	if fake.expectedStatus != currentStatus {
// 		fake.t.Errorf("Current Batch Status is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedStatus, currentStatus)
// 	}
// 	return fake.code, fake.body
// }

// func (fake fakeAction) terminate(_ string, request *model.TerminateRequest, _ auth.HriAzClaims, _ *elasticsearch.Client, _ kafka.Writer, currentStatus status.BatchStatus) (int, interface{}) {
// 	if !reflect.DeepEqual(fake.expectedRequest, request) {
// 		fake.t.Errorf("Request is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedRequest, request)
// 	}
// 	if fake.expectedStatus != currentStatus {
// 		fake.t.Errorf("Current Batch Status is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedStatus, currentStatus)
// 	}
// 	return fake.code, fake.body
// }

// func (fake fakeAction) processingComplete(_ string, request *model.ProcessingCompleteRequest, _ auth.HriAzClaims, _ *elasticsearch.Client, _ kafka.Writer, currentStatus status.BatchStatus) (int, interface{}) {
// 	if !reflect.DeepEqual(fake.expectedRequest, request) {
// 		fake.t.Errorf("Request is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedRequest, request)
// 	}
// 	if fake.expectedStatus != currentStatus {
// 		fake.t.Errorf("Current Batch Status is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedStatus, currentStatus)
// 	}
// 	return fake.code, fake.body
// }

// func (fake fakeAction) fail(_ string, request *model.FailRequest, _ auth.HriAzClaims, _ *elasticsearch.Client, _ kafka.Writer, currentStatus status.BatchStatus) (int, interface{}) {
// 	if !reflect.DeepEqual(fake.expectedRequest, request) {
// 		fake.t.Errorf("Request is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedRequest, request)
// 	}
// 	if fake.expectedStatus != currentStatus {
// 		fake.t.Errorf("Current Batch Status is not equal expected:\n\tExpected: %v\n\tActual:   %v", fake.expectedStatus, currentStatus)
// 	}
// 	return fake.code, fake.body
// }

// const topicBase = "awesomeTopic"
// const defaultTenantId = test.ValidTenantId
// const defaultBatchId = test.ValidBatchId
// const defaultBatchName = "monkeeBatch2"
// const defaultBatchDataType = "claims"
// const defaultTopicBase = "awesomeTopic"
// const defaultInputTopic = topicBase + inputSuffix
// const defaultBatchStatus = status.Started //Started
