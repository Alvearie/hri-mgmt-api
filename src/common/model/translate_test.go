/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package model

import (
	"errors"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestRegisterCustomTranslations(t *testing.T) {
	en := en.New()
	uni := ut.New(en, en)
	translator, found := uni.GetTranslator("en")
	assert.Equal(t, true, found)

	validate := validator.New()
	validate.RegisterValidation(InjectionCheckValidatorTag, InjectionCheckValidator)
	validate.RegisterValidation(TenantIdValidatorTag, TenantIdValidator)
	validate.RegisterValidation(StreamIdValidatorTag, StreamIdValidator)
	validate.RegisterTagNameFunc(getNameFromStructField) // replace struct field names with tag names

	// Add built in translations from the validator library
	entranslations.RegisterDefaultTranslations(validate, translator)

	// Add custom translations for custom validations and overwrites of existing translations
	err := RegisterCustomTranslations(validate, translator)
	assert.NoError(t, err)

	v := CustomValidator{
		Validator:  validate,
		translator: translator,
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
