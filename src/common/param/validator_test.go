/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package param

import (
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"reflect"
	"testing"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name     string
		inputs   map[string]interface{}
		required []Info
		expected map[string]interface{}
	}{
		{
			name:     "one-missing",
			inputs:   map[string]interface{}{"p1": "p1"},
			required: []Info{{"p1", reflect.String}, {"p2", reflect.Int}},
			expected: response.MissingParams("p2"),
		},
		{
			name:     "two-missing",
			inputs:   map[string]interface{}{"p1": "v1", "p3": "v3"},
			required: []Info{{"p1", reflect.String}, {"p2", reflect.String}, {"p3", reflect.String}, {"p4", reflect.String}},
			expected: response.MissingParams("p2", "p4"),
		},
		{
			name:     "none-missing",
			inputs:   map[string]interface{}{"p1": "v1", "p2": "v2", "p3": "v3", "p4": "p4"},
			required: []Info{{"p1", reflect.String}, {"p3", reflect.String}, {"p2", reflect.String}, {"p4", reflect.String}},
		},
		{
			name:     "one-invalid",
			inputs:   map[string]interface{}{"p1": 1},
			required: []Info{{"p1", reflect.String}},
			expected: response.InvalidParams(fmt.Sprintf(invalidTypeMsg, "p1", reflect.String, reflect.Int)),
		},
		{
			name:     "two-invalid",
			inputs:   map[string]interface{}{"p1": 3.1, "p2": 3.1, "p3": 3.1},
			required: []Info{{"p1", reflect.Int}, {"p2", reflect.Float32}, {"p3", reflect.Float64}},
			expected: response.InvalidParams(
				fmt.Sprintf(invalidTypeMsg, "p1", reflect.Int, reflect.Float64),
				fmt.Sprintf(invalidTypeMsg, "p2", reflect.Float32, reflect.Float64),
			),
		},
	}

	validator := ParamValidator{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := validator.Validate(tc.inputs, tc.required...)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Unexpected Error. Expected: [%v], Actual: [%v]", tc.expected, actual)
			}
		})
	}
}

func TestValidateOptional(t *testing.T) {
	testCases := []struct {
		name     string
		inputs   map[string]interface{}
		optional []Info
		expected map[string]interface{}
	}{
		{
			name:     "one-missing",
			inputs:   map[string]interface{}{"p1": "v1"},
			optional: []Info{{"p1", reflect.String}, {"p2", reflect.Int}},
		},
		{
			name:     "none-missing",
			inputs:   map[string]interface{}{"p1": "v1", "p2": "v2", "p3": "v3", "p4": "p4"},
			optional: []Info{{"p1", reflect.String}, {"p3", reflect.String}, {"p2", reflect.String}, {"p4", reflect.String}},
		},
		{
			name:     "one-invalid",
			inputs:   map[string]interface{}{"p1": 1},
			optional: []Info{{"p1", reflect.String}},
			expected: response.InvalidParams(fmt.Sprintf(invalidTypeMsg, "p1", reflect.String, reflect.Int)),
		},
		{
			name:     "two-invalid",
			inputs:   map[string]interface{}{"p1": 3.1, "p2": 3.1, "p3": 3.1},
			optional: []Info{{"p1", reflect.Int}, {"p2", reflect.Float32}, {"p3", reflect.Float64}},
			expected: response.InvalidParams(
				fmt.Sprintf(invalidTypeMsg, "p1", reflect.Int, reflect.Float64),
				fmt.Sprintf(invalidTypeMsg, "p2", reflect.Float32, reflect.Float64),
			),
		},
	}

	validator := ParamValidator{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := validator.ValidateOptional(tc.inputs, tc.optional...)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Unexpected Error. Expected: [%v], Actual: [%v]", tc.expected, actual)
			}
		})
	}
}
