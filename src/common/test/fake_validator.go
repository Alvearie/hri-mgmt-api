/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package test

import (
	"github.com/Alvearie/hri-mgmt-api/common/param"
	"reflect"
	"testing"
)

type FakeValidator struct {
	T        *testing.T
	Required []param.Info
	Optional []param.Info
	Response map[string]interface{}
}

func (fv FakeValidator) Validate(params map[string]interface{}, requiredParams ...param.Info) map[string]interface{} {
	if !reflect.DeepEqual(fv.Required, requiredParams) {
		fv.T.Errorf("Unexpected required params. Expected: [%s], Actual: [%s]", fv.Required, requiredParams)
	}
	return fv.Response
}

func (fv FakeValidator) ValidateOptional(params map[string]interface{}, optionalParams ...param.Info) map[string]interface{} {
	if !reflect.DeepEqual(fv.Optional, optionalParams) {
		fv.T.Errorf("Unexpected optional params. Expected: [%s], Actual: [%s]", fv.Optional, optionalParams)
	}
	return fv.Response
}
