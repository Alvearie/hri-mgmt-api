package test

import (
	"testing"

	"github.com/Alvearie/hri-mgmt-api/common/param"
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

	// copy before deleting start data
	expected := copyValue(fw.ExpectedValue)
	actual := copyValue(val)

	// ignore start date when comparing values
	delete(expected, param.StartDate)
	delete(actual, param.StartDate)

	return fw.Error
}

func copyValue(value map[string]interface{}) map[string]interface{} {
	rtn := make(map[string]interface{})
	for key, value := range value {
		rtn[key] = value
	}
	return rtn
}

func (fw FakeWriter) Close() {
	return
}
