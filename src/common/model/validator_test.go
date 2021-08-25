/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

const CustomRegexTag = "custom-regex"

type ValidateTestStruct struct {
	TestStr string `validate:"custom-regex"`
}

type TestEmptyOkBindStruct struct {
	InjectionRejectionJson string `json:"mustBeSqlSafe" validate:"omitempty,injection-check-validator"`
	UntaggedParam          string
}

type TestAlwaysRequiredBindStruct struct {
	RequiredParam string `param:"importantParam" validate:"required"`
}

type TestRequiredWithoutBindStruct struct {
	RequiredWithout        string `param:"thing1" validate:"required_without=RequiredWithoutsFriend"`
	RequiredWithoutsFriend string `param:"thing2"`
}

type TestEmbeddedStruct struct {
	TestAlwaysRequiredBindStruct
	AnotherElement int `json:"regularDegularInt"`
}

func TestValidate(t *testing.T) {
	v, err := GetValidator()
	if err != nil {
		t.Errorf("Unexpected Error Getting Validator:  %s", err.Error())
	}

	justAnExistingStr := "I'm here don't worry"
	testCases := []struct {
		name          string
		boundInput    interface{}
		expectedError error
	}{
		{
			name:       "valid input empty",
			boundInput: TestEmptyOkBindStruct{},
		},
		{
			name: "valid input present",
			boundInput: TestAlwaysRequiredBindStruct{
				RequiredParam: justAnExistingStr,
			},
		},
		{
			name: "valid embedded struct",
			boundInput: TestEmbeddedStruct{
				TestAlwaysRequiredBindStruct: TestAlwaysRequiredBindStruct{
					RequiredParam: justAnExistingStr,
				},
				AnotherElement: 875,
			},
		},
		{
			name:          "missing always required param",
			boundInput:    TestAlwaysRequiredBindStruct{},
			expectedError: errors.New("invalid request arguments:\n- importantParam (url path parameter) is a required field"),
		},
		{
			name: "InjectionCheckValidatorTag (aka injection check) failure",
			boundInput: TestEmptyOkBindStruct{
				InjectionRejectionJson: "[injection]",
			},
			expectedError: errors.New("invalid request arguments:\n- mustBeSqlSafe (json field in request body) must not contain the following characters: \"=<>[]{}"),
		},
		{
			name:          "required without failure",
			boundInput:    TestRequiredWithoutBindStruct{},
			expectedError: errors.New("invalid request arguments:\n- thing1 (url path parameter) must be present if thing2 (url path parameter) is not present"),
		},
		{
			name:          "embedded struct failures",
			boundInput:    TestEmbeddedStruct{},
			expectedError: errors.New("invalid request arguments:\n- importantParam (url path parameter) is a required field"),
		},
		{
			name:          "bad validation input",
			boundInput:    []string{"whoops", "not a struct"},
			expectedError: errors.New("unexpected bad validation input failure validator: (nil []string)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Validate(tc.boundInput)
			if !reflect.DeepEqual(tc.expectedError, err) {
				t.Errorf("Validate\nexpected\n%v\nactual\n%v", tc.expectedError, err)
			}
		})
	}
}

func TestFindAndReplaceStructFieldNames(t *testing.T) {
	testCases := []struct {
		name           string
		structInstance interface{}
		originalStr    string
		expectedStr    string
	}{
		{
			name:           "no replacements necessary",
			structInstance: TestEmptyOkBindStruct{},
			originalStr:    "invalid request arguments:\n- importantParam (url path parameter) is a required field",
			expectedStr:    "invalid request arguments:\n- importantParam (url path parameter) is a required field",
		},
		{
			name:           "replacement necessary",
			structInstance: TestEmptyOkBindStruct{},
			originalStr:    "invalid request arguments:\n- mustBeCustomAlphaNum (request query parameter) must be equal to " + structFieldMarker + "InjectionRejectionJson]",
			expectedStr:    "invalid request arguments:\n- mustBeCustomAlphaNum (request query parameter) must be equal to mustBeSqlSafe (json field in request body)",
		},
		{
			name:           "replacing nonexistent struct param",
			structInstance: TestEmptyOkBindStruct{},
			originalStr:    "invalid request arguments:\n- mustBeCustomAlphaNum (request query parameter) must be equal to " + structFieldMarker + "SpookyVar]",
			expectedStr:    "invalid request arguments:\n- mustBeCustomAlphaNum (request query parameter) must be equal to SpookyVar (error: field is not part of model.TestEmptyOkBindStruct definition)",
		},
		{
			name:           "replacing param with no tagged name",
			structInstance: TestEmptyOkBindStruct{},
			originalStr:    "invalid request arguments:\n- mustBeCustomAlphaNum (request query parameter) must be equal to " + structFieldMarker + "UntaggedParam]",
			expectedStr:    "invalid request arguments:\n- mustBeCustomAlphaNum (request query parameter) must be equal to UntaggedParam (field exists in struct but has no tag)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualStr := findAndReplaceStructFieldNames(tc.originalStr, tc.structInstance)
			assert.Equal(t, tc.expectedStr, actualStr)
		})
	}
}

func TestValidateInjectionCheck(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		isValid bool
	}{
		{
			name:    "valid-string-AlphaNumeric",
			input:   "datatypeNoSpecialChars1245",
			isValid: true,
		},
		{
			name:    "valid-string-DashUnderscoreDot",
			input:   "valid-string_with.dot",
			isValid: true,
		},
		{
			name:    "valid-string-UTF8",
			input:   "valid-string_with中文",
			isValid: true,
		},
		{
			name:    "invalid-string-1",
			input:   "invalid-._{}$%##",
			isValid: false,
		},
		{
			name:    "invalid-string-2",
			input:   "INvalID889-[]()*&^",
			isValid: false,
		},
		{
			name:    "invalid-string-3",
			input:   "INvalID889-=()*&^",
			isValid: false,
		},
		{
			name:    "invalid-string-4",
			input:   "INvalID889-<>()*&^",
			isValid: false,
		},
	}

	validate := validator.New()
	validate.RegisterValidation(CustomRegexTag, InjectionCheckValidator)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := ValidateTestStruct{TestStr: tc.input}
			err := validate.Struct(s)
			if tc.isValid && err != nil {
				t.Errorf("Unexpected Error. Expected validation result (No Error)| Actual Error Result: [%v]", err)
			} else if !tc.isValid && err == nil {
				t.Errorf("Did NOT get Expected Error. Expected Error Returned for string match: [%v]", s.TestStr)
			}
		})
	}
}

func TestValidateTenantId(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		isValid bool
	}{
		{
			name:    "Good with alpha-numeric",
			input:   "tenant1234",
			isValid: true,
		},
		{
			name:    "Good with -",
			input:   "tenant-1234",
			isValid: true,
		},
		{
			name:    "Good with _",
			input:   "tenant_1234",
			isValid: true,
		},
		{
			name:    "Error with capital",
			input:   "tenantId1234",
			isValid: false,
		},
		{
			name:    "Error with $",
			input:   "tenant$1234",
			isValid: false,
		},
		{
			name:    "Error with .",
			input:   "tenant.1234",
			isValid: false,
		},
	}

	validate := validator.New()
	validate.RegisterValidation(CustomRegexTag, TenantIdValidator)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := ValidateTestStruct{TestStr: tc.input}
			err := validate.Struct(s)
			if tc.isValid && err != nil {
				t.Errorf("Unexpected Error. Expected validation result (No Error)| Actual Error Result: [%v]", err)
			} else if !tc.isValid && err == nil {
				t.Errorf("Did NOT get Expected Error. Expected Error Returned for string match: [%v]", s.TestStr)
			}
		})
	}
}

func TestValidateStreamId(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		isValid bool
	}{
		{
			name:    "Good with alpha-numeric",
			input:   "stream1234",
			isValid: true,
		},
		{
			name:    "Good with -",
			input:   "stream-1234",
			isValid: true,
		},
		{
			name:    "Good with _",
			input:   "stream_1234",
			isValid: true,
		},
		{
			name:    "Error with capital",
			input:   "streamId1234",
			isValid: false,
		},
		{
			name:    "Error with ~",
			input:   "stream~1234",
			isValid: false,
		},
		{
			name:    "Good with one.",
			input:   "stream.1234",
			isValid: true,
		},
		{
			name:    "Error with 2 .",
			input:   "stream.123.4",
			isValid: false,
		},
	}

	validate := validator.New()
	validate.RegisterValidation(CustomRegexTag, StreamIdValidator)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := ValidateTestStruct{TestStr: tc.input}
			err := validate.Struct(s)
			if tc.isValid && err != nil {
				t.Errorf("Unexpected Error. Expected validation result (No Error)| Actual Error Result: [%v]", err)
			} else if !tc.isValid && err == nil {
				t.Errorf("Did NOT get Expected Error. Expected Error Returned for string match: [%v]", s.TestStr)
			}
		})
	}
}
