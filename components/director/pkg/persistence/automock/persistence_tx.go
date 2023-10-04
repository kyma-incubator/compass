// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	sql "database/sql"
)

// PersistenceTx is an autogenerated mock type for the PersistenceTx type
type PersistenceTx struct {
	mock.Mock
}

// Commit provides a mock function with given fields:
func (_m *PersistenceTx) Commit() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ExecContext provides a mock function with given fields: ctx, query, args
func (_m *PersistenceTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 sql.Result
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) (sql.Result, error)); ok {
		return rf(ctx, query, args...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, ...interface{}) sql.Result); ok {
		r0 = rf(ctx, query, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sql.Result)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, ...interface{}) error); ok {
		r1 = rf(ctx, query, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetContext provides a mock function with given fields: ctx, dest, query, args
func (_m *PersistenceTx) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, ctx, dest, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string, ...interface{}) error); ok {
		r0 = rf(ctx, dest, query, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NamedExecContext provides a mock function with given fields: ctx, query, arg
func (_m *PersistenceTx) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	ret := _m.Called(ctx, query, arg)

	var r0 sql.Result
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}) (sql.Result, error)); ok {
		return rf(ctx, query, arg)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}) sql.Result); ok {
		r0 = rf(ctx, query, arg)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sql.Result)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, interface{}) error); ok {
		r1 = rf(ctx, query, arg)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Rollback provides a mock function with given fields:
func (_m *PersistenceTx) Rollback() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SelectContext provides a mock function with given fields: ctx, dest, query, args
func (_m *PersistenceTx) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, ctx, dest, query)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string, ...interface{}) error); ok {
		r0 = rf(ctx, dest, query, args...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewPersistenceTx creates a new instance of PersistenceTx. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPersistenceTx(t interface {
	mock.TestingT
	Cleanup(func())
}) *PersistenceTx {
	mock := &PersistenceTx{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
