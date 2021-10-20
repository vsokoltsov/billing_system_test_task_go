// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/operations/query_params.go

// Package reports is a generated GoMock package.
package reports

import (
	gomock "github.com/golang/mock/gomock"
	url "net/url"
	reflect "reflect"
)

// MockQueryReaderManager is a mock of QueryReaderManager interface
type MockQueryReaderManager struct {
	ctrl     *gomock.Controller
	recorder *MockQueryReaderManagerMockRecorder
}

// MockQueryReaderManagerMockRecorder is the mock recorder for MockQueryReaderManager
type MockQueryReaderManagerMockRecorder struct {
	mock *MockQueryReaderManager
}

// NewMockQueryReaderManager creates a new mock instance
func NewMockQueryReaderManager(ctrl *gomock.Controller) *MockQueryReaderManager {
	mock := &MockQueryReaderManager{ctrl: ctrl}
	mock.recorder = &MockQueryReaderManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockQueryReaderManager) EXPECT() *MockQueryReaderManagerMockRecorder {
	return m.recorder
}

// Parse mocks base method
func (m *MockQueryReaderManager) Parse(query url.Values) (*QueryParams, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Parse", query)
	ret0, _ := ret[0].(*QueryParams)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Parse indicates an expected call of Parse
func (mr *MockQueryReaderManagerMockRecorder) Parse(query interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Parse", reflect.TypeOf((*MockQueryReaderManager)(nil).Parse), query)
}
