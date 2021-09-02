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

func TestTerminate_CheckAuth(t *testing.T) {
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

	terminate := Terminate{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := terminate.CheckAuth(tt.claims)
			if (err == nil && tt.expectedErr != "") || (err != nil && err.Error() != tt.expectedErr) {
				t.Errorf("GetAuth() = '%v', expected '%v'", err, tt.expectedErr)
			}
		})
	}
}

func TestTerminate_GetUpdateScript(t *testing.T) {
	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: "integratorId"}

	tests := []struct {
		name   string
		params map[string]interface{}
		claims auth.HriClaims
		// Note that the following chars must be escaped because expectedScript is used as a regex pattern: ., ), (
		expectedRequest map[string]interface{}
		expectedErr     map[string]interface{}
	}{
		{
			name:   "UpdateScript returns expected script without metadata",
			params: map[string]interface{}{},
			claims: validClaims,
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}`,
				},
			},
		},
		{
			name: "UpdateScript returns expected script with metadata",
			params: map[string]interface{}{
				param.Metadata: map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5},
			},
			claims: validClaims,
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `'; ctx\._source\.metadata = params\.metadata;} else {ctx\.op = 'none'}`,
					"lang":   "painless",
					"params": map[string]interface{}{"metadata": map[string]interface{}{"compression": "gzip", "userMetaField1": "metadata", "userMetaField2": -5}},
				},
			},
		},
		{
			name:   "Missing claim.Subject",
			params: map[string]interface{}{},
			claims: auth.HriClaims{},
			expectedRequest: map[string]interface{}{
				"script": map[string]interface{}{
					"source": `if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == ''\) {ctx\._source\.status = 'sendCompleted'; ctx\._source\.expectedRecordCount = 200;} else {ctx\.op = 'none'}`,
				},
			},
		},
		{
			name: "Bad Metadata type",
			params: map[string]interface{}{
				param.Metadata: "nil",
			},
			claims:      auth.HriClaims{},
			expectedErr: response.InvalidParams("metadata must be a map, got string instead."),
		},
	}

	terminate := Terminate{}
	logger := log.New(os.Stdout, fmt.Sprintf("batches/%s: ", terminate.GetAction()), log.Llongfile)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			request, errResp := terminate.GetUpdateScript(tt.params, param.ParamValidator{}, tt.claims, logger)
			if !reflect.DeepEqual(errResp, tt.expectedErr) {
				t.Errorf("GetUpdateScript().errResp = '%v', expected '%v'", errResp, tt.expectedErr)
			} else if tt.expectedRequest != nil {
				if err := RequestCompareScriptTest(tt.expectedRequest, request); err != nil {
					t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", request, tt.expectedRequest, err)
				} else if tt.params[param.Metadata] != nil {
					if err := RequestCompareWithMetadataTest(tt.expectedRequest, request); err != nil {
						t.Errorf("GetUpdateScript().udpateRequest = \n\t'%s' \nDoesn't match expected \n\t'%s'\n%v", request, tt.expectedRequest, err)
					}
				}
			}
		})
	}

}

func TestUpdateStatus_Terminate(t *testing.T) {
	activationId := "activationId"
	_ = os.Setenv(response.EnvOwActivationId, activationId)

	const (
		scriptTerminate string = `{"script":{"source":"if \(ctx\._source\.status == 'started' && ctx\._source\.integratorId == 'integratorId'\) {ctx\._source\.status = 'terminated'; ctx\._source\.endDate = '` + test.DatePattern + `';} else {ctx\.op = 'none'}"}}` + "\n"
	)

	validClaims := auth.HriClaims{Scope: auth.HriIntegrator, Subject: "integratorId"}

	terminatedBatch := map[string]interface{}{
		param.BatchId:             batchId,
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
		params               map[string]interface{}
		claims               auth.HriClaims
		ft                   *test.FakeTransport
		writerError          error
		expectedNotification map[string]interface{}
		expectedResponse     map[string]interface{}
	}{
		{
			name: "simple Terminate",
			params: map[string]interface{}{
				path.ParamOwPath: "/hri/tenants/1234/batches/test-batch/action/terminate",
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminate,
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
			name: "Terminate fails on terminated batch",
			params: map[string]interface{}{
				path.ParamOwPath:          "/hri/tenants/1234/batches/test-batch/action/terminate",
				param.Validation:          true,
				param.ExpectedRecordCount: batchExpectedRecordCount,
			},
			claims: validClaims,
			ft: test.NewFakeTransport(t).AddCall(
				"/1234-batches/_doc/test-batch/_update",
				test.ElasticCall{
					RequestQuery: transportQueryParams,
					RequestBody:  scriptTerminate,
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
			expectedResponse: response.Error(http.StatusConflict, "The 'terminate' endpoint failed, batch is in 'terminated' state"),
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

			if result := UpdateStatus(tt.params, param.ParamValidator{}, tt.claims, Terminate{}, esClient, writer); !reflect.DeepEqual(result, tt.expectedResponse) {
				t.Errorf("UpdateStatus() = \n\t%v,\nexpected: \n\t%v", result, tt.expectedResponse)
			}
		})
	}
}
