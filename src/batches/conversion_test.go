/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"ibm.com/watson/health/foundation/hri/batches/status"
	"reflect"
	"testing"
)

func TestEsDocToBatch(t *testing.T) {
	tests := []struct {
		name     string
		esDoc    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "example1",
			esDoc: map[string]interface{}{
				"_index":        "test-batches",
				"_type":         "_doc",
				"_id":           "1",
				"_version":      1,
				"_seq_no":       0,
				"_primary_term": 1,
				"found":         true,
				"_source": map[string]interface{}{
					"name":         "batch-2019-10-07",
					"topic":        "ingest.1.fhir",
					"dataType":     "claims",
					"integratorId": "dataIntegrator1",
					"status":       "started",
					"recordCount":  100,
					"startDate":    "2019-10-30T12:34:00Z",
				},
			},
			expected: map[string]interface{}{
				"id":                  "1",
				"name":                "batch-2019-10-07",
				"topic":               "ingest.1.fhir",
				"dataType":            "claims",
				"integratorId":        "dataIntegrator1",
				"status":              "started",
				"recordCount":         100,
				"expectedRecordCount": 100,
				"startDate":           "2019-10-30T12:34:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if batch := EsDocToBatch(tt.esDoc); !reflect.DeepEqual(batch, tt.expected) {
				t.Errorf("EsDocToBatch() = %v, expected %v", batch, tt.expected)
			}
		})
	}
}

func TestExtractBatchStatus(t *testing.T) {
	tests := []struct {
		name           string
		inputBatch     map[string]interface{}
		expectedStatus status.BatchStatus
		expectedErr    string
	}{
		{
			name: "success extracting batch status 'started'",
			inputBatch: map[string]interface{}{
				"id":                  "1",
				"name":                "batch-2020-10-23",
				"topic":               "ingest.1.fhir",
				"dataType":            "claims",
				"integratorId":        "dataIntegrator1",
				"status":              "started",
				"recordCount":         20,
				"expectedRecordCount": 20,
				"startDate":           "2019-10-30T12:34:00Z",
			},
			expectedStatus: status.Started,
		},
		{
			name: "success extracting batch status 'failed'",
			inputBatch: map[string]interface{}{
				"id":                  "3",
				"name":                "batch-terminated",
				"topic":               "ingest.3.fhir",
				"integratorId":        "dataIntegrator2",
				"status":              "failed",
				"recordCount":         20,
				"expectedRecordCount": 20,
				"startDate":           "2021-07-12T12:34:00Z",
			},
			expectedStatus: status.Failed,
		},
		{
			name: "extract batch status failed missing status field",
			inputBatch: map[string]interface{}{
				"id":                  "1",
				"name":                "batch-2020-10-23",
				"topic":               "ingest.1.fhir",
				"dataType":            "claims",
				"recordCount":         20,
				"expectedRecordCount": 20,
				"startDate":           "2019-10-30T12:34:00Z",
			},
			expectedStatus: status.Unknown,
			expectedErr:    "Error extracting Batch Status: 'status' field missing",
		},
		{
			name: "extract batch status failed missing status field",
			inputBatch: map[string]interface{}{
				"id":                  "1",
				"name":                "batch-2020-10-23",
				"topic":               "ingest.1.fhir",
				"dataType":            "claims",
				"integratorId":        "dataIntegrator1",
				"status":              "BellBivDeVoe",
				"recordCount":         20,
				"expectedRecordCount": 20,
				"startDate":           "2019-10-30T12:34:00Z",
			},
			expectedErr: "Error extracting Batch Status: Invalid 'status' value: BellBivDeVoe",
		},
		{
			name:        "error extracting batch status from empty Batch map",
			inputBatch:  map[string]interface{}{},
			expectedErr: "Error extracting Batch Status: 'status' field missing",
		},
		{
			name:        "error extracting batch status from nil Batch map",
			inputBatch:  nil,
			expectedErr: "Error extracting Batch Status: 'status' field missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualBatchStatus, err := ExtractBatchStatus(tt.inputBatch)
			if len(tt.expectedErr) > 0 && (err == nil || err.Error() != tt.expectedErr) {
				t.Errorf("ExtractBatchStatus() Error does Not Match expected Error = %v, expected %v", err, tt.expectedErr)
			}
			if actualBatchStatus != tt.expectedStatus {
				t.Errorf("ExtractBatchStatus() = %v, expected %v", actualBatchStatus, tt.expectedStatus)
			}
		})
	}
}

func TestNormalizeBatchRecordCountValues(t *testing.T) {
	tests := []struct {
		name       string
		inputBatch map[string]interface{}
		expected   map[string]interface{}
	}{
		{
			name: "set-record-count-from-expected-record-count",
			inputBatch: map[string]interface{}{
				"id":                  "1",
				"name":                "batch-2019-10-07",
				"topic":               "ingest.1.fhir",
				"dataType":            "claims",
				"integratorId":        "dataIntegrator1",
				"status":              "started",
				"expectedRecordCount": 100,
				"startDate":           "2019-10-30T12:34:00Z",
			},
			expected: map[string]interface{}{
				"id":                  "1",
				"name":                "batch-2019-10-07",
				"topic":               "ingest.1.fhir",
				"dataType":            "claims",
				"integratorId":        "dataIntegrator1",
				"status":              "started",
				"recordCount":         100,
				"expectedRecordCount": 100,
				"startDate":           "2019-10-30T12:34:00Z",
			},
		},
		{
			name: "set-expected-record-count-from-record-count",
			inputBatch: map[string]interface{}{
				"id":           "1",
				"name":         "batch-2019-10-07",
				"topic":        "ingest.1.fhir",
				"dataType":     "claims",
				"integratorId": "dataIntegrator1",
				"status":       "started",
				"recordCount":  100,
				"startDate":    "2019-10-30T12:34:00Z",
			},
			expected: map[string]interface{}{
				"id":                  "1",
				"name":                "batch-2019-10-07",
				"topic":               "ingest.1.fhir",
				"dataType":            "claims",
				"integratorId":        "dataIntegrator1",
				"status":              "started",
				"recordCount":         100,
				"expectedRecordCount": 100,
				"startDate":           "2019-10-30T12:34:00Z",
			},
		},
		{
			name: "no-change-when-neither-set",
			inputBatch: map[string]interface{}{
				"id":           "1",
				"name":         "batch-2019-10-07",
				"topic":        "ingest.1.fhir",
				"dataType":     "claims",
				"integratorId": "dataIntegrator1",
				"status":       "started",
				"startDate":    "2019-10-30T12:34:00Z",
			},
			expected: map[string]interface{}{
				"id":           "1",
				"name":         "batch-2019-10-07",
				"topic":        "ingest.1.fhir",
				"dataType":     "claims",
				"integratorId": "dataIntegrator1",
				"status":       "started",
				"startDate":    "2019-10-30T12:34:00Z",
			},
		},
		{
			name: "set-expected-record-count-from-record-count-0",
			inputBatch: map[string]interface{}{
				"id":           "1",
				"name":         "batch-2019-10-07",
				"topic":        "ingest.1.fhir",
				"dataType":     "claims",
				"integratorId": "dataIntegrator1",
				"status":       "started",
				"recordCount":  0,
				"startDate":    "2019-10-30T12:34:00Z",
			},
			expected: map[string]interface{}{
				"id":                  "1",
				"name":                "batch-2019-10-07",
				"topic":               "ingest.1.fhir",
				"dataType":            "claims",
				"integratorId":        "dataIntegrator1",
				"status":              "started",
				"recordCount":         0,
				"expectedRecordCount": 0,
				"startDate":           "2019-10-30T12:34:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if values := NormalizeBatchRecordCountValues(tt.inputBatch); !reflect.DeepEqual(values, tt.expected) {
				t.Errorf("EsDocToBatch() = %v, expected %v", values, tt.expected)
			}
		})
	}
}

func TestInputTopicToNotificationTopic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "has-input-suffix",
			input:    "ingest.1.claims.in",
			expected: "ingest.1.claims.notification",
		},
		{
			name:     "missing-input-suffix",
			input:    "ingest.1.claims.foo",
			expected: "ingest.1.claims.foo.notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if actual := InputTopicToNotificationTopic(tt.input); actual != tt.expected {
				t.Errorf("InputTopicToNotification() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}
