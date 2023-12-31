// Code generated by MockGen. DO NOT EDIT.
// Source: models/datastore.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	models "github.com/mara-labs/transactionsigner/models"
	gomock "go.uber.org/mock/gomock"
)

// Store is a mock of Datastore interface.
type Store struct {
	ctrl     *gomock.Controller
	recorder *StoreMockRecorder
}

// StoreMockRecorder is the mock recorder for Store.
type StoreMockRecorder struct {
	mock *Store
}

// NewStore creates a new mock instance.
func NewStore(ctrl *gomock.Controller) *Store {
	mock := &Store{ctrl: ctrl}
	mock.recorder = &StoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *Store) EXPECT() *StoreMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *Store) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *StoreMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*Store)(nil).Close))
}

// CreateTransaction mocks base method.
func (m *Store) CreateTransaction(arg0 context.Context, arg1 *models.Transaction) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateTransaction", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateTransaction indicates an expected call of CreateTransaction.
func (mr *StoreMockRecorder) CreateTransaction(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTransaction", reflect.TypeOf((*Store)(nil).CreateTransaction), arg0, arg1)
}

// GetDerivationPath mocks base method.
func (m *Store) GetDerivationPath(arg0 context.Context, arg1 models.FindDerivationPathOptions) (*models.DerivationPath, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDerivationPath", arg0, arg1)
	ret0, _ := ret[0].(*models.DerivationPath)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDerivationPath indicates an expected call of GetDerivationPath.
func (mr *StoreMockRecorder) GetDerivationPath(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDerivationPath", reflect.TypeOf((*Store)(nil).GetDerivationPath), arg0, arg1)
}

// GetWallet mocks base method.
func (m *Store) GetWallet(arg0 context.Context, arg1 int64) (*models.SenderWallet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWallet", arg0, arg1)
	ret0, _ := ret[0].(*models.SenderWallet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWallet indicates an expected call of GetWallet.
func (mr *StoreMockRecorder) GetWallet(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWallet", reflect.TypeOf((*Store)(nil).GetWallet), arg0, arg1)
}
