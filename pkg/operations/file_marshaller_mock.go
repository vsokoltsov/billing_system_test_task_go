// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/operations/file_marshaller.go

// Package operations is a generated GoMock package.
package operations

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockFileMarshallingManager is a mock of FileMarshallingManager interface
type MockFileMarshallingManager struct {
	ctrl     *gomock.Controller
	recorder *MockFileMarshallingManagerMockRecorder
}

// MockFileMarshallingManagerMockRecorder is the mock recorder for MockFileMarshallingManager
type MockFileMarshallingManagerMockRecorder struct {
	mock *MockFileMarshallingManager
}

// NewMockFileMarshallingManager creates a new mock instance
func NewMockFileMarshallingManager(ctrl *gomock.Controller) *MockFileMarshallingManager {
	mock := &MockFileMarshallingManager{ctrl: ctrl}
	mock.recorder = &MockFileMarshallingManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFileMarshallingManager) EXPECT() *MockFileMarshallingManagerMockRecorder {
	return m.recorder
}

// MarshallOperation mocks base method
func (m *MockFileMarshallingManager) MarshallOperation(operation *WalletOperation) (*MarshalledResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarshallOperation", operation)
	ret0, _ := ret[0].(*MarshalledResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarshallOperation indicates an expected call of MarshallOperation
func (mr *MockFileMarshallingManagerMockRecorder) MarshallOperation(operation interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarshallOperation", reflect.TypeOf((*MockFileMarshallingManager)(nil).MarshallOperation), operation)
}

// WriteToFile mocks base method
func (m *MockFileMarshallingManager) WriteToFile(mr *MarshalledResult) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteToFile", mr)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteToFile indicates an expected call of WriteToFile
func (mr_2 *MockFileMarshallingManagerMockRecorder) WriteToFile(mr interface{}) *gomock.Call {
	mr_2.mock.ctrl.T.Helper()
	return mr_2.mock.ctrl.RecordCallWithMethodType(mr_2.mock, "WriteToFile", reflect.TypeOf((*MockFileMarshallingManager)(nil).WriteToFile), mr)
}

// MockCSVWriter is a mock of CSVWriter interface
type MockCSVWriter struct {
	ctrl     *gomock.Controller
	recorder *MockCSVWriterMockRecorder
}

// MockCSVWriterMockRecorder is the mock recorder for MockCSVWriter
type MockCSVWriterMockRecorder struct {
	mock *MockCSVWriter
}

// NewMockCSVWriter creates a new mock instance
func NewMockCSVWriter(ctrl *gomock.Controller) *MockCSVWriter {
	mock := &MockCSVWriter{ctrl: ctrl}
	mock.recorder = &MockCSVWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCSVWriter) EXPECT() *MockCSVWriterMockRecorder {
	return m.recorder
}

// Write mocks base method
func (m *MockCSVWriter) Write(record []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", record)
	ret0, _ := ret[0].(error)
	return ret0
}

// Write indicates an expected call of Write
func (mr *MockCSVWriterMockRecorder) Write(record interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockCSVWriter)(nil).Write), record)
}
