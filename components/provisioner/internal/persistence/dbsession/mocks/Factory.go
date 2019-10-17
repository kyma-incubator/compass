// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	dberrors "github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	dbsession "github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dbsession"

	mock "github.com/stretchr/testify/mock"
)

// Factory is an autogenerated mock type for the Factory type
type Factory struct {
	mock.Mock
}

// NewReadSession provides a mock function with given fields:
func (_m *Factory) NewReadSession() dbsession.ReadSession {
	ret := _m.Called()

	var r0 dbsession.ReadSession
	if rf, ok := ret.Get(0).(func() dbsession.ReadSession); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dbsession.ReadSession)
		}
	}

	return r0
}

// NewSessionWithinTransaction provides a mock function with given fields:
func (_m *Factory) NewSessionWithinTransaction() (dbsession.WriteSessionWithinTransaction, dberrors.Error) {
	ret := _m.Called()

	var r0 dbsession.WriteSessionWithinTransaction
	if rf, ok := ret.Get(0).(func() dbsession.WriteSessionWithinTransaction); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dbsession.WriteSessionWithinTransaction)
		}
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func() dberrors.Error); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// NewWriteSession provides a mock function with given fields:
func (_m *Factory) NewWriteSession() dbsession.WriteSession {
	ret := _m.Called()

	var r0 dbsession.WriteSession
	if rf, ok := ret.Get(0).(func() dbsession.WriteSession); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dbsession.WriteSession)
		}
	}

	return r0
}
