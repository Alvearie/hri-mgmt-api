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

type FakeWriter struct {
	T             *testing.T
	ExpectedTopic string
	ExpectedKey   string
	ExpectedValue map[string]interface{}
	Error         error
}

func (fw FakeWriter) Write(topic string, key string, val map[string]interface{}) error {
	if topic != fw.ExpectedTopic {
		fw.T.Errorf("Unexpected topic. Expected: [%s], Actual: [%s]", fw.ExpectedTopic, topic)
	}
	if key != fw.ExpectedKey {
		fw.T.Errorf("Unexpected key. Expected: [%s], Actual: [%s]", fw.ExpectedKey, key)
	}

	// copy before deleting start data
	expected := copyValue(fw.ExpectedValue)
	actual := copyValue(val)

	// ignore start date when comparing values
	delete(expected, param.StartDate)
	delete(actual, param.StartDate)

	if !reflect.DeepEqual(actual, expected) {
		fw.T.Errorf("Unexpected val. \n\tExpected: [%v]\n\tActual:   [%v]", fw.ExpectedValue, val)
	}
	return fw.Error
}

func copyValue(value map[string]interface{}) map[string]interface{} {
	rtn := make(map[string]interface{})
	for key, value := range value {
		rtn[key] = value
	}
	return rtn
}
