// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// IntSysSvc is an autogenerated mock type for the intSysSvc type
type IntSysSvc struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *IntSysSvc) Create(ctx context.Context, in model.IntegrationSystemInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.IntegrationSystemInput) (string, error)); ok {
		return rf(ctx, in)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.IntegrationSystemInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.IntegrationSystemInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, pageSize, cursor
func (_m *IntSysSvc) List(ctx context.Context, pageSize int, cursor string) (model.IntegrationSystemPage, error) {
	ret := _m.Called(ctx, pageSize, cursor)

	var r0 model.IntegrationSystemPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int, string) (model.IntegrationSystemPage, error)); ok {
		return rf(ctx, pageSize, cursor)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int, string) model.IntegrationSystemPage); ok {
		r0 = rf(ctx, pageSize, cursor)
	} else {
		r0 = ret.Get(0).(model.IntegrationSystemPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int, string) error); ok {
		r1 = rf(ctx, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewIntSysSvc creates a new instance of IntSysSvc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIntSysSvc(t interface {
	mock.TestingT
	Cleanup(func())
}) *IntSysSvc {
	mock := &IntSysSvc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
