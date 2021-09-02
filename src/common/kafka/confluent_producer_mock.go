// Code generated by MockGen. DO NOT EDIT.
// Source: src/common/kafka/writer.go

// Package mock_kafka is a generated GoMock package.
package kafka

import (
	reflect "reflect"

	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
	gomock "github.com/golang/mock/gomock"
)

// MockconfluentProducer is a mock of confluentProducer interface.
type MockconfluentProducer struct {
	ctrl     *gomock.Controller
	recorder *MockconfluentProducerMockRecorder
}

// MockconfluentProducerMockRecorder is the mock recorder for MockconfluentProducer.
type MockconfluentProducerMockRecorder struct {
	mock *MockconfluentProducer
}

// NewMockconfluentProducer creates a new mock instance.
func NewMockconfluentProducer(ctrl *gomock.Controller) *MockconfluentProducer {
	mock := &MockconfluentProducer{ctrl: ctrl}
	mock.recorder = &MockconfluentProducerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockconfluentProducer) EXPECT() *MockconfluentProducerMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockconfluentProducer) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockconfluentProducerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockconfluentProducer)(nil).Close))
}

// Events mocks base method.
func (m *MockconfluentProducer) Events() chan kafka.Event {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Events")
	ret0, _ := ret[0].(chan kafka.Event)
	return ret0
}

// Events indicates an expected call of Events.
func (mr *MockconfluentProducerMockRecorder) Events() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Events", reflect.TypeOf((*MockconfluentProducer)(nil).Events))
}

// Flush mocks base method.
func (m *MockconfluentProducer) Flush(arg0 int) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Flush", arg0)
	ret0, _ := ret[0].(int)
	return ret0
}

// Flush indicates an expected call of Flush.
func (mr *MockconfluentProducerMockRecorder) Flush(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Flush", reflect.TypeOf((*MockconfluentProducer)(nil).Flush), arg0)
}

// Produce mocks base method.
func (m *MockconfluentProducer) Produce(arg0 *kafka.Message, arg1 chan kafka.Event) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Produce", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Produce indicates an expected call of Produce.
func (mr *MockconfluentProducerMockRecorder) Produce(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Produce", reflect.TypeOf((*MockconfluentProducer)(nil).Produce), arg0, arg1)
}
