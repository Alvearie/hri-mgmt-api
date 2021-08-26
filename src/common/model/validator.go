/**
 * (C) Copyright IBM Corp. 2021
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package model

import (
	"errors"
	"fmt"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	"reflect"
	"sort"
	"strings"
	"unicode"
)

// CustomValidator is our type setting of third party validator (go-playground Validator) that is compatible with Echo
type CustomValidator struct {
	Validator  *validator.Validate
	translator ut.Translator
}

func GetValidator() (*CustomValidator, error) {
	en := en.New()
	uni := ut.New(en, en)
	translator, found := uni.GetTranslator("en")
	if !found {
		return nil, errors.New("could not find the translator")
	}
	validate := validator.New()

	// Register custom validators
	validate.RegisterValidation(InjectionCheckValidatorTag, InjectionCheckValidator)
	validate.RegisterValidation(TenantIdValidatorTag, TenantIdValidator)
	validate.RegisterValidation(StreamIdValidatorTag, StreamIdValidator)
	validate.RegisterTagNameFunc(getNameFromStructField) // replace struct field names with tag names

	// Add built in translations from the validator library
	entranslations.RegisterDefaultTranslations(validate, translator)

	// Add custom translations for custom validations and overwrites of existing translations
	err := RegisterCustomTranslations(validate, translator)
	if err != nil {
		return nil, errors.New("could not register custom translations " + err.Error())
	}

	custom := CustomValidator{
		Validator:  validate,
		translator: translator,
	}
	return &custom, nil
}

func (cv *CustomValidator) Validate(structInstance interface{}) error {
	rawValidationError := cv.Validator.Struct(structInstance)
	if rawValidationError == nil {
		return nil
	}

	// Validator.Struct() will return InvalidValidationError for bad validation input
	if _, ok := rawValidationError.(*validator.InvalidValidationError); ok {
		return errors.New("unexpected bad validation input failure " + rawValidationError.Error())
	}
	// Proper errors will be returned as ValidationErrors ( []FieldError )
	errs := rawValidationError.(validator.ValidationErrors)
	validationErrTranslations := errs.Translate(cv.translator)

	errMsg := "invalid request arguments:\n- "
	// Maintain the same order of errs every time to allow for unit testing
	sortedErrs := make([]string, 0, len(validationErrTranslations))
	for _, errStr := range validationErrTranslations {
		sortedErrs = append(sortedErrs, errStr)
	}
	sort.Strings(sortedErrs)
	errMsg += strings.Join(sortedErrs, "\n- ")
	errMsg = findAndReplaceStructFieldNames(errMsg, structInstance)
	return errors.New(errMsg)
}

// Given a Struct Field, find and return the parameter name in the tag & where in the request it comes from
func getNameFromStructField(fld reflect.StructField) string {
	var possibleArgumentLocations = [][]string{
		{"json", "json field in request body"},
		{"query", "request query parameter"},
		{"param", "url path parameter"},
	}
	for _, location := range possibleArgumentLocations {
		name, exists := fld.Tag.Lookup(location[0])
		if exists {
			return fmt.Sprintf("%s (%s)", name, location[1])
		}
	}
	if len(fld.Name) > 0 {
		return fld.Name + " (field exists in struct but has no tag)"
	}
	return ""
}

// Look for any Struct Field names and replace them with something that is relevant to the user: the corresponding
// argument's name and its location in the request.  The struct fields will appear in the message surrounded by
// '[STRUCT_FIELD' and ']'.
func findAndReplaceStructFieldNames(originalStr string, structInstance interface{}) string {
	splitByMarker := strings.Split(originalStr, structFieldMarker)
	rebuiltStr := splitByMarker[0]
	for _, str := range splitByMarker[1:] {
		splitFieldNameAndErrMsg := strings.SplitN(str, "]", 2)
		structFieldName := splitFieldNameAndErrMsg[0]
		t := reflect.TypeOf(structInstance)
		field, _ := t.FieldByName(structFieldName)
		name := getNameFromStructField(field)
		if len(name) == 0 {
			rebuiltStr += structFieldName + " (error: field is not part of " + reflect.TypeOf(structInstance).String() + " definition)"
		}
		rebuiltStr += getNameFromStructField(field)
		rebuiltStr += splitFieldNameAndErrMsg[1]
	}
	return rebuiltStr
}

// VALIDATION FUNCTIONS TO BE REGISTERED IN CUSTOM VALIDATOR ////////////////

// InjectionCheckValidator A valid String may not contain characters that can be
// used to exploit the query via SQL Injection.  Specifically "=<>[]{}
func InjectionCheckValidator(flv validator.FieldLevel) bool {
	return !strings.ContainsAny(flv.Field().String(), InjectionCheckContainsString)
}

// TenantIdValidator checks that the tenantId contains only lower case characters, numbers, '-', or '_'
func TenantIdValidator(flv validator.FieldLevel) bool {
	tenantId := flv.Field().String()
	for _, r := range tenantId {
		if !(unicode.IsLetter(r) && unicode.IsLower(r)) && !unicode.IsNumber(r) && r != '-' && r != '_' {
			return false
		}
	}
	return true
}

// StreamIdValidator checks that the streamId contains only lower case characters, numbers, '-', '_', and no more than one '.'
func StreamIdValidator(flv validator.FieldLevel) bool {
	streamId := flv.Field().String()
	for _, r := range streamId {
		if !(unicode.IsLetter(r) && unicode.IsLower(r)) && !unicode.IsNumber(r) && r != '-' && r != '_' && r != '.' {
			return false
		}
	}
	if strings.Count(streamId, ".") > 1 {
		return false
	}
	return true
}
