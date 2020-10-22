// Code generated by MockGen. DO NOT EDIT.
// Source: src/common/elastic/client.go

// Package test is a generated GoMock package.
package test

import (
	context "context"
	generated "github.com/IBM/resource-controller-go-sdk-generator/build/generated"
	gomock "github.com/golang/mock/gomock"
	http "net/http"
	reflect "reflect"
)

// MockResourceControllerService is a mock of ResourceControllerService interface.
type MockResourceControllerService struct {
	ctrl     *gomock.Controller
	recorder *MockResourceControllerServiceMockRecorder
}

// MockResourceControllerServiceMockRecorder is the mock recorder for MockResourceControllerService.
type MockResourceControllerServiceMockRecorder struct {
	mock *MockResourceControllerService
}

// NewMockResourceControllerService creates a new mock instance.
func NewMockResourceControllerService(ctrl *gomock.Controller) *MockResourceControllerService {
	mock := &MockResourceControllerService{ctrl: ctrl}
	mock.recorder = &MockResourceControllerServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResourceControllerService) EXPECT() *MockResourceControllerServiceMockRecorder {
	return m.recorder
}

// GetResourceInstance mocks base method.
func (m *MockResourceControllerService) GetResourceInstance(ctx context.Context, authorization, id string) (generated.ResourceInstance, *http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResourceInstance", ctx, authorization, id)
	ret0, _ := ret[0].(generated.ResourceInstance)
	ret1, _ := ret[1].(*http.Response)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetResourceInstance indicates an expected call of GetResourceInstance.
func (mr *MockResourceControllerServiceMockRecorder) GetResourceInstance(ctx, authorization, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResourceInstance", reflect.TypeOf((*MockResourceControllerService)(nil).GetResourceInstance), ctx, authorization, id)
}
