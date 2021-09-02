/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package batches

import (
	"reflect"
	"testing"
)

func TestEsDocToBatch(t *testing.T) {
	tests := []struct {
		name  string
		esDoc map[string]interface{}
		want  map[string]interface{}
	}{
		{"example1",
			map[string]interface{}{
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
			map[string]interface{}{
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
			if got := EsDocToBatch(tt.esDoc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EsDocToBatch() = %v, want %v", got, tt.want)
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
		{"set-record-count-from-expected-record-count",
			map[string]interface{}{
				"id":                  "1",
				"name":                "batch-2019-10-07",
				"topic":               "ingest.1.fhir",
				"dataType":            "claims",
				"integratorId":        "dataIntegrator1",
				"status":              "started",
				"expectedRecordCount": 100,
				"startDate":           "2019-10-30T12:34:00Z",
			},
			map[string]interface{}{
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
		{"set-expected-record-count-from-record-count",
			map[string]interface{}{
				"id":           "1",
				"name":         "batch-2019-10-07",
				"topic":        "ingest.1.fhir",
				"dataType":     "claims",
				"integratorId": "dataIntegrator1",
				"status":       "started",
				"recordCount":  100,
				"startDate":    "2019-10-30T12:34:00Z",
			},
			map[string]interface{}{
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
		{"no-change-when-neither-set",
			map[string]interface{}{
				"id":           "1",
				"name":         "batch-2019-10-07",
				"topic":        "ingest.1.fhir",
				"dataType":     "claims",
				"integratorId": "dataIntegrator1",
				"status":       "started",
				"startDate":    "2019-10-30T12:34:00Z",
			},
			map[string]interface{}{
				"id":           "1",
				"name":         "batch-2019-10-07",
				"topic":        "ingest.1.fhir",
				"dataType":     "claims",
				"integratorId": "dataIntegrator1",
				"status":       "started",
				"startDate":    "2019-10-30T12:34:00Z",
			},
		},
		{"set-expected-record-count-from-record-count-0",
			map[string]interface{}{
				"id":           "1",
				"name":         "batch-2019-10-07",
				"topic":        "ingest.1.fhir",
				"dataType":     "claims",
				"integratorId": "dataIntegrator1",
				"status":       "started",
				"recordCount":  0,
				"startDate":    "2019-10-30T12:34:00Z",
			},
			map[string]interface{}{
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
			if got := NormalizeBatchRecordCountValues(tt.inputBatch); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("EsDocToBatch() = %v, want %v", got, tt.expected)
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
