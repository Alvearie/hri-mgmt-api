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

	// ignore start date when comparing values
	delete(fw.ExpectedValue, param.StartDate)
	delete(val, param.StartDate)

	if !reflect.DeepEqual(val, fw.ExpectedValue) {
		fw.T.Errorf("Unexpected val. Expected: [%v], Actual: [%v]", fw.ExpectedValue, val)
	}
	return fw.Error
}
