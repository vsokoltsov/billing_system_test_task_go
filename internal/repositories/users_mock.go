// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/repositories/users.go

// Package repositories is a generated GoMock package.
package repositories

import (
	tx "billing_system_test_task/internal/adapters/tx"
	entities "billing_system_test_task/internal/entities"
	context "context"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockUsersManager is a mock of UsersManager interface
type MockUsersManager struct {
	ctrl     *gomock.Controller
	recorder *MockUsersManagerMockRecorder
}

// MockUsersManagerMockRecorder is the mock recorder for MockUsersManager
type MockUsersManagerMockRecorder struct {
	mock *MockUsersManager
}

// NewMockUsersManager creates a new mock instance
func NewMockUsersManager(ctrl *gomock.Controller) *MockUsersManager {
	mock := &MockUsersManager{ctrl: ctrl}
	mock.recorder = &MockUsersManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockUsersManager) EXPECT() *MockUsersManagerMockRecorder {
	return m.recorder
}

// WithTx mocks base method
func (m *MockUsersManager) WithTx(t tx.Tx) UsersManager {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithTx", t)
	ret0, _ := ret[0].(UsersManager)
	return ret0
}

// WithTx indicates an expected call of WithTx
func (mr *MockUsersManagerMockRecorder) WithTx(t interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithTx", reflect.TypeOf((*MockUsersManager)(nil).WithTx), t)
}

// GetByID mocks base method
func (m *MockUsersManager) GetByID(ctx context.Context, userID int) (*entities.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, userID)
	ret0, _ := ret[0].(*entities.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID
func (mr *MockUsersManagerMockRecorder) GetByID(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUsersManager)(nil).GetByID), ctx, userID)
}

// GetByWalletID mocks base method
func (m *MockUsersManager) GetByWalletID(ctx context.Context, walletID int) (*entities.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByWalletID", ctx, walletID)
	ret0, _ := ret[0].(*entities.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByWalletID indicates an expected call of GetByWalletID
func (mr *MockUsersManagerMockRecorder) GetByWalletID(ctx, walletID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByWalletID", reflect.TypeOf((*MockUsersManager)(nil).GetByWalletID), ctx, walletID)
}

// Create mocks base method
func (m *MockUsersManager) Create(ctx context.Context, email string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, email)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create
func (mr *MockUsersManagerMockRecorder) Create(ctx, email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockUsersManager)(nil).Create), ctx, email)
}
