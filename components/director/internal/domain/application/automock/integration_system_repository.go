// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// IntegrationSystemRepository is an autogenerated mock type for the IntegrationSystemRepository type
type IntegrationSystemRepository struct {
	mock.Mock
}

// Exists provides a mock function with given fields: ctx, id
func (_m *IntegrationSystemRepository) Exists(ctx context.Context, id string) (bool, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (bool, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewIntegrationSystemRepository creates a new instance of IntegrationSystemRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIntegrationSystemRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *IntegrationSystemRepository {
	mock := &IntegrationSystemRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
