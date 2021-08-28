// Code generated by MockGen. DO NOT EDIT.
// Source: ./pkg/operations/file_marshaller.go

// Package operations is a generated GoMock package.
package operations

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockIFileMarshaller is a mock of IFileMarshaller interface
type MockIFileMarshaller struct {
	ctrl     *gomock.Controller
	recorder *MockIFileMarshallerMockRecorder
}

// MockIFileMarshallerMockRecorder is the mock recorder for MockIFileMarshaller
type MockIFileMarshallerMockRecorder struct {
	mock *MockIFileMarshaller
}

// NewMockIFileMarshaller creates a new mock instance
func NewMockIFileMarshaller(ctrl *gomock.Controller) *MockIFileMarshaller {
	mock := &MockIFileMarshaller{ctrl: ctrl}
	mock.recorder = &MockIFileMarshallerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIFileMarshaller) EXPECT() *MockIFileMarshallerMockRecorder {
	return m.recorder
}

// MarshallOperation mocks base method
func (m *MockIFileMarshaller) MarshallOperation(operation *WalletOperation) (*MarshalledResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarshallOperation", operation)
	ret0, _ := ret[0].(*MarshalledResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MarshallOperation indicates an expected call of MarshallOperation
func (mr *MockIFileMarshallerMockRecorder) MarshallOperation(operation interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarshallOperation", reflect.TypeOf((*MockIFileMarshaller)(nil).MarshallOperation), operation)
}

// WriteToFile mocks base method
func (m *MockIFileMarshaller) WriteToFile(mr *MarshalledResult) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteToFile", mr)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteToFile indicates an expected call of WriteToFile
func (mr_2 *MockIFileMarshallerMockRecorder) WriteToFile(mr interface{}) *gomock.Call {
	mr_2.mock.ctrl.T.Helper()
	return mr_2.mock.ctrl.RecordCallWithMethodType(mr_2.mock, "WriteToFile", reflect.TypeOf((*MockIFileMarshaller)(nil).WriteToFile), mr)
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
