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
)

const invalidTypeMsg string = "%s must be a %s, got %s instead."

type Info struct {
	Name string
	Kind reflect.Kind
}

type Validator interface {
	Validate(params map[string]interface{}, requiredParams ...Info) map[string]interface{}
	ValidateOptional(params map[string]interface{}, optionalParams ...Info) map[string]interface{}
}

type ParamValidator struct{}

func (v ParamValidator) Validate(params map[string]interface{}, requiredParams ...Info) map[string]interface{} {
	missingParams := make([]string, 0, len(requiredParams)+1)
	invalidTypes := make([]string, 0, len(requiredParams)+1)

	for _, info := range requiredParams {
		// check that param exists
		if val := params[info.Name]; val != nil {
			if reflect.TypeOf(val).Kind() != info.Kind {
				invalidTypes = append(invalidTypes, fmt.Sprintf(invalidTypeMsg, info.Name, info.Kind, reflect.TypeOf(val).Kind()))
			}
		} else {
			missingParams = append(missingParams, info.Name)
		}
	}

	if len(missingParams) > 0 {
		return response.MissingParams(missingParams...)
	} else if len(invalidTypes) > 0 {
		return response.InvalidParams(invalidTypes...)
	} else {
		return nil
	}
}

func (v ParamValidator) ValidateOptional(params map[string]interface{}, optionalParams ...Info) map[string]interface{} {
	invalidTypes := make([]string, 0, len(optionalParams)+1)

	for _, info := range optionalParams {
		// check the type for each param that exists
		if val := params[info.Name]; val != nil {
			if reflect.TypeOf(val).Kind() != info.Kind {
				invalidTypes = append(invalidTypes, fmt.Sprintf(invalidTypeMsg, info.Name, info.Kind, reflect.TypeOf(val).Kind()))
			}
		}
	}

	if len(invalidTypes) > 0 {
		return response.InvalidParams(invalidTypes...)
	} else {
		return nil
	}
}
