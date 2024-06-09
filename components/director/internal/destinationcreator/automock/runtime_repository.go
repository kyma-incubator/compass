// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// RuntimeRepository is an autogenerated mock type for the runtimeRepository type
type RuntimeRepository struct {
	mock.Mock
}

// OwnerExists provides a mock function with given fields: ctx, tenant, id
func (_m *RuntimeRepository) OwnerExists(ctx context.Context, tenant string, id string) (bool, error) {
	ret := _m.Called(ctx, tenant, id)

	if len(ret) == 0 {
		panic("no return value specified for OwnerExists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (bool, error)); ok {
		return rf(ctx, tenant, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRuntimeRepository creates a new instance of RuntimeRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRuntimeRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *RuntimeRepository {
	mock := &RuntimeRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
