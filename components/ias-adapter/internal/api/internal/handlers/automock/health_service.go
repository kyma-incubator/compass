// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

// HealthService is an autogenerated mock type for the HealthService type
type HealthService struct {
	mock.Mock
}

// CheckHealth provides a mock function with given fields: ctx
func (_m *HealthService) CheckHealth(ctx context.Context) (types.HealthStatus, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for CheckHealth")
	}

	var r0 types.HealthStatus
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (types.HealthStatus, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) types.HealthStatus); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(types.HealthStatus)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewHealthService creates a new instance of HealthService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewHealthService(t interface {
	mock.TestingT
	Cleanup(func())
}) *HealthService {
	mock := &HealthService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
